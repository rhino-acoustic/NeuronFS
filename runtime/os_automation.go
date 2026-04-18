// auto_accept.go — auto-accept.mjs Go 완전 포팅
// CDP 기반 버튼 자동 클릭 (Run/Accept/Retry)
// + NEURON 명령 감지 (전사 파일 tail)
// + Groq 배치 분석 (5분 유휴 시)
// + 진단 스크립트 (30초마다)
// + Git snapshot (유휴 시)
// 외부 의존성: 0 (cdp_client.go + llm_groq.go 활용)
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	aaCDPPort        = 9000
	aaPollMs         = 1000
	aaClickCoolMs    = 1500
	aaDiagIntervalMs = 30000
	aaIdleThresholdS = 300 // 5분
	aaIdleCheckS     = 30
)

var aaRejectTexts = []string{"always run", "skip", "reject", "cancel", "close", "refine", "running command"}
var aaAcceptTexts = []string{"run", "accept", "accept all", "send all", "retry", "apply", "confirm", "allow once", "allow"}

// ★ BUTTON_ONLY: 이 텍스트는 BUTTON 태그에서만 클릭 (SPAN/A에서 클릭 금지)
var aaButtonOnlyTexts = []string{"run"}

// ── AI 출력 버퍼 (Groq 배치 분석용) ──
var aaAIBuffer []string
var aaAIBufferMu sync.Mutex
var aaLastAIOutput int64 // unix millis
var aaBatchRunning int32 // atomic-like guard

// ── NEURON 명령 중복 방지 ──
var aaNeuronProcessed sync.Map

// JavaScript 클릭 스크립트 (원본 auto-accept.mjs와 100% 동일 안전장치)
func aaClickScript() string {
	rejectJSON, _ := json.Marshal(aaRejectTexts)
	acceptJSON, _ := json.Marshal(aaAcceptTexts)
	buttonOnlyJSON, _ := json.Marshal(aaButtonOnlyTexts)
	return fmt.Sprintf(`(() => {
    function collectAll(root) {
        const found = [];
        const walk = node => {
            if (!node) return;
            if (node.shadowRoot) walk(node.shadowRoot);
            const children = node.children || [];
            for (let i = 0; i < children.length; i++) {
                if (children[i].nodeType === 1) { found.push(children[i]); walk(children[i]); }
            }
        };
        walk(root);
        return found;
    }
    function isMenubar(el) {
        let p = el;
        while (p) {
            const cls = (p.className||'').toString().toLowerCase();
            if (cls.includes('menubar')||cls.includes('titlebar')||cls.includes('tabs-container')||cls.includes('tab ')||cls.includes('monaco-icon')) return true;
            const role = (p.getAttribute('role')||'').toLowerCase();
            if (role === 'menubar' || role === 'menuitem' || role === 'tab' || role === 'tablist') return true;
            p = p.parentElement;
        }
        return false;
    }
    function hasOpacity70Ancestor(el) {
        let p = el.parentElement;
        while (p) { if ((p.className||'').toString().includes('opacity-70')) return true; p = p.parentElement; }
        return false;
    }
    function isInsideChatMessage(el) {
        let p = el.parentElement;
        while (p) { const cls = (p.className||'').toString().toLowerCase(); if (cls.includes('markdown')||cls.includes('message-content')||cls.includes('chat-message')||cls.includes('rendered-markdown')) return true; p = p.parentElement; }
        return false;
    }
    function forceClick(el) {
        const opts = { view: window, bubbles: true, cancelable: true };
        try { el.dispatchEvent(new PointerEvent('pointerdown', { ...opts, pointerId: 1 })); } catch {}
        try { el.dispatchEvent(new MouseEvent('mousedown', opts)); } catch {}
        try { el.dispatchEvent(new MouseEvent('mouseup', opts)); } catch {}
        try { el.click(); } catch {}
        try { el.dispatchEvent(new MouseEvent('click', opts)); } catch {}
        try { el.dispatchEvent(new PointerEvent('pointerup', { ...opts, pointerId: 1 })); } catch {}
    }
    // 비활성 창에서도 Runtime.evaluate로 동작 (Input domain 미사용)
    const REJECT = %s;
    const ACCEPT = %s;
    const BUTTON_ONLY = %s;
    const allEls = collectAll(document.body);
    const candidates = [];
    for (const el of allEls) {
        const tag = el.tagName;
        if (!tag || tag === 'DIV') continue;

        // ★ 모든 요소에 대해 menubar/tab/titlebar 체크
        if (isMenubar(el)) continue;

        const cls = (el.className||'').toString().toLowerCase();
        const isButton = tag === 'BUTTON';
        const hasRole = el.getAttribute('role') === 'button';
        const hasButtonClass = cls.includes('ide-button') || cls.includes('monaco-button');
        const hasCursorPointer = cls.includes('cursor-pointer');
        if (!(isButton || hasRole || hasButtonClass || hasCursorPointer)) continue;
        const text = (el.innerText || el.textContent || '').trim().toLowerCase();
        if (!text || text.length > 20) continue;
        if (el.offsetParent === null) continue;
        if (!isButton && hasOpacity70Ancestor(el)) continue;
        if (REJECT.some(r => text === r || text.includes(r))) continue;

        // ★ label-name/action-label/codicon 필터 (원본과 동일)
        if (cls.includes('label-name') || cls.includes('action-label') || cls.includes('codicon')) continue;

        if (isButton) {
            const matched = ACCEPT.find(a => text === a || text.startsWith(a));
            if (matched) {
                const pri = matched === 'accept all' ? 0 : matched === 'run' ? 1 : matched.includes('accept') ? 2 : 4;
                candidates.push({ el, text: matched, tag, priority: pri });
            }
        } else if (tag === 'SPAN') {
            if (!hasCursorPointer || isInsideChatMessage(el)) continue;
            const matched = ACCEPT.find(a => text === a);
            if (matched && BUTTON_ONLY.includes(matched)) continue; // ★ 'run'은 SPAN에서 절대 클릭 안함
            if (matched) {
                const pri = 100 + (matched.includes('accept') ? 2 : 4);
                candidates.push({ el, text: matched, tag, priority: pri });
            }
        } else if (tag === 'A' && (hasRole || hasButtonClass)) {
            const matched = ACCEPT.find(a => text === a);
            if (matched && BUTTON_ONLY.includes(matched)) continue; // ★ 'run'은 A에서도 절대 클릭 안함
            if (matched) {
                candidates.push({ el, text: matched, tag, priority: 54 });
            }
        }
    }
    candidates.sort((a, b) => a.priority - b.priority);
    if (candidates.length > 0) { const best = candidates[0]; forceClick(best.el); return { found: true, text: best.text, tag: best.tag, total: candidates.length }; }
    return { found: false };
})()`, string(rejectJSON), string(acceptJSON), string(buttonOnlyJSON))
}

// ── 진단 스크립트 (클릭 안함, 보기만) ──
func aaDiagScript() string {
	return `(() => {
    function collectAll(root) {
        const found = [];
        const walk = n => { if (!n) return; if (n.shadowRoot) walk(n.shadowRoot); for (const c of (n.children||[])) { if (c.nodeType===1) { found.push(c); walk(c); } } };
        walk(root); return found;
    }
    function ancestors(el, d) {
        const r = []; let p = el.parentElement; let i = 0;
        while (p && i < d) { const c = (p.className||'').toString().replace(/\s+/g,' ').trim(); if(c) r.push(c.slice(0,60)); p=p.parentElement; i++; }
        return r;
    }
    const kw = ['run','accept','allow','retry','apply','confirm','send'];
    const results = [];
    for (const el of collectAll(document.body)) {
        if (!el.tagName) continue;
        let text = (el.innerText||el.textContent||'').trim();
        if (!text) text = (el.getAttribute('aria-label')||'').trim();
        if (!text) text = (el.getAttribute('title')||'').trim();
        const t = text.toLowerCase();
        if (!t || t.length > 30) continue;
        if (!kw.some(k => t === k || t.startsWith(k))) continue;
        if (el.offsetParent === null) continue;
        results.push({
            tag: el.tagName, text: t.slice(0,25),
            cls: (el.className||'').toString().slice(0,80),
            role: el.getAttribute('role')||'',
            anc: ancestors(el, 3)
        });
    }
    return results;
})()`
}

// ── AI 출력 스크래핑 스크립트 (원본과 동일) ──
func aaScrapeAIScript() string {
	return `(() => {
    const msgs = [];
    document.querySelectorAll('div').forEach(el => {
        const cls = (el.className||'').toString();
        if (cls.includes('whitespace-pre-wrap') && cls.includes('text-sm') && !cls.includes('font-mono')) {
            const parent = el.parentElement;
            if (parent && (parent.className||'').toString().includes('max-h-')) {
                const text = (el.innerText||'').trim();
                if (text && text.length > 2 && text.length < 50000) {
                    msgs.push({role:'PD', text: text.slice(0, 20000)});
                }
            }
        }
    });
    document.querySelectorAll('div').forEach(el => {
        const cls = (el.className||'').toString();
        if (cls.includes('leading-relaxed') && cls.includes('select-text')) {
            const op = parseFloat(getComputedStyle(el).opacity);
            const text = (el.innerText||'').trim();
            if (text && text.length > 20 && text.length < 50000) {
                const role = op < 0.9 ? 'THINK' : 'AI';
                msgs.push({role, text: text.slice(0, 20000)});
            }
        }
    });
    return msgs.slice(-6);
})()`
}

type aaAgent struct {
	name      string
	client    *CDPClient
	targetID  string
	lastClick time.Time
}

var aaAgents sync.Map // targetID → *aaAgent

// runMacroWorker — 순수 매크로(UI 클릭/진단) 워커
func runMacroWorker(brainRoot string) {
	script := aaClickScript()
	diagScript := aaDiagScript()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		aaScanTargets()

		// 1. 순수 매크로 버튼 클릭
		go aaPollButtons(script)

		// 2. 진단 매크로 패널 클릭 (선택적)
		go aaRunDiag(diagScript)
	}
}
func aaLog(format string, args ...interface{}) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("[%s] [AA] %s\n", ts, fmt.Sprintf(format, args...))
}

func aaScanTargets() {
	targets, err := cdpListTargets(aaCDPPort)
	if err != nil {
		return
	}

	active := make(map[string]bool)
	for _, t := range targets {
		active[t.ID] = true
		if _, exists := aaAgents.Load(t.ID); exists {
			continue
		}
		if t.Type == "worker" {
			continue
		}
		// ★ 원본과 동일: workbench.html만 연결 (webview/설정 페이지 제외)
		if !strings.Contains(t.URL, "workbench.html") || t.WebSocketDebuggerURL == "" {
			continue
		}

		client, err := NewCDPClient(t.WebSocketDebuggerURL)
		if err != nil {
			continue
		}
		// Enable Runtime
		client.Call("Runtime.enable", map[string]interface{}{})

		name := t.Title
		if idx := strings.Index(name, " - "); idx > 0 {
			name = name[:idx]
		}

		aaAgents.Store(t.ID, &aaAgent{name: name, client: client, targetID: t.ID})
		aaLog("✅ 연결: [%s]", name)
	}

	// 종료된 타겟 정리
	aaAgents.Range(func(key, value interface{}) bool {
		if !active[key.(string)] {
			a := value.(*aaAgent)
			a.client.Close()
			aaAgents.Delete(key)
			aaLog("🗑️ 종료: [%s]", a.name)
		}
		return true
	})
}

func aaPollButtons(script string) {
	now := time.Now()
	aaAgents.Range(func(key, value interface{}) bool {
		a := value.(*aaAgent)
		if now.Sub(a.lastClick) < time.Duration(aaClickCoolMs)*time.Millisecond {
			return true
		}

		result, err := a.client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    script,
			"returnByValue": true,
		})
		if err != nil {
			return true
		}

		var evalResult struct {
			Result struct {
				Value struct {
					Found bool   `json:"found"`
					Text  string `json:"text"`
					Tag   string `json:"tag"`
					Total int    `json:"total"`
				} `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalResult)

		if evalResult.Result.Value.Found {
			a.lastClick = time.Now()
			aaLog("🖱️ [%s] 클릭: \"%s\" (%s)", a.name, evalResult.Result.Value.Text, evalResult.Result.Value.Tag)
		}
		return true
	})
}

// ── 진단 (30초마다, 클릭 안함) ──
func aaRunDiag(diagScript string) {
	aaAgents.Range(func(key, value interface{}) bool {
		a := value.(*aaAgent)
		result, err := a.client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    diagScript,
			"returnByValue": true,
		})
		if err != nil {
			return true
		}

		type diagItem struct {
			Tag  string   `json:"tag"`
			Text string   `json:"text"`
			Cls  string   `json:"cls"`
			Role string   `json:"role"`
			Anc  []string `json:"anc"`
		}
		var evalRes struct {
			Result struct {
				Value []diagItem `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)

		items := evalRes.Result.Value
		if len(items) > 0 {
			aaLog("🔎 [%s] 버튼 후보 %d개:", a.name, len(items))
			limit := len(items)
			if limit > 10 {
				limit = 10
			}
			for _, it := range items[:limit] {
				aaLog("   <%s> \"%s\" role=\"%s\" cls=\"%.50s\"", it.Tag, it.Text, it.Role, it.Cls)
			}
		}
		return true
	})
}

// ── AI 출력 수집 + NEURON 명령 감지 ──
