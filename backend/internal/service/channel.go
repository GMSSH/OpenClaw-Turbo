package service

import (
	"encoding/json"
	"fmt"
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
	"wecom-app": "@openclaw-china/wecom-app",
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
		cmd := exec.Command("openclaw", args...)
		cmd.Dir = getLocalDeployDir()
		return cmd.CombinedOutput()
	}
	// Docker: 先尝试执行
	dockerArgs := append([]string{"exec", containerName, "openclaw"}, args...)
	out, err := exec.Command("docker", dockerArgs...).CombinedOutput()
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
			retryArgs := append([]string{"exec", containerName, "openclaw"}, args...)
			return exec.Command("docker", retryArgs...).CombinedOutput()
		}
	}
	return out, err
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
		cmd := exec.Command("npx", args...)
		cmd.Dir = getLocalDeployDir()
		return cmd.CombinedOutput()
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

	out, err := runClawCmd("plugins", "install", pluginName)
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

	// 特殊处理：针对钉钉通道，固定写入 enableAICard: false 关闭实验性卡片功能
	if channelKey == "dingtalk" {
		channelConfig["enableAICard"] = false
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
