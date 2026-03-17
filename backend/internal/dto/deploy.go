package dto

// CheckEnvResp 环境检测响应
type CheckEnvResp struct {
	DockerReady        bool   `json:"dockerReady"`
	DockerComposeReady bool   `json:"dockerComposeReady"`
	NodeReady          bool   `json:"nodeReady"`
	PnpmReady          bool   `json:"pnpmReady"`
	NodeVersion        string `json:"nodeVersion"`
	AllReady           bool   `json:"allReady"`
}

// GenerateTokenResp 生成Token响应
type GenerateTokenResp struct {
	Token string `json:"token"`
}

// DeployReq 部署请求
type DeployReq struct {
	Lang          string `json:"lang"`
	Token         string `json:"token"`
	WebPort       int    `json:"webPort"`
	BridgePort    int    `json:"bridgePort"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	ApiKey        string `json:"apiKey"`
	ApiMode       string `json:"apiMode"`       // openai | anthropic | anthropic-messages | 空=按供应商默认
	CustomBaseUrl string `json:"customBaseUrl"`
	DeployMode    string `json:"deployMode"` // docker | local
}

// DeployResp 部署响应
type DeployResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeployLogReq 部署日志请求
type DeployLogReq struct {
	Lang string `json:"lang"`
}

// DeployLogResp 部署日志响应
type DeployLogResp struct {
	Logs     []string `json:"logs"`
	Finished bool     `json:"finished"`
	Success  bool     `json:"success"`
}

// ClawStatusResp OpenClaw状态响应
type ClawStatusResp struct {
	Installed     bool   `json:"installed"`
	Running       bool   `json:"running"`
	ContainerName string `json:"containerName"`
	Status        string `json:"status"`
	WebPort       int    `json:"webPort"`
	BridgePort    int    `json:"bridgePort"`
	Uptime        string `json:"uptime"`
	Image         string `json:"image"`
}

// CheckPortsReq 端口检测请求
type CheckPortsReq struct {
	Ports []int `json:"ports"`
}

// PortStatus 单个端口状态
type PortStatus struct {
	Port      int    `json:"port"`
	Available bool   `json:"available"`
	Process   string `json:"process"`
}

// CheckPortsResp 端口检测响应
type CheckPortsResp struct {
	Results []PortStatus `json:"results"`
}

// ClawConfigResp OpenClaw配置响应
type ClawConfigResp struct {
	Provider     string            `json:"provider"`
	ModelName    string            `json:"modelName"`
	PrimaryModel string           `json:"primaryModel"`
	BaseUrl      string            `json:"baseUrl"`
	ApiKeyMasked string            `json:"apiKeyMasked"`
	ContextWindow int             `json:"contextWindow"`
	MaxTokens    int              `json:"maxTokens"`
	GatewayPort  int              `json:"gatewayPort"`
	AuthMode     string            `json:"authMode"`
	GatewayBind  string            `json:"gatewayBind"`
	GatewayMode  string            `json:"gatewayMode"`
	GatewayToken string            `json:"gatewayToken"`
	WebPort      int               `json:"webPort"`
	DeployMode   string            `json:"deployMode"`
	ConfigPath   string            `json:"configPath"`
	MemoryFlushEnabled   bool     `json:"memoryFlushEnabled"`
	SessionMemoryEnabled bool     `json:"sessionMemoryEnabled"`
}

// TestApiReq API连通性测试请求
type TestApiReq struct {
	BaseUrl  string `json:"baseUrl"`
	ApiKey   string `json:"apiKey"`
	Provider string `json:"provider"`
	ApiMode  string `json:"apiMode"` // openai | anthropic | gemini
}

// UpdateModelReq 切换AI模型请求
type UpdateModelReq struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	ApiKey   string `json:"apiKey"`
	BaseUrl  string `json:"baseUrl"`
	ApiMode  string `json:"apiMode"` // openai | anthropic
}
// TestApiResp API连通性测试响应
type TestApiResp struct {
	Reachable bool   `json:"reachable"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latencyMs"`
}

// ========== Agent 管理 ==========

// AgentFileItem 单个Agent文件
type AgentFileItem struct {
	Name    string `json:"name"`    // IDENTITY / USER / SOUL
	Content string `json:"content"` // markdown 内容
}

// AgentFilesResp 获取Agent文件响应
type AgentFilesResp struct {
	Files []AgentFileItem `json:"files"`
}

// AgentFileReq 读取单个Agent文件请求
type AgentFileReq struct {
	AgentID string `json:"agentId"` // Agent ID
	Name    string `json:"name"`    // IDENTITY / USER / SOUL 等
}

// AgentFileResp 单个文件响应
type AgentFileResp struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// AgentSaveReq 保存Agent文件请求
type AgentSaveReq struct {
	AgentID string `json:"agentId"` // Agent ID (空 or "main" = 主工作区)
	Name    string `json:"name"`    // IDENTITY / USER / SOUL 等
	Content string `json:"content"` // markdown 内容
}

// AgentTemplate 预设模板
type AgentTemplate struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Identity    string `json:"identity"`
	User        string `json:"user"`
	Soul        string `json:"soul"`
}

// AgentTemplatesResp 模板列表响应
type AgentTemplatesResp struct {
	Templates []AgentTemplate `json:"templates"`
}

// ApplyTemplateReq 应用模板请求
type ApplyTemplateReq struct {
	Key string `json:"key"` // 模板 key
}

// ========== 多 Agent 管理 ==========

// AgentInfo 单个 Agent 的完整信息
type AgentInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`        // main | specialist
	Description  string   `json:"description"`
	Status       string   `json:"status"`      // idle | thinking | acting | error
	Avatar       string   `json:"avatar"`
	Model        string   `json:"model"`
	ParentID     string   `json:"parentId"`
	AllowAgents  []string `json:"allowAgents,omitempty"` // 主控允许委派的子 Agent ID 列表
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// ListAgentsResp Agent 列表响应
type ListAgentsResp struct {
	Agents []AgentInfo `json:"agents"`
}

// CreateAgentReq 创建 Agent 请求
type CreateAgentReq struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Model       string `json:"model"`
	ParentID    string `json:"parentId"`
}

// UpdateAgentReq 更新 Agent 请求
type UpdateAgentReq struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Description  string   `json:"description"`
	Avatar       string   `json:"avatar"`
	Model        string   `json:"model"`
	ParentID     string   `json:"parentId"`
	AllowAgents  []string `json:"allowAgents"` // 允许委派的子 Agent ID 列表（仅主控有效）
}

// DeleteAgentReq 删除 Agent 请求
type DeleteAgentReq struct {
	ID string `json:"id"`
}

// GetAgentDetailReq 获取 Agent 详情请求
type GetAgentDetailReq struct {
	ID string `json:"id"`
}

// AgentDetailResp Agent 详情响应（含人格文件）
type AgentDetailResp struct {
	Agent AgentInfo       `json:"agent"`
	Files []AgentFileItem `json:"files"`
}

// ClawWsInfoResp WS 代理连接信息
type ClawWsInfoResp struct {
	Port  int    `json:"port"`  // WS 代理端口
	Token string `json:"token"` // WS 代理认证令牌
}
