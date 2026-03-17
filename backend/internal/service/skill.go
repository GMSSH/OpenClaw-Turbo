package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// SkillService 技能管理服务
type SkillService struct{}

func NewSkillService() *SkillService {
	return &SkillService{}
}

// ========== 通用 clawhub 命令执行 ==========

// isClawHubGlobal 检测 clawhub 是否可用（全局安装或 npx 缓存中）
func isClawHubGlobal() bool {
	if getDeployMode() == "local" {
		// 先检测全局安装
		out, err := exec.Command("bash", "-lc", "which clawhub").CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return true
		}
		// 再检测 npx 缓存（npx clawhub -V 不会重新下载如果已缓存）
		err = exec.Command("bash", "-lc", "npx clawhub -V").Run()
		return err == nil
	}
	out, err := exec.Command("docker", "exec", containerName, "sh", "-c", "which clawhub || npx clawhub -V").CombinedOutput()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// IsClawHubInstalled 检测 clawhub 是否可用
func (s *SkillService) IsClawHubInstalled() map[string]any {
	return map[string]any{"installed": isClawHubGlobal()}
}

// InstallClawHub 预安装 clawhub 到 npx 缓存
func (s *SkillService) InstallClawHub() (map[string]any, error) {
	var cmd *exec.Cmd
	if getDeployMode() == "local" {
		cmd = exec.Command("bash", "-lc", "npx -y clawhub -V")
	} else {
		cmd = exec.Command("docker", "exec", containerName, "sh", "-c", "npx -y clawhub -V")
	}
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return nil, fmt.Errorf("安装失败: %s", output)
	}
	return map[string]any{"success": true, "message": "clawhub 安装成功"}, nil
}

// runClawHubCmd 在 ~/.openclaw 目录下执行 clawhub 命令
// 优先使用全局安装的 clawhub（快），否则回退到 npx -y clawhub（慢）
func runClawHubCmd(args ...string) ([]byte, error) {
	// 对参数进行 shell 安全转义
	var safeArgs []string
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			safeArgs = append(safeArgs, fmt.Sprintf("'%s'", arg))
		} else {
			safeArgs = append(safeArgs, arg)
		}
	}

	// 判断 clawhub 是否在 PATH 中（全局安装）
	clawCmd := "npx -y clawhub"
	if getDeployMode() == "local" {
		// 优先检测 which clawhub（PATH 中存在）
		out, err := exec.Command("bash", "-lc", "which clawhub").CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			clawCmd = "clawhub" // 全局安装，直接调用（最快）
		} else {
			clawCmd = "npx --no-install clawhub" // npx 缓存中，使用离线模式避免限速
		}
	} else {
		out, err := exec.Command("docker", "exec", containerName, "sh", "-c", "which clawhub").CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			clawCmd = "clawhub"
		} else {
			clawCmd = "npx --no-install clawhub"
		}
	}

	if getDeployMode() == "local" {
		cmdStr := fmt.Sprintf("mkdir -p ~/.openclaw && cd ~/.openclaw && %s %s", clawCmd, strings.Join(safeArgs, " "))
		cmd := exec.Command("bash", "-lc", cmdStr)
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) == 0 {
			return []byte(err.Error()), err
		}
		return out, err
	}
	// Docker
	cmdStr := fmt.Sprintf("mkdir -p ~/.openclaw && cd ~/.openclaw && %s %s", clawCmd, strings.Join(safeArgs, " "))
	dockerArgs := []string{"exec", containerName, "sh", "-c", cmdStr}
	return exec.Command("docker", dockerArgs...).CombinedOutput()
}

// ========== 市场技能 ==========

// SearchSkills 搜索技能
func (s *SkillService) SearchSkills(req map[string]any) (map[string]any, error) {
	query, _ := req["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	out, err := runClawHubCmd("search", query)
	output := strings.TrimSpace(string(out))
	if err != nil && !strings.Contains(output, "v") {
		return nil, fmt.Errorf("搜索失败: %s", output)
	}

	skills := parseSearchResults(output)
	return map[string]any{"skills": skills}, nil
}

// InspectSkill 查看技能详情
func (s *SkillService) InspectSkill(req map[string]any) (map[string]any, error) {
	slug, _ := req["slug"].(string)
	if slug == "" {
		return nil, fmt.Errorf("技能 slug 不能为空")
	}

	out, err := runClawHubCmd("inspect", slug)
	output := strings.TrimSpace(string(out))
	if err != nil && !strings.Contains(output, "Summary") {
		return nil, fmt.Errorf("查看详情失败: %s", output)
	}

	info := parseInspectResult(output)
	return info, nil
}

// InstallSkill 安装技能（市场 + 内置通用）
func (s *SkillService) InstallSkill(req map[string]any) (map[string]any, error) {
	slug, _ := req["slug"].(string)
	if slug == "" {
		return nil, fmt.Errorf("技能 slug 不能为空")
	}

	args := []string{"install", slug, "--force"}

	out, err := runClawHubCmd(args...)
	output := strings.TrimSpace(string(out))

	if strings.Contains(output, "already installed") || strings.Contains(output, "Already") {
		return map[string]any{"success": true, "message": "技能已安装"}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("安装失败: %s", output)
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 安装成功", slug)}, nil
}

// extractSuspiciousWarning 从输出中提取风险警告信息
func extractSuspiciousWarning(output string) string {
	var warnings []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "npm") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "warning") || strings.Contains(lower, "suspicious") ||
			strings.Contains(lower, "risky") || strings.Contains(lower, "crypto") ||
			strings.Contains(lower, "eval") || strings.Contains(lower, "flagged") ||
			strings.Contains(lower, "external") || strings.Contains(lower, "error") {
			cleaned := strings.TrimLeft(line, "⚠️ ")
			if cleaned != "" {
				warnings = append(warnings, cleaned)
			}
		}
	}
	if len(warnings) == 0 {
		return output
	}
	return strings.Join(warnings, "\n")
}

// UninstallSkill 卸载技能（市场 + 内置通用）
func (s *SkillService) UninstallSkill(req map[string]any) (map[string]any, error) {
	slug, _ := req["slug"].(string)
	if slug == "" {
		return nil, fmt.Errorf("技能 slug 不能为空")
	}

	out, err := runClawHubCmd("uninstall", slug, "--yes")
	output := strings.TrimSpace(string(out))
	if err != nil {
		return nil, fmt.Errorf("卸载失败: %s", output)
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 已卸载", slug)}, nil
}

// getInstalledSkillsDir 获取已安装技能目录。
// 优先读取环境变量 OPENCLAW_SKILLS_DIR，兜底使用 ~/.openclaw/skills。
func getInstalledSkillsDir() string {
	if dir := os.Getenv("OPENCLAW_SKILLS_DIR"); dir != "" {
		return dir
	}
	// Docker 模式：使用宿主机挂载目录 /opt/gmclaw/conf/skills
	if getDeployMode() == "docker" {
		return filepath.Join(getDataDir(), "conf", "skills")
	}
	// local 模式：使用 ~/.openclaw/skills
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "skills")
}

// skillMetaFromDir 读取技能目录下的 _meta.json，提取元数据。
// 若文件不存在或解析失败，则以目录名为 slug，版本返回 "unknown"。
func skillMetaFromDir(dirPath, dirName string) map[string]any {
	item := map[string]any{
		"slug":    dirName,
		"name":    dirName,
		"version": "unknown",
	}

	metaPath := filepath.Join(dirPath, "_meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return item
	}

	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		return item
	}

	if v, ok := meta["name"].(string); ok && v != "" {
		item["name"] = v
	}
	if v, ok := meta["version"].(string); ok && v != "" {
		item["version"] = v
	}
	if v, ok := meta["description"].(string); ok && v != "" {
		item["description"] = v
	}
	if v, ok := meta["author"].(string); ok && v != "" {
		item["author"] = v
	}
	if v, ok := meta["slug"].(string); ok && v != "" {
		item["slug"] = v
	}

	return item
}

// ListInstalledSkills 直接扫描 ~/.openclaw/skills/ 目录列出已安装技能。
// 每个子目录对应一个技能，优先读取 _meta.json 获取元数据。
func (s *SkillService) ListInstalledSkills() (map[string]any, error) {
	skillsDir := getInstalledSkillsDir()

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"skills": []any{}}, nil
		}
		return nil, fmt.Errorf("读取技能目录失败 (%s): %v", skillsDir, err)
	}

	var skills []map[string]any
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 跳过隐藏目录
		if strings.HasPrefix(name, ".") {
			continue
		}
		dirPath := filepath.Join(skillsDir, name)
		skills = append(skills, skillMetaFromDir(dirPath, name))
	}

	if skills == nil {
		skills = []map[string]any{}
	}
	return map[string]any{"skills": skills, "dir": skillsDir}, nil
}

// ExploreSkills 浏览最新技能
func (s *SkillService) ExploreSkills() (map[string]any, error) {
	out, err := runClawHubCmd("explore", "--limit", "20")
	output := strings.TrimSpace(string(out))
	if err != nil && len(output) == 0 {
		return nil, fmt.Errorf("浏览失败: %s", output)
	}

	skills := parseExploreResults(output)
	return map[string]any{"skills": skills}, nil
}

// ========== 内置技能（通过 OpenClaw WS 网关 skills.status / skills.update） ==========

// clawWsRequest 按需连接 OpenClaw WS 网关，完成认证，发送一条请求，返回 payload，然后断开。
// 不持久保持连接，适合低频管理操作。
func clawWsRequest(method string, params any) (map[string]any, error) {
	// 1. 读取 gateway 连接信息
	clawPort, clawToken, err := getClawConnInfo()
	if err != nil {
		return nil, fmt.Errorf("读取 OpenClaw 配置失败: %v", err)
	}
	if clawPort == 0 {
		return nil, fmt.Errorf("OpenClaw gateway 端口未配置")
	}

	// 2. 建立 WebSocket 连接
	clawUrl := fmt.Sprintf("ws://127.0.0.1:%d", clawPort)
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	requestHeader := http.Header{}
	requestHeader.Set("Origin", fmt.Sprintf("http://127.0.0.1:%d", clawPort))

	conn, _, err := dialer.Dial(clawUrl, requestHeader)
	if err != nil {
		return nil, fmt.Errorf("连接 OpenClaw gateway 失败: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(15 * time.Second))

	// 3. 读取 connect.challenge
	_, rawChallenge, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("读取 challenge 失败: %v", err)
	}
	var challengeMsg map[string]any
	if err := json.Unmarshal(rawChallenge, &challengeMsg); err != nil {
		return nil, fmt.Errorf("解析 challenge 失败: %v", err)
	}
	if challengeMsg["type"] != "event" || challengeMsg["event"] != "connect.challenge" {
		return nil, fmt.Errorf("期望 connect.challenge，实际收到: type=%v event=%v", challengeMsg["type"], challengeMsg["event"])
	}

	// 4. 发送 connect 认证请求
	connectReq := map[string]any{
		"type":   "req",
		"id":     genUUID(),
		"method": "connect",
		"params": map[string]any{
			"minProtocol": 3,
			"maxProtocol": 3,
			"client": map[string]any{
				"id":       "openclaw-control-ui",
				"version":  "dev",
				"platform": "linux",
				"mode":     "webchat",
			},
			"role":   "operator",
			"scopes": []string{"operator.admin", "operator.approvals", "operator.pairing"},
			"caps":   []any{},
			"auth": map[string]any{
				"token": clawToken,
			},
			"userAgent": "Mozilla/5.0 (Linux) GMClaw-SkillManager/1.0",
			"locale":    "zh-CN",
		},
	}
	if err := conn.WriteJSON(connectReq); err != nil {
		return nil, fmt.Errorf("发送 connect 请求失败: %v", err)
	}

	// 5. 读取认证响应，期望 hello-ok
	_, rawAuth, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("读取认证响应失败: %v", err)
	}
	var authRes map[string]any
	if err := json.Unmarshal(rawAuth, &authRes); err != nil {
		return nil, fmt.Errorf("解析认证响应失败: %v", err)
	}
	authPayload, _ := authRes["payload"].(map[string]any)
	if authPayload == nil || authPayload["type"] != "hello-ok" {
		return nil, fmt.Errorf("OpenClaw 认证失败，响应: %s", string(rawAuth))
	}
	log.Printf("[SkillWS] ✅ 认证成功，调用 %s", method)

	// 6. 发送业务请求
	reqMsg := map[string]any{
		"type":   "req",
		"id":     genUUID(),
		"method": method,
		"params": params,
	}
	if err := conn.WriteJSON(reqMsg); err != nil {
		return nil, fmt.Errorf("发送 %s 请求失败: %v", method, err)
	}

	// 7. 循环读取，跳过事件消息，直到收到对应的 res 消息
	for {
		_, rawRes, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("读取 %s 响应失败: %v", method, err)
		}
		var res map[string]any
		if err := json.Unmarshal(rawRes, &res); err != nil {
			return nil, fmt.Errorf("解析 %s 响应失败: %v", method, err)
		}
		// 跳过事件推送，等待 res 类型
		if res["type"] != "res" {
			continue
		}
		if ok, _ := res["ok"].(bool); !ok {
			return nil, fmt.Errorf("%s 请求失败: %s", method, string(rawRes))
		}
		payload, _ := res["payload"].(map[string]any)
		if payload == nil {
			payload = map[string]any{}
		}
		log.Printf("[SkillWS] ✅ %s 调用成功", method)
		return payload, nil
	}
}

// ListBuiltinSkills 通过 OpenClaw WS 网关 skills.status 获取内置技能列表
func (s *SkillService) ListBuiltinSkills() (map[string]any, error) {
	payload, err := clawWsRequest("skills.status", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("获取内置技能列表失败: %v", err)
	}

	rawSkills, _ := payload["skills"].([]any)
	var skills []map[string]any
	for _, item := range rawSkills {
		s, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// 只保留内置技能（bundled == true）
		bundled, _ := s["bundled"].(bool)
		if !bundled {
			continue
		}
		disabled, _ := s["disabled"].(bool)
		skills = append(skills, map[string]any{
			"name":        s["name"],
			"description": s["description"],
			"icon":        s["emoji"],
			"enabled":     !disabled,
			"source":      "openclaw-bundled",
		})
	}
	if skills == nil {
		skills = []map[string]any{}
	}
	return map[string]any{"skills": skills}, nil
}

// ToggleBuiltinSkill 通过 OpenClaw WS 网关 skills.update 启用/禁用内置技能
func (s *SkillService) ToggleBuiltinSkill(req map[string]any) (map[string]any, error) {
	skillKey, _ := req["skillKey"].(string)
	if skillKey == "" {
		return nil, fmt.Errorf("skillKey 不能为空")
	}
	enabled, _ := req["enabled"].(bool)

	payload, err := clawWsRequest("skills.update", map[string]any{
		"skillKey": skillKey,
		"enabled":  enabled,
	})
	if err != nil {
		return nil, fmt.Errorf("更新技能状态失败: %v", err)
	}

	status := "已禁用"
	if enabled {
		status = "已启用"
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s %s", skillKey, status), "config": payload["config"]}, nil
}

// InstallBuiltinSkill 安装内置技能（保留兼容，内部使用 ToggleBuiltinSkill）
func (s *SkillService) InstallBuiltinSkill(req map[string]any) (map[string]any, error) {
	name, _ := req["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("技能名不能为空")
	}
	return s.ToggleBuiltinSkill(map[string]any{"skillKey": name, "enabled": true})
}

// UninstallBuiltinSkill 卸载内置技能（保留兼容，内部使用 ToggleBuiltinSkill）
func (s *SkillService) UninstallBuiltinSkill(req map[string]any) (map[string]any, error) {
	name, _ := req["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("技能名不能为空")
	}
	return s.ToggleBuiltinSkill(map[string]any{"skillKey": name, "enabled": false})
}

// ========== 解析函数 ==========

// parseSearchResults 解析搜索结果
// 格式: "slug  Name  (score)" 或 "slug vVersion  Name  (score)"
func parseSearchResults(output string) []map[string]any {
	var results []map[string]any
	re := regexp.MustCompile(`^(\S+)\s+(?:v([\d.]+)\s+)?(.+?)\s+\(([^)]+)\)$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m != nil {
			item := map[string]any{
				"slug":  m[1],
				"name":  strings.TrimSpace(m[3]),
				"score": m[4],
			}
			if m[2] != "" {
				item["version"] = m[2]
			}
			results = append(results, item)
		}
	}
	return results
}

// parseInspectResult 解析详情
func parseInspectResult(output string) map[string]any {
	info := map[string]any{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			switch strings.ToLower(key) {
			case "slug":
				info["slug"] = val
			case "name":
				info["name"] = val
			case "version":
				info["version"] = val
			case "summary", "description":
				info["summary"] = val
			case "owner", "author":
				info["owner"] = val
			case "updated", "date":
				info["updated"] = val
			case "tags", "keywords":
				info["tags"] = val
			}
		}
	}
	return info
}

// parseListResults 解析 clawhub list 输出
// 格式: "slug  version"
func parseListResults(output string) []map[string]any {
	var results []map[string]any
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "npm") || strings.Contains(line, "No installed") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			item := map[string]any{"slug": parts[0]}
			if len(parts) >= 2 {
				item["version"] = strings.TrimPrefix(parts[1], "v")
			}
			results = append(results, item)
		}
	}
	return results
}

// parseExploreResults 解析 explore 输出
func parseExploreResults(output string) []map[string]any {
	var results []map[string]any
	re := regexp.MustCompile(`^(\S+)\s+v([\d.]+)(?:\s+(.+?))?(?:\s+(\d+\w+ ago))?$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m != nil {
			item := map[string]any{
				"slug":    m[1],
				"version": m[2],
			}
			if m[3] != "" {
				item["description"] = strings.TrimSpace(m[3])
			}
			if m[4] != "" {
				item["timeAgo"] = m[4]
			}
			results = append(results, item)
		}
	}
	return results
}

// parseBuiltinSkillList 解析 openclaw skills list 的 Unicode 表格输出
func parseBuiltinSkillList(output string) []map[string]any {
	var results []map[string]any
	var current map[string]any

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "│") {
			continue
		}

		cols := strings.Split(line, "│")
		if len(cols) < 5 {
			continue
		}

		status := strings.TrimSpace(cols[1])
		skillName := strings.TrimSpace(cols[2])
		desc := strings.TrimSpace(cols[3])
		source := strings.TrimSpace(cols[4])

		if status == "Status" || skillName == "Skill" {
			continue
		}

		if status != "" {
			if current != nil {
				results = append(results, current)
			}
			// 分离 emoji 前缀
			cleanName := skillName
			icon := ""
			if idx := strings.Index(skillName, " "); idx >= 0 {
				icon = strings.TrimSpace(skillName[:idx])
				candidate := strings.TrimSpace(skillName[idx:])
				if candidate != "" {
					cleanName = candidate
				}
			}

			ready := strings.Contains(status, "✓") || strings.Contains(status, "ready")
			current = map[string]any{
				"name":        cleanName,
				"icon":        icon,
				"status":      status,
				"enabled":     ready,
				"description": desc,
				"source":      source,
			}
		} else if current != nil {
			if skillName != "" {
				current["name"] = current["name"].(string) + skillName
			}
			if desc != "" {
				current["description"] = current["description"].(string) + " " + desc
			}
		}
	}
	if current != nil {
		results = append(results, current)
	}
	return results
}

// stripAnsi 去除 ANSI 转义序列
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// GetSkillsDir 获取技能目录路径（用于手动安装提示）
func (s *SkillService) GetSkillsDir() map[string]any {
	skillsDir := getInstalledSkillsDir()
	mode := getDeployMode()
	return map[string]any{
		"path": skillsDir,
		"mode": mode,
	}
}

// GetActiveSkillCount 获取开启/安装的能力数总计
func (s *SkillService) GetActiveSkillCount() (map[string]any, error) {
	var count int

	if builtin, err := s.ListBuiltinSkills(); err == nil {
		if skills, ok := builtin["skills"].([]map[string]any); ok {
			for _, skill := range skills {
				if enabled, ok := skill["enabled"].(bool); ok && enabled {
					count++
				}
			}
		}
	}

	if installed, err := s.ListInstalledSkills(); err == nil {
		if skills, ok := installed["skills"].([]map[string]any); ok {
			count += len(skills)
		}
	}

	return map[string]any{"count": count}, nil
}

// ========== 推荐技能（从 scripts/ 目录安装） ==========

// recommendedSkillMeta skill.json 中单个记录的结构
type recommendedSkillMeta struct {
	File   string `json:"file"`
	En     string `json:"en"`
	DescEn string `json:"descEn"`
	DescCn string `json:"descCn"`
}

// getSkillJsonPath 返回 skill.json 路径（位于 scripts/ 目录内）
func getSkillJsonPath() string {
	return filepath.Join(getScriptsDir(), "skill.json")
}

// loadSkillMeta 读取 skill.json 列表
func loadSkillMeta() ([]recommendedSkillMeta, error) {
	data, err := os.ReadFile(getSkillJsonPath())
	if err != nil {
		return nil, fmt.Errorf("读取 skill.json 失败: %v", err)
	}
	var list []recommendedSkillMeta
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("解析 skill.json 失败: %v", err)
	}
	return list, nil
}

// installedSlugSet 返回已安装技能的 slug 集合（目录名即 slug）
func installedSlugSet() map[string]bool {
	skillsDir := getInstalledSkillsDir()
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return map[string]bool{}
	}
	set := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			set[e.Name()] = true
		}
	}
	return set
}

// ListRecommendedSkills 返回 skill.json 内所有推荐技能，并标注安装状态
func (s *SkillService) ListRecommendedSkills() (map[string]any, error) {
	metaList, err := loadSkillMeta()
	if err != nil {
		return nil, err
	}

	installed := installedSlugSet()
	var skills []map[string]any
	for _, m := range metaList {
		// 检查 zip 包是否存在于 scripts 目录
		zipPath := filepath.Join(getScriptsDir(), m.File)
		_, zipErr := os.Stat(zipPath)
		skills = append(skills, map[string]any{
			"slug":        m.En,
			"file":        m.File,
			"descEn":      m.DescEn,
			"descCn":      m.DescCn,
			"installed":   installed[m.En],
			"zipExists":   zipErr == nil,
		})
	}
	if skills == nil {
		skills = []map[string]any{}
	}
	return map[string]any{"skills": skills, "scriptsDir": getScriptsDir()}, nil
}

// InstallRecommendedSkill 将 skill.json 中对应的 zip 解压到 ~/.openclaw/skills/
func (s *SkillService) InstallRecommendedSkill(req map[string]any) (map[string]any, error) {
	slug, _ := req["slug"].(string)
	if slug == "" {
		return nil, fmt.Errorf("slug 不能为空")
	}

	// 找到元数据
	metaList, err := loadSkillMeta()
	if err != nil {
		return nil, err
	}
	var meta *recommendedSkillMeta
	for i := range metaList {
		if metaList[i].En == slug {
			meta = &metaList[i]
			break
		}
	}
	if meta == nil {
		return nil, fmt.Errorf("未找到推荆技能: %s", slug)
	}

	zipPath := filepath.Join(getScriptsDir(), meta.File)
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("安装包不存在: %s", zipPath)
	}

	skillsDir := getInstalledSkillsDir()
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建技能目录失败: %v", err)
	}

	// 确保目标目录干净
	destDir := filepath.Join(skillsDir, slug)
	os.RemoveAll(destDir)

	// 解压 zip 到 skills 目录
	if err := unzipSkill(zipPath, skillsDir, slug); err != nil {
		return nil, fmt.Errorf("解压失败: %v", err)
	}

	log.Printf("[RecommendSkill] ✅ 安装成功: %s -> %s", zipPath, destDir)
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 安装成功", slug)}, nil
}

// UninstallRecommendedSkill 删除 ~/.openclaw/skills/ 中对应 slug 目录
func (s *SkillService) UninstallRecommendedSkill(req map[string]any) (map[string]any, error) {
	slug, _ := req["slug"].(string)
	if slug == "" {
		return nil, fmt.Errorf("slug 不能为空")
	}
	// 防止路径穿越
	if strings.Contains(slug, "..") || strings.Contains(slug, "/") {
		return nil, fmt.Errorf("无效的 slug: %s", slug)
	}

	destDir := filepath.Join(getInstalledSkillsDir(), slug)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("技能未安装: %s", slug)
	}
	if err := os.RemoveAll(destDir); err != nil {
		return nil, fmt.Errorf("卸载失败: %v", err)
	}
	log.Printf("[RecommendSkill] 🗑️ 已卸载: %s", destDir)
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 已卸载", slug)}, nil
}

// unzipSkill 解压 zip 包到 destSkillsDir/<slug>/ 目录。
// zip 包内容结构可能是:
//   A) 直接包含 SKILL.md 等文件（平层）
//   B) 外面有一层目录（品版号-名称/）
// 统一写入 destSkillsDir/slug/
func unzipSkill(zipPath, destSkillsDir, slug string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开 zip 失败: %v", err)
	}
	defer r.Close()

	// 发现根目录：若所有条目均共享同一前缀目录，则剪去它
	var rootPrefix string
	if len(r.File) > 0 {
		firstPath := r.File[0].Name
		parts := strings.SplitN(firstPath, "/", 2)
		if len(parts) >= 1 {
			candidate := parts[0] + "/"
			allMatch := true
			for _, f := range r.File {
				if !strings.HasPrefix(f.Name, candidate) {
					allMatch = false
					break
				}
			}
			if allMatch && len(r.File) > 1 {
				rootPrefix = candidate
			}
		}
	}

	destDir := filepath.Join(destSkillsDir, slug)
	for _, f := range r.File {
		// 剪去公共前缀目录
		relPath := f.Name
		if rootPrefix != "" {
			relPath = strings.TrimPrefix(relPath, rootPrefix)
		}
		if relPath == "" {
			continue
		}

		// 防止路径穿越
		destPath := filepath.Join(destDir, relPath)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.Mode())
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", filepath.Dir(destPath), err)
		}

		// 写入文件
		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("创建文件 %s 失败: %v", destPath, err)
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("打开 zip 条目 %s 失败: %v", f.Name, err)
		}
		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if copyErr != nil {
			return fmt.Errorf("写入文件 %s 失败: %v", destPath, copyErr)
		}
	}
	return nil
}
