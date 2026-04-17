// agent_bridge.go — agent-bridge.mjs Go 포팅
// PROVIDES: runAgentBridge, abInjectToAgent, abCheckOutboxToTelegram
// DEPENDS ON: cdp_client.go (cdpListTargets, NewCDPClient), telegram_bridge.go (sendTelegramSafe)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const abPollMs = 3000

var agentTargets = map[string]string{
	"bot1":     "bot1",
	"entp":     "entp",
	"enfp":     "enfp",
	"NeuronFS": "NeuronFS",
}

func runAgentBridge(brainRoot string) {
	agentsDir := filepath.Join(brainRoot, "_agents")
	logFile := filepath.Join(filepath.Dir(brainRoot), "logs", "bridge.log")

	// NAS brain
	nasBrain := os.Getenv("NEURONFS_NAS_BRAIN")
	var nasAgentsDir string
	if nasBrain != "" {
		nasAgentsDir = filepath.Join(nasBrain, "_agents")
	}

	// inbox 폴더 확보
	for agentID := range agentTargets {
		os.MkdirAll(filepath.Join(agentsDir, agentID, "inbox"), 0750)
		os.MkdirAll(filepath.Join(agentsDir, agentID, "outbox"), 0750)
	}

	abLog(logFile, "=== Agent Bridge v2.0 (Go native) ===")
	abLog(logFile, "Agents: bot1, entp, enfp, NeuronFS")

	processed := make(map[string]bool)
	retryCounts := make(map[string]int)
	outProcessed := make(map[string]bool)

	for {
		abCheckInboxes(agentsDir, nasAgentsDir, logFile, processed, retryCounts)
		abCheckOutboxToTelegram(agentsDir, logFile, outProcessed)
		time.Sleep(time.Duration(abPollMs) * time.Millisecond)
	}
}

func abLog(logFile, msg string) {
	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s", ts, msg)
	fmt.Println(line)
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, line)
			f.Close()
		}
	}
}

func abCheckInboxes(agentsDir, nasAgentsDir, logFile string, processed map[string]bool, retryCounts map[string]int) {
	sources := []string{agentsDir}
	if nasAgentsDir != "" {
		if _, err := os.Stat(nasAgentsDir); err == nil {
			sources = append(sources, nasAgentsDir)
		}
	}

	for agentID := range agentTargets {
		for _, baseDir := range sources {
			inbox := filepath.Join(baseDir, agentID, "inbox")
			entries, err := os.ReadDir(inbox)
			if err != nil {
				continue
			}

			for _, e := range entries {
				name := e.Name()
				if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "_") {
					continue
				}
				if processed[name] {
					continue
				}

				fp := filepath.Join(inbox, name)
				data, err := os.ReadFile(fp)
				if err != nil {
					continue
				}

				content := string(data)
				isNas := baseDir != agentsDir
				prefix := ""
				if isNas {
					prefix = "[NAS] "
				}
				abLog(logFile, fmt.Sprintf("📨 %s%s/inbox/%s", prefix, agentID, name))

				// 헤더 파싱
				fromRe := regexp.MustCompile(`(?m)^# from: (.+)$`)
				priorityRe := regexp.MustCompile(`(?m)^# priority: (.+)$`)
				from := "unknown"
				priority := "normal"
				if m := fromRe.FindStringSubmatch(content); len(m) > 1 {
					from = m[1]
				}
				if m := priorityRe.FindStringSubmatch(content); len(m) > 1 {
					priority = m[1]
				}

				// 헤더 제거
				headerRe := regexp.MustCompile(`(?m)^#.*$`)
				body := strings.TrimSpace(headerRe.ReplaceAllString(content, ""))

				urgentPrefix := ""
				if priority == "urgent" {
					urgentPrefix = "🚨 URGENT: "
				}
				injection := fmt.Sprintf("[%s → %s] %s%s", from, agentID, urgentPrefix, body)

				// 마스터 프롬프트 감지 → 넛지로 변환 또는 생략
				if strings.Contains(body, "NeuronFS 자율 진화 명령") || strings.Contains(body, "마스터 프롬프트") {
					brainRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(fp)))) // file → inbox → agentID → _agents → brain_v4
					if nudge, hasWork := hlBuildContextualPrompt(brainRoot); hasWork {
						injection = nudge
					} else {
						// 시스템 안정 — inbox 처리만 하고 주입 생략
						processed[name] = true
						os.Rename(fp, filepath.Join(inbox, "_"+name))
						abLog(logFile, fmt.Sprintf("✅ %s 마스터 프롬프트 → 시스템 안정, 주입 생략", name))
						continue
					}
				}

				ok := abInjectToAgent(agentID, injection)
				if ok {
					processed[name] = true
					os.Rename(fp, filepath.Join(inbox, "_"+name))
					abLog(logFile, fmt.Sprintf("✅ %s → %s injected%s", name, agentID, func() string {
						if isNas {
							return " (from NAS)"
						}
						return ""
					}()))
				} else {
					// 재시도 카운터
					retryKey := agentID + "/" + name
					retryCounts[retryKey]++
					if retryCounts[retryKey] >= 3 {
						processed[name] = true
						os.Rename(fp, filepath.Join(inbox, "_"+name))
						abLog(logFile, fmt.Sprintf("⚠️ %s → %s 3회 실패 포기", name, agentID))
						delete(retryCounts, retryKey)
					}
				}
			}
		}
	}
}

func abInjectToAgent(agentID, message string) bool {
	keyword, ok := agentTargets[agentID]
	if !ok {
		return false
	}

	targets, err := cdpListTargets(hlCDPPort)
	if err != nil {
		return false
	}

	var target *CDPTarget
	for i, t := range targets {
		if strings.Contains(t.URL, "workbench.html") &&
			strings.HasPrefix(strings.ToLower(t.Title), strings.ToLower(keyword)) &&
			t.Type == "page" {
			target = &targets[i]
			break
		}
	}
	if target == nil || target.WebSocketDebuggerURL == "" {
		return false
	}

	client, err := NewCDPClient(target.WebSocketDebuggerURL)
	if err != nil {
		return false
	}
	defer client.Close()

	client.Call("Runtime.enable", map[string]interface{}{})
	time.Sleep(300 * time.Millisecond)

	// Runtime.evaluate로 DOM 직접 조작 (비활성 탭에서도 동작)
	escaped := strings.ReplaceAll(message, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	escaped = strings.ReplaceAll(escaped, "'", `\'`)

	injectExpr := fmt.Sprintf(`(() => {
		const all = Array.from(document.querySelectorAll("[contenteditable]"));
		const el = all.reverse().find(e => {
			const r = e.getBoundingClientRect();
			return r.height > 0 && r.height < 300 && r.width > 100;
		}) || all[0];
		if(el) {
			el.focus();
			document.execCommand("insertText", false, "%s");
			setTimeout(() => {
				const btn = document.querySelector('button[aria-label="Send message"]') ||
					Array.from(document.querySelectorAll("button")).find(b => b.textContent.includes("Send"));
				if(btn) btn.click();
			}, 50);
			return "Injected";
		}
		return "NoTarget";
	})()`, escaped)

	result, err := client.Call("Runtime.evaluate", map[string]interface{}{
		"expression":    injectExpr,
		"returnByValue": true,
	})
	if err != nil {
		return false
	}

	var evalRes struct {
		Result struct {
			Value string `json:"value"`
		} `json:"result"`
	}
	json.Unmarshal(result, &evalRes)
	return evalRes.Result.Value == "Injected"
}

// ── outbox → Telegram 전송 ──
func abCheckOutboxToTelegram(agentsDir, logFile string, outProcessed map[string]bool) {
	if hlTgToken == "" || hlTgChatID == "" {
		return
	}

	// NeuronFS 에이전트의 outbox만 감시
	outbox := filepath.Join(agentsDir, "NeuronFS", "outbox")
	entries, err := os.ReadDir(outbox)
	if err != nil {
		return
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "_") {
			continue
		}
		if outProcessed[name] {
			continue
		}
		outProcessed[name] = true

		fp := filepath.Join(outbox, name)
		data, err := os.ReadFile(fp)
		if err != nil {
			delete(outProcessed, name)
			continue
		}

		content := string(data)
		// 헤더 제거
		headerRe := regexp.MustCompile(`(?m)^#.*$`)
		body := strings.TrimSpace(headerRe.ReplaceAllString(content, ""))
		if body == "" {
			continue
		}

		// sendTelegramSafe로 분할 전송 (글자 제한 자동 처리)
		sendTelegramSafe(hlTgToken, hlTgChatID, fmt.Sprintf("🤖 [NeuronFS]\n%s", body))
		os.Rename(fp, filepath.Join(outbox, "_"+name))
		abLog(logFile, fmt.Sprintf("📤 NeuronFS/outbox/%s → Telegram", name))
	}
}
