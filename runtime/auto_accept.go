// auto_accept.go — auto-accept.mjs Go 포팅
// CDP 기반 버튼 자동 클릭 (Run/Accept/Retry)
// 외부 의존성: 0 (cdp_client.go의 순수 Go WebSocket 사용)
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	aaCDPPort     = 9000
	aaPollMs      = 1000
	aaClickCoolMs = 1500
)

var aaRejectTexts = []string{"always run", "skip", "reject", "cancel", "close", "refine", "running command"}
var aaAcceptTexts = []string{"run", "accept", "accept all", "retry", "apply", "execute", "confirm", "allow once", "allow", "send all"}

// JavaScript 클릭 스크립트 (auto-accept.mjs와 동일)
func aaClickScript() string {
	rejectJSON, _ := json.Marshal(aaRejectTexts)
	acceptJSON, _ := json.Marshal(aaAcceptTexts)
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
        while (p) { const cls = (p.className||'').toString(); if (cls.includes('menubar')||cls.includes('titlebar')) return true; p = p.parentElement; }
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
    const REJECT = %s;
    const ACCEPT = %s;
    const allEls = collectAll(document.body);
    const candidates = [];
    for (const el of allEls) {
        const tag = el.tagName;
        if (!tag || tag === 'DIV') continue;
        const cls = (el.className||'').toString().toLowerCase();
        const isButton = tag === 'BUTTON';
        const hasRole = el.getAttribute('role') === 'button';
        const hasButtonClass = cls.includes('ide-button') || cls.includes('monaco-button');
        const hasCursorPointer = cls.includes('cursor-pointer');
        if (!(isButton || hasRole || hasButtonClass || hasCursorPointer)) continue;
        const text = (el.innerText || el.textContent || '').trim().toLowerCase();
        if (!text || text.length > 20) continue;
        if (el.offsetParent === null) continue;
        if (isMenubar(el)) continue;
        if (!isButton && hasOpacity70Ancestor(el)) continue;
        if (REJECT.some(r => text === r || text.includes(r))) continue;
        if (isButton) {
            const matched = ACCEPT.find(a => text === a || text.startsWith(a));
            if (matched) candidates.push({ el, text: matched, tag, priority: matched === 'run' ? 1 : matched.includes('accept') ? 2 : 4 });
        } else if (tag === 'SPAN') {
            if (!hasCursorPointer || isInsideChatMessage(el)) continue;
            const matched = ACCEPT.find(a => text === a);
            if (matched) candidates.push({ el, text: matched, tag, priority: 100 + (matched === 'run' ? 1 : matched.includes('accept') ? 2 : 4) });
        } else if (tag === 'A' && (hasRole || hasButtonClass)) {
            const matched = ACCEPT.find(a => text === a);
            if (matched) candidates.push({ el, text: matched, tag, priority: 54 });
        }
    }
    candidates.sort((a, b) => a.priority - b.priority);
    if (candidates.length > 0) { const best = candidates[0]; forceClick(best.el); return { found: true, text: best.text, tag: best.tag }; }
    return { found: false };
})()`, string(rejectJSON), string(acceptJSON))
}

type aaAgent struct {
	name      string
	client    *CDPClient
	targetID  string
	lastClick time.Time
}

var aaAgents sync.Map // targetID → *aaAgent

func runAutoAccept(brainRoot string) {
	script := aaClickScript()

	aaLog("🚀 Auto-Accept (Go native) 시작 CDP:%d", aaCDPPort)

	// 10초마다 새 탭 스캔
	go func() {
		for {
			aaScanTargets()
			time.Sleep(10 * time.Second)
		}
	}()

	// 버튼 폴링
	for {
		aaPollButtons(script)
		time.Sleep(time.Duration(aaPollMs) * time.Millisecond)
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
		isWB := strings.Contains(t.URL, "workbench.html") || strings.Contains(t.Title, "Antigravity") || strings.Contains(t.Title, "Cursor")
		if !isWB || t.WebSocketDebuggerURL == "" {
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
