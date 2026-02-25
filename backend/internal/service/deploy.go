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
	"strings"
	"sync"
	"guanxi/eazy-claw/internal/dto"
	"time"
)

// DeployService 部署服务
type DeployService struct{}

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

// 部署模式持久化
func getDeployModeFile() string {
	return filepath.Join(getTmpDir(), "deploy_mode")
}

func getDeployMode() string {
	data, err := os.ReadFile(getDeployModeFile())
	if err != nil {
		return "docker"
	}
	mode := strings.TrimSpace(string(data))
	if mode == "local" {
		return "local"
	}
	return "docker"
}

func saveDeployMode(mode string) {
	os.MkdirAll(getTmpDir(), 0755)
	os.WriteFile(getDeployModeFile(), []byte(mode), 0644)
}

const (
	pluginBinPath = "/.__gmssh/plugin/official/docker/app/bin"
)

// getDataDir 获取数据目录 /opt/gmclaw
func getDataDir() string {
	return "/opt/gmclaw"
}

// CheckEnvironment 检测Docker环境
func (s *DeployService) CheckEnvironment() (*dto.CheckEnvResp, error) {
	resp := &dto.CheckEnvResp{}

	// 1. 检测插件路径
	if _, err := os.Stat(pluginBinPath); err == nil {
		resp.PluginPathExists = true
	}

	// 2. 检测 docker 命令
	if err := exec.Command("docker", "--version").Run(); err == nil {
		resp.DockerReady = true
	}

	// 3. 检测 docker compose 命令
	if err := exec.Command("docker", "compose", "version").Run(); err == nil {
		resp.DockerComposeReady = true
	}

	// 4. 检测 Node.js
	if out, err := exec.Command("node", "--version").Output(); err == nil {
		resp.NodeReady = true
		resp.NodeVersion = strings.TrimSpace(string(out))
	}

	// 5. 检测 pnpm
	if err := exec.Command("pnpm", "--version").Run(); err == nil {
		resp.PnpmReady = true
	}

	resp.AllReady = resp.PluginPathExists && resp.DockerReady && resp.DockerComposeReady
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
	envContent := fmt.Sprintf("OPENCLAW_GATEWAY_MODE=local\nOPENCLAW_GATEWAY_TOKEN=%s\n", req.Token)
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

// getLocalDeployDir 本地部署目录 = tmpDir/openclaw-cn
func getLocalDeployDir() string {
	return filepath.Join(getTmpDir(), "openclaw-cn")
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
		cmd := exec.Command("git", "clone", "https://gitee.com/OpenClaw-CN/openclaw-cn.git", cloneDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			addDeployLog(fmt.Sprintf("❌ 克隆仓库失败: %s", strings.TrimSpace(string(out))))
			deployLock.Lock()
			deployFinished = true
			deploySuccess = false
			deployLock.Unlock()
			return
		}
		addDeployLog("✅ 仓库克隆完成")

		// 切换到稳定标签
		addDeployLog("🏷️ 切换到 v2026.2.2-cn 分支...")
		checkoutCmd := exec.Command("git", "-C", cloneDir, "checkout", "v2026.2.2-cn")
		if out, err := checkoutCmd.CombinedOutput(); err != nil {
			addDeployLog(fmt.Sprintf("⚠️ 切换分支失败，使用 main: %s", strings.TrimSpace(string(out))))
		}
	} else {
		addDeployLog("📦 项目已存在，跳过克隆")
	}

	// ====== 第二步: pnpm install ======
	addDeployLog("📥 正在安装依赖 (pnpm install)...")
	s.runStreamCmd(cloneDir, "pnpm", "install")

	if !s.checkDeployOK() {
		return
	}

	// ====== 第三步: pnpm ui:build ======
	addDeployLog("🎨 正在构建 UI 依赖 (pnpm ui:build)...")
	s.runStreamCmd(cloneDir, "pnpm", "ui:build")

	if !s.checkDeployOK() {
		return
	}

	// ====== 第四步: pnpm build ======
	addDeployLog("🔨 正在构建项目 (pnpm build)...")
	s.runStreamCmd(cloneDir, "pnpm", "build")

	if !s.checkDeployOK() {
		return
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
	nodePath, err := exec.LookPath("node")
	if err != nil {
		addDeployLog("❌ 未找到 node 命令，请确认 Node.js 已安装")
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

// runStreamCmd 执行命令并流式输出到部署日志
func (s *DeployService) runStreamCmd(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 命令启动失败: %v", err))
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

	if err := cmd.Wait(); err != nil {
		addDeployLog(fmt.Sprintf("❌ 命令执行失败: %v", err))
		deployLock.Lock()
		deployFinished = true
		deploySuccess = false
		deployLock.Unlock()
	}
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
		var config map[string]any
		if json.Unmarshal(data, &config) == nil {
			if gw, ok := config["gateway"].(map[string]any); ok {
				if port, ok := gw["port"].(float64); ok {
					resp.WebPort = int(port)
				}
			}
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
		resp.Running = false
		resp.Status = "stopped"
		resp.Uptime = "-"
		return resp, nil
	}

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
	exec.Command("docker", "stop", "gmssh-openclaw").CombinedOutput()
	exec.Command("docker", "rm", "-f", "gmssh-openclaw").CombinedOutput()
	exec.Command("docker", "rmi", "-f", "gmssh/openclaw:2026.02.17").CombinedOutput()

	dataDir := getDataDir()
	if err := os.RemoveAll(dataDir); err != nil {
		return nil, fmt.Errorf("清理数据目录失败: %v", err)
	}
	os.Remove(getDeployModeFile())
	return map[string]any{"success": true, "message": "已完全卸载"}, nil
}

// ===== 本地模式操作（systemd） =====

func (s *DeployService) stopLocalClaw() (map[string]any, error) {
	out, err := exec.Command("systemctl", "stop", "openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("停止服务失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "服务已停止"}, nil
}

func (s *DeployService) restartLocalClaw() (map[string]any, error) {
	out, err := exec.Command("systemctl", "restart", "openclaw").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("重启服务失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"success": true, "message": "服务已重启"}, nil
}

func (s *DeployService) uninstallLocalClaw() (map[string]any, error) {
	// 停止并禁用服务
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

	// 确定 API 协议
	apiProtocol := "openai-completions"
	if req.ApiMode == "anthropic" {
		apiProtocol = "anthropic-messages"
	}

	// 更新 agents.defaults.model.primary
	if agents, ok := config["agents"].(map[string]any); ok {
		if defaults, ok := agents["defaults"].(map[string]any); ok {
			if model, ok := defaults["model"].(map[string]any); ok {
				model["primary"] = modelRef
			}
		}
	}

	// 替换 models.providers（只保留新的 provider）
	if models, ok := config["models"].(map[string]any); ok {
		models["providers"] = map[string]any{
			req.Provider: map[string]any{
				"api":     apiProtocol,
				"apiKey":  req.ApiKey,
				"baseUrl": req.BaseUrl,
				"models": []map[string]any{
					{
						"contextWindow": 128000,
						"cost": map[string]any{
							"cacheRead": 0, "cacheWrite": 0,
							"input": 0, "output": 0,
						},
						"id":        modelName,
						"input":     []string{"text"},
						"maxTokens": 8192,
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
		testUrl = baseUrl + "/v1beta/models?key=" + req.ApiKey
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
		// Gemini 用 query param 传 key，不设 header
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
    image: gmssh/openclaw:2026.02.17
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
	apiProtocol := "openai-completions"
	if req.Provider == "anthropic" || req.Provider == "minimax" {
		apiProtocol = "anthropic-messages"
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
				"allowInsecureAuth": true,
			},
			"mode": "local",
			"port": req.WebPort,
		},
		"models": map[string]any{
			"mode": "merge",
			"providers": map[string]any{
				req.Provider: map[string]any{
					"api":     apiProtocol,
					"apiKey":  req.ApiKey,
					"baseUrl": baseUrl,
					"models": []map[string]any{
						{
							"contextWindow": 128000,
							"cost": map[string]any{
								"cacheRead":  0,
								"cacheWrite": 0,
								"input":      0,
								"output":     0,
							},
							"id":        modelName,
							"input":     []string{"text"},
							"maxTokens": 8192,
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

	pullCmd := exec.Command("docker", "compose", "-f", composeFile, "-p", "gmclaw", "pull")
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

	upCmd := exec.Command("docker", "compose", "-f", composeFile, "-p", "gmclaw", "up", "-d")
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
