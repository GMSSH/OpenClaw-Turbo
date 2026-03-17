package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ChannelService 通道管理服务
type ChannelService struct{}

func NewChannelService() *ChannelService {
	return &ChannelService{}
}

var channelLock sync.Mutex

// 通道 → 需要安装的插件映射
var channelPluginMap = map[string]string{
	"wecom": "@openclaw-china/wecom",
	"qqbot":     "@openclaw-china/qqbot",
	"dingtalk":  "@openclaw-china/dingtalk",
}



// 通道 → 自带插件映射（只需 enable）
var channelEnableMap = map[string]string{
	"feishu": "feishu",
}

// containerName OpenClaw 容器名 (Docker 模式)
const containerName = "gmssh-openclaw"

// runClawCmd 根据部署模式执行 openclaw 命令
func runClawCmd(args ...string) ([]byte, error) {
	if getDeployMode() == "local" {
		// 通过 bash -lc 执行，确保完整 shell 环境、PATH 和 Node.js 可用
		// CI=true 跳过 npm/openclaw 的交互确认提示，避免 pipe 下卡死
		var safeArgs []string
		for _, a := range args {
			if strings.Contains(a, " ") || strings.Contains(a, "@") {
				safeArgs = append(safeArgs, fmt.Sprintf("'%s'", a))
			} else {
				safeArgs = append(safeArgs, a)
			}
		}
		cmdStr := fmt.Sprintf("cd %s && CI=true openclaw %s", getLocalDeployDir(), strings.Join(safeArgs, " "))
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "bash", "-lc", cmdStr).CombinedOutput()
		// openclaw 内部 npm/npx 崩溃时进程可能挂起，超时后检测真实错误
		if err != nil && strings.Contains(string(out), "npm pack failed") {
			return out, fmt.Errorf("npm/node 环境异常（可能存在多套 npm 冲突）: %s", extractNpmError(string(out)))
		}
		return out, err
	}
	// Docker: 先尝试执行
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	dockerArgs := append([]string{"exec", "-e", "CI=true", containerName, "openclaw"}, args...)
	log.Printf("[runClawCmd] docker %s", strings.Join(dockerArgs, " "))
	out, err := exec.CommandContext(ctx, "docker", dockerArgs...).CombinedOutput()
	if err != nil && strings.Contains(string(out), "pairing required") {
		// 自动审批配对 + 重启容器让 gateway 重载 paired 列表
		if autoApproveDevicePairing() {
			exec.Command("docker", "restart", containerName).Run()
			// 等待容器重新启动就绪（最多 30 秒）
			for i := 0; i < 15; i++ {
				time.Sleep(2 * time.Second)
				check, _ := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", containerName).Output()
				if strings.TrimSpace(string(check)) == "running" {
					// 再等 3 秒确保 gateway 完成初始化
					time.Sleep(3 * time.Second)
					break
				}
			}
			// 重试一次
			ctx2, cancel2 := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel2()
			retryArgs := append([]string{"exec", "-e", "CI=true", containerName, "openclaw"}, args...)
			return exec.CommandContext(ctx2, "docker", retryArgs...).CombinedOutput()
		}
	}
	return out, err
}

// nodeBinDir 缓存：进程生命周期内只探测一次 node 所在目录
var (
	nodeBinDirOnce  sync.Once
	nodeBinDirCache string
)

// resolveNodeBinDir 返回登录 shell 中 node 可执行文件所在目录（懒加载，仅执行一次）
func resolveNodeBinDir() string {
	nodeBinDirOnce.Do(func() {
		out, err := exec.Command("bash", "-lc", "dirname $(which node) 2>/dev/null").Output()
		if err == nil {
			dir := strings.TrimSpace(string(out))
			if dir != "" && dir != "." {
				nodeBinDirCache = dir
			}
		}
	})
	return nodeBinDirCache
}

// repairNpmSymlink 检测系统 /usr/bin/npm 是否为坏掉的 apt 残留脚本，
// 若是则把它重新指向 nvm 管理的正确 npm，避免 openclaw 内部 spawn npm 时选错。
// 无 root 权限时静默跳过（不影响正常流程）。
var repairNpmOnce sync.Once

func repairNpmSymlink() {
	repairNpmOnce.Do(func() {
		const sysnpm = "/usr/bin/npm"
		target, err := os.Readlink(sysnpm)
		if err != nil {
			return // 不是 symlink 或不存在，无需处理
		}
		// apt 装的坏 npm symlink 指向 .js 文件
		if !strings.HasSuffix(target, ".js") {
			return // 已经指向二进制，无需修复
		}
		// 找 nvm 管理的正确 npm 路径
		nvmNpm, err := exec.Command("bash", "-lc", "which npm 2>/dev/null").Output()
		if err != nil {
			return
		}
		correctNpm := strings.TrimSpace(string(nvmNpm))
		if correctNpm == "" || correctNpm == sysnpm {
			return
		}
		// 重新指向正确的 npm（忽略权限错误，不影响正常流程）
		if err := os.Remove(sysnpm); err != nil {
			return
		}
		if err := os.Symlink(correctNpm, sysnpm); err != nil {
			// 修复失败时还原（尽力）
			os.Symlink(target, sysnpm) //nolint
			return
		}
		log.Printf("[channel] 已自动修复 %s → %s（原指向损坏的 apt npm）", sysnpm, correctNpm)
	})
}

// runPluginInstall 专用于插件安装：修复 npm 环境后，在 PATH 最前面注入正确的 node/npm 目录，
// 避免系统 apt npm（/usr/share/nodejs）与 nvm npm 共存时 openclaw 选错 npm
func runPluginInstall(pluginName string) ([]byte, error) {
	// 安装前先修复可能损坏的 /usr/bin/npm symlink（仅执行一次）
	repairNpmSymlink()
	if getDeployMode() == "local" {
		pathPrefix := ""
		if dir := resolveNodeBinDir(); dir != "" {
			pathPrefix = fmt.Sprintf("export PATH=%s:$PATH && ", dir)
		}
		cmdStr := fmt.Sprintf("%scd %s && CI=true openclaw plugins install '%s'",
			pathPrefix, getLocalDeployDir(), pluginName)
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "bash", "-lc", cmdStr).CombinedOutput()
		if err != nil && strings.Contains(string(out), "npm pack failed") {
			return out, fmt.Errorf("npm/node 环境异常（可能存在多套 npm 冲突）: %s", extractNpmError(string(out)))
		}
		return out, err
	}
	// Docker 模式
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	dockerArgs := []string{"exec", "-e", "CI=true", containerName, "openclaw", "plugins", "install", pluginName}
	out, err := exec.CommandContext(ctx, "docker", dockerArgs...).CombinedOutput()
	if err != nil && strings.Contains(string(out), "npm pack failed") {
		return out, fmt.Errorf("npm/node 环境异常（可能存在多套 npm 冲突）: %s", extractNpmError(string(out)))
	}
	return out, err
}


// extractNpmError 从 npm 崩溃输出中提取关键错误行
func extractNpmError(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Error:") || strings.HasPrefix(line, "npm pack failed:") {
			return line
		}
	}
	return strings.TrimSpace(output)
}

// autoApproveDevicePairing 自动将 pending 配对请求合并到 paired 列表
// 返回 true 表示有新的配对被审批（需要重启容器）
func autoApproveDevicePairing() bool {
	devicesDir := filepath.Join(getDataDir(), "conf", "devices")
	pendingFile := filepath.Join(devicesDir, "pending.json")
	pairedFile := filepath.Join(devicesDir, "paired.json")

	pendingData, err := os.ReadFile(pendingFile)
	if err != nil {
		return false
	}

	var pending map[string]any
	if err := json.Unmarshal(pendingData, &pending); err != nil || len(pending) == 0 {
		return false
	}

	// 读取已配对列表
	paired := map[string]any{}
	if pairedData, err := os.ReadFile(pairedFile); err == nil {
		json.Unmarshal(pairedData, &paired)
	}

	// 将所有 pending 合并到 paired
	for id, device := range pending {
		paired[id] = device
	}

	// 写回 paired.json
	pairedJSON, _ := json.MarshalIndent(paired, "", "  ")
	os.WriteFile(pairedFile, pairedJSON, 0644)

	// 清空 pending.json
	os.WriteFile(pendingFile, []byte("{}"), 0644)
	return true
}

// runNpxCmd 根据部署模式执行 npx 命令
func runNpxCmd(args ...string) ([]byte, error) {
	if getDeployMode() == "local" {
		// 通过 bash -lc 执行，加载完整登录 shell 环境（与终端一致）
		var safeArgs []string
		for _, a := range args {
			if strings.Contains(a, " ") {
				safeArgs = append(safeArgs, fmt.Sprintf("'%s'", a))
			} else {
				safeArgs = append(safeArgs, a)
			}
		}
		cmdStr := fmt.Sprintf("cd %s && npx %s", getLocalDeployDir(), strings.Join(safeArgs, " "))
		return exec.Command("bash", "-lc", cmdStr).CombinedOutput()
	}
	dockerArgs := append([]string{"exec", containerName, "npx"}, args...)
	return exec.Command("docker", dockerArgs...).CombinedOutput()
}

// isPluginInstalled 检查插件是否已安装
func isPluginInstalled(pluginName string) bool {
	out, err := runClawCmd("plugins", "list")
	if err != nil {
		return false
	}
	return strings.Contains(string(out), pluginName)
}

// checkPluginStatus 检查插件加载状态
func checkPluginStatus(channelKey string) error {
	out, err := runClawCmd("plugins", "list")
	if err != nil {
		return nil
	}
	output := string(out)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, channelKey) && strings.Contains(line, "error") {
			errDetail := extractPluginError(channelKey)
			if errDetail != "" {
				return fmt.Errorf("插件 %s 安装成功但加载失败: %s", channelKey, errDetail)
			}
			return fmt.Errorf("插件 %s 安装成功但加载失败，请检查日志", channelKey)
		}
	}
	return nil
}

// extractPluginError 从日志中提取插件加载错误
func extractPluginError(channelKey string) string {
	var out []byte
	var err error
	if getDeployMode() == "local" {
		out, err = exec.Command("journalctl", "-u", "openclaw", "--no-pager", "-n", "30").CombinedOutput()
	} else {
		out, err = exec.Command("docker", "logs", "--tail", "30", containerName).CombinedOutput()
	}
	if err != nil {
		return ""
	}
	output := string(out)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, channelKey) && (strings.Contains(line, "failed to load") || strings.Contains(line, "Cannot find module") || strings.Contains(line, "Error:")) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// ensurePluginInstalled 确保插件已安装或已启用
func ensurePluginInstalled(channelKey string) (string, error) {
	// 先检查是否是自带插件（只需 enable）
	if enableId, ok := channelEnableMap[channelKey]; ok {
		out, err := runClawCmd("plugins", "enable", enableId)
		output := strings.TrimSpace(string(out))
		if strings.Contains(output, "already") || strings.Contains(output, "enabled") {
			if loadErr := checkPluginStatus(channelKey); loadErr != nil {
				return "", loadErr
			}
			return "插件已启用", nil
		}
		if err != nil {
			return "", fmt.Errorf("启用插件失败: %s", output)
		}
		if loadErr := checkPluginStatus(channelKey); loadErr != nil {
			return "", loadErr
		}
		return fmt.Sprintf("插件 %s 已启用", enableId), nil
	}

	// 检查是否需要安装第三方插件
	pluginName, ok := channelPluginMap[channelKey]
	if !ok {
		return "", nil
	}

	if isPluginInstalled(channelKey) {
		if loadErr := checkPluginStatus(channelKey); loadErr != nil {
			return "", loadErr
		}
		return "插件已存在", nil
	}

	out, err := runPluginInstall(pluginName)
	output := strings.TrimSpace(string(out))

	if strings.Contains(output, "already exists") || strings.Contains(output, "already") {
		if loadErr := checkPluginStatus(channelKey); loadErr != nil {
			return "", loadErr
		}
		return "插件已存在", nil
	}

	if err != nil {
		return "", fmt.Errorf("安装插件失败: %s", output)
	}

	if loadErr := checkPluginStatus(channelKey); loadErr != nil {
		return "", loadErr
	}
	return fmt.Sprintf("插件 %s 安装成功", pluginName), nil
}

// readOpenClawConfig 读取 openclaw.json
func readOpenClawConfig() (map[string]any, error) {
	var configPath string
	if getDeployMode() == "local" {
		configPath = getOpenClawConfigPath()
	} else {
		configPath = filepath.Join(getDataDir(), "conf", "openclaw.json")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	return config, nil
}

// writeOpenClawConfig 写入 openclaw.json
func writeOpenClawConfig(config map[string]any) error {
	var configPath string
	if getDeployMode() == "local" {
		configPath = getOpenClawConfigPath()
	} else {
		configPath = filepath.Join(getDataDir(), "conf", "openclaw.json")
	}
	// 移除 meta 字段，避免版本冲突
	delete(config, "meta")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

// GetChannels 获取所有通道配置
func (s *ChannelService) GetChannels() (map[string]any, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, err
	}
	channels, _ := config["channels"].(map[string]any)
	if channels == nil {
		channels = map[string]any{}
	}
	// 转换为列表格式返回
	var list []map[string]any
	for key, val := range channels {
		ch, ok := val.(map[string]any)
		if !ok {
			continue
		}
		ch["key"] = key
		list = append(list, ch)
	}
	return map[string]any{"channels": list}, nil
}

// SaveChannel 保存/更新通道配置
func (s *ChannelService) SaveChannel(req map[string]any) (map[string]any, error) {
	channelLock.Lock()
	defer channelLock.Unlock()

	channelKey, _ := req["channelKey"].(string)
	if channelKey == "" {
		return nil, fmt.Errorf("channelKey 不能为空")
	}

	// 自动安装所需插件
	pluginMsg, err := ensurePluginInstalled(channelKey)
	if err != nil {
		return nil, err
	}

	config, err := readOpenClawConfig()
	if err != nil {
		return nil, err
	}

	channels, _ := config["channels"].(map[string]any)
	if channels == nil {
		channels = map[string]any{}
	}

	// 构建通道配置（不含 channelKey 本身）
	channelConfig := map[string]any{}
	for k, v := range req {
		if k != "channelKey" {
			channelConfig[k] = v
		}
	}

	// Telegram: 只保留 botToken, allowFrom, enabled
	if channelKey == "telegram" {
		channelConfig = map[string]any{
			"botToken":  req["botToken"],
			"allowFrom": req["allowFrom"],
			"enabled":   req["enabled"],
		}
	}

	// 企业微信：新格式 channels.wecom，只需 botId + secret，mode 固定为 ws
	if channelKey == "wecom" {
		channelConfig = map[string]any{
			"enabled": req["enabled"],
			"mode":    "ws",
			"botId":   req["botId"],
			"secret":  req["secret"],
		}
	}

	// 钉钉：多账号结构，自动从 openclaw.json 读取网关 token 注入每个账号
	if channelKey == "dingtalk" {
		// 读取网关 token
		gatewayToken := ""
		if gw, ok := config["gateway"].(map[string]any); ok {
			if auth, ok := gw["auth"].(map[string]any); ok {
				if t, ok := auth["token"].(string); ok {
					gatewayToken = t
				}
			}
		}

		accounts := map[string]any{}
		// 前端传来 accounts 数组：[{ name, clientId, clientSecret }...]
		if rawAccounts, ok := req["accounts"].([]any); ok && len(rawAccounts) > 0 {
			for i, item := range rawAccounts {
				if acc, ok := item.(map[string]any); ok {
					botKey := fmt.Sprintf("bot%d", i+1)
					accounts[botKey] = map[string]any{
						"name":         acc["name"],
						"clientId":     acc["clientId"],
						"clientSecret": acc["clientSecret"],
						"gatewayToken": gatewayToken,
						"enableAICard": true,
					}
				}
			}
		}
		if len(accounts) == 0 {
			// 兼容旧格式（单账号直接传字段）
			accounts["bot1"] = map[string]any{
				"name":         req["name"],
				"clientId":     req["clientId"],
				"clientSecret": req["clientSecret"],
				"gatewayToken": gatewayToken,
				"enableAICard": true,
			}
		}
		channelConfig = map[string]any{
			"enabled":               req["enabled"],
			"accounts":              accounts,
			"enableAICard":          true,
			"longTaskNoticeDelayMs": 30000,
			"maxFileSizeMB":         100,
			"inboundMedia": map[string]any{
				"dir":      "~/.openclaw/media/dingtalk/inbound",
				"keepDays": 7,
			},
		}
	}

	if channelKey == "discord" {
		token, _ := req["token"].(string)
		guildId, _ := req["guildId"].(string)
		enabled, _ := req["enabled"].(bool)
		channelConfig = map[string]any{
			"token":       token,
			"groupPolicy": "allowlist",
			"dm": map[string]any{
				"enabled": false,
			},
			"retry": map[string]any{
				"attempts":    3,
				"minDelayMs":  500,
				"maxDelayMs":  30000,
				"jitter":      0.1,
			},
			"guilds": map[string]any{
				guildId: map[string]any{
					"users":          []string{"*"},
					"requireMention": true,
					"channels": map[string]any{
						"*": map[string]any{
							"allow":          true,
							"requireMention": true,
						},
					},
				},
			},
			"enabled": enabled,
		}
	}

	channels[channelKey] = channelConfig
	config["channels"] = channels

	if err := writeOpenClawConfig(config); err != nil {
		return nil, err
	}

	message := "通道配置已保存"
	if pluginMsg != "" {
		message = message + "，" + pluginMsg
	}
	return map[string]any{"success": true, "message": message}, nil
}

// DeleteChannel 删除通道
func (s *ChannelService) DeleteChannel(req map[string]any) (map[string]any, error) {
	channelLock.Lock()
	defer channelLock.Unlock()

	channelKey, _ := req["channelKey"].(string)
	if channelKey == "" {
		return nil, fmt.Errorf("channelKey 不能为空")
	}

	config, err := readOpenClawConfig()
	if err != nil {
		return nil, err
	}

	channels, _ := config["channels"].(map[string]any)
	if channels == nil {
		return nil, fmt.Errorf("通道不存在: %s", channelKey)
	}

	delete(channels, channelKey)
	config["channels"] = channels

	if err := writeOpenClawConfig(config); err != nil {
		return nil, err
	}

	return map[string]any{"success": true, "message": "通道已删除"}, nil
}

// ToggleChannel 启用/禁用通道
func (s *ChannelService) ToggleChannel(req map[string]any) (map[string]any, error) {
	channelLock.Lock()
	defer channelLock.Unlock()

	channelKey, _ := req["channelKey"].(string)
	enabled, _ := req["enabled"].(bool)

	config, err := readOpenClawConfig()
	if err != nil {
		return nil, err
	}

	channels, _ := config["channels"].(map[string]any)
	if channels == nil {
		return nil, fmt.Errorf("通道不存在")
	}
	ch, ok := channels[channelKey].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("通道不存在: %s", channelKey)
	}

	ch["enabled"] = enabled
	channels[channelKey] = ch
	config["channels"] = channels

	if err := writeOpenClawConfig(config); err != nil {
		return nil, err
	}

	status := "已禁用"
	if enabled {
		status = "已启用"
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("通道%s", status)}, nil
}

// ApprovePairing 批准 Telegram 配对码
func (s *ChannelService) ApprovePairing(req map[string]any) (map[string]any, error) {
	code, _ := req["code"].(string)
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("配对码不能为空")
	}

	var cmd *exec.Cmd
	if getDeployMode() == "local" {
		clawBin := filepath.Join(getLocalDeployDir(), "node_modules", ".bin", "openclaw")
		cmd = exec.Command(clawBin, "pairing", "approve", "telegram", code)
		cmd.Dir = getLocalDeployDir()
	} else {
		cmd = exec.Command("docker", "exec", "gmssh-openclaw", "openclaw", "pairing", "approve", "telegram", code)
	}

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			return nil, fmt.Errorf("配对失败: %s", output)
		}
		return nil, fmt.Errorf("配对失败: %v", err)
	}

	return map[string]any{"success": true, "message": "配对码已批准，约一小时内生效"}, nil
}
