// hijack_launcher.go — hijack-launcher.mjs Go 포팅
// 통합 브릿지: TG polling + CDP 캡처 + 전사 + Groq 뉴런 추출
// 외부 의존성: 0 (cdp_client.go + Go stdlib)
package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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

var hlTgSkipRoles = map[string]bool{"TOOL": true, "HIJACK_START": true, "AI_RESP": true}

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

// hlTgSendDocument — SVG나 파일 등 Document를 Telegram API로 전송
func hlTgSendDocument(filePath string, caption string) error {
	if hlTgToken == "" || hlTgChatID == "" {
		return fmt.Errorf("no telegram token or chatid")
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	_ = writer.WriteField("chat_id", hlTgChatID)
	// Add caption
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	// Add document
	part, err := writer.CreateFormFile("document", fileInfo.Name())
	if err != nil {
		return err
	}
	part.Write(fileData)

	writer.Close()

	targetURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", hlTgToken)
	req, err := http.NewRequest("POST", targetURL, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send document, status: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
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
		return fmt.Sprintf("🧠 %s\n<pre>%s</pre>", label, esc.Replace(truncated))
	case "CMD":
		return fmt.Sprintf("⚡ %s\n<pre>%s</pre>", label, esc.Replace(truncated))
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

	if role == "ATTACH" {
		filePath := strings.TrimSpace(text)
		if fileExists(filePath) {
			sendTelegramFileSafe(hlTgToken, hlTgChatID, filePath, label)
			appendDebugLog(hlTgNfsRoot, fmt.Sprintf("ATTACH sent: %s", filePath))
		} else {
			appendDebugLog(hlTgNfsRoot, fmt.Sprintf("ATTACH skip (not found): %s", filePath))
		}
		return
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
			case "USER":
				return "👤"
			case "THINK":
				return "🧠"
			case "CMD":
				return "⚡"
			case "NEURON":
				return "🧬"
			default:
				return "💬"
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

			// ── 기능 누락 복구: 기존 MJS 기능 재이식 ──
			if text == "/brain" || text == "/neurons" {
				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/brain", hlCDPPort+90)) // api port: 9090 (hlCDPPort=9000)
				if err == nil {
					var bJSON struct {
						TotalNeurons int `json:"totalNeurons"`
						TotalCounter int `json:"totalCounter"`
						Regions      []struct {
							Name    string        `json:"name"`
							Neurons []interface{} `json:"neurons"`
						} `json:"regions"`
					}
					json.NewDecoder(resp.Body).Decode(&bJSON)
					resp.Body.Close()

					msgText := fmt.Sprintf("🧠 <b>Brain State</b>\n뉴런: %d | 활성: %d\n\n", bJSON.TotalNeurons, bJSON.TotalCounter)
					for _, r := range bJSON.Regions {
						msgText += fmt.Sprintf("  %s: %d\n", r.Name, len(r.Neurons))
					}
					hlTgSafeSend("sendMessage", map[string]string{
						"chat_id": chatID, "text": msgText, "parse_mode": "HTML",
					})
				} else {
					hlTgSend(chatID, "❌ API 연결 실패")
				}
				continue
			}

			if text == "/inject" {
				http.Post(fmt.Sprintf("http://127.0.0.1:%d/api/inject", hlCDPPort+90), "application/json", nil)
				hlTgSend(chatID, "🔄 Inject 트리거 완료")
				continue
			}

			if text == "/stale" {
				stale := collectStaleCodemaps(brainRoot)
				if len(stale) > 0 {
					msg := fmt.Sprintf("⚠️ STALE %d건:\n", len(stale))
					for _, s := range stale { msg += "- " + s + "\n" }
					hlTgSend(chatID, msg)
					go hlCDPInject(hlTgMountedRoom, msg)
				} else {
					hlTgSend(chatID, "✅ 코드맵 전부 최신")
				}
				continue
			}

			if text == "/rooms" {
				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json/list", hlCDPPort))
				if err == nil {
					var targets []struct {
						Type  string `json:"type"`
						URL   string `json:"url"`
						Title string `json:"title"`
					}
					json.NewDecoder(resp.Body).Decode(&targets)
					resp.Body.Close()

					list := ""
					for _, t := range targets {
						if t.Type == "page" && strings.Contains(t.URL, "workbench.html") {
							name := strings.Split(t.Title, " - ")[0]
							name = strings.TrimSpace(name)
							prefix := "  "
							if name == hlTgMountedRoom {
								prefix = "📌"
							}
							list += fmt.Sprintf("%s %s\n", prefix, name)
						}
					}
					hlTgSend(chatID, fmt.Sprintf("🏠 열린 창:\n\n%s\n\n/mount [이름] 으로 전환", list))
				} else {
					hlTgSend(chatID, "❌ CDP 연결 안 됨")
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

			// 마스터 프롬프트 감지 → inbox 저장 전에 넛지로 변환 또는 생략
			cdpPayload := payload
			if strings.Contains(payload, "NeuronFS 자율 진화 명령") || strings.Contains(payload, "마스터 프롬프트") {
				if nudge, hasWork := hlBuildContextualPrompt(brainRoot); hasWork {
					cdpPayload = nudge
					payload = nudge // inbox에도 넛지로 저장
				} else {
					fmt.Println("[TG→IDE] ✅ 시스템 안정 — 마스터 프롬프트 전체 생략 (inbox+CDP)")
					hlTgSend(chatID, "✅ 시스템 안정 — 주입 생략")
					continue
				}
			}

			inboxDir := filepath.Join(agentsDir, targetRoom, "inbox")
			os.MkdirAll(inboxDir, 0750)
			fname := fmt.Sprintf("tg_%d.md", time.Now().UnixMilli())
			content := fmt.Sprintf("# from: telegram\n# priority: normal\n\n%s", payload)
			os.WriteFile(filepath.Join(inboxDir, fname), []byte(content), 0600)
			hlTgSend(chatID, fmt.Sprintf("✅ [%s] 전달됨", targetRoom))

			go hlCDPInject(targetRoom, cdpPayload)
		}

		time.Sleep(1 * time.Second)
	}
}
