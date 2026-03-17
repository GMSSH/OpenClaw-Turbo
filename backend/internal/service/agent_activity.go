package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ---- Types ----

type AgentActivityItem struct {
	AgentID    string `json:"agentId"`
	Name       string `json:"name"`
	State      string `json:"state"` // "working"|"online"|"idle"|"offline"
	LastActive int64  `json:"lastActive"`
}

type GetAgentActivityResult struct {
	Agents []AgentActivityItem `json:"agents"`
}

const (
	actWorkingThresholdMs = 3 * 60 * 1000
	actOnlineThresholdMs  = 10 * 60 * 1000
	actIdleThresholdMs    = 24 * 60 * 60 * 1000
)

// ---- GetAgentActivity ----

func (s *AgentService) GetAgentActivity() (*GetAgentActivityResult, error) {
	// 直接用和 ListAgents 相同的来源，避免重复解析 openclaw.json
	agents, err := readAgentsList()
	if err != nil {
		return &GetAgentActivityResult{Agents: []AgentActivityItem{}}, nil
	}

	result := &GetAgentActivityResult{Agents: []AgentActivityItem{}}

	var agentsBaseDir string
	if getDeployMode() == "local" {
		home, _ := os.UserHomeDir()
		agentsBaseDir = filepath.Join(home, ".openclaw", "agents")
	} else {
		agentsBaseDir = filepath.Join(getDataDir(), "agents")
	}

	now := time.Now().UnixMilli()

	for _, agent := range agents {
		sessionsDir := filepath.Join(agentsBaseDir, agent.ID, "sessions")
		lastActive, lastAssistantTs := actScanSessionActivity(sessionsDir)

		diff := now - lastActive
		var state string
		switch {
		case lastActive == 0 || diff > actIdleThresholdMs:
			state = "offline"
		case lastAssistantTs > 0 && now-lastAssistantTs < actWorkingThresholdMs:
			state = "working"
		case diff < actOnlineThresholdMs:
			state = "online"
		default:
			state = "idle"
		}

		result.Agents = append(result.Agents, AgentActivityItem{
			AgentID:    agent.ID,
			Name:       agent.Name,
			State:      state,
			LastActive: lastActive,
		})
	}

	return result, nil
}

// actScanSessionActivity returns (lastActive ms, lastAssistantMsg ms)
func actScanSessionActivity(sessionsDir string) (int64, int64) {
	var lastActive, lastAssistantTs int64

	sessionsJSON, err := os.ReadFile(filepath.Join(sessionsDir, "sessions.json"))
	if err == nil {
		var sessions map[string]struct {
			UpdatedAt int64 `json:"updatedAt"`
		}
		if json.Unmarshal(sessionsJSON, &sessions) == nil {
			for _, s := range sessions {
				if s.UpdatedAt > lastActive {
					lastActive = s.UpdatedAt
				}
			}
		}
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return lastActive, lastAssistantTs
	}

	now := time.Now().UnixMilli()

	type fileInfo struct {
		path  string
		mtime int64
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		mtime := info.ModTime().UnixMilli()
		if mtime > lastActive {
			lastActive = mtime
		}
		if now-mtime <= 3*60*1000 {
			files = append(files, fileInfo{
				path:  filepath.Join(sessionsDir, e.Name()),
				mtime: mtime,
			})
		}
	}

	sort.Slice(files, func(i, j int) bool { return files[i].mtime > files[j].mtime })
	if len(files) > 5 {
		files = files[:5]
	}

	for _, f := range files {
		raw, err := os.ReadFile(f.path)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
		start := len(lines) - 20
		if start < 0 {
			start = 0
		}
		for i := len(lines) - 1; i >= start; i-- {
			var entry struct {
				Type      string `json:"type"`
				Timestamp string `json:"timestamp"`
				Message   struct {
					Role string `json:"role"`
				} `json:"message"`
			}
			if json.Unmarshal([]byte(lines[i]), &entry) != nil {
				continue
			}
			if entry.Type == "message" && entry.Message.Role == "assistant" && entry.Timestamp != "" {
				ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
				if err == nil {
					ms := ts.UnixMilli()
					if ms > lastAssistantTs {
						lastAssistantTs = ms
					}
					if ms > lastActive {
						lastActive = ms
					}
				}
			}
		}
	}

	return lastActive, lastAssistantTs
}
