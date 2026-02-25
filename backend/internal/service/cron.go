package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CronService 定时任务管理服务
type CronService struct{}

func NewCronService() *CronService {
	return &CronService{}
}

// ========== 公共辅助 ==========

// getCronDir 获取 cron 数据目录
func getCronDir() string {
	if getDeployMode() == "local" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".openclaw", "cron")
	}
	return filepath.Join(getDataDir(), "conf", "cron")
}

// getJobsFilePath 获取 jobs.json 路径
func getJobsFilePath() string {
	return filepath.Join(getCronDir(), "jobs.json")
}

// readJobsFile 读取 jobs.json
func readJobsFile() (map[string]any, error) {
	data, err := os.ReadFile(getJobsFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"version": float64(1), "jobs": []any{}}, nil
		}
		return nil, fmt.Errorf("读取任务文件失败: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析任务文件失败: %v", err)
	}
	return result, nil
}

// writeJobsFile 写入 jobs.json
func writeJobsFile(data map[string]any) error {
	os.MkdirAll(getCronDir(), 0755)
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务数据失败: %v", err)
	}
	return os.WriteFile(getJobsFilePath(), jsonBytes, 0644)
}

// getJobsList 从 jobs.json 提取 jobs 数组
func getJobsList(data map[string]any) []any {
	if jobs, ok := data["jobs"].([]any); ok {
		return jobs
	}
	return []any{}
}

// generateJobID 生成随机 Job ID
func generateJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// reloadGateway 写入 jobs.json 后重启 gateway 使调度器生效
func reloadGateway() {
	if getDeployMode() == "local" {
		exec.Command("systemctl", "restart", "openclaw").Run()
	} else {
		exec.Command("docker", "restart", containerName).Run()
	}
}

// parseEveryToMs 将 "30m" / "2h" / "1d" 格式转换为毫秒
func parseEveryToMs(every string) (float64, error) {
	every = strings.TrimSpace(every)
	if every == "" {
		return 0, fmt.Errorf("间隔不能为空")
	}
	unit := every[len(every)-1:]
	valStr := every[:len(every)-1]
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, fmt.Errorf("无效的间隔值: %s", every)
	}
	switch unit {
	case "s":
		return val * 1000, nil
	case "m":
		return val * 60 * 1000, nil
	case "h":
		return val * 3600 * 1000, nil
	case "d":
		return val * 86400 * 1000, nil
	default:
		return 0, fmt.Errorf("未知时间单位: %s (支持 s/m/h/d)", unit)
	}
}

// ========== 定时任务操作 ==========

// CronStatus 获取调度器状态 — 直接读取 jobs 文件返回摘要
func (s *CronService) CronStatus() (map[string]any, error) {
	jobsData, err := readJobsFile()
	if err != nil {
		return nil, err
	}
	jobs := getJobsList(jobsData)
	enabled := 0
	for _, j := range jobs {
		if jobMap, ok := j.(map[string]any); ok {
			if e, ok := jobMap["enabled"].(bool); ok && e {
				enabled++
			}
		}
	}
	return map[string]any{
		"total":   len(jobs),
		"enabled": enabled,
	}, nil
}

// ListCronJobs 列出所有定时任务
func (s *CronService) ListCronJobs() (map[string]any, error) {
	return readJobsFile()
}

// AddCronJob 新增定时任务（直接写 jobs.json + 重启 gateway）
func (s *CronService) AddCronJob(req map[string]any) (map[string]any, error) {
	name, _ := req["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}

	jobID := generateJobID()
	now := time.Now().UnixMilli()

	// 构建 schedule
	schedule := map[string]any{}
	scheduleKind, _ := req["scheduleKind"].(string)
	switch scheduleKind {
	case "at":
		at, _ := req["at"].(string)
		if at == "" {
			return nil, fmt.Errorf("定时时间不能为空")
		}
		schedule["at"] = at
	case "every":
		every, _ := req["every"].(string)
		if every == "" {
			return nil, fmt.Errorf("间隔时间不能为空")
		}
		ms, err := parseEveryToMs(every)
		if err != nil {
			return nil, err
		}
		schedule["everyMs"] = ms
	case "cron":
		expr, _ := req["cron"].(string)
		if expr == "" {
			return nil, fmt.Errorf("Cron 表达式不能为空")
		}
		schedule["cron"] = expr
		if tz, ok := req["tz"].(string); ok && tz != "" {
			schedule["tz"] = tz
		}
	default:
		return nil, fmt.Errorf("未知调度类型: %s", scheduleKind)
	}

	// 构建 payload
	payload := map[string]any{}
	session, _ := req["session"].(string)
	if session == "" {
		session = "isolated"
	}
	if session == "main" {
		sysEvent, _ := req["systemEvent"].(string)
		if sysEvent == "" {
			sysEvent, _ = req["message"].(string)
		}
		if sysEvent != "" {
			payload["systemEvent"] = sysEvent
		}
	} else {
		message, _ := req["message"].(string)
		if message != "" {
			payload["message"] = message
		}
	}

	// 构建 job 对象
	job := map[string]any{
		"id":        jobID,
		"name":      name,
		"enabled":   true,
		"session":   session,
		"schedule":  schedule,
		"payload":   payload,
		"createdAt": now,
		"updatedAt": now,
	}

	if desc, ok := req["description"].(string); ok && desc != "" {
		job["description"] = desc
	}
	if wake, ok := req["wakeMode"].(string); ok && wake != "" {
		if wake == "heartbeat" {
			wake = "next-heartbeat"
		}
		job["wake"] = wake
	}
	if del, ok := req["deleteAfterRun"].(bool); ok && del {
		job["deleteAfterRun"] = true
	}
	if disabled, ok := req["disabled"].(bool); ok && disabled {
		job["enabled"] = false
	}

	deliveryMode, _ := req["deliveryMode"].(string)
	if deliveryMode != "" && deliveryMode != "none" {
		delivery := map[string]any{"mode": deliveryMode}
		if to, ok := req["deliveryTo"].(string); ok && to != "" {
			delivery["to"] = to
		}
		if ch, ok := req["deliveryChannel"].(string); ok && ch != "" {
			delivery["channel"] = ch
		}
		job["delivery"] = delivery
	}
	if model, ok := req["model"].(string); ok && model != "" {
		job["model"] = model
	}
	if thinking, ok := req["thinking"].(string); ok && thinking != "" {
		job["thinking"] = thinking
	}

	// 读取现有 jobs，追加新 job
	jobsData, err := readJobsFile()
	if err != nil {
		return nil, err
	}
	jobs := getJobsList(jobsData)
	jobs = append(jobs, job)
	jobsData["jobs"] = jobs

	if err := writeJobsFile(jobsData); err != nil {
		return nil, fmt.Errorf("写入任务失败: %v", err)
	}

	// 重启 gateway 让调度器加载新任务
	go reloadGateway()

	return map[string]any{"ok": true, "job": job}, nil
}

// EditCronJob 编辑定时任务
func (s *CronService) EditCronJob(req map[string]any) (map[string]any, error) {
	jobId, _ := req["jobId"].(string)
	if jobId == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	jobsData, err := readJobsFile()
	if err != nil {
		return nil, err
	}
	jobs := getJobsList(jobsData)

	found := false
	for i, j := range jobs {
		jobMap, ok := j.(map[string]any)
		if !ok {
			continue
		}
		if jobMap["id"] == jobId {
			if name, ok := req["name"].(string); ok && name != "" {
				jobMap["name"] = name
			}
			if message, ok := req["message"].(string); ok {
				if payload, ok := jobMap["payload"].(map[string]any); ok {
					payload["message"] = message
				} else {
					jobMap["payload"] = map[string]any{"message": message}
				}
			}
			if sysEvent, ok := req["systemEvent"].(string); ok {
				if payload, ok := jobMap["payload"].(map[string]any); ok {
					payload["systemEvent"] = sysEvent
				}
			}

			schedule, _ := jobMap["schedule"].(map[string]any)
			if schedule == nil {
				schedule = map[string]any{}
			}
			if cronExpr, ok := req["cron"].(string); ok && cronExpr != "" {
				delete(schedule, "everyMs")
				delete(schedule, "at")
				schedule["cron"] = cronExpr
			}
			if tz, ok := req["tz"].(string); ok && tz != "" {
				schedule["tz"] = tz
			}
			if every, ok := req["every"].(string); ok && every != "" {
				delete(schedule, "cron")
				delete(schedule, "at")
				ms, err := parseEveryToMs(every)
				if err == nil {
					schedule["everyMs"] = ms
				}
			}
			if at, ok := req["at"].(string); ok && at != "" {
				delete(schedule, "cron")
				delete(schedule, "everyMs")
				schedule["at"] = at
			}
			jobMap["schedule"] = schedule
			jobMap["updatedAt"] = time.Now().UnixMilli()

			jobs[i] = jobMap
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("任务 %s 不存在", jobId)
	}

	jobsData["jobs"] = jobs
	if err := writeJobsFile(jobsData); err != nil {
		return nil, fmt.Errorf("写入任务失败: %v", err)
	}

	go reloadGateway()

	return map[string]any{"ok": true}, nil
}

// RemoveCronJob 删除定时任务
func (s *CronService) RemoveCronJob(req map[string]any) (map[string]any, error) {
	jobId, _ := req["jobId"].(string)
	if jobId == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	jobsData, err := readJobsFile()
	if err != nil {
		return nil, err
	}
	jobs := getJobsList(jobsData)

	newJobs := []any{}
	found := false
	for _, j := range jobs {
		jobMap, ok := j.(map[string]any)
		if !ok {
			newJobs = append(newJobs, j)
			continue
		}
		if jobMap["id"] == jobId {
			found = true
			continue
		}
		newJobs = append(newJobs, j)
	}

	if !found {
		return nil, fmt.Errorf("任务 %s 不存在", jobId)
	}

	jobsData["jobs"] = newJobs
	if err := writeJobsFile(jobsData); err != nil {
		return nil, fmt.Errorf("写入任务失败: %v", err)
	}

	go reloadGateway()

	return map[string]any{"ok": true}, nil
}

// EnableCronJob 启用定时任务
func (s *CronService) EnableCronJob(req map[string]any) (map[string]any, error) {
	return s.toggleCronJob(req, true)
}

// DisableCronJob 禁用定时任务
func (s *CronService) DisableCronJob(req map[string]any) (map[string]any, error) {
	return s.toggleCronJob(req, false)
}

// toggleCronJob 修改 enabled 字段
func (s *CronService) toggleCronJob(req map[string]any, enabled bool) (map[string]any, error) {
	jobId, _ := req["jobId"].(string)
	if jobId == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	jobsData, err := readJobsFile()
	if err != nil {
		return nil, err
	}
	jobs := getJobsList(jobsData)

	found := false
	for i, j := range jobs {
		jobMap, ok := j.(map[string]any)
		if !ok {
			continue
		}
		if jobMap["id"] == jobId {
			jobMap["enabled"] = enabled
			jobMap["updatedAt"] = time.Now().UnixMilli()
			jobs[i] = jobMap
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("任务 %s 不存在", jobId)
	}

	jobsData["jobs"] = jobs
	if err := writeJobsFile(jobsData); err != nil {
		return nil, fmt.Errorf("写入任务失败: %v", err)
	}

	go reloadGateway()

	return map[string]any{"ok": true}, nil
}

// RunCronJob 手动执行定时任务（需要 gateway 在线）
func (s *CronService) RunCronJob(req map[string]any) (map[string]any, error) {
	jobId, _ := req["jobId"].(string)
	if jobId == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	out, err := runClawCmd("cron", "run", jobId)
	if err != nil {
		return nil, fmt.Errorf("执行任务失败: %s", strings.TrimSpace(string(out)))
	}
	return map[string]any{"ok": true, "output": strings.TrimSpace(string(out))}, nil
}

// GetCronRuns 获取任务运行历史
func (s *CronService) GetCronRuns(req map[string]any) (map[string]any, error) {
	jobId, _ := req["jobId"].(string)
	if jobId == "" {
		return nil, fmt.Errorf("任务 ID 不能为空")
	}

	limit := 20
	if l, ok := req["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	runsFile := filepath.Join(getCronDir(), "runs", jobId+".jsonl")
	data, err := os.ReadFile(runsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"runs": []any{}}, nil
		}
		return nil, fmt.Errorf("读取运行记录失败: %v", err)
	}

	var runs []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err == nil {
			runs = append(runs, record)
		}
	}

	for i, j := 0, len(runs)-1; i < j; i, j = i+1, j-1 {
		runs[i], runs[j] = runs[j], runs[i]
	}
	if limit > 0 && len(runs) > limit {
		runs = runs[:limit]
	}

	result := make([]any, len(runs))
	for i, r := range runs {
		result[i] = r
	}
	return map[string]any{"runs": result}, nil
}
