// hijack_launcher.go — hijack-launcher.mjs Go 포팅
// 통합 브릿지: TG polling + CDP 캡처 + 전사 + Groq 뉴런 추출
// 외부 의존성: 0 (cdp_client.go + Go stdlib)
package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const hlCDPPort = 9000

// ── 텔레그램 양방향 브릿지 ──
var hlTgToken, hlTgChatID string
var hlTgOffset int
var hlTgMountedRoom = "NeuronFS"
var hlTgInboundHash sync.Map

var hlTgNfsRoot string

// ── TG 발신: 연속 메시지 병합 (원본 mjs 포팅) ──
var hlTgStartTime = time.Now()
var hlTgLastMsgID int
var hlTgLastRole string
var hlTgLastRaw string
var hlTgLastLabel string
var hlTgLastTs time.Time
var hlTgSentHashes sync.Map
var hlTgEditMu sync.Mutex
var hlTgEditPending bool
var hlTgEditTimer *time.Timer

var hlTgSkipRoles = map[string]bool{"TOOL": true, "ATTACH": true, "HIJACK_START": true, "AI_RESP": true}

func hlLoadTelegram(nfsRoot string) {
	hlTgNfsRoot = nfsRoot
	dir := filepath.Join(nfsRoot, "telegram-bridge")
	if d, err := os.ReadFile(filepath.Join(dir, ".token")); err == nil {
		hlTgToken = strings.TrimSpace(string(d))
	}
	if d, err := os.ReadFile(filepath.Join(dir, ".chat_id")); err == nil {
		hlTgChatID = strings.TrimSpace(string(d))
	}
	if d, err := os.ReadFile(filepath.Join(dir, ".mount")); err == nil {
		hlTgMountedRoom = strings.TrimSpace(string(d))
	}
	// offset 복원 (없으면 -1 = 새 메시지만)
	if d, err := os.ReadFile(filepath.Join(dir, ".offsets")); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(d)), "%d", &hlTgOffset)
	} else {
		hlTgOffset = -1 // 첫 실행: 기존 메시지 스킵
	}
}

func hlTgSend(chatID, text string) {
	sendTelegramSafe(hlTgToken, chatID, text)
}

// hlTgAPI — Telegram Bot API 호출 (JSON 응답 반환)
func hlTgAPI(method string, payload map[string]string) (map[string]interface{}, error) {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/%s", hlTgToken, method),
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result, nil
}

// hlTgSafeSend — HTML 파싱 에러 시 plaintext 폴백 (원본 _tgSafeSend 포팅)
func hlTgSafeSend(method string, payload map[string]string) (map[string]interface{}, error) {
	resp, err := hlTgAPI(method, payload)
	if err != nil {
		return resp, err
	}
	if resp != nil {
		if ok, _ := resp["ok"].(bool); !ok {
			if desc, _ := resp["description"].(string); strings.Contains(strings.ToLower(desc), "parse") {
				// HTML 파싱 에러 → parse_mode 제거 후 재시도
				fallback := make(map[string]string)
				for k, v := range payload {
					if k != "parse_mode" {
						fallback[k] = v
					}
				}
				return hlTgAPI(method, fallback)
			}
		}
	}
	return resp, nil
}

// hlFormatTgMsg — 원본 mjs _formatTgMsg 포팅 (HTML 포맷)
func hlFormatTgMsg(role, label, rawText string) string {
	// HTML escape
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	truncated := rawText

	// Markdown → HTML 변환
	escaped := esc.Replace(truncated)
	// 코드블록
	codeRe := regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+)?\n(.*?)```")
	escaped = codeRe.ReplaceAllString(escaped, "<pre><code>$1</code></pre>")
	codeRe2 := regexp.MustCompile("(?s)```(.*?)```")
	escaped = codeRe2.ReplaceAllString(escaped, "<pre><code>$1</code></pre>")
	// 인라인 코드
	inlineRe := regexp.MustCompile("`([^`\n]+)`")
	escaped = inlineRe.ReplaceAllString(escaped, "<code>$1</code>")
	// 볼드
	boldRe := regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	escaped = boldRe.ReplaceAllString(escaped, "<b>$1</b>")

	switch role {
	case "USER":
		return fmt.Sprintf("👤 %s%s", label, escaped)
	case "THINK":
		return fmt.Sprintf("🧠 %s\n<pre><code class=\"language-text\">%s</code></pre>", label, esc.Replace(truncated))
	case "CMD":
		return fmt.Sprintf("⚡ %s\n<pre><code class=\"language-powershell\">%s</code></pre>", label, esc.Replace(truncated))
	case "NEURON":
		return fmt.Sprintf("🧬 %s%s", label, escaped)
	default: // AI
		return fmt.Sprintf("💬 %s%s", label, escaped)
	}
}

// hlSendToTelegram — 원본 mjs _sendToTelegramInner 충실 포팅
func hlSendToTelegram(entry, proj string) {
	if hlTgToken == "" || hlTgChatID == "" {
		appendDebugLog(hlTgNfsRoot, fmt.Sprintf("NO TOKEN/CHATID: token=%d chatid=%d", len(hlTgToken), len(hlTgChatID)))
		return
	}
	// 워밍업 60초 — 재시작 시 DOM에 남아있는 기존 메시지를 dedup에 사전 등록
	if time.Since(hlTgStartTime) < 60*time.Second {
		// 전송은 안 하지만 해시는 등록 → 워밍업 후에도 중복으로 인식
		// SHA256 전체 해시 사용 (text[:150]은 다른 메시지와 충돌 가능)
		h := sha256.Sum256([]byte(entry))
		hlTgSentHashes.Store(hex.EncodeToString(h[:12]), true)
		return
	}

	// role/text 파싱
	roleRe := regexp.MustCompile(`(?s)\] (\w+)(?:@[^:]*)?: (.*)`)
	m := roleRe.FindStringSubmatch(entry)
	if len(m) < 3 {
		appendDebugLog(hlTgNfsRoot, fmt.Sprintf("regex fail: entry=%s", entry[:min(len(entry), 60)]))
		return
	}
	role := m[1]
	text := m[2]

	if len(text) < 5 {
		return
	}
	if strings.Contains(text, "[telegram") {
		return
	}
	if strings.Contains(text, "신호 기록됨") || strings.Contains(text, "교정 반영") {
		return
	}

	// NEURON 감지
	neuronRe := regexp.MustCompile(`(cortex|brainstem|limbic|hippocampus|sensors|ego|prefrontal)/\S*/`)
	if neuronRe.MatchString(text) {
		role = "NEURON"
	}

	// 스킵 role
	if hlTgSkipRoles[role] {
		appendDebugLog(hlTgNfsRoot, fmt.Sprintf("skip role=%s", role))
		return
	}

	// 중복 해시 — timestamp 제거한 본문 기반 (고도화)
	tgBodyForHash := entry
	if idx := strings.Index(entry, "] "); idx > 0 && idx < 20 {
		tgBodyForHash = entry[idx+2:]
	}
	dedupH := sha256.Sum256([]byte(tgBodyForHash))
	dedupKey := hex.EncodeToString(dedupH[:12])
	if _, loaded := hlTgSentHashes.LoadOrStore(dedupKey, true); loaded {
		appendDebugLog(hlTgNfsRoot, fmt.Sprintf("dedup hit: role=%s len=%d", role, len(text)))
		return
	}
	appendDebugLog(hlTgNfsRoot, fmt.Sprintf("PASS: role=%s len=%d → sending", role, len(text)))

	label := ""
	if proj != "" && proj != "global" {
		label = fmt.Sprintf("[%s] ", proj)
	}
	now := time.Now()

	hlTgEditMu.Lock()
	defer hlTgEditMu.Unlock()

	// 연속 같은 role → editMessageText로 병합 (30초 이내, 3900자 미만)
	if hlTgLastMsgID != 0 && hlTgLastRole == role && hlTgLastLabel == label &&
		now.Sub(hlTgLastTs) < 30*time.Second &&
		len([]rune(hlTgLastRaw))+len([]rune(text)) < 3900 {
		hlTgLastRaw += "\n" + text
		hlTgLastTs = now
		hlTgEditPending = true

		// 2초 throttle (API rate limit 회피)
		if hlTgEditTimer == nil {
			hlTgEditTimer = time.AfterFunc(2*time.Second, func() {
				hlTgEditMu.Lock()
				defer hlTgEditMu.Unlock()
				hlTgEditTimer = nil
				if !hlTgEditPending {
					return
				}
				hlTgEditPending = false
				merged := hlFormatTgMsg(role, label, hlTgLastRaw)
				hlTgSafeSend("editMessageText", map[string]string{
					"chat_id":    hlTgChatID,
					"message_id": fmt.Sprintf("%d", hlTgLastMsgID),
					"text":       merged,
					"parse_mode": "HTML",
				})
			})
		}
		return
	}

	// 새 메시지 — 4000자 초과 시 sendTelegramSafe로 분할 전송
	msg := hlFormatTgMsg(role, label, text)
	if len([]rune(msg)) > 4000 {
		// 분할 전송 (plaintext — HTML 태그 분할 시 깨짐 방지)
		plain := fmt.Sprintf("%s %s%s", func() string {
			switch role {
			case "USER": return "👤"
			case "THINK": return "🧠"
			case "CMD": return "⚡"
			case "NEURON": return "🧬"
			default: return "💬"
			}
		}(), label, text)
		sendTelegramSafe(hlTgToken, hlTgChatID, plain)
		hlTgLastMsgID = 0 // 분할 전송 후 edit 불가
		hlTgLastRole = role
		hlTgLastRaw = text
		hlTgLastLabel = label
		hlTgLastTs = now
		return
	}
	resp, err := hlTgSafeSend("sendMessage", map[string]string{
		"chat_id":    hlTgChatID,
		"text":       msg,
		"parse_mode": "HTML",
	})
	if err == nil && resp != nil {
		if ok, _ := resp["ok"].(bool); ok {
			if result, _ := resp["result"].(map[string]interface{}); result != nil {
				if msgID, _ := result["message_id"].(float64); msgID > 0 {
					hlTgLastMsgID = int(msgID)
					hlTgLastRole = role
					hlTgLastRaw = text
					hlTgLastLabel = label
					hlTgLastTs = now
				}
			}
		}
	}
}

func hlTgPoll(brainRoot string) {
	if hlTgToken == "" {
		return
	}
	agentsDir := filepath.Join(brainRoot, "_agents")

	for {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=5&allowed_updates=[\"message\"]", hlTgToken, hlTgOffset)
		client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{}}}
		resp, err := client.Get(url)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result struct {
			OK     bool `json:"ok"`
			Result []struct {
				UpdateID int `json:"update_id"`
				Message  *struct {
					MessageID int    `json:"message_id"`
					Text      string `json:"text"`
					Chat      struct {
						ID int64 `json:"id"`
					} `json:"chat"`
				} `json:"message"`
			} `json:"result"`
		}
		if json.Unmarshal(data, &result) != nil || !result.OK {
			time.Sleep(2 * time.Second)
			continue
		}

		for _, u := range result.Result {
			hlTgOffset = u.UpdateID + 1
			// offset 즉시 저장 (재시작 시 복원)
			if hlTgNfsRoot != "" {
				os.WriteFile(filepath.Join(hlTgNfsRoot, "telegram-bridge", ".offsets"), []byte(fmt.Sprintf("%d", hlTgOffset)), 0600)
			}
			msg := u.Message
			if msg == nil || msg.Text == "" {
				continue
			}
			chatID := fmt.Sprintf("%d", msg.Chat.ID)

			// chat_id 자동 갱신
			if hlTgChatID != chatID {
				hlTgChatID = chatID
				nfsRoot := filepath.Dir(brainRoot)
				os.WriteFile(filepath.Join(nfsRoot, "telegram-bridge", ".chat_id"), []byte(chatID), 0600)
			}

			// 중복 방지
			hash := fmt.Sprintf("%d", msg.MessageID)
			if _, loaded := hlTgInboundHash.LoadOrStore(hash, true); loaded {
				continue
			}

			text := msg.Text

			// 명령어
			if text == "/start" || text == "/status" || text == "/system" {
				autoState := "✅ ON"
				if fileExists(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".auto_evolve_disabled")) {
					autoState = "❌ OFF"
				}
				hlTgSend(chatID, fmt.Sprintf("🧠 NeuronFS Bridge (Go native)\n현재 방: 📌 %s\n자율진화 봇: %s", hlTgMountedRoom, autoState))
				continue
			}
			if text == "/auto off" {
				os.WriteFile(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".auto_evolve_disabled"), []byte("1"), 0600)
				hlTgSend(chatID, "⏸️ 자율진화(Idle Loop)를 껐습니다. 사용자의 명시적 명령에만 반응합니다.")
				continue
			}
			if text == "/auto on" {
				os.Remove(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".auto_evolve_disabled"))
				hlTgSend(chatID, "▶️ 자율진화(Idle Loop)를 켰습니다. (5분 무활동 시 전사분석 + 자율 트리거 실행)")
				continue
			}
			if strings.HasPrefix(text, "/mount") {
				room := strings.TrimSpace(strings.TrimPrefix(text, "/mount"))
				if room != "" {
					hlTgMountedRoom = room
					nfsRoot := filepath.Dir(brainRoot)
					os.WriteFile(filepath.Join(nfsRoot, "telegram-bridge", ".mount"), []byte(room), 0600)
					hlTgSend(chatID, fmt.Sprintf("✅ 방 전환: 📌 %s", room))
				} else {
					hlTgSend(chatID, fmt.Sprintf("현재 방: 📌 %s", hlTgMountedRoom))
				}
				continue
			}

			// ── 159487: 재시작 코드 ──
			if strings.TrimSpace(text) == "159487" {
				nfsRoot := filepath.Dir(brainRoot)
				hlTgSend(chatID, "🔄 재시작 요청 수신. 현재 대화 스크래핑 중...")

				// CDP로 현재 활성 대화 제목 스크래핑
				convTitle := hlScrapeCurrentConversation()
				ctx := map[string]string{
					"ts":    time.Now().Format(time.RFC3339),
					"title": convTitle,
					"room":  hlTgMountedRoom,
				}
				ctxJSON, _ := json.Marshal(ctx)
				os.WriteFile(filepath.Join(nfsRoot, ".restart_context"), ctxJSON, 0600)

				// _reboot_request 생성 (start.bat가 감지)
				os.WriteFile(filepath.Join(nfsRoot, "_reboot_request"), []byte(time.Now().Format(time.RFC3339)), 0600)

				hlTgSend(chatID, fmt.Sprintf("🔄 재시작 중...\n📌 대화: %s\n⏳ 3초 후 자동 복귀", convTitle))

				// offset 저장 (재시작 후 중복 방지)
				if hlTgNfsRoot != "" {
					os.WriteFile(filepath.Join(hlTgNfsRoot, "telegram-bridge", ".offsets"), []byte(fmt.Sprintf("%d", hlTgOffset)), 0600)
				}

				// 1초 대기 (텔레그램 메시지 전송 보장)
				time.Sleep(1 * time.Second)
				os.Exit(0)
			}

			// 일반 메시지 → inbox
			targetRoom := hlTgMountedRoom
			payload := text
			mentionRe := regexp.MustCompile(`(?s)^@([a-zA-Z0-9_\-\s]+)\s+(.*)`)
			if m := mentionRe.FindStringSubmatch(text); len(m) > 2 {
				mention := strings.ToLower(strings.TrimSpace(m[1]))
				entries, _ := os.ReadDir(agentsDir)
				for _, e := range entries {
					if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), mention) {
						targetRoom = e.Name()
						payload = m[2]
						break
					}
				}
			}

			inboxDir := filepath.Join(agentsDir, targetRoom, "inbox")
			os.MkdirAll(inboxDir, 0750)
			fname := fmt.Sprintf("tg_%d.md", time.Now().UnixMilli())
			content := fmt.Sprintf("# from: telegram\n# priority: normal\n\n%s", payload)
			os.WriteFile(filepath.Join(inboxDir, fname), []byte(content), 0600)
			hlTgSend(chatID, fmt.Sprintf("✅ [%s] 전달됨", targetRoom))

			// CDP 인젝션
			go hlCDPInject(targetRoom, payload)
		}

		time.Sleep(1 * time.Second)
	}
}

func hlCDPInject(targetRoom, payload string) {
	targets, err := cdpListTargets(hlCDPPort)
	if err != nil {
		return
	}
	for _, t := range targets {
		if !strings.Contains(t.URL, "workbench.html") {
			continue
		}
		if !strings.Contains(strings.ToLower(t.Title), strings.ToLower(targetRoom)) {
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

		escaped := strings.ReplaceAll(payload, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		injectCode := fmt.Sprintf(`(() => { const all = Array.from(document.querySelectorAll("[contenteditable]")); const el = all.reverse().find(e => { const r = e.getBoundingClientRect(); return r.height > 0 && r.height < 300 && r.width > 100; }) || all[0]; if(el) { el.focus(); document.execCommand("insertText", false, "[telegram → NeuronFS] %s"); return "Injected"; } return "NoTarget"; })()`, escaped)
		client.Call("Runtime.evaluate", map[string]interface{}{"expression": injectCode, "returnByValue": true})
		time.Sleep(500 * time.Millisecond)
		// Enter 키 dispatch → 자동 submit
		enterScript := `(() => { const el = document.activeElement; if(el) { el.dispatchEvent(new KeyboardEvent("keydown", {key:"Enter",code:"Enter",keyCode:13,which:13,bubbles:true})); } return "Enter"; })()`
		client.Call("Runtime.evaluate", map[string]interface{}{"expression": enterScript, "returnByValue": true})
		time.Sleep(100 * time.Millisecond)
		client.Close()
		break
	}
}

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
var hlActiveNetScrapers sync.Map // wsURL → true

func hlStartCDPMonitor(brainRoot string) {
	for {
		targets, err := cdpListTargets(hlCDPPort)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		for _, t := range targets {
			if t.WebSocketDebuggerURL == "" {
				continue
			}
			if strings.Contains(t.URL, "workbench.html") {
				// DOM scraper (AI/THINK)
				if _, already := hlActiveScrapers.LoadOrStore(t.WebSocketDebuggerURL, true); !already {
					go hlAttachDOMScraper(t.WebSocketDebuggerURL, t.Title, brainRoot)
				}
				// Network monitor (USER/CMD)
				netKey := "net:" + t.WebSocketDebuggerURL
				if _, already := hlActiveNetScrapers.LoadOrStore(netKey, true); !already {
					go hlAttachNetworkMonitor(t.WebSocketDebuggerURL, t.Title, brainRoot)
				}
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func hlAttachDOMScraper(wsURL, title, brainRoot string) {
	// ★ 종료 시 활성 목록에서 제거 — 재연결 허용
	defer hlActiveScrapers.Delete(wsURL)

	proj := title
	if idx := strings.Index(proj, " - "); idx > 0 {
		proj = proj[:idx]
	}

	client, err := NewCDPClient(wsURL)
	if err != nil {
		return
	}
	defer client.Close()

	client.Call("Runtime.enable", map[string]interface{}{})
	time.Sleep(500 * time.Millisecond)

	scrapeScript := `(() => {
		const msgs = [];
		document.querySelectorAll('div').forEach(el => {
			const cls = (el.className||'').toString();
			if (cls.includes('leading-relaxed') && cls.includes('select-text')) {
				const op = parseFloat(getComputedStyle(el).opacity);
				const text = (el.innerText||'').trim();
				if (text && text.length > 5) {
					if (!el.dataset.nid) el.dataset.nid = Math.random().toString(36).substring(2, 9);
					const role = op < 0.9 ? 'THINK' : 'AI';
					msgs.push({role, text: text.slice(0, 10000), id: el.dataset.nid});
				}
			}
		});
		return msgs.slice(-10);
	})()`

	type scrapedMsg struct {
		Role string `json:"role"`
		Text string `json:"text"`
		ID   string `json:"id"`
	}

	tracked := make(map[string]struct{ text string; ts time.Time; logged string })

	for i := 0; i < 600; i++ { // ~30분
		result, err := client.Call("Runtime.evaluate", map[string]interface{}{"expression": scrapeScript, "returnByValue": true})
		if err != nil {
			return
		}

		var evalRes struct {
			Result struct {
				Value []scrapedMsg `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)

		now := time.Now()
		for _, m := range evalRes.Result.Value {
			if m.Text == "" {
				continue
			}
			t, exists := tracked[m.ID]
			if !exists {
				tracked[m.ID] = struct{ text string; ts time.Time; logged string }{m.Text, now, ""}
			} else if t.text != m.Text {
				tracked[m.ID] = struct{ text string; ts time.Time; logged string }{m.Text, now, t.logged}
			}
		}

		// 안정화 판단 (4초 무변경 → 확정 기록)
		for id, t := range tracked {
			if t.text != t.logged && now.Sub(t.ts) > 4*time.Second {
				// role 확인
				role := "AI"
				for _, m := range evalRes.Result.Value {
					if m.ID == id {
						role = m.Role
						break
					}
				}
				kst := now.UTC().Add(9 * time.Hour).Format("15:04:05")
				hlAppendTranscript(fmt.Sprintf("[%s] %s: %s", kst, role, t.text), proj, brainRoot)
				t.logged = t.text
				tracked[id] = t
			}
			// 10분 eviction (3분→10분 확장 — 재캡처 중복 방지)
			if now.Sub(t.ts) > 10*time.Minute {
				delete(tracked, id)
			}
		}

		time.Sleep(3 * time.Second)
	}
}

// ── CDP Network 모니터 (USER 입력 + CMD 캡처) ──
// 원본 mjs attachCDPNetwork 포팅: Network.requestWillBeSent → LanguageServerService postData 파싱
func hlAttachNetworkMonitor(wsURL, title, brainRoot string) {
	netKey := "net:" + wsURL
	defer hlActiveNetScrapers.Delete(netKey)

	proj := title
	if idx := strings.Index(proj, " - "); idx > 0 {
		proj = proj[:idx]
	}

	client, err := NewCDPClient(wsURL)
	if err != nil {
		return
	}
	defer client.Close()

	// Network.enable
	client.Call("Network.enable", map[string]interface{}{"maxTotalBufferSize": 50000000})

	fmt.Fprintf(os.Stderr, "[HL] 🌐 [NET-%s] Network monitor started\n", proj)

	// 이벤트 수신 루프 — CDPClient의 event channel 사용
	// CDPClient는 현재 request-response만 지원하므로, 폴링 방식으로 구현
	// → Network.requestWillBeSent를 직접 수신할 수 없으므로,
	// 대안: DOM에서 사용자 입력을 직접 스크래핑

	// ★ DOM 기반 USER 입력 스크래핑 — Network 이벤트 대신 채팅 메시지 DOM 캡처
	userScrapeScript := `(() => {
		const msgs = [];
		// Antigravity 채팅 사용자 메시지 DOM
		document.querySelectorAll('div').forEach(el => {
			const cls = (el.className||'').toString();
			// 사용자 메시지: whitespace-pre-wrap + text-sm 조합 (PD 입력)
			if (cls.includes('whitespace-pre-wrap') && cls.includes('text-sm') && !cls.includes('font-mono')) {
				const parent = el.parentElement;
				if (parent && (parent.className||'').toString().includes('max-h-')) {
					const text = (el.innerText||'').trim();
					if (text && text.length > 2 && text.length < 5000) {
						if (!el.dataset.uid) el.dataset.uid = Math.random().toString(36).substring(2, 9);
						msgs.push({role:'USER', text: text.slice(0, 3000), id: el.dataset.uid});
					}
				}
			}
		});
		return msgs.slice(-5);
	})()`

	type userMsg struct {
		Role string `json:"role"`
		Text string `json:"text"`
		ID   string `json:"id"`
	}

	tracked := make(map[string]struct{ text string; ts time.Time; logged string })

	for i := 0; i < 600; i++ { // ~30분
		result, err := client.Call("Runtime.evaluate", map[string]interface{}{"expression": userScrapeScript, "returnByValue": true})
		if err != nil {
			return
		}

		var evalRes struct {
			Result struct {
				Value []userMsg `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)

		now := time.Now()
		for _, m := range evalRes.Result.Value {
			if m.Text == "" {
				continue
			}
			t, exists := tracked[m.ID]
			if !exists {
				tracked[m.ID] = struct{ text string; ts time.Time; logged string }{m.Text, now, ""}
			} else if t.text != m.Text {
				tracked[m.ID] = struct{ text string; ts time.Time; logged string }{m.Text, now, t.logged}
			}
		}

		// 안정화 판단 (2초 — 사용자 입력은 즉시 확정되므로 AI보다 짧게)
		for id, t := range tracked {
			if t.text != t.logged && now.Sub(t.ts) > 2*time.Second {
				if len(t.text) > len(t.logged) {
					kst := now.UTC().Add(9 * time.Hour).Format("15:04:05")
					hlAppendTranscript(fmt.Sprintf("[%s] USER: %s", kst, t.text), proj, brainRoot)
				}
				t.logged = t.text
				tracked[id] = t
			}
			if now.Sub(t.ts) > 3*time.Minute {
				delete(tracked, id)
			}
		}

		time.Sleep(3 * time.Second)
	}
}

// ── 자동 통합 스케줄러 (30분마다) ──
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
			"expression":   scrapeExpr,
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
