package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// SkillService 技能管理服务
type SkillService struct{}

func NewSkillService() *SkillService {
	return &SkillService{}
}

// ========== 通用 clawhub 命令执行 ==========

// findNpx 查找 npx 的绝对路径（兼容 nvm 环境）
func findNpx() string {
	// 先从 PATH 查找
	if p, err := exec.LookPath("npx"); err == nil {
		return p
	}
	// 扫描常见 nvm 安装位置
	nvmDirs := []string{
		"/usr/local/nvm/versions/node",
		filepath.Join(os.Getenv("HOME"), ".nvm/versions/node"),
	}
	for _, base := range nvmDirs {
		entries, _ := os.ReadDir(base)
		for i := len(entries) - 1; i >= 0; i-- { // 倒序 = 最新版本优先
			p := filepath.Join(base, entries[i].Name(), "bin", "npx")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return "npx" // fallback
}

// runClawHubCmd 在 ~/.openclaw 目录下执行 npx clawhub 命令
// 同时支持 local 和 docker 两种部署模式
func runClawHubCmd(args ...string) ([]byte, error) {
	if getDeployMode() == "local" {
		home, _ := os.UserHomeDir()
		workDir := filepath.Join(home, ".openclaw")
		os.MkdirAll(workDir, 0755)

		npxPath := findNpx()
		cmd := exec.Command(npxPath, append([]string{"-y", "clawhub"}, args...)...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) == 0 {
			return []byte(err.Error()), err
		}
		return out, err
	}
	// Docker: 通过 sh -c 动态解析 ~, 避免 hardcode /root 路径导致权限或路径错误
	var safeArgs []string
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			safeArgs = append(safeArgs, fmt.Sprintf("'%s'", arg))
		} else {
			safeArgs = append(safeArgs, arg)
		}
	}
	cmdStr := fmt.Sprintf("mkdir -p ~/.openclaw && cd ~/.openclaw && npx -y clawhub %s", strings.Join(safeArgs, " "))
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

	force, _ := req["force"].(bool)

	args := []string{"install", slug}
	if force {
		args = append(args, "--force")
	}

	out, err := runClawHubCmd(args...)
	output := strings.TrimSpace(string(out))

	if strings.Contains(output, "already installed") || strings.Contains(output, "Already") {
		return map[string]any{"success": true, "message": "技能已安装"}, nil
	}

	// 检测 VirusTotal suspicious 警告
	if strings.Contains(output, "suspicious") || strings.Contains(output, "flagged as suspicious") {
		warning := extractSuspiciousWarning(output)
		return map[string]any{
			"success":    false,
			"suspicious": true,
			"warning":    warning,
			"slug":       slug,
			"message":    "检测到该技能存在安全风险",
		}, nil
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

// ListInstalledSkills 列出已安装技能
func (s *SkillService) ListInstalledSkills() (map[string]any, error) {
	out, err := runClawHubCmd("list")
	output := strings.TrimSpace(string(out))
	if err != nil && !strings.Contains(output, "No installed") {
		return nil, fmt.Errorf("获取已安装技能失败: %s", output)
	}

	if strings.Contains(output, "No installed") {
		return map[string]any{"skills": []any{}}, nil
	}

	skills := parseListResults(output)
	return map[string]any{"skills": skills}, nil
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

// ========== 内置技能（openclaw skills list 获取列表，clawhub 安装/卸载） ==========

// ListBuiltinSkills 通过 openclaw skills list 获取所有技能及状态
func (s *SkillService) ListBuiltinSkills() (map[string]any, error) {
	out, err := runClawCmd("skills", "list", "--no-color")
	output := stripAnsi(strings.TrimSpace(string(out)))
	if err != nil && output == "" {
		return nil, fmt.Errorf("获取技能列表失败: %s", output)
	}

	skills := parseBuiltinSkillList(output)
	return map[string]any{"skills": skills}, nil
}

// InstallBuiltinSkill 安装内置技能（复用 clawhub install）
func (s *SkillService) InstallBuiltinSkill(req map[string]any) (map[string]any, error) {
	name, _ := req["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("技能名不能为空")
	}

	out, err := runClawHubCmd("install", name, "--force")
	output := strings.TrimSpace(string(out))
	if err != nil && !strings.Contains(output, "Already installed") {
		return nil, fmt.Errorf("安装失败: %s", output)
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 已安装", name)}, nil
}

// UninstallBuiltinSkill 卸载内置技能（复用 clawhub uninstall）
func (s *SkillService) UninstallBuiltinSkill(req map[string]any) (map[string]any, error) {
	name, _ := req["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("技能名不能为空")
	}

	out, err := runClawHubCmd("uninstall", name, "--yes")
	output := strings.TrimSpace(string(out))
	if err != nil {
		return nil, fmt.Errorf("卸载失败: %s", output)
	}
	return map[string]any{"success": true, "message": fmt.Sprintf("技能 %s 已卸载", name)}, nil
}

// ========== 解析函数 ==========

// parseSearchResults 解析搜索结果
// 格式: "slug vVersion  Name  (score)"
func parseSearchResults(output string) []map[string]any {
	var results []map[string]any
	re := regexp.MustCompile(`^(\S+)\s+v([\d.]+)\s+(.+?)(?:\s+\(([^)]+)\))?$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m != nil {
			item := map[string]any{
				"slug":    m[1],
				"version": m[2],
				"name":    strings.TrimSpace(m[3]),
			}
			if m[4] != "" {
				item["score"] = m[4]
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
