// hijack_orchestrator.go — hijack_launcher 분할 (Orchestrator 파트)
// 외부 의존성: 0 (cdp_client.go + Go stdlib)

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ── 전사 기록 ──
var hlTranscriptDedup sync.Map

// ── 활성 scraper 추적 (중복 goroutine 방지) ──
var hlActiveScrapers sync.Map // wsURL → true

func hlAppendTranscript(entry, projectLabel, brainRoot string) {
	transcriptDir := filepath.Join(brainRoot, "_transcripts")
	os.MkdirAll(transcriptDir, 0750)

	now := time.Now().UTC().Add(9 * time.Hour) // KST
	timeKey := now.Format("2006-01-02_15") + "h"
	proj := projectLabel
	if proj == "" {
		proj = "global"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	proj = re.ReplaceAllString(proj, "_")
	if len(proj) > 30 {
		proj = proj[:30]
	}
	file := filepath.Join(transcriptDir, fmt.Sprintf("%s_%s.txt", proj, timeKey))

	// Dedup — timestamp 제거한 본문 기반 (고도화)
	// entry format: "[HH:MM:SS] ROLE: text" → ROLE+text로 해시
	bodyForHash := entry
	if idx := strings.Index(entry, "] "); idx > 0 && idx < 20 {
		bodyForHash = entry[idx+2:] // timestamp 제거
	}
	h := sha256.Sum256([]byte(bodyForHash))
	hashKey := hex.EncodeToString(h[:12])
	if _, loaded := hlTranscriptDedup.LoadOrStore(hashKey, true); loaded {
		return
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		fmt.Fprintln(f, entry)
		f.Close()
	}

	// ── 텔레그램 전송 (원본 mjs _sendToTelegramInner 충실 포팅) ──
	hlSendToTelegram(entry, proj)

	// ── 세션 rolling buffer (최근 20턴 JSONL) ──
	hlUpdateSessionTranscript(entry, brainRoot)
}

func hlUpdateSessionTranscript(entry, brainRoot string) {
	roleRe := regexp.MustCompile(`(?s)\] (\w+)(?:@[^:]*)?: (.*)`)
	m := roleRe.FindStringSubmatch(entry)
	if len(m) < 3 {
		return
	}
	role := m[1]
	text := m[2]
	if role != "USER" && role != "AI" {
		return
	}

	// 1. JSONL append + rolling
	jsonlFile := filepath.Join(brainRoot, "_agents", "global_inbox", "transcript_latest.jsonl")
	os.MkdirAll(filepath.Dir(jsonlFile), 0750)
	ts := time.Now().UTC().Add(9 * time.Hour).Format(time.RFC3339)
	entryJSON := fmt.Sprintf(`{"ts":"%s","role":"%s","text":"%s"}`,
		ts, role, strings.ReplaceAll(strings.ReplaceAll(text, `\`, `\\`), `"`, `\"`))
	if len([]rune(entryJSON)) > 2200 {
		entryJSON = string([]rune(entryJSON)[:2200]) + `"}`
	}

	f, err := os.OpenFile(jsonlFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		fmt.Fprintln(f, entryJSON)
		f.Close()
	}

	// Rolling: 최근 20줄만 유지
	data, err := os.ReadFile(jsonlFile)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
			os.WriteFile(jsonlFile, []byte(strings.Join(lines, "\n")+"\n"), 0600)
		}

		// 2. hippocampus 뉴런 업데이트
		neuronFile := filepath.Join(brainRoot, "hippocampus", "session_log", "절대_최근대화_전사기록_동기화.neuron")
		os.MkdirAll(filepath.Dir(neuronFile), 0750)
		var sb strings.Builder
		sb.WriteString("# 최근 대화 콘텍스트 주입 (실시간 동기화)\n\n이전 대화 맥락을 파악하고 대답을 연계할 것:\n\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			// 간단 파싱
			tsIdx := strings.Index(l, `"ts":"`)
			roleIdx := strings.Index(l, `"role":"`)
			textIdx := strings.Index(l, `"text":"`)
			if tsIdx >= 0 && roleIdx >= 0 && textIdx >= 0 {
				tsVal := l[tsIdx+6:]
				if i := strings.Index(tsVal, `"`); i > 0 {
					tsVal = tsVal[:i]
				}
				roleVal := l[roleIdx+8:]
				if i := strings.Index(roleVal, `"`); i > 0 {
					roleVal = roleVal[:i]
				}
				textVal := l[textIdx+8:]
				if i := strings.LastIndex(textVal, `"`); i > 0 {
					textVal = textVal[:i]
				}
				timeOnly := tsVal
				if len(tsVal) > 11 {
					timeOnly = tsVal[11:]
					if len(timeOnly) > 8 {
						timeOnly = timeOnly[:8]
					}
				}
				truncText := textVal
				if len([]rune(truncText)) > 400 {
					truncText = string([]rune(truncText)[:400])
				}
				sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", timeOnly, strings.ToUpper(roleVal), truncText))
			}
		}
		os.WriteFile(neuronFile, []byte(sb.String()), 0600)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── CDP DOM 스크래핑 + Network 모니터 ──
// 활성 Network scraper 추적
func hlAutoEvolve(brainRoot string) {
	nfsExe, _ := os.Executable()
	for {
		time.Sleep(30 * time.Minute)
		signalDir := filepath.Join(brainRoot, "hippocampus", "_signals")
		entries, err := os.ReadDir(signalDir)
		if err != nil {
			continue
		}
		count := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				count++
			}
		}
		if count > 0 {
			fmt.Printf("[HL] 🧠 자동통합: %d개 신호 → evolve\n", count)
			cmd := exec.Command(nfsExe, brainRoot, "--evolve")
			cmd.Run()
		}
	}
}

// ── 메인 런처 ──
func runHijackLauncher(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	hlLoadTelegram(nfsRoot)

	fmt.Println("[HL] 🚀 Hijack Launcher (Go native) 시작")

	// 텔레그램 양방향 polling
	go hlTgPoll(brainRoot)

	// CDP 모니터
	go hlStartCDPMonitor(brainRoot)

	// 자동 evolve
	go hlAutoEvolve(brainRoot)

	// keep alive
	select {}
}

// hlScrapeCurrentConversation — CDP로 현재 활성 대화 제목 스크래핑
func hlScrapeCurrentConversation() string {
	targets, err := cdpListTargets(hlCDPPort)
	if err != nil {
		return "(CDP 연결 실패)"
	}
	for _, t := range targets {
		if !strings.Contains(t.URL, "workbench.html") {
			continue
		}
		if t.WebSocketDebuggerURL == "" {
			continue
		}
		client, err := NewCDPClient(t.WebSocketDebuggerURL)
		if err != nil {
			continue
		}
		client.Call("Runtime.enable", map[string]interface{}{})
		time.Sleep(300 * time.Millisecond)

		// 현재 대화 제목 추출 (Antigravity 채팅 패널의 제목)
		scrapeExpr := `(() => {
			// 방법1: 탭 타이틀에서 추출
			const title = document.title || '';
			// 방법2: 활성 채팅 헤더에서 추출
			const chatHeader = document.querySelector('[class*="chat"] [class*="title"]') ||
				document.querySelector('.conversation-header') ||
				document.querySelector('[aria-label*="conversation"]');
			const headerText = chatHeader ? chatHeader.textContent.trim() : '';
			return headerText || title || 'unknown';
		})()`

		result, err := client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    scrapeExpr,
			"returnByValue": true,
		})
		client.Close()

		if err != nil {
			continue
		}

		var evalRes struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)
		if evalRes.Result.Value != "" && evalRes.Result.Value != "unknown" {
			return evalRes.Result.Value
		}
		// fallback: 탭 타이틀 사용
		if t.Title != "" {
			return t.Title
		}
	}
	return "(대화 정보 없음)"
}

// appendDebugLog writes TG debug entries to tg_debug.log for troubleshooting
func appendDebugLog(nfsRoot, msg string) {
	logPath := filepath.Join(nfsRoot, "dist", "neuronfs", "logs", "tg_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
}
