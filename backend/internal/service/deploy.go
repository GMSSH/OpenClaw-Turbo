package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"guanxi/eazy-claw/internal/dto"
	"time"
)

// DeployService 部署服务
type DeployService struct{}

func init() {
	// GMSSH 插件进程可能 PATH 受限，确保常见系统路径可用
	extraPaths := []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}
	currentPath := os.Getenv("PATH")
	for _, p := range extraPaths {
		if !strings.Contains(currentPath, p) {
			currentPath = currentPath + ":" + p
		}
	}
	os.Setenv("PATH", currentPath)
}

// NewDeployService 创建部署服务实例
func NewDeployService() *DeployService {
	return &DeployService{}
}

// 部署状态管理（进程内缓存）
var (
	deployLock     sync.Mutex
	deployLogs     []string
	deployFinished bool
	deploySuccess  bool
)

// 部署模式持久化 — 放在 GMSSH 不会清理的持久目录
func getDeployModeFile() string {
	dir := "/.__gmssh/tmp/GMClaw"
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "deploy_mode")
}

func getDeployMode() string {
	// 1) 先读缓存文件
	data, err := os.ReadFile(getDeployModeFile())
	if err == nil {
		mode := strings.TrimSpace(string(data))
		// 验证缓存是否与实际状态匹配（避免卸载后残留旧缓存）
		if mode == "docker" {
			out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "gmssh-openclaw").Output()
			if err == nil && len(strings.TrimSpace(string(out))) > 0 {
				return "docker" // 容器确实存在
			}
			// 容器不存在，缓存失效，清除并重检测
			os.Remove(getDeployModeFile())
		} else if mode == "local" {
			_, binErr := exec.LookPath("openclaw")
			outSvc, _ := exec.Command("systemctl", "is-enabled", "openclaw").Output()
			if binErr == nil || strings.TrimSpace(string(outSvc)) == "enabled" {
				return "local" // 二进制或服务确实存在
			}
			// 本地未安装，缓存失效，清除并重检测
			os.Remove(getDeployModeFile())
		}
	}
	// 2) 缓存丢失或失效，自动检测
	detected := detectDeployMode()
	if detected != "" {
		saveDeployMode(detected)
	}
	return detected
}

// IsClawInstalled 检查 OpenClaw 是否已安装（验证实际部署状态，忽略残留文件）
func (s *DeployService) IsClawInstalled() map[string]any {
	// 1) openclaw 命令存在 → shell/本地安装
	if _, err := exec.LookPath("openclaw"); err == nil {
		return map[string]any{"installed": true, "reason": "binary"}
	}
	// 2) Docker 容器存在（运行中或已停止）→ Docker 部署
	if out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "gmssh-openclaw").Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return map[string]any{"installed": true, "reason": "docker"}
	}
	// 都不存在 → 未安装，清理残留文件
	os.Remove(getDeployModeFile())
	return map[string]any{"installed": false}
}

// detectDeployMode 根据实际环境判断部署模式
func detectDeployMode() string {
	// 检查 Docker 容器是否存在
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "gmssh-openclaw").Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return "docker"
	}
	// 检查本地 openclaw 二进制是否安装
	if _, err := exec.LookPath("openclaw"); err == nil {
		return "local"
	}
	// 检查 systemd 服务
	if out, err := exec.Command("systemctl", "is-enabled", "openclaw").Output(); err == nil {
		if strings.TrimSpace(string(out)) == "enabled" {
			return "local"
		}
	}
	return "" // 什么都没安装，不设置默认值
}

func saveDeployMode(mode string) {
	os.MkdirAll(getTmpDir(), 0755)
	os.WriteFile(getDeployModeFile(), []byte(mode), 0644)
}

// getDockerComposeCmd 检测可用的 docker compose 命令形式
// 优先 "docker compose"（CLI 插件），回退 "docker-compose"（独立二进制）
func getDockerComposeCmd() (name string, args []string, ok bool) {
	// 尝试 docker compose (V2 plugin)
	if err := exec.Command("docker", "compose", "version").Run(); err == nil {
		return "docker", []string{"compose"}, true
	}
	// 尝试 docker-compose (standalone)
	if err := exec.Command("docker-compose", "version").Run(); err == nil {
		return "docker-compose", nil, true
	}
	return "", nil, false
}

// getDataDir 获取数据目录 /opt/gmclaw
func getDataDir() string {
	return "/opt/gmclaw"
}

// CheckEnvironment 检测Docker环境
func (s *DeployService) CheckEnvironment() (*dto.CheckEnvResp, error) {
	resp := &dto.CheckEnvResp{}

	// 1. 检测 docker 命令
	if err := exec.Command("docker", "--version").Run(); err == nil {
		resp.DockerReady = true
	}

	// 2. 检测 docker compose 命令（兼容 docker compose 和 docker-compose）
	_, _, composeOk := getDockerComposeCmd()
	resp.DockerComposeReady = composeOk

	// 3. 检测 Node.js
	// 通过 bash -lc 执行，加载登录 shell 环境（nvm/profile.d），
	// 确保能检测到 nvm 安装的 node，与用户终端里看到的版本一致
	if out, err := exec.Command("bash", "-lc", "node --version 2>/dev/null").Output(); err == nil {
		v := strings.TrimSpace(string(out))
		if v != "" {
			resp.NodeVersion = v
			// 版本号校验：必须 >= 22
			major := parseNodeMajorVersion(v)
			if major >= 22 {
				resp.NodeReady = true
			}
		}
	}

	// 4. 检测 pnpm（同样通过 bash -lc，确保 PNPM_HOME 已加载）
	if err := exec.Command("bash", "-lc", "pnpm --version >/dev/null 2>&1").Run(); err == nil {
		resp.PnpmReady = true
	}

	resp.AllReady = resp.DockerReady && resp.DockerComposeReady
	return resp, nil
}


// CheckPorts 检测端口是否被占用
func (s *DeployService) CheckPorts(req dto.CheckPortsReq) (*dto.CheckPortsResp, error) {
	resp := &dto.CheckPortsResp{}
	for _, port := range req.Ports {
		ps := dto.PortStatus{Port: port, Available: true}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			ps.Available = false
			// 尝试获取占用进程
			out, e := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t").Output()
			if e == nil && len(strings.TrimSpace(string(out))) > 0 {
				pid := strings.TrimSpace(strings.Split(string(out), "\n")[0])
				cmdOut, _ := exec.Command("ps", "-p", pid, "-o", "comm=").Output()
				ps.Process = strings.TrimSpace(string(cmdOut))
			}
		} else {
			ln.Close()
		}
		resp.Results = append(resp.Results, ps)
	}
	return resp, nil
}

// GenerateToken 生成32位随机Token
func (s *DeployService) GenerateToken() (*dto.GenerateTokenResp, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return &dto.GenerateTokenResp{Token: string(b)}, nil
}

// GetClawConfig 读取 openclaw.json 配置信息
func (s *DeployService) GetClawConfig() (*dto.ClawConfigResp, error) {
	resp := &dto.ClawConfigResp{}

	var configPath string
	if getDeployMode() == "local" {
		configPath = getOpenClawConfigPath()
	} else {
		configPath = filepath.Join(getDataDir(), "conf", "openclaw.json")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return resp, nil // 配置文件不存在，返回空
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return resp, nil
	}

	// 解析 agents.defaults.model.primary
	if agents, ok := config["agents"].(map[string]any); ok {
		if defaults, ok := agents["defaults"].(map[string]any); ok {
			if model, ok := defaults["model"].(map[string]any); ok {
				if primary, ok := model["primary"].(string); ok {
					resp.PrimaryModel = primary
				}
			}
		}
	}

	// 解析 models.providers
	if models, ok := config["models"].(map[string]any); ok {
		if providers, ok := models["providers"].(map[string]any); ok {
			for providerName, pv := range providers {
				resp.Provider = providerName
				p, ok := pv.(map[string]any)
				if !ok {
					continue
				}
				if baseUrl, ok := p["baseUrl"].(string); ok {
					resp.BaseUrl = baseUrl
				}
				if apiKey, ok := p["apiKey"].(string); ok {
					if len(apiKey) > 8 {
						resp.ApiKeyMasked = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
					} else {
						resp.ApiKeyMasked = "****"
					}
				}
				if modelsArr, ok := p["models"].([]any); ok && len(modelsArr) > 0 {
					if m, ok := modelsArr[0].(map[string]any); ok {
						if name, ok := m["name"].(string); ok {
							resp.ModelName = name
						}
						if cw, ok := m["contextWindow"].(float64); ok {
							resp.ContextWindow = int(cw)
						}
						if mt, ok := m["maxTokens"].(float64); ok {
							resp.MaxTokens = int(mt)
						}
					}
				}
				break // 只取第一个 provider
			}
		}
	}

	// 解析 gateway
	if gw, ok := config["gateway"].(map[string]any); ok {
		if port, ok := gw["port"].(float64); ok {
			resp.GatewayPort = int(port)
		}
		if bind, ok := gw["bind"].(string); ok {
			resp.GatewayBind = bind
		}
		if mode, ok := gw["mode"].(string); ok {
			resp.GatewayMode = mode
		}
		if auth, ok := gw["auth"].(map[string]any); ok {
			if mode, ok := auth["mode"].(string); ok {
				resp.AuthMode = mode
			}
			if token, ok := auth["token"].(string); ok {
				resp.GatewayToken = token
			}
		}
	}

	// Docker 模式：从 docker-compose.yml 获取端口 + 容器 CPU/内存
	if getDeployMode() != "local" {
		composeFile := filepath.Join(getDataDir(), "docker-compose.yml")
		if composeData, err := os.ReadFile(composeFile); err == nil {
			content := string(composeData)
			for _, line := range strings.Split(content, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "- \"") && strings.Contains(trimmed, ":") {
					trimmed = strings.Trim(trimmed, "- \"")
					parts := strings.SplitN(trimmed, ":", 2)
					if len(parts) == 2 {
						if port, err := fmt.Sscanf(parts[0], "%d", &resp.WebPort); err == nil && port == 1 {
							break
						}
					}
				}
			}
		}
	}

	resp.DeployMode = getDeployMode()
	resp.ConfigPath = configPath

	// 解析 agents.defaults 下的记忆配置
	if agents, ok := config["agents"].(map[string]any); ok {
		if defs, ok := agents["defaults"].(map[string]any); ok {
			// compaction.memoryFlush.enabled
			if compaction, ok := defs["compaction"].(map[string]any); ok {
				if mf, ok := compaction["memoryFlush"].(map[string]any); ok {
					if enabled, ok := mf["enabled"].(bool); ok {
						resp.MemoryFlushEnabled = enabled
					}
				}
			}
			// memorySearch.experimental.sessionMemory
			if ms, ok := defs["memorySearch"].(map[string]any); ok {
				if exp, ok := ms["experimental"].(map[string]any); ok {
					if _, ok := exp["sessionMemory"]; ok {
						resp.SessionMemoryEnabled = true
					}
				}
			}
		}
	}

	return resp, nil
}

// Deploy 执行部署
func (s *DeployService) Deploy(req dto.DeployReq) (*dto.DeployResp, error) {
	// 重置部署状态
	deployLock.Lock()
	deployLogs = []string{}
	deployFinished = false
	deploySuccess = false
	deployLock.Unlock()

	dataDir := getDataDir()
	confDir := filepath.Join(dataDir, "conf")
	workspaceDir := filepath.Join(dataDir, "workspace")
	composeFile := filepath.Join(dataDir, "docker-compose.yml")
	envFile := filepath.Join(dataDir, ".env")
	configFile := filepath.Join(confDir, "openclaw.json")

	// 创建所需目录
	for _, dir := range []string{dataDir, confDir, workspaceDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}

	// 生成 docker-compose.yml
	composeContent := s.generateComposeFile(req, dataDir)
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		return nil, fmt.Errorf("写入 docker-compose.yml 失败: %v", err)
	}

	// 生成 .env 文件
	envContent := fmt.Sprintf("OPENCLAW_GATEWAY_MODE=local\nOPENCLAW_GATEWAY_TOKEN=%s\n\n# === PLUGINS ===\n", req.Token)
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		return nil, fmt.Errorf("写入 .env 失败: %v", err)
	}

	// 生成 openclaw.json 配置
	openclawConfig := s.generateOpenClawConfig(req)
	configJSON, err := json.MarshalIndent(openclawConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %v", err)
	}
	if err := os.WriteFile(configFile, configJSON, 0644); err != nil {
		return nil, fmt.Errorf("写入 openclaw.json 失败: %v", err)
	}
	// Gemini：补写 auth 配置、.env 和 auth-profiles.json
	if req.Provider == "gemini" {
		if err := writeGeminiExtraFiles(openclawConfig, req.ApiKey, configFile); err != nil {
			return nil, fmt.Errorf("写入 Gemini 额外配置失败: %v", err)
		}
	}
	// 赋予容器内 node 用户(UID 1000)读写权限
	chownCmd := exec.Command("chown", "-R", "1000:1000", dataDir)
	if out, err := chownCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("设置目录权限失败: %v, output: %s", err, string(out))
	}
	chmodCmd := exec.Command("chmod", "-R", "775", dataDir)
	if out, err := chmodCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("设置目录权限失败: %v, output: %s", err, string(out))
	}

	saveDeployMode("docker")
	// 异步执行 docker compose up
	go s.runDockerCompose(dataDir, composeFile, req.WebPort)

	return &dto.DeployResp{
		Success: true,
		Message: "部署任务已启动",
	}, nil
}

// DeployLocal 本地 Shell 部署入口
func (s *DeployService) DeployLocal(req dto.DeployReq) (*dto.DeployResp, error) {
	// 重置部署状态
	deployLock.Lock()
	deployLogs = []string{}
	deployFinished = false
	deploySuccess = false
	deployLock.Unlock()

	saveDeployMode("local")
	go s.runLocalDeploy(req)

	return &dto.DeployResp{
		Success: true,
		Message: "本地部署任务已启动",
	}, nil
}

// getTmpDir 获取 tmp 目录 (与 main.go 同逻辑)
func getTmpDir() string {
	workDir, _ := os.Getwd()
	absPath, _ := filepath.Abs(filepath.Join(workDir, "..", "..", "tmp"))
	return absPath
}

// getLocalDeployDir 本地部署源码目录
func getLocalDeployDir() string {
	dir := "/.__gmssh/tmp/GMClaw"
	os.MkdirAll(dir, 0755)
	return dir
}

// getScriptsDir 脚本目录 = tmpDir 同级的 scripts/
func getScriptsDir() string {
	return filepath.Join(filepath.Dir(getTmpDir()), "scripts")
}

// getOpenClawConfigDir 获取 ~/.openclaw 目录
func getOpenClawConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw")
}

// getOpenClawConfigPath 获取 ~/.openclaw/openclaw.json
func getOpenClawConfigPath() string {
	return filepath.Join(getOpenClawConfigDir(), "openclaw.json")
}

// writeGeminiExtraFiles 在写完 openclaw.json 之后，追加写入 Gemini 所需的三项额外配置：
//  1. openclaw.json 中补充 auth.profiles.google:default（固定值）
//  2. 配置目录下的 .env 文件（GEMINI_API_KEY + GOOGLE_API_KEY）
//  3. agents/main/agent/auth-profiles.json
func writeGeminiExtraFiles(config map[string]any, apiKey string, configPath string) error {
	confDir := filepath.Dir(configPath)

	// --- 1. 补充 auth.profiles 并重新写入 openclaw.json ---
	authMap, _ := config["auth"].(map[string]any)
	if authMap == nil {
		authMap = map[string]any{}
	}
	profilesMap, _ := authMap["profiles"].(map[string]any)
	if profilesMap == nil {
		profilesMap = map[string]any{}
	}
	profilesMap["google:default"] = map[string]any{
		"provider": "google",
		"mode":     "api_key",
	}
	authMap["profiles"] = profilesMap
	config["auth"] = authMap

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入 openclaw.json 失败: %v", err)
	}

	// --- 2. 写入 .env ---
	envContent := fmt.Sprintf("GEMINI_API_KEY=%s\nGOOGLE_API_KEY=%s\n", apiKey, apiKey)
	envPath := filepath.Join(confDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return fmt.Errorf("写入 .env 失败: %v", err)
	}

	// --- 3. 写入 agents/main/agent/auth-profiles.json ---
	agentDir := filepath.Join(confDir, "agents", "main", "agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("创建 agent 目录失败: %v", err)
	}
	authProfiles := map[string]any{
		"version": 1,
		"profiles": map[string]any{
			"google:default": map[string]any{
				"type":     "api_key",
				"provider": "google",
				"key":      apiKey,
			},
		},
	}
	apData, err := json.MarshalIndent(authProfiles, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 auth-profiles 失败: %v", err)
	}
	apPath := filepath.Join(agentDir, "auth-profiles.json")
	if err := os.WriteFile(apPath, apData, 0600); err != nil {
		return fmt.Errorf("写入 auth-profiles.json 失败: %v", err)
	}

	return nil
}

// InstallNodeEnv 执行 Node.js 环境安装脚本
func (s *DeployService) InstallNodeEnv() (map[string]any, error) {
	// 重置日志
	deployLock.Lock()
	deployLogs = []string{}
	deployFinished = false
	deploySuccess = false
	deployLock.Unlock()

	go s.runInstallNode()

	return map[string]any{"success": true, "message": "Node 环境安装已启动"}, nil
}

func (s *DeployService) runInstallNode() {
	addDeployLog("🔧 正在安装 Node.js 环境...")

	scriptPath := filepath.Join(getScriptsDir(), "install_node.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		addDeployLog(fmt.Sprintf("❌ 安装脚本不存在: %s", scriptPath))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	cmd := exec.Command("bash", scriptPath)
	stderrPipe, _ := cmd.StderrPipe()
	stdoutPipe, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 启动安装脚本失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()

	err := cmd.Wait()

	deployLock.Lock()
	defer deployLock.Unlock()
	if err != nil {
		addDeployLogLocked(fmt.Sprintf("❌ Node 环境安装失败: %v", err))
		deployFinished = true
		deploySuccess = false
	} else {
		addDeployLogLocked("✅ Node.js 环境安装完成")
		deployFinished = true
		deploySuccess = true
	}
}

// runLocalDeploy 本地部署流程
func (s *DeployService) runLocalDeploy(req dto.DeployReq) {
	cloneDir := getLocalDeployDir()

	// ====== 第一步: Git Clone ======
	if _, err := os.Stat(filepath.Join(cloneDir, "package.json")); os.IsNotExist(err) {
		addDeployLog("📦 正在克隆 OpenClaw 仓库...")
		os.MkdirAll(filepath.Dir(cloneDir), 0755)

		const maxCloneRetries = 3
		var cloneErr error
		for attempt := 1; attempt <= maxCloneRetries; attempt++ {
			// 每次尝试前清理可能存在的残留目录
			if _, dirErr := os.Stat(cloneDir); dirErr == nil {
				addDeployLog("🗑️ 清理旧目录...")
				os.RemoveAll(cloneDir)
			}
			if attempt > 1 {
				addDeployLog(fmt.Sprintf("🔄 第 %d/%d 次重试克隆仓库...", attempt, maxCloneRetries))
			}
			cmd := exec.Command("git", "clone", "https://gitee.com/OpenClaw-CN/openclaw-cn.git", cloneDir)
			out, err := cmd.CombinedOutput()
			if err == nil {
				cloneErr = nil
				break
			}
			cloneErr = err
			addDeployLog(fmt.Sprintf("⚠️ 克隆失败 (第 %d/%d 次): %s", attempt, maxCloneRetries, strings.TrimSpace(string(out))))
			if attempt < maxCloneRetries {
				addDeployLog("⏳ 5 秒后重试...")
				time.Sleep(5 * time.Second)
			}
		}
		if cloneErr != nil {
			addDeployLog("❌ 克隆仓库失败，已重试 3 次，部署终止")
			deployLock.Lock()
			deployFinished = true
			deploySuccess = false
			deployLock.Unlock()
			return
		}
		addDeployLog("✅ 仓库克隆完成")

		// 切换到稳定标签
		// addDeployLog("🏷️ 切换到 v2026.2.2-cn 分支...")
		// checkoutCmd := exec.Command("git", "-C", cloneDir, "checkout", "v2026.2.2-cn")
		// if out, err := checkoutCmd.CombinedOutput(); err != nil {
		// 	addDeployLog(fmt.Sprintf("⚠️ 切换分支失败，使用 main: %s", strings.TrimSpace(string(out))))
		// }
	} else {
		addDeployLog("📦 项目已存在，跳过克隆")
	}

	// ====== 第二步: pnpm install ======
	addDeployLog("📥 正在安装依赖 (pnpm install)...")
	if !s.runStreamCmd(cloneDir, "pnpm", "install") {
		return
	}

	// ====== 第三步: pnpm ui:build ======
	addDeployLog("🎨 正在构建 UI 依赖 (pnpm ui:build)...")
	if !s.runStreamCmd(cloneDir, "pnpm", "ui:build") {
		addDeployLog("⚠️ pnpm ui:build 失败，请到 OpenClaw 目录手动执行: pnpm ui:build")
		return
	}

	// ====== 第四步: pnpm build ======
	addDeployLog("🔨 正在构建项目 (pnpm build)...")
	if !s.runStreamCmd(cloneDir, "pnpm", "build") {
		return
	}

	// ====== 预防性检查: dist/control-ui 是否存在 ======
	// 如果重部署场景下 ui:build 产物丢失，运行时会报 "Control UI assets not found"
	controlUiDir := filepath.Join(cloneDir, "dist", "control-ui")
	if _, err := os.Stat(controlUiDir); os.IsNotExist(err) {
		addDeployLog("⚠️ 未找到 dist/control-ui，重新执行 pnpm ui:build...")
		if !s.runStreamCmd(cloneDir, "pnpm", "ui:build") {
			addDeployLog("❌ pnpm ui:build 重试失败，部署终止")
			return
		}
		// 再次验证
		if _, err2 := os.Stat(controlUiDir); os.IsNotExist(err2) {
			addDeployLog("❌ dist/control-ui 仍不存在，部署终止")
			deployLock.Lock()
			deployFinished = true
			deploySuccess = false
			deployLock.Unlock()
			return
		}
		addDeployLog("✅ dist/control-ui 构建完成")
	} else {
		addDeployLog("✅ dist/control-ui 已存在")
	}

	// ====== 第五步: 生成配置 ======
	addDeployLog("⚙️ 正在生成配置文件...")

	openclawConfig := s.generateOpenClawConfig(req)
	configJSON, err := json.MarshalIndent(openclawConfig, "", "  ")
	if err != nil {
		addDeployLog(fmt.Sprintf("❌ 生成配置失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}
	// 配置文件写到 ~/.openclaw/openclaw.json
	os.MkdirAll(getOpenClawConfigDir(), 0755)
	configPath := getOpenClawConfigPath()
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		addDeployLog(fmt.Sprintf("❌ 写入配置失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}
	// Gemini：补写 auth 配置、.env 和 auth-profiles.json
	if req.Provider == "gemini" {
		if err := writeGeminiExtraFiles(openclawConfig, req.ApiKey, getOpenClawConfigPath()); err != nil {
			addDeployLog(fmt.Sprintf("⚠️ 写入 Gemini 额外配置失败: %v", err))
		}
	}
	addDeployLog("✅ 配置文件已生成")

	// ====== 第 5.5 步: 创建 openclaw 命令链接 ======
	openclawBin := filepath.Join(cloneDir, "openclaw.mjs")
	symlinkTarget := "/usr/local/bin/openclaw"
	addDeployLog("🔗 正在创建 openclaw 命令...")
	os.Remove(symlinkTarget) // 先清除旧链接
	if err := os.Symlink(openclawBin, symlinkTarget); err != nil {
		addDeployLog(fmt.Sprintf("⚠️ 创建 openclaw 命令链接失败: %v（可手动执行 ln -sf %s %s）", err, openclawBin, symlinkTarget))
	} else {
		addDeployLog("✅ openclaw 命令已可用")
	}

	// ====== 第六步: 创建 systemd 服务并启动 ======
	addDeployLog("🚀 正在配置 OpenClaw 系统服务...")

	// 动态获取 node 的实际路径
	// 通过 bash -lc 加载 nvm/profile.d 环境，与 CheckEnvironment 检测逻辑保持一致
	nodePathBytes, nodePathErr := exec.Command("bash", "-lc", "which node 2>/dev/null").Output()
	nodePath := strings.TrimSpace(string(nodePathBytes))
	if nodePathErr != nil || nodePath == "" {
		addDeployLog("❌ 未找到 node 命令，请先安装 Node.js（点击「安装 Node.js 环境」）")
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=OpenClaw Gateway
After=network.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s %s gateway --bind lan --port %d
Environment=OPENCLAW_GATEWAY_TOKEN=%s
Environment=OPENCLAW_GATEWAY_MODE=local
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, cloneDir, nodePath, openclawBin, req.WebPort, req.Token)

	if err := os.WriteFile("/etc/systemd/system/openclaw.service", []byte(serviceContent), 0644); err != nil {
		addDeployLog(fmt.Sprintf("❌ 创建服务文件失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}
	addDeployLog("✅ systemd 服务已创建")

	// daemon-reload + enable + start
	addDeployLog("🔄 正在启动 OpenClaw 服务...")
	exec.Command("systemctl", "daemon-reload").Run()
	if out, err := exec.Command("systemctl", "enable", "--now", "openclaw").CombinedOutput(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 启动服务失败: %s", strings.TrimSpace(string(out))))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	// 等待几秒确认服务正常运行
	exec.Command("sleep", "3").Run()

	deployLock.Lock()
	defer deployLock.Unlock()
	addDeployLogLocked("✅ OpenClaw Gateway 已启动（systemd 管理）")
	addDeployLogLocked(fmt.Sprintf("🌐 访问地址: http://<服务器IP>:%d", req.WebPort))
	addDeployLogLocked(fmt.Sprintf("🔥 请确保防火墙已放开端口 %d 的访问", req.WebPort))
	deployFinished = true
	deploySuccess = true
}

// runStreamCmd 通过 bash login shell 执行命令并流式输出到部署日志。
// 使用 bash -lc 确保 /etc/profile.d/ 下的 nvm/pnpm 环境变量被加载。
// 返回 true 表示成功，false 表示失败（失败时已设置 deployFinished=true）。
func (s *DeployService) runStreamCmd(dir string, name string, args ...string) bool {
	// 拼接完整命令字符串，通过 bash login shell 执行
	rawCmd := name
	if len(args) > 0 {
		rawCmd = name + " " + strings.Join(args, " ")
	}
	cmd := exec.Command("bash", "-lc", rawCmd)
	cmd.Dir = dir

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 命令启动失败 [%s]: %v", rawCmd, err))
		// 必须标记失败，否则 checkDeployOK() 返回 true 导致后续步骤继续执行
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return false
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 命令执行失败 [%s]: %v", rawCmd, err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return false
	}
	return true
}

// checkDeployOK 检查部署是否仍在进行中（未失败）
func (s *DeployService) checkDeployOK() bool {
	deployLock.Lock()
	defer deployLock.Unlock()
	return !deployFinished
}

// GetDeployLogs 获取部署日志
func (s *DeployService) GetDeployLogs() (*dto.DeployLogResp, error) {
	deployLock.Lock()
	defer deployLock.Unlock()

	// 复制日志
	logs := make([]string, len(deployLogs))
	copy(logs, deployLogs)

	// 清空已读日志
	deployLogs = []string{}

	return &dto.DeployLogResp{
		Logs:     logs,
		Finished: deployFinished,
		Success:  deploySuccess,
	}, nil
}

// GetClawStatus 获取OpenClaw运行状态
func (s *DeployService) GetClawStatus() (*dto.ClawStatusResp, error) {
	if getDeployMode() == "local" {
		return s.getLocalClawStatus()
	}
	return s.getDockerClawStatus()
}

// getLocalClawStatus 获取本地部署状态
func (s *DeployService) getLocalClawStatus() (*dto.ClawStatusResp, error) {
	resp := &dto.ClawStatusResp{
		ContainerName: "openclaw",
		Image:         "本地编译",
	}

	// 读取配置获取端口
	configPath := getOpenClawConfigPath()
	data, err := os.ReadFile(configPath)
	if err == nil {
		resp.Installed = true // 配置文件存在即认为已安装
		var config map[string]any
		if json.Unmarshal(data, &config) == nil {
			if gw, ok := config["gateway"].(map[string]any); ok {
				if port, ok := gw["port"].(float64); ok {
					resp.WebPort = int(port)
				}
			}
		}
	} else {
		// 配置文件不存在，但 openclaw 二进制存在也算已安装
		if _, err := exec.LookPath("openclaw"); err == nil {
			resp.Installed = true
		}
	}

	// 用 systemctl is-active 检测服务状态
	out, err := exec.Command("systemctl", "is-active", "openclaw").Output()
	status := strings.TrimSpace(string(out))
	if err == nil && status == "active" {
		resp.Running = true
		resp.Status = "running"
		resp.Uptime = "-"
	} else {
		resp.Running = false
		resp.Status = status // inactive / failed / etc.
		resp.Uptime = "-"
	}

	return resp, nil
}

// WS 代理信息（运行时设置）
var wsProxyPort int
var wsProxyToken string

// SetWsProxyInfo 设置 WS 代理端口和认证令牌
func SetWsProxyInfo(port int, token string) {
	wsProxyPort = port
	wsProxyToken = token
}

// GetClawWsInfo 获取 WS 代理连接信息（前端用）
func (s *DeployService) GetClawWsInfo() (*dto.ClawWsInfoResp, error) {
	if wsProxyPort == 0 {
		return nil, fmt.Errorf("WS 代理未启动")
	}
	return &dto.ClawWsInfoResp{
		Port:  wsProxyPort,
		Token: wsProxyToken,
	}, nil
}

// === WS 代理配置持久化 ===

type wsProxyConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	UseSSL  bool `json:"useSSL"`
}

func getWsProxyConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gmclaw_ws.json")
}

func loadWsProxyConfig() wsProxyConfig {
	cfg := wsProxyConfig{Enabled: false, Port: 37300}
	data, err := os.ReadFile(getWsProxyConfigPath())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	if cfg.Port == 0 {
		cfg.Port = 37300
	}
	return cfg
}

func saveWsProxyConfig(cfg wsProxyConfig) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(getWsProxyConfigPath(), data, 0644)
}

// AutoStartWsProxy 服务启动时自动恢复 WS 代理（如果之前已启用）
func AutoStartWsProxy() {
	cfg := loadWsProxyConfig()
	if !cfg.Enabled {
		fmt.Println("WS 代理未启用（可在对话菜单中开启）")
		return
	}
	proxy := GetGlobalProxy()
	port, err := proxy.Start(cfg.Port, cfg.UseSSL)
	if err != nil {
		fmt.Printf("WS 代理自动启动失败: %v\n", err)
		return
	}
	fmt.Printf("WS 代理已启动, port: %d\n", port)
}

// GetWsProxyStatus 获取 WS 代理状态
func (s *DeployService) GetWsProxyStatus() (map[string]any, error) {
	cfg := loadWsProxyConfig()
	proxy := GetGlobalProxy()
	return map[string]any{
		"enabled": cfg.Enabled,
		"port":    cfg.Port,
		"running": proxy.IsRunning(),
	}, nil
}

// ToggleWsProxy 开启或关闭 WS 代理
func (s *DeployService) ToggleWsProxy(req map[string]any) (map[string]any, error) {
	enabled, _ := req["enabled"].(bool)
	portF, hasPort := req["port"].(float64)
	force, _ := req["force"].(bool)

	useSSL, _ := req["useSSL"].(bool)

	cfg := loadWsProxyConfig()
	if hasPort && int(portF) > 0 {
		cfg.Port = int(portF)
	}
	cfg.Enabled = enabled
	cfg.UseSSL = useSSL
	saveWsProxyConfig(cfg)

	proxy := GetGlobalProxy()

	if enabled {
		// 已经在目标端口运行且不强制重启，直接返回成功
		if proxy.IsRunning() && proxy.GetPort() == cfg.Port && !force {
			return map[string]any{
				"success": true,
				"port":    cfg.Port,
				"message": fmt.Sprintf("WS 代理已在端口 %d 运行", cfg.Port),
			}, nil
		}
		// 端口变了或未运行，先停后启
		if proxy.IsRunning() {
			proxy.Stop()
		}
		port, err := proxy.Start(cfg.Port, cfg.UseSSL)
		if err != nil {
			return map[string]any{
				"success": false,
				"message": fmt.Sprintf("启动失败: %v", err),
			}, nil
		}
		return map[string]any{
			"success": true,
			"port":    port,
			"message": fmt.Sprintf("WS 代理已启动在端口 %d，请确保防火墙已放行该端口", port),
		}, nil
	}

	// 关闭
	proxy.Stop()
	return map[string]any{
		"success": true,
		"message": "WS 代理已关闭",
	}, nil
}

// GetRecentLogs 获取 OpenClaw 近期日志
func (s *DeployService) GetRecentLogs(req map[string]any) (map[string]any, error) {
	countF, _ := req["count"].(float64)
	count := int(countF)
	if count <= 0 || count > 50 {
		count = 10
	}

	// 优先使用前端传入的 mode，避免 getDeployMode 误判
	mode, _ := req["mode"].(string)
	if mode == "" {
		mode = getDeployMode()
	}
	var cmd *exec.Cmd
	if mode == "" {
		// 未安装任何服务，直接返回空日志
		return map[string]any{"logs": []string{}}, nil
	} else if mode == "local" {
		cmd = exec.Command("journalctl", "-u", "openclaw", "-n", fmt.Sprintf("%d", count), "--no-pager", "-o", "short-iso")
	} else {
		cmd = exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", count), "gmssh-openclaw")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[GetRecentLogs] mode=%s, cmd error: %v, output: %s\n", mode, err, string(out))
		return map[string]any{
			"logs":  []string{fmt.Sprintf("获取日志失败 (mode=%s): %v", mode, err)},
		}, nil
	}

	// 去除 ANSI 转义码
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleaned := ansiRe.ReplaceAllString(string(out), "")

	lines := strings.Split(strings.TrimSpace(cleaned), "\n")
	var filtered []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			filtered = append(filtered, l)
		}
	}

	return map[string]any{
		"logs": filtered,
	}, nil
}

// getDockerClawStatus 获取 Docker 容器状态
func (s *DeployService) getDockerClawStatus() (*dto.ClawStatusResp, error) {
	resp := &dto.ClawStatusResp{
		ContainerName: "gmssh-openclaw",
		Image:         "gmssh/openclaw:2026.02.17",
	}

	// 检测容器状态
	out, err := exec.Command("docker", "inspect", "--format",
		"{{.State.Status}}|{{.State.StartedAt}}", "gmssh-openclaw").Output()
	if err != nil {
		// 容器不存在，检查配置文件判断是否曾安装过
		if _, cfgErr := os.Stat(getOpenClawConfigPath()); cfgErr == nil {
			resp.Installed = true
		}
		resp.Running = false
		resp.Status = "stopped"
		resp.Uptime = "-"
		return resp, nil
	}

	// 容器存在即已安装
	resp.Installed = true

	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) >= 2 {
		resp.Status = parts[0]
		resp.Running = parts[0] == "running"

		// 计算运行时间
		if startedAt, err := time.Parse(time.RFC3339Nano, parts[1]); err == nil {
			duration := time.Since(startedAt)
			if duration.Hours() >= 24 {
				resp.Uptime = fmt.Sprintf("%.0f 天", duration.Hours()/24)
			} else if duration.Hours() >= 1 {
				resp.Uptime = fmt.Sprintf("%.0f 小时", duration.Hours())
			} else {
				resp.Uptime = fmt.Sprintf("%.0f 分钟", duration.Minutes())
			}
		}
	}

	// 读取端口配置 - 动态从容器获取实际映射端口
	portOut, err := exec.Command("docker", "inspect", "--format",
		"{{range $p, $conf := .NetworkSettings.Ports}}{{(index $conf 0).HostPort}} {{end}}",
		"gmssh-openclaw").Output()
	if err == nil {
		var ports []int
		for _, p := range strings.Fields(strings.TrimSpace(string(portOut))) {
			var port int
			if _, err := fmt.Sscanf(p, "%d", &port); err == nil && port > 0 {
				ports = append(ports, port)
			}
		}
		// 较小的端口是 Web 端口，较大的是 Bridge 端口
		if len(ports) >= 2 {
			if ports[0] < ports[1] {
				resp.WebPort = ports[0]
				resp.BridgePort = ports[1]
			} else {
				resp.WebPort = ports[1]
				resp.BridgePort = ports[0]
			}
		} else if len(ports) == 1 {
			resp.WebPort = ports[0]
		}
	}

	return resp, nil
}

// StopClaw 停止 OpenClaw
func (s *DeployService) StopClaw() (map[string]any, error) {
	if getDeployMode() == "local" {
		return s.stopLocalClaw()
	}
	out, err := exec.Command("docker", "stop", "gmssh-openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("停止容器失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "容器已停止"}, nil
}

// RestartClaw 重启 OpenClaw
func (s *DeployService) RestartClaw() (map[string]any, error) {
	if getDeployMode() == "local" {
		return s.restartLocalClaw()
	}
	out, err := exec.Command("docker", "restart", "gmssh-openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("重启容器失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "容器已重启"}, nil
}

// UninstallClaw 卸载 OpenClaw
func (s *DeployService) UninstallClaw() (map[string]any, error) {
	if getDeployMode() == "local" {
		return s.uninstallLocalClaw()
	}
	// Docker 卸载
	// 先获取容器使用的镜像名（避免硬编码）
	imageOut, _ := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", "gmssh-openclaw").Output()
	imageName := strings.TrimSpace(string(imageOut))

	exec.Command("docker", "stop", "gmssh-openclaw").CombinedOutput()
	exec.Command("docker", "rm", "-f", "gmssh-openclaw").CombinedOutput()

	// 仅在获取到镜像名时才删除镜像
	if imageName != "" {
		exec.Command("docker", "rmi", "-f", imageName).CombinedOutput()
	}

	dataDir := getDataDir()
	if err := os.RemoveAll(dataDir); err != nil {
		return nil, fmt.Errorf("清理数据目录失败: %v", err)
	}
	os.Remove(getDeployModeFile())
	return map[string]any{"success": true, "message": "已完全卸载"}, nil
}

// ===== 本地模式操作（systemd） =====

// killOpenclawProcess 强制杀死所有 openclaw 相关进程（忽略失败）
// 使用 -f 匹配完整命令行，确保 openclaw.mjs / openclaw-gateway 等子进程均被清除。
// kill 后等待 2 秒，让内核释放端口，避免 systemctl restart 时端口仍被占用。
func killOpenclawProcess() {
	exec.Command("pkill", "-9", "-f", "openclaw").Run()
	time.Sleep(2 * time.Second)
}

func (s *DeployService) stopLocalClaw() (map[string]any, error) {
	killOpenclawProcess()
	out, err := exec.Command("systemctl", "stop", "openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("停止服务失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "服务已停止"}, nil
}

func (s *DeployService) restartLocalClaw() (map[string]any, error) {
	killOpenclawProcess()
	out, err := exec.Command("systemctl", "restart", "openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("重启服务失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "服务已重启"}, nil
}

func (s *DeployService) uninstallLocalClaw() (map[string]any, error) {
	// 强制杀进程 + 停止并禁用服务
	killOpenclawProcess()
	exec.Command("systemctl", "stop", "openclaw").CombinedOutput()
	exec.Command("systemctl", "disable", "openclaw").CombinedOutput()
	os.Remove("/etc/systemd/system/openclaw.service")
	exec.Command("systemctl", "daemon-reload").Run()

	// 清理部署目录
	cloneDir := getLocalDeployDir()
	if err := os.RemoveAll(cloneDir); err != nil {
		return nil, fmt.Errorf("清理部署目录失败: %v", err)
	}

	// 清理配置和链接
	os.RemoveAll(getOpenClawConfigDir())
	os.Remove("/usr/local/bin/openclaw")
	os.Remove(getDeployModeFile())

	return map[string]any{"success": true, "message": "已完全卸载"}, nil
}

// UpdateModelConfig 切换AI模型配置
func (s *DeployService) UpdateModelConfig(req dto.UpdateModelReq) (map[string]any, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	// 处理模型名
	modelRef := req.Model
	modelName := req.Model
	if parts := strings.SplitN(req.Model, "/", 2); len(parts) == 2 {
		modelName = parts[1]
	} else {
		modelRef = req.Provider + "/" + req.Model
	}

	// ====== Gemini 特殊处理：使用 OpenClaw 内置 google 供应商 ======
	if req.Provider == "gemini" {
		// 设置 env.GOOGLE_API_KEY
		envMap, _ := config["env"].(map[string]any)
		if envMap == nil {
			envMap = map[string]any{}
		}
		envMap["GOOGLE_API_KEY"] = req.ApiKey
		config["env"] = envMap

		// 更新 agents.defaults.model.primary（modelRef 已经是 google/xxx 格式）
		if agents, ok := config["agents"].(map[string]any); ok {
			if defaults, ok := agents["defaults"].(map[string]any); ok {
				if model, ok := defaults["model"].(map[string]any); ok {
					model["primary"] = modelRef
				}
			}
		}

		// 清除 models.providers（Gemini 使用内置供应商，不需要自定义 provider）
		if models, ok := config["models"].(map[string]any); ok {
			models["providers"] = map[string]any{}
		}

		var geminiConfPath string
		if getDeployMode() == "local" {
			geminiConfPath = getOpenClawConfigPath()
		} else {
			geminiConfPath = filepath.Join(getDataDir(), "conf", "openclaw.json")
		}
		if err := writeGeminiExtraFiles(config, req.ApiKey, geminiConfPath); err != nil {
			return nil, fmt.Errorf("写入 Gemini 配置失败: %v", err)
		}

		// 重启服务使配置生效
		if getDeployMode() == "local" {
			killOpenclawProcess()
			if out, err := exec.Command("systemctl", "restart", "openclaw").CombinedOutput(); err != nil {
				return nil, fmt.Errorf("重启服务失败: %s", strings.TrimSpace(string(out)))
			}
		} else {
			if out, err := exec.Command("docker", "restart", "gmssh-openclaw").CombinedOutput(); err != nil {
				return nil, fmt.Errorf("重启容器失败: %s", strings.TrimSpace(string(out)))
			}
		}
		return map[string]any{"success": true, "message": "模型已切换为 " + modelRef}, nil
	}

	// ====== 非 Gemini 供应商：通过 models.providers 配置 ======

	// 确定 API 协议
	apiProtocol := "openai-completions"
	switch req.ApiMode {
	case "anthropic", "anthropic-messages":
		apiProtocol = "anthropic-messages"
	}

	// Ollama 使用本地推理，apiKey 不能为空
	apiKey := req.ApiKey
	if req.Provider == "ollama" && apiKey == "" {
		apiKey = "ollama-local"
	}

	// Ollama 本地模型上下文窗口和 maxTokens 参数
	contextWindow := 131072 // 128k
	maxTokens := 8192
	if req.Provider == "ollama" {
		maxTokens = 81920 // 8192 * 10，本地无限制
	}

	// 更新 agents.defaults.model.primary
	if agents, ok := config["agents"].(map[string]any); ok {
		if defaults, ok := agents["defaults"].(map[string]any); ok {
			if model, ok := defaults["model"].(map[string]any); ok {
				model["primary"] = modelRef
			}
		}
	}

	// 清除之前可能遗留的 Gemini env key
	if envMap, ok := config["env"].(map[string]any); ok {
		delete(envMap, "GOOGLE_API_KEY")
	}

	// 替换 models.providers（只保留新的 provider）
	if models, ok := config["models"].(map[string]any); ok {
		models["providers"] = map[string]any{
			req.Provider: map[string]any{
				"api":     apiProtocol,
				"apiKey":  apiKey,
				"baseUrl": req.BaseUrl,
				"models": []map[string]any{
					{
						"contextWindow": contextWindow,
						"cost": map[string]any{
							"cacheRead": 0, "cacheWrite": 0,
							"input": 0, "output": 0,
						},
						"id":        modelName,
						"input":     []string{"text"},
						"maxTokens": maxTokens,
						"name":      modelName,
						"reasoning": false,
					},
				},
			},
		}
	}

	if err := writeOpenClawConfig(config); err != nil {
		return nil, fmt.Errorf("写入配置失败: %v", err)
	}

	// 重启服务使配置生效（根据部署模式选择重启方式）
	if getDeployMode() == "local" {
		killOpenclawProcess()
		if out, err := exec.Command("systemctl", "restart", "openclaw").CombinedOutput(); err != nil {
			return nil, fmt.Errorf("重启服务失败: %s", strings.TrimSpace(string(out)))
		}
	} else {
		if out, err := exec.Command("docker", "restart", "gmssh-openclaw").CombinedOutput(); err != nil {
			return nil, fmt.Errorf("重启容器失败: %s", strings.TrimSpace(string(out)))
		}
	}

	return map[string]any{"success": true, "message": "模型已切换为 " + modelRef}, nil
}

// UpdateMemoryConfig 更新记忆相关配置
func (s *DeployService) UpdateMemoryConfig(req map[string]any) (map[string]any, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	// 确保 agents.defaults 存在
	agents, _ := config["agents"].(map[string]any)
	if agents == nil {
		agents = map[string]any{}
	}
	defaults, _ := agents["defaults"].(map[string]any)
	if defaults == nil {
		defaults = map[string]any{}
	}

	// 处理 agents.defaults.compaction.memoryFlush.enabled
	if memFlush, ok := req["memoryFlushEnabled"]; ok {
		enabled, _ := memFlush.(bool)
		compaction, _ := defaults["compaction"].(map[string]any)
		if compaction == nil {
			compaction = map[string]any{}
		}
		compaction["memoryFlush"] = map[string]any{"enabled": enabled}
		defaults["compaction"] = compaction
	}

	// 处理 agents.defaults.memorySearch.experimental.sessionMemory
	if memSearch, ok := req["sessionMemoryEnabled"]; ok {
		enabled, _ := memSearch.(bool)
		memorySearch, _ := defaults["memorySearch"].(map[string]any)
		if memorySearch == nil {
			memorySearch = map[string]any{}
		}
		experimental, _ := memorySearch["experimental"].(map[string]any)
		if experimental == nil {
			experimental = map[string]any{}
		}
		if enabled {
			experimental["sessionMemory"] = true
		} else {
			delete(experimental, "sessionMemory")
		}
		memorySearch["experimental"] = experimental
		defaults["memorySearch"] = memorySearch
	}

	agents["defaults"] = defaults
	config["agents"] = agents

	if err := writeOpenClawConfig(config); err != nil {
		return nil, fmt.Errorf("写入配置失败: %v", err)
	}

	return map[string]any{"success": true, "message": "记忆配置已更新"}, nil
}

// TestApiConnection 测试AI API连通性
func (s *DeployService) TestApiConnection(req dto.TestApiReq) (*dto.TestApiResp, error) {
	resp := &dto.TestApiResp{}

	baseUrl := strings.TrimRight(req.BaseUrl, "/")
	apiMode := req.ApiMode
	if apiMode == "" {
		apiMode = "openai" // 默认 OpenAI 协议
	}

	// 根据 API 协议模式构造测试请求
	testUrl := baseUrl + "/models"
	testMethod := "GET"
	var testBody *strings.Reader

	switch apiMode {
	case "anthropic":
		// Anthropic Messages 协议: POST /v1/messages
		testUrl = baseUrl + "/v1/messages"
		testMethod = "POST"
		testBody = strings.NewReader(`{"model":"test","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`)
	case "gemini":
		// Google Generative AI: GET /v1beta/models（用 header 认证）
		testUrl = baseUrl + "/v1beta/models"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var httpReq *http.Request
	var err error
	if testBody != nil {
		httpReq, err = http.NewRequest(testMethod, testUrl, testBody)
	} else {
		httpReq, err = http.NewRequest(testMethod, testUrl, nil)
	}
	if err != nil {
		resp.Reachable = false
		resp.Message = "请求构造失败: " + err.Error()
		return resp, nil
	}

	// 根据协议模式设置认证头
	switch apiMode {
	case "anthropic":
		httpReq.Header.Set("x-api-key", req.ApiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	case "gemini":
		// Google Generative AI 使用 x-goog-api-key header
		httpReq.Header.Set("x-goog-api-key", req.ApiKey)
	default:
		httpReq.Header.Set("Authorization", "Bearer "+req.ApiKey)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	latency := time.Since(start).Milliseconds()
	resp.LatencyMs = latency

	if err != nil {
		resp.Reachable = false
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			resp.Message = "连接超时，请检查网络或API地址"
		} else if strings.Contains(err.Error(), "no such host") {
			resp.Message = "域名无法解析，请检查API地址"
		} else if strings.Contains(err.Error(), "connection refused") {
			resp.Message = "连接被拒绝，请确认服务是否运行"
		} else {
			resp.Message = "连接失败: " + err.Error()
		}
		return resp, nil
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 200 {
		resp.Reachable = true
		resp.Message = fmt.Sprintf("连接成功 (%dms)", latency)
	} else if httpResp.StatusCode == 401 || httpResp.StatusCode == 403 || httpResp.StatusCode == 405 {
		// 401/403/405 说明地址可达但需要认证或方法不对，算连通成功
		resp.Reachable = true
		resp.Message = fmt.Sprintf("API 地址可达 (%dms)", latency)
	} else if httpResp.StatusCode == 404 {
		resp.Reachable = false
		resp.Message = "接口路径不存在 (404)，请检查 API 地址是否正确"
	} else {
		resp.Reachable = false
		resp.Message = fmt.Sprintf("API返回异常状态码: %d", httpResp.StatusCode)
	}

	return resp, nil
}

// generateComposeFile 生成docker-compose.yml内容
// 所有文件和 volumes 统一使用 /opt/gmclaw
func (s *DeployService) generateComposeFile(req dto.DeployReq, dataDir string) string {
	return fmt.Sprintf(`services:
  gmssh-openclaw:
    container_name: gmssh-openclaw
    image: docker-rep.gmssh.com/openclaw/gmssh-openclaw:2026.03.11
    restart: unless-stopped
    environment:
      - HOME=/home/node
      - TERM=xterm-256color
      - OPENCLAW_GATEWAY_TOKEN=${OPENCLAW_GATEWAY_TOKEN}
      - NODE_ENV=production
    volumes:
      - %s/conf:/home/node/.openclaw
      - %s/workspace:/home/node/.openclaw/workspace
    ports:
      - "%d:%d"
      - "%d:%d"
    init: true
    command:
      [
        "openclaw",
        "gateway",
        "--bind",
        "lan",
        "--port",
        "%d"
      ]

networks:
  gmssh-network:
    external: true
`, dataDir, dataDir, req.WebPort, req.WebPort, req.BridgePort, req.BridgePort, req.WebPort)
}

// generateOpenClawConfig 生成openclaw.json配置
func (s *DeployService) generateOpenClawConfig(req dto.DeployReq) map[string]any {
	// 从model ID中提取实际模型名 (如 "deepseek/deepseek-chat" -> "deepseek-chat")
	modelRef := req.Model
	modelName := req.Model
	if parts := strings.SplitN(req.Model, "/", 2); len(parts) == 2 {
		modelName = parts[1]
	} else {
		// 模型名没有 provider 前缀时自动添加（如自定义输入 "mimo-v2-flash" → "custom/mimo-v2-flash"）
		modelRef = req.Provider + "/" + req.Model
	}

	// 获取 baseUrl 映射
	providerBaseUrls := map[string]string{
		"deepseek":  "https://api.deepseek.com/v1",
		"openai":    "https://api.openai.com/v1",
		"alibaba":   "https://dashscope.aliyuncs.com/compatible-mode/v1",
		"anthropic": "https://api.anthropic.com",
		"gemini":    "https://generativelanguage.googleapis.com",
		"kimi":      "https://api.moonshot.cn/v1",
		"minimax":   "https://api.minimaxi.com/anthropic",
		"ollama":    "http://localhost:11434/v1",
	}
	baseUrl := providerBaseUrls[req.Provider]
	// 自定义 baseUrl 覆盖（Ollama 端口、自定义接口等）
	if req.CustomBaseUrl != "" {
		baseUrl = req.CustomBaseUrl
	}
	if baseUrl == "" {
		baseUrl = "https://api.openai.com/v1"
	}

	// 根据提供商确定 API 协议
	// 优先使用前端显式传入的 ApiMode（自定义模式下用户可以选择 anthropic-messages 等）
	apiProtocol := req.ApiMode
	if apiProtocol == "" {
		// 未传入时按供应商默认
		apiProtocol = "openai-completions"
		if req.Provider == "anthropic" || req.Provider == "minimax" {
			apiProtocol = "anthropic-messages"
		}
	}
	// 规范化协议值：前端可能传 "anthropic" / "openai" 简写，OpenClaw 配置需要完整屗面量
	switch apiProtocol {
	case "anthropic":
		apiProtocol = "anthropic-messages"
	case "openai":
		apiProtocol = "openai-completions"
	case "gemini":
		// gemini 协议值直接透传，无需变换
	}

	// Ollama 使用本地推理，apiKey 不能为空，否则部分版本 OpenClaw 会拒绝
	apiKey := req.ApiKey
	if req.Provider == "ollama" && apiKey == "" {
		apiKey = "ollama-local"
	}

	// Ollama 本地模型上下文窗口和 maxTokens 参数
	contextWindow := 131072 // 128k
	maxTokens := 8192
	if req.Provider == "ollama" {
		maxTokens = 81920 // 8192 * 10，本地无限制
	}

	// ====== Gemini 特殊处理：使用 OpenClaw 内置 google 供应商 ======
	if req.Provider == "gemini" {
		return map[string]any{
			"env": map[string]any{
				"GOOGLE_API_KEY": req.ApiKey,
			},
			"agents": map[string]any{
				"defaults": map[string]any{
					"model": map[string]any{
						"primary": modelRef,
					},
				},
			},
			"gateway": map[string]any{
				"auth": map[string]any{
					"mode":  "token",
					"token": req.Token,
				},
				"bind": "lan",
				"controlUi": map[string]any{
					"allowInsecureAuth":                        true,
					"dangerouslyAllowHostHeaderOriginFallback": true,
					"dangerouslyDisableDeviceAuth":             true,
				},
				"mode": "local",
				"port": req.WebPort,
			},
		}
	}

	return map[string]any{
		"agents": map[string]any{
			"defaults": map[string]any{
				"model": map[string]any{
					"primary": modelRef,
				},
			},
		},
		"gateway": map[string]any{
			"auth": map[string]any{
				"mode":  "token",
				"token": req.Token,
			},
			"bind": "lan",
			"controlUi": map[string]any{
				"allowInsecureAuth":                        true,
				"dangerouslyAllowHostHeaderOriginFallback": true,
				"dangerouslyDisableDeviceAuth":             true,
			},
			"mode": "local",
			"port": req.WebPort,
		},
		"models": map[string]any{
			"mode": "merge",
			"providers": map[string]any{
				req.Provider: map[string]any{
					"api":     apiProtocol,
					"apiKey":  apiKey,
					"baseUrl": baseUrl,
					"models": []map[string]any{
						{
							"contextWindow": contextWindow,
							"cost": map[string]any{
								"cacheRead":  0,
								"cacheWrite": 0,
								"input":      0,
								"output":     0,
							},
							"id":        modelName,
							"input":     []string{"text"},
							"maxTokens": maxTokens,
							"name":      modelName,
							"reasoning": false,
						},
					},
				},
			},
		},
	}
}

// runDockerCompose 异步执行docker compose部署（实时流式输出）
func (s *DeployService) runDockerCompose(dataDir, composeFile string, webPort int) {
	addDeployLog("📁 配置文件已生成")

	// ====== 第一步：拉取镜像（独立步骤，避免长时间拉取影响容器创建）======
	addDeployLog("🐳 正在拉取镜像，请耐心等待...")

	composeName, composePrefix, _ := getDockerComposeCmd()
	pullArgs := make([]string, 0, len(composePrefix)+5)
	pullArgs = append(pullArgs, composePrefix...)
	pullArgs = append(pullArgs, "-f", composeFile, "-p", "gmclaw", "pull")
	pullCmd := exec.Command(composeName, pullArgs...)
	pullCmd.Dir = dataDir

	pullStderr, _ := pullCmd.StderrPipe()
	pullStdout, _ := pullCmd.StdoutPipe()

	if err := pullCmd.Start(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 拉取镜像命令启动失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	// 实时读取 pull 输出
	go func() {
		scanner := bufio.NewScanner(pullStderr)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()
	go func() {
		if pullStdout != nil {
			scanner := bufio.NewScanner(pullStdout)
			scanner.Buffer(make([]byte, 64*1024), 64*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					addDeployLog(line)
				}
			}
		}
	}()

	if err := pullCmd.Wait(); err != nil {
		addDeployLog(fmt.Sprintf("⚠️ 镜像拉取异常: %v（将尝试继续启动）", err))
		// 不立刻失败，可能本地已有缓存镜像
	} else {
		addDeployLog("✅ 镜像拉取完成")
	}

	// ====== 第二步：清理可能的残留容器 ======
	exec.Command("docker", "rm", "-f", "gmssh-openclaw").CombinedOutput()

	// ====== 第三步：创建并启动容器 ======
	addDeployLog("🚀 正在创建并启动容器...")

	upArgs := make([]string, 0, len(composePrefix)+6)
	upArgs = append(upArgs, composePrefix...)
	upArgs = append(upArgs, "-f", composeFile, "-p", "gmclaw", "up", "-d")
	upCmd := exec.Command(composeName, upArgs...)
	upCmd.Dir = dataDir

	upStderr, err := upCmd.StderrPipe()
	if err != nil {
		addDeployLog(fmt.Sprintf("❌ 创建输出管道失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}
	upStdout, _ := upCmd.StdoutPipe()

	if err := upCmd.Start(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 启动命令失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
		return
	}

	go func() {
		scanner := bufio.NewScanner(upStderr)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				addDeployLog(line)
			}
		}
	}()
	go func() {
		if upStdout != nil {
			scanner := bufio.NewScanner(upStdout)
			scanner.Buffer(make([]byte, 64*1024), 64*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					addDeployLog(line)
				}
			}
		}
	}()

	err = upCmd.Wait()

	deployLock.Lock()
	defer deployLock.Unlock()

	if err != nil {
		addDeployLogLocked(fmt.Sprintf("❌ 容器启动失败: %v", err))
		deployFinished = true
		deploySuccess = false
	} else {
		// 验证容器是否真正运行
		verifyOut, verifyErr := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "gmssh-openclaw").Output()
		if verifyErr != nil || strings.TrimSpace(string(verifyOut)) != "running" {
			addDeployLogLocked("⚠️ 容器创建完成但未正常运行，请检查 Docker 日志")
			deployFinished = true
			deploySuccess = false
		} else {
			addDeployLogLocked("✅ 容器已成功启动")

			// 设置容器内 npm/pnpm 镜像加速
			addDeployLogLocked("⚙️ 正在配置 npm 镜像加速...")
			exec.Command("docker", "exec", containerName, "npm", "config", "set", "registry", "https://registry.npmmirror.com").Run()
			exec.Command("docker", "exec", containerName, "sh", "-c", "yes | pnpm config set registry https://registry.npmmirror.com").Run()
			addDeployLogLocked("✅ npm/pnpm 镜像加速已配置")

			addDeployLogLocked(fmt.Sprintf("🔥 请确保防火墙已放开端口 %d 的访问，否则外部将无法连接 Web UI", webPort))
			deployFinished = true
			deploySuccess = true
		}
	}
}

func addDeployLog(msg string) {
	deployLock.Lock()
	defer deployLock.Unlock()
	deployLogs = append(deployLogs, msg)
}

func addDeployLogLocked(msg string) {
	deployLogs = append(deployLogs, msg)
}

// GetOpenClawConfigDirPath 返回 OpenClaw 配置目录路径（区分部署模式）
func (s *DeployService) GetOpenClawConfigDirPath() map[string]any {
	var dir string
	if getDeployMode() == "local" {
		dir = getOpenClawConfigDir() // ~/.openclaw
	} else {
		dir = filepath.Join(getDataDir(), "conf") // /opt/gmclaw/conf
	}
	return map[string]any{"path": dir, "mode": getDeployMode()}
}

// parseNodeMajorVersion 从 "v22.22.1" 类似字符串中解析出主要版本号
func parseNodeMajorVersion(v string) int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		var major int
		fmt.Sscanf(parts[0], "%d", &major)
		return major
	}
	return 0
}
