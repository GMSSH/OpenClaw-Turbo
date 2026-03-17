package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"guanxi/eazy-claw/internal/dto"
)

// AgentStatService Agent 实时状态服务（借鉴 bot-view 文件扫描机制）
type AgentStatService struct{}

// NewAgentStatService 创建 AgentStat 服务实例
func NewAgentStatService() *AgentStatService {
	return &AgentStatService{}
}

// openclawAgentsDir 获取 ~/.openclaw/agents 目录
func openclawAgentsDir() string {
	return filepath.Join(getOpenClawConfigDir(), "agents")
}

// openclawAgentSessionsDir 获取指定 agent 的 sessions 目录
func openclawAgentSessionsDir(agentID string) string {
	return filepath.Join(openclawAgentsDir(), agentID, "sessions")
}

// listAgentIDsFromFilesystem 从 ~/.openclaw/agents/ 目录枚举所有 agent ID
func listAgentIDsFromFilesystem() []string {
	agentsDir := openclawAgentsDir()
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		// 降级：只有 main
		return []string{"main"}
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			ids = append(ids, e.Name())
		}
	}
	if len(ids) == 0 {
		return []string{"main"}
	}
	return ids
}

// getAgentState 检测单个 agent 的实时状态
// 借鉴 bot-view agent-status/route.ts 逻辑，低 I/O 高效实现
func getAgentState(agentID string) dto.AgentStatus {
	sessionsDir := openclawAgentSessionsDir(agentID)
	now := time.Now().UnixMilli()
	var lastActive int64
	var lastAssistantMs int64

	// Step 1: 从 sessions.json 获取最近活跃时间（updatedAt 字段）
	sessionsJSON := filepath.Join(sessionsDir, "sessions.json")
	if data, err := os.ReadFile(sessionsJSON); err == nil {
		var sessions map[string]any
		if json.Unmarshal(data, &sessions) == nil {
			for _, v := range sessions {
				if m, ok := v.(map[string]any); ok {
					if ts, ok := m["updatedAt"].(float64); ok {
						if int64(ts) > lastActive {
							lastActive = int64(ts)
						}
					}
				}
			}
		}
	}

	// Step 2: 扫描最近 5 个 .jsonl 文件，找最近 3 分钟内的 assistant 消息
	type fileInfo struct {
		name  string
		mtime int64
	}
	entries, err := os.ReadDir(sessionsDir)
	if err == nil {
		var jsonlFiles []fileInfo
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") || strings.Contains(e.Name(), ".deleted.") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			jsonlFiles = append(jsonlFiles, fileInfo{
				name:  e.Name(),
				mtime: info.ModTime().UnixMilli(),
			})
		}
		// 按修改时间降序
		sort.Slice(jsonlFiles, func(i, j int) bool {
			return jsonlFiles[i].mtime > jsonlFiles[j].mtime
		})

		threeMinutesAgo := now - 3*60*1000
		limit := 5
		if len(jsonlFiles) < limit {
			limit = len(jsonlFiles)
		}

		for i := 0; i < limit; i++ {
			f := jsonlFiles[i]
			// 只扫描 3 分钟内修改过的文件（快速路径）
			if f.mtime < threeMinutesAgo {
				continue
			}
			filePath := filepath.Join(sessionsDir, f.name)
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			// 从后往前扫描最多 20 行
			start := len(lines) - 20
			if start < 0 {
				start = 0
			}
			for j := len(lines) - 1; j >= start; j-- {
				var entry map[string]any
				if json.Unmarshal([]byte(lines[j]), &entry) != nil {
					continue
				}
				if entry["type"] != "message" {
					continue
				}
				msg, ok := entry["message"].(map[string]any)
				if !ok {
					continue
				}
				if msg["role"] != "assistant" {
					continue
				}
				tsStr, _ := entry["timestamp"].(string)
				if tsStr == "" {
					continue
				}
				t, err := time.Parse(time.RFC3339Nano, tsStr)
				if err != nil {
					t, err = time.Parse(time.RFC3339, tsStr)
					if err != nil {
						continue
					}
				}
				ms := t.UnixMilli()
				if ms > lastAssistantMs {
					lastAssistantMs = ms
				}
				if ms > lastActive {
					lastActive = ms
				}
			}
		}
	}

	// 判断状态
	state := dto.AgentStateOffline
	if lastActive > 0 {
		diff := now - lastActive
		if lastAssistantMs > 0 && (now-lastAssistantMs) < 3*60*1000 {
			state = dto.AgentStateWorking
		} else if diff < 10*60*1000 {
			state = dto.AgentStateOnline
		} else if diff < 24*60*60*1000 {
			state = dto.AgentStateIdle
		}
	}

	return dto.AgentStatus{
		AgentID:    agentID,
		State:      state,
		LastActive: lastActive,
	}
}

// GetAgentStatuses 获取所有 Agent 的实时状态
func (s *AgentStatService) GetAgentStatuses() (*dto.GetAgentStatusesResp, error) {
	agentIDs := listAgentIDsFromFilesystem()
	statuses := make([]dto.AgentStatus, 0, len(agentIDs))
	for _, id := range agentIDs {
		statuses = append(statuses, getAgentState(id))
	}
	return &dto.GetAgentStatusesResp{Statuses: statuses}, nil
}

// ========== Token 统计 ==========

// GetAgentTokenStats 获取 Token 消耗统计（最近 N 天）
func (s *AgentStatService) GetAgentTokenStats(req dto.GetAgentTokenStatsReq) (*dto.GetAgentTokenStatsResp, error) {
	days := req.Days
	if days <= 0 {
		days = 7
	}

	var agentIDs []string
	if req.AgentID != "" {
		agentIDs = []string{req.AgentID}
	} else {
		agentIDs = listAgentIDsFromFilesystem()
	}

	resp := &dto.GetAgentTokenStatsResp{
		Stats: make([]dto.AgentTokenStats, 0, len(agentIDs)),
	}

	// 生成日期列表
	dates := make([]string, days)
	for i := days - 1; i >= 0; i-- {
		dates[days-1-i] = time.Now().AddDate(0, 0, -i).Format("2006-01-02")
	}

	cutoffMs := time.Now().AddDate(0, 0, -days).UnixMilli()

	for _, agentID := range agentIDs {
		stats := computeAgentTokenStats(agentID, dates, cutoffMs)
		resp.Stats = append(resp.Stats, stats)
		resp.TotalTokens += stats.TotalTokens
		resp.TotalMessages += stats.TotalMessages
	}

	return resp, nil
}

// computeAgentTokenStats 解析 JSONL 计算单个 agent 的 token 统计
func computeAgentTokenStats(agentID string, dates []string, cutoffMs int64) dto.AgentTokenStats {
	sessionsDir := openclawAgentSessionsDir(agentID)
	dailyTokens := make(map[string]int)
	dailyResponseTimes := make(map[string][]int)
	dailyMessages := make(map[string]int)
	for _, d := range dates {
		dailyTokens[d] = 0
		dailyResponseTimes[d] = nil
		dailyMessages[d] = 0
	}
	total := dto.AgentTokenStats{
		AgentID:   agentID,
		Daily:     make([]dto.DailyStat, 0, len(dates)),
		Platforms: []string{},
	}

	// ① 从 sessions.json 读取平台信息和会话数
	sessionsJSON := filepath.Join(sessionsDir, "sessions.json")
	platformSet := map[string]bool{}
	if data, err := os.ReadFile(sessionsJSON); err == nil {
		var sessions map[string]any
		if json.Unmarshal(data, &sessions) == nil {
			total.SessionCount = len(sessions)
			// 解析平台类型
			for key := range sessions {
				switch {
				case strings.Contains(key, ":feishu:"):
					platformSet["feishu"] = true
				case strings.Contains(key, ":discord:"):
					platformSet["discord"] = true
				case strings.Contains(key, ":telegram:"):
					platformSet["telegram"] = true
				case strings.Contains(key, ":whatsapp:"):
					platformSet["whatsapp"] = true
				case strings.Contains(key, ":cron:"):
					platformSet["cron"] = true
				case strings.HasSuffix(key, ":main"):
					platformSet["main"] = true
				}
			}
		}
	}
	for p := range platformSet {
		total.Platforms = append(total.Platforms, p)
	}
	sort.Strings(total.Platforms)

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		// 目录不存在，返回空数据
		for _, d := range dates {
			total.Daily = append(total.Daily, dto.DailyStat{Date: d})
		}
		return total
	}

	// 全局响应时间列表（用于计算整体平均值）
	var allResponseTimes []int

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") || strings.Contains(e.Name(), ".deleted.") {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().UnixMilli() < cutoffMs {
			continue
		}

		filePath := filepath.Join(sessionsDir, e.Name())
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		// 流式解析 JSONL，O(n)
		type msgRecord struct {
			role       string
			date       string
			ts         string
			stopReason string
			tokens     int
		}
		var msgs []msgRecord

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 512*1024), 512*1024)
		for scanner.Scan() {
			var entry map[string]any
			if json.Unmarshal(scanner.Bytes(), &entry) != nil {
				continue
			}
			if entry["type"] != "message" {
				continue
			}
			msg, ok := entry["message"].(map[string]any)
			if !ok {
				continue
			}
			tsStr, _ := entry["timestamp"].(string)
			if tsStr == "" {
				continue
			}
			date := tsStr[:10] // YYYY-MM-DD
			if _, ok := dailyTokens[date]; !ok {
				continue // 超出统计区间
			}
			role, _ := msg["role"].(string)

			rec := msgRecord{role: role, date: date, ts: tsStr}
			if role == "assistant" {
				if usage, ok := msg["usage"].(map[string]any); ok {
					inp, _ := usage["input"].(float64)
					out, _ := usage["output"].(float64)
					rec.tokens = int(inp + out)
				}
				rec.stopReason, _ = msg["stopReason"].(string)
			}
			msgs = append(msgs, rec)
		}
		f.Close()

		// 计算 token 和响应时间
		var lastUserTs string
		for _, m := range msgs {
			if m.role == "assistant" && m.tokens > 0 {
				dailyTokens[m.date] += m.tokens
				total.TotalTokens += m.tokens
				total.TotalMessages++
				dailyMessages[m.date]++
			}
			if m.role == "user" {
				lastUserTs = m.ts
			} else if m.role == "assistant" && m.stopReason == "stop" && lastUserTs != "" {
				t1, e1 := time.Parse(time.RFC3339Nano, lastUserTs)
				t2, e2 := time.Parse(time.RFC3339Nano, m.ts)
				if e1 != nil {
					t1, e1 = time.Parse(time.RFC3339, lastUserTs)
				}
				if e2 != nil {
					t2, e2 = time.Parse(time.RFC3339, m.ts)
				}
				if e1 == nil && e2 == nil {
					diffMs := int(t2.UnixMilli() - t1.UnixMilli())
					if diffMs > 0 && diffMs < 600000 {
						dailyResponseTimes[m.date] = append(dailyResponseTimes[m.date], diffMs)
						allResponseTimes = append(allResponseTimes, diffMs)
					}
				}
				lastUserTs = ""
			}
		}
	}

	// 计算整体平均响应时间
	if len(allResponseTimes) > 0 {
		sum := 0
		for _, t := range allResponseTimes {
			sum += t
		}
		total.AvgResponseMs = sum / len(allResponseTimes)
	}

	for _, d := range dates {
		avgMs := 0
		if times := dailyResponseTimes[d]; len(times) > 0 {
			sum := 0
			for _, t := range times {
				sum += t
			}
			avgMs = sum / len(times)
		}
		total.Daily = append(total.Daily, dto.DailyStat{
			Date:          d,
			Tokens:        dailyTokens[d],
			AvgResponseMs: avgMs,
			MessageCount:  dailyMessages[d],
		})
	}
	return total
}

// ========== Gateway 健康检测 ==========

// GetGatewayHealth 探测 Gateway 是否健康
func (s *AgentStatService) GetGatewayHealth() (*dto.GatewayHealthResp, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return &dto.GatewayHealthResp{
			OK:        false,
			Status:    "down",
			Error:     "配置文件读取失败: " + err.Error(),
			CheckedAt: time.Now().UnixMilli(),
		}, nil
	}

	port := 18789
	token := ""
	if gw, ok := config["gateway"].(map[string]any); ok {
		if p, ok := gw["port"].(float64); ok {
			port = int(p)
		}
		if auth, ok := gw["auth"].(map[string]any); ok {
			if t, ok := auth["token"].(string); ok {
				token = t
			}
		}
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/api/health", port)
	startedAt := time.Now()

	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	checkedAt := time.Now().UnixMilli()
	responseMs := time.Since(startedAt).Milliseconds()

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "Gateway 未运行"
		} else if strings.Contains(errMsg, "timeout") {
			errMsg = "请求超时"
		}
		return &dto.GatewayHealthResp{
			OK:        false,
			Status:    "down",
			Error:     errMsg,
			CheckedAt: checkedAt,
			ResponseMs: responseMs,
		}, nil
	}
	resp.Body.Close()

	status := "healthy"
	if responseMs > 1500 {
		status = "degraded"
	}
	if resp.StatusCode >= 400 {
		return &dto.GatewayHealthResp{
			OK:        false,
			Status:    "down",
			Error:     fmt.Sprintf("HTTP %d", resp.StatusCode),
			CheckedAt: checkedAt,
			ResponseMs: responseMs,
		}, nil
	}


	return &dto.GatewayHealthResp{
		OK:         true,
		Status:     status,
		CheckedAt:  checkedAt,
		ResponseMs: responseMs,
	}, nil
}

// ========== 平台连通性测试 ==========

// TestPlatformConn 测试平台连通性（HTTP 直接测试 Feishu / Discord / Telegram）
func (s *AgentStatService) TestPlatformConn(req dto.TestPlatformConnReq) (*dto.TestPlatformConnResp, error) {
	config, err := readOpenClawConfig()
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	channels, _ := config["channels"].(map[string]any)
	if channels == nil {
		channels = map[string]any{}
	}

	var results []dto.PlatformTestResult

	// 测试飞书
	if req.Platform == "" || req.Platform == "feishu" {
		if feishu, ok := channels["feishu"].(map[string]any); ok {
			feishuResults := testFeishuPlatform(req.AgentID, feishu, config)
			results = append(results, feishuResults...)
		}
	}

	// 测试 Discord
	if req.Platform == "" || req.Platform == "discord" {
		if discord, ok := channels["discord"].(map[string]any); ok {
			result := testDiscordPlatform(req.AgentID, discord)
			if result != nil {
				results = append(results, *result)
			}
		}
	}

	// 测试 Telegram
	if req.Platform == "" || req.Platform == "telegram" {
		if telegram, ok := channels["telegram"].(map[string]any); ok {
			result := testTelegramPlatform(req.AgentID, telegram)
			if result != nil {
				results = append(results, *result)
			}
		}
	}

	return &dto.TestPlatformConnResp{Results: results}, nil
}

// testFeishuPlatform 测试飞书平台连通性（获取 tenant_access_token 验证）
func testFeishuPlatform(agentID string, feishuConfig map[string]any, fullConfig map[string]any) []dto.PlatformTestResult {
	var results []dto.PlatformTestResult

	domain, _ := feishuConfig["domain"].(string)
	baseURL := "https://open.feishu.cn"
	if domain == "lark" {
		baseURL = "https://open.larksuite.com"
	}

	// 从 feishu.accounts 获取账号配置
	accounts, _ := feishuConfig["accounts"].(map[string]any)
	if accounts == nil {
		// 尝试顶层配置（单账号模式）
		accounts = map[string]any{"main": feishuConfig}
	}

	bindings, _ := fullConfig["bindings"].([]any)

	testedAccounts := map[string]bool{}
	for _, bindAny := range bindings {
		bind, ok := bindAny.(map[string]any)
		if !ok {
			continue
		}
		match, _ := bind["match"].(map[string]any)
		if match == nil {
			continue
		}
		if match["channel"] != "feishu" {
			continue
		}
		targetAgentID, _ := bind["agentId"].(string)
		if agentID != "" && targetAgentID != agentID {
			continue
		}
		accountID, _ := match["accountId"].(string)
		if accountID == "" {
			accountID = targetAgentID
		}
		if testedAccounts[accountID] {
			continue
		}
		acc, ok := accounts[accountID].(map[string]any)
		if !ok {
			continue
		}
		appID, _ := acc["appId"].(string)
		appSecret, _ := acc["appSecret"].(string)
		if appID == "" || appSecret == "" {
			continue
		}
		testedAccounts[accountID] = true
		result := probeFeishuToken(targetAgentID, accountID, appID, appSecret, baseURL)
		results = append(results, result)
	}

	// 如果没有绑定，尝试 main 账号
	if len(testedAccounts) == 0 {
		if mainAcc, ok := accounts["main"].(map[string]any); ok {
			appID, _ := mainAcc["appId"].(string)
			appSecret, _ := mainAcc["appSecret"].(string)
			if appID != "" && appSecret != "" {
				result := probeFeishuToken("main", "main", appID, appSecret, baseURL)
				results = append(results, result)
			}
		}
	}

	return results
}

func probeFeishuToken(agentID, accountID, appID, appSecret, baseURL string) dto.PlatformTestResult {
	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second}

	tokenURL := baseURL + "/open-apis/auth/v3/tenant_access_token/internal"
	body := strings.NewReader(fmt.Sprintf(`{"app_id":"%s","app_secret":"%s"}`, appID, appSecret))
	req, _ := http.NewRequest("POST", tokenURL, body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return dto.PlatformTestResult{
			AgentID: agentID, Platform: "feishu", AccountID: accountID,
			OK: false, Error: err.Error(), Elapsed: elapsed,
		}
	}
	defer resp.Body.Close()

	var data map[string]any
	json.NewDecoder(resp.Body).Decode(&data)
	code, _ := data["code"].(float64)
	if code != 0 || data["tenant_access_token"] == nil {
		msg, _ := data["msg"].(string)
		return dto.PlatformTestResult{
			AgentID: agentID, Platform: "feishu", AccountID: accountID,
			OK: false, Error: "获取 token 失败: " + msg, Elapsed: elapsed,
		}
	}

	return dto.PlatformTestResult{
		AgentID: agentID, Platform: "feishu", AccountID: accountID,
		OK:      true,
		Detail:  fmt.Sprintf("Feishu 账号 %s 认证成功 (%dms)", accountID, elapsed),
		Elapsed: elapsed,
	}
}

func testDiscordPlatform(agentID string, discordConfig map[string]any) *dto.PlatformTestResult {
	botToken, _ := discordConfig["token"].(string)
	if botToken == "" {
		return nil
	}
	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	req.Header.Set("Authorization", "Bot "+botToken)
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	result := &dto.PlatformTestResult{
		AgentID: agentID, Platform: "discord", Elapsed: elapsed,
	}
	if err != nil {
		result.OK = false
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	var data map[string]any
	json.NewDecoder(resp.Body).Decode(&data)
	if !resp.ProtoAtLeast(1, 0) || resp.StatusCode >= 400 {
		msg, _ := data["message"].(string)
		result.OK = false
		result.Error = "Discord API 错误: " + msg
		return result
	}
	username, _ := data["username"].(string)
	result.OK = true
	result.Detail = fmt.Sprintf("Bot @%s 认证成功 (%dms)", username, elapsed)
	return result
}

func testTelegramPlatform(agentID string, telegramConfig map[string]any) *dto.PlatformTestResult {
	botToken, _ := telegramConfig["botToken"].(string)
	if botToken == "" {
		return nil
	}
	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)
	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()
	result := &dto.PlatformTestResult{
		AgentID: agentID, Platform: "telegram", Elapsed: elapsed,
	}
	if err != nil {
		result.OK = false
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	var data map[string]any
	json.NewDecoder(resp.Body).Decode(&data)
	if ok, _ := data["ok"].(bool); !ok {
		desc, _ := data["description"].(string)
		result.OK = false
		result.Error = "Telegram API 错误: " + desc
		return result
	}
	r, _ := data["result"].(map[string]any)
	username, _ := r["username"].(string)
	result.OK = true
	result.Detail = fmt.Sprintf("Bot @%s 认证成功 (%dms)", username, elapsed)
	return result
}

// ========== Agent 会话列表 ==========

// GetAgentSessions 获取指定 Agent 的会话列表
func (s *AgentStatService) GetAgentSessions(req dto.GetAgentSessionsReq) (*dto.GetAgentSessionsResp, error) {
	if req.AgentID == "" {
		return nil, fmt.Errorf("agentId 不能为空")
	}
	sessionsPath := filepath.Join(openclawAgentSessionsDir(req.AgentID), "sessions.json")
	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		return &dto.GetAgentSessionsResp{
			AgentID:  req.AgentID,
			Sessions: []dto.SessionItem{},
		}, nil
	}

	var sessions map[string]any
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("解析 sessions.json 失败: %v", err)
	}

	var list []dto.SessionItem
	for key, val := range sessions {
		item := parseSessionKey(key, val)
		list = append(list, item)
	}

	// 按最近活跃降序
	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt > list[j].UpdatedAt
	})

	return &dto.GetAgentSessionsResp{
		AgentID:  req.AgentID,
		Sessions: list,
	}, nil
}

// parseSessionKey 将 session key 解析为 SessionItem
// 格式：agent:{id}:{channel}:{type}:{target}
func parseSessionKey(key string, val any) dto.SessionItem {
	item := dto.SessionItem{Key: key, Type: "unknown"}

	m, ok := val.(map[string]any)
	if ok {
		if ts, ok := m["updatedAt"].(float64); ok {
			item.UpdatedAt = int64(ts)
		}
		if t, ok := m["totalTokens"].(float64); ok {
			item.TotalTokens = int(t)
		}
		if c, ok := m["contextTokens"].(float64); ok {
			item.ContextTokens = int(c)
		}
		if s, ok := m["systemSent"].(bool); ok {
			item.SystemSent = s
		}
	}

	// 解析 channel 和 type
	channelPatterns := []struct {
		suffix string
		typ    string
		sep    string
	}{
		{":main", "main", ""},
		{":feishu:direct:", "feishu-dm", ":feishu:direct:"},
		{":feishu:group:", "feishu-group", ":feishu:group:"},
		{":discord:direct:", "discord-dm", ":discord:direct:"},
		{":discord:channel:", "discord-channel", ":discord:channel:"},
		{":telegram:direct:", "telegram-dm", ":telegram:direct:"},
		{":telegram:group:", "telegram-group", ":telegram:group:"},
		{":whatsapp:direct:", "whatsapp-dm", ":whatsapp:direct:"},
		{":whatsapp:group:", "whatsapp-group", ":whatsapp:group:"},
		{":cron:", "cron", ":cron:"},
	}

	for _, p := range channelPatterns {
		if p.sep == "" {
			if strings.HasSuffix(key, p.suffix) {
				item.Type = p.typ
				break
			}
		} else if idx := strings.Index(key, p.sep); idx >= 0 {
			item.Type = p.typ
			item.Target = key[idx+len(p.sep):]
			break
		}
	}

	return item
}
