// hijack_launcher.go — hijack-launcher.mjs Go 포팅
// 통합 브릿지: TG polling + CDP 캡처 + 전사 + Groq 뉴런 추출
// 외부 의존성: 0 (cdp_client.go + Go stdlib)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

func hlCDPInject(targetRoom, payload string) {
	targets, err := cdpListTargets(hlCDPPort)
	if err != nil {
		return
	}

	escaped := strings.ReplaceAll(payload, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)

	// 채팅 입력창 감지 + 인젝션 스크립트 (아티팩트/미리보기 탭에는 채팅창이 없으므로 NoTarget 반환)
	injectCode := fmt.Sprintf(`(() => {
		const all = Array.from(document.querySelectorAll("[contenteditable]"));
		const el = all.reverse().find(e => {
			const r = e.getBoundingClientRect();
			const tag = (e.getAttribute("role")||"").toLowerCase();
			return r.height > 0 && r.height < 300 && r.width > 100
				&& (tag === "textbox" || e.closest("[class*='chat']") || e.closest("[class*='input']") || e.closest("[class*='editor-input']"));
		}) || all.reverse().find(e => {
			const r = e.getBoundingClientRect();
			return r.height > 0 && r.height < 300 && r.width > 100;
		});
		if(el) { el.focus(); document.execCommand("insertText", false, "[telegram → NeuronFS] %s"); return "Injected"; }
		return "NoTarget";
	})()`, escaped)

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

		result, err := client.Call("Runtime.evaluate", map[string]interface{}{"expression": injectCode, "returnByValue": true})
		if err != nil {
			client.Close()
			continue
		}

		// 결과 확인: "Injected"면 성공 → Enter 전송 후 종료. "NoTarget"이면 다음 탭 시도.
		resultStr := string(result)
		if strings.Contains(resultStr, "NoTarget") {
			client.Close()
			continue // 이 탭에는 채팅 입력창이 없음 (아티팩트 미리보기 등)
		}

		time.Sleep(500 * time.Millisecond)

		// 제출 전략 1: submit 버튼 직접 클릭 (Antigravity 채팅 UI의 전송 버튼)
		submitScript := `(() => {
			const btn = document.querySelector('button[aria-label*="Send"], button[aria-label*="send"], button[class*="submit"], button[class*="send"]')
				|| Array.from(document.querySelectorAll('button')).find(b => b.querySelector('svg') && b.closest('[class*="chat"], [class*="input"]'));
			if(btn) { btn.click(); return "Clicked"; }
			return "NoButton";
		})()`
		submitResult, _ := client.Call("Runtime.evaluate", map[string]interface{}{"expression": submitScript, "returnByValue": true})
		submitStr := string(submitResult)

		if strings.Contains(submitStr, "NoButton") {
			// 제출 전략 2: CDP 네이티브 Input.dispatchKeyEvent
			client.Call("Input.dispatchKeyEvent", map[string]interface{}{
				"type": "rawKeyDown", "key": "Enter", "code": "Enter",
				"windowsVirtualKeyCode": 13, "nativeVirtualKeyCode": 13,
			})
			time.Sleep(30 * time.Millisecond)
			client.Call("Input.dispatchKeyEvent", map[string]interface{}{
				"type": "char", "text": "\r",
				"windowsVirtualKeyCode": 13, "nativeVirtualKeyCode": 13,
			})
			time.Sleep(30 * time.Millisecond)
			client.Call("Input.dispatchKeyEvent", map[string]interface{}{
				"type": "keyUp", "key": "Enter", "code": "Enter",
				"windowsVirtualKeyCode": 13, "nativeVirtualKeyCode": 13,
			})
		}

		time.Sleep(100 * time.Millisecond)
		client.Close()
		return // 성공적으로 인젝션됨
	}
}

// ── 전사 기록 ──

// (min func removed to prevent redeclaration)

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

	tracked := make(map[string]struct {
		text   string
		ts     time.Time
		logged string
	})

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
				tracked[m.ID] = struct {
					text   string
					ts     time.Time
					logged string
				}{m.Text, now, ""}
			} else if t.text != m.Text {
				tracked[m.ID] = struct {
					text   string
					ts     time.Time
					logged string
				}{m.Text, now, t.logged}
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

	tracked := make(map[string]struct {
		text   string
		ts     time.Time
		logged string
	})

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
				tracked[m.ID] = struct {
					text   string
					ts     time.Time
					logged string
				}{m.Text, now, ""}
			} else if t.text != m.Text {
				tracked[m.ID] = struct {
					text   string
					ts     time.Time
					logged string
				}{m.Text, now, t.logged}
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
