package dto

// ========== Agent 实时状态 ==========

// AgentStateValue Agent 状态枚举
// working: 3分钟内有assistant消息
// online: 10分钟内有session活动
// idle: 24小时内有活动
// offline: 超过24小时无活动
type AgentStateValue string

const (
	AgentStateWorking AgentStateValue = "working"
	AgentStateOnline  AgentStateValue = "online"
	AgentStateIdle    AgentStateValue = "idle"
	AgentStateOffline AgentStateValue = "offline"
)

// AgentStatus 单个 Agent 的实时状态
type AgentStatus struct {
	AgentID    string          `json:"agentId"`
	State      AgentStateValue `json:"state"`
	LastActive int64           `json:"lastActive"` // Unix 毫秒时间戳，0 表示从未活跃
}

// GetAgentStatusesResp 获取所有 Agent 状态响应
type GetAgentStatusesResp struct {
	Statuses []AgentStatus `json:"statuses"`
}

// ========== Agent Token 统计 ==========

// GetAgentTokenStatsReq 获取 Token 统计请求
type GetAgentTokenStatsReq struct {
	Lang    string `json:"lang"`
	AgentID string `json:"agentId"` // 空字符串 = 返回所有 agent 聚合数据
	Days    int    `json:"days"`    // 统计天数，默认 7
}

// DailyStat 每天的统计数据
type DailyStat struct {
	Date          string `json:"date"`          // YYYY-MM-DD
	Tokens        int    `json:"tokens"`        // 当天 token 消耗
	AvgResponseMs int    `json:"avgResponseMs"` // 当天平均响应时间(ms)，0 表示无数据
	MessageCount  int    `json:"messageCount"`  // 当天消息数
}

// AgentTokenStats 单个 Agent 的 Token 统计
type AgentTokenStats struct {
	AgentID       string      `json:"agentId"`
	TotalTokens   int         `json:"totalTokens"`   // 统计区间内总 token
	TotalMessages int         `json:"totalMessages"` // 统计区间内消息数
	SessionCount  int         `json:"sessionCount"`  // 活跃会话数
	AvgResponseMs int         `json:"avgResponseMs"` // 整体平均响应时间(ms)
	Platforms     []string    `json:"platforms"`     // 活跃平台类型列表，如 ["feishu", "discord"]
	Daily         []DailyStat `json:"daily"`         // 按天统计（最近 N 天）
}

// GetAgentTokenStatsResp Token 统计响应
type GetAgentTokenStatsResp struct {
	Stats         []AgentTokenStats `json:"stats"`
	TotalTokens   int               `json:"totalTokens"`   // 所有 agent 总和
	TotalMessages int               `json:"totalMessages"` // 所有 agent 消息总和
}

// ========== Gateway 健康检测 ==========

// GatewayHealthResp Gateway 健康检测响应
type GatewayHealthResp struct {
	OK         bool   `json:"ok"`
	Status     string `json:"status"`      // healthy | degraded | down
	ResponseMs int64  `json:"responseMs"`  // 响应延迟(ms)
	Error      string `json:"error,omitempty"`
	CheckedAt  int64  `json:"checkedAt"`   // Unix 毫秒
}

// ========== 平台连通性测试 ==========

// TestPlatformConnReq 平台连通性测试请求
type TestPlatformConnReq struct {
	Lang     string `json:"lang"`
	AgentID  string `json:"agentId"`  // 空 = 测试所有 agent
	Platform string `json:"platform"` // feishu | discord | telegram | whatsapp | "" = 全部
}

// PlatformTestResult 单个平台测试结果
type PlatformTestResult struct {
	AgentID   string `json:"agentId"`
	Platform  string `json:"platform"`
	AccountID string `json:"accountId,omitempty"`
	OK        bool   `json:"ok"`
	Detail    string `json:"detail,omitempty"`
	Error     string `json:"error,omitempty"`
	Elapsed   int64  `json:"elapsed"` // 耗时(ms)
}

// TestPlatformConnResp 平台连通性测试响应
type TestPlatformConnResp struct {
	Results []PlatformTestResult `json:"results"`
}

// ========== Agent 会话列表 ==========

// GetAgentSessionsReq 获取会话列表请求
type GetAgentSessionsReq struct {
	Lang    string `json:"lang"`
	AgentID string `json:"agentId"` // 必填
}

// SessionItem 单个会话
type SessionItem struct {
	Key          string `json:"key"`
	Type         string `json:"type"`         // feishu-dm | feishu-group | discord-dm | discord-channel | telegram-dm | telegram-group | whatsapp-dm | whatsapp-group | cron | main | unknown
	Target       string `json:"target"`       // 对端 ID（open_id / channel_id 等）
	UpdatedAt    int64  `json:"updatedAt"`    // Unix 毫秒
	TotalTokens  int    `json:"totalTokens"`
	ContextTokens int   `json:"contextTokens"`
	SystemSent   bool   `json:"systemSent"`
}

// GetAgentSessionsResp 会话列表响应
type GetAgentSessionsResp struct {
	AgentID  string        `json:"agentId"`
	Sessions []SessionItem `json:"sessions"`
}
