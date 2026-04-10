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

func hlLoadTelegram(nfsRoot string) {
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
}

func hlTgSend(chatID, text string) {
	if hlTgToken == "" || chatID == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", hlTgToken),
		"application/json",
		strings.NewReader(string(body)),
	)
	if err == nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
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
			if text == "/start" || text == "/status" {
				hlTgSend(chatID, fmt.Sprintf("🧠 NeuronFS Bridge (Go native)\n현재 방: 📌 %s", hlTgMountedRoom))
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
		time.Sleep(100 * time.Millisecond)
		client.Close()
		break
	}
}

// ── 전사 기록 ──
var hlTranscriptDedup sync.Map

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

	// Dedup
	h := sha256.Sum256([]byte(entry[:min(len(entry), 200)]))
	hashKey := file + "|" + hex.EncodeToString(h[:8])
	if _, loaded := hlTranscriptDedup.LoadOrStore(hashKey, true); loaded {
		return
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		fmt.Fprintln(f, entry)
		f.Close()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── CDP DOM 스크래핑 ──
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
				go hlAttachDOMScraper(t.WebSocketDebuggerURL, t.Title, brainRoot)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func hlAttachDOMScraper(wsURL, title, brainRoot string) {
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

		// 안정화 판단 (4초)
		for id, t := range tracked {
			if t.text != t.logged && now.Sub(t.ts) > 4*time.Second {
				if len(t.text) > len(t.logged) {
					kst := now.UTC().Add(9 * time.Hour).Format("15:04:05")
					// tracked에서 role 가져오기
					role := "AI"
					for _, m := range evalRes.Result.Value {
						if m.ID == id {
							role = m.Role
							break
						}
					}
					hlAppendTranscript(fmt.Sprintf("[%s] %s: %s", kst, role, t.text), proj, brainRoot)
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
