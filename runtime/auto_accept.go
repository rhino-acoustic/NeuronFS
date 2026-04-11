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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
var aaAIBuffer     []string
var aaAIBufferMu   sync.Mutex
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
                if (text && text.length > 2 && text.length < 5000) {
                    msgs.push({role:'PD', text: text.slice(0, 2000)});
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
                msgs.push({role, text: text.slice(0, 2000)});
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

func runAutoAccept(brainRoot string) {
	script := aaClickScript()
	diagScript := aaDiagScript()
	scrapeScript := aaScrapeAIScript()

	aaLog("🚀 Auto-Accept (Go native) 시작 CDP:%d", aaCDPPort)

	// 10초마다 새 탭 스캔
	go func() {
		for {
			aaScanTargets()
			time.Sleep(10 * time.Second)
		}
	}()

	// 30초마다 진단
	go func() {
		time.Sleep(3 * time.Second) // 초기 대기
		for {
			aaRunDiag(diagScript)
			time.Sleep(time.Duration(aaDiagIntervalMs) * time.Millisecond)
		}
	}()

	// 10초마다 AI 출력 수집
	go func() {
		for {
			aaPollAIOutput(scrapeScript, brainRoot)
			time.Sleep(10 * time.Second)
		}
	}()

	// 30초마다 유휴 체크 → Groq 배치 분석 + Git snapshot
	go func() {
		for {
			aaIdleCheck(brainRoot)
			time.Sleep(time.Duration(aaIdleCheckS) * time.Second)
		}
	}()

	// 버튼 폴링 (메인 루프)
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
func aaPollAIOutput(scrapeScript string, brainRoot string) {
	aaAgents.Range(func(key, value interface{}) bool {
		a := value.(*aaAgent)
		result, err := a.client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    scrapeScript,
			"returnByValue": true,
		})
		if err != nil {
			return true
		}

		type scrapedMsg struct {
			Role string `json:"role"`
			Text string `json:"text"`
		}
		var evalRes struct {
			Result struct {
				Value []scrapedMsg `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)

		msgs := evalRes.Result.Value
		if len(msgs) == 0 {
			return true
		}

		added := 0
		aaAIBufferMu.Lock()
		for _, m := range msgs {
			if m.Text == "" || len(m.Text) < 3 {
				continue
			}
			// 중복 방지
			keyStr := m.Role + ":" + m.Text[:aaMin(100, len(m.Text))]
			if len(aaAIBuffer) > 0 && aaAIBuffer[len(aaAIBuffer)-1] == keyStr {
				continue
			}
			aaAIBuffer = append(aaAIBuffer, keyStr)
			aaLastAIOutput = time.Now().UnixMilli()
			added++

			// AI 응답에서만 NEURON 명령 감지
			if m.Role == "AI" {
				go aaDetectNeuronCommands(m.Text, brainRoot)
				go aaDetectEvolveRequest(m.Text, brainRoot)
			}
		}
		if len(aaAIBuffer) > 5000 {
			aaAIBuffer = aaAIBuffer[len(aaAIBuffer)-5000:]
		}
		aaAIBufferMu.Unlock()

		if added > 0 {
			aaLog("🧠 [전사] +%d개 수집 (버퍼: %d)", added, len(aaAIBuffer))
		}
		return true
	})
}

func aaMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── NEURON 명령 즉시 감지 (원본과 동일) ──
var aaNeuronCmdRe = regexp.MustCompile(`\[NEURON:(\{[^}]+\})\]`)

func aaDetectNeuronCommands(text string, brainRoot string) {
	nfsExe, _ := os.Executable()

	matches := aaNeuronCmdRe.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		var cmd map[string]string
		if err := json.Unmarshal([]byte(match[1]), &cmd); err != nil {
			continue
		}

		key := match[1]
		if _, loaded := aaNeuronProcessed.LoadOrStore(key, true); loaded {
			continue
		}

		if growPath, ok := cmd["grow"]; ok && growPath != "" {
			aaLog("🧬 [NEURON] grow 감지: %s — %s", growPath, cmd["reason"])
			out, err := exec.Command(nfsExe, brainRoot, "--grow", growPath).CombinedOutput()
			if err != nil {
				aaLog("🧬 [NEURON] ❌ grow 실패: %v %s", err, string(out))
			} else {
				aaLog("🧬 [NEURON] ✅ grow 완료: %s", growPath)
			}
		}
		if firePath, ok := cmd["fire"]; ok && firePath != "" {
			aaLog("🔥 [NEURON] fire 감지: %s", firePath)
			out, err := exec.Command(nfsExe, brainRoot, "--fire", firePath).CombinedOutput()
			if err != nil {
				aaLog("🔥 [NEURON] ❌ fire 실패: %v %s", err, string(out))
			} else {
				aaLog("🔥 [NEURON] ✅ fire 완료: %s", firePath)
			}
		}
	}
}

// ── 유휴 체크 → Groq 배치 분석 + Git snapshot ──
func aaIdleCheck(brainRoot string) {
	aaAIBufferMu.Lock()
	bufLen := len(aaAIBuffer)
	lastOutput := aaLastAIOutput
	aaAIBufferMu.Unlock()

	if lastOutput == 0 || bufLen == 0 {
		return
	}

	idleMs := time.Now().UnixMilli() - lastOutput
	if idleMs < int64(aaIdleThresholdS*1000) {
		return
	}

	aaLog("🧠 [NeuronFS] %d초 유휴 → 배치 분석 트리거", idleMs/1000)
	aaBatchAnalyze(brainRoot)

	// Git snapshot
	nfsExe, _ := os.Executable()
	if out, err := exec.Command(nfsExe, brainRoot, "--snapshot").CombinedOutput(); err == nil {
		aaLog("🧠 [NeuronFS] Git snapshot 완료")
	} else {
		_ = out // git 없으면 조용히 스킵
	}
}

// ── Groq 배치 분석 (원본 batchAnalyze와 동일 로직) ──
func aaBatchAnalyze(brainRoot string) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		aaLog("🧠 [NeuronFS] GROQ_API_KEY 미설정 — 배치 분석 스킵")
		return
	}

	// 이미 실행 중이면 스킵
	aaAIBufferMu.Lock()
	if aaBatchRunning != 0 {
		aaAIBufferMu.Unlock()
		return
	}
	aaBatchRunning = 1
	bufCopy := make([]string, len(aaAIBuffer))
	copy(bufCopy, aaAIBuffer)
	aaAIBufferMu.Unlock()

	defer func() {
		aaAIBufferMu.Lock()
		aaBatchRunning = 0
		aaAIBufferMu.Unlock()
	}()

	// brain_state.json에서 규칙 로드
	rules := aaLoadBrainRules(brainRoot)
	if len(rules) == 0 {
		aaLog("🧠 [NeuronFS] brain_state.json 규칙 없음 — 스킵")
		return
	}

	transcript := strings.Join(bufCopy, "\n---\n")
	if len(transcript) > 8000 {
		transcript = transcript[:8000]
	}
	aaLog("🧠 [NeuronFS] 배치 분석 시작: %d개 메시지, %d개 규칙", len(bufCopy), len(rules))

	systemPrompt := fmt.Sprintf(`당신은 NeuronFS 교정 분석기입니다. AI 출력에서 PD(사용자) 교정 패턴과 새 인사이트를 추출합니다.

기존 규칙:
%s

분석 항목:
1. CORRECTIONS: PD가 AI를 교정한 것 ("하지마", "이렇게 해", 불만 표현)
2. VIOLATED: AI가 기존 규칙을 위반한 것
3. REINFORCED: AI가 기존 규칙을 잘 따른 것 (최대 5개)
4. INSIGHTS: AI가 발견한 새 패턴/지식

뉴런 경로 규칙:
- 7개 영역: brainstem, limbic, hippocampus, sensors, cortex, ego, prefrontal
- 한글/한자 사용 가능 (예: cortex/코딩/禁console_log)
- 금지: 禁 접두어, 권장: 推 접두어
- snake_case 또는 한글 조합

JSON으로만 응답:
{
  "corrections": [{"path": "cortex/코딩/禁console_log", "reason": "PD 3회 교정", "counter_add": 3}],
  "violated": [{"rule": "cortex/frontend/coding/no_console_log", "reason": "used console.log"}],
  "reinforced": ["brainstem/verify_before_deliver"],
  "insights": [{"path": "cortex/NAS/推robocopy_MT", "reason": "병렬 전송에 효과적"}]
}

CRITICAL: violated/reinforced의 rule 경로는 기존 규칙과 정확히 일치해야 합니다.`, strings.Join(rules, "\n"))

	content, err := callGroqRaw(apiKey, groqRequest{
		Model:          "llama-3.3-70b-versatile",
		Messages:       []groqMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: "대화 전사:\n" + transcript}},
		Temperature:    0.1,
		MaxTokens:      1000,
		TopP:           1.0,
		Stream:         false,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	})
	if err != nil {
		aaLog("🧠 [NeuronFS] Groq 호출 실패: %v", err)
		return
	}

	// JSON 파싱
	type correction struct {
		Path       string `json:"path"`
		Reason     string `json:"reason"`
		CounterAdd int    `json:"counter_add"`
	}
	type violated struct {
		Rule   string `json:"rule"`
		Reason string `json:"reason"`
	}
	type insight struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}
	var analysis struct {
		Corrections []correction `json:"corrections"`
		Violated    []violated   `json:"violated"`
		Reinforced  []string     `json:"reinforced"`
		Insights    []insight    `json:"insights"`
	}

	jsonRe := regexp.MustCompile(`\{[\s\S]*\}`)
	jsonMatch := jsonRe.FindString(content)
	if jsonMatch == "" {
		aaLog("🧠 [NeuronFS] Groq 응답 파싱 실패")
		return
	}
	if err := json.Unmarshal([]byte(jsonMatch), &analysis); err != nil {
		aaLog("🧠 [NeuronFS] JSON 파싱 실패: %v", err)
		return
	}

	nfsExe, _ := os.Executable()
	nfsRoot := filepath.Dir(brainRoot)
	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	os.MkdirAll(filepath.Dir(inboxPath), 0750)

	validRegions := map[string]bool{"brainstem": true, "limbic": true, "hippocampus": true, "sensors": true, "cortex": true, "ego": true, "prefrontal": true}

	cleanPath := func(raw string) string {
		p := regexp.MustCompile(`\[.*?\]\s*`).ReplaceAllString(raw, "")
		p = regexp.MustCompile(`\s*\(\d+\)$`).ReplaceAllString(p, "")
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		parts := strings.SplitN(p, "/", 2)
		if len(parts) < 1 || !validRegions[parts[0]] {
			return ""
		}
		return p
	}

	// 교정 → _inbox에 저장
	for _, c := range analysis.Corrections {
		cPath := cleanPath(c.Path)
		if cPath == "" {
			continue
		}
		aaLog("🧠 [교정] %s — %s", cPath, c.Reason)
		entry, _ := json.Marshal(map[string]interface{}{
			"ts": time.Now().Format(time.RFC3339), "type": "correction",
			"text": c.Reason, "source": "groq-batch", "path": cPath,
			"counter_add": c.CounterAdd,
		})
		f, err := os.OpenFile(inboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, string(entry))
			f.Close()
		}
	}

	// 위반 → --fire
	for _, v := range analysis.Violated {
		rulePath := cleanPath(v.Rule)
		if rulePath == "" {
			continue
		}
		aaLog("🧠 [위반] %s — %s", rulePath, v.Reason)
		exec.Command(nfsExe, brainRoot, "--fire", rulePath).Run()
	}

	// 준수 → --fire (최대 5개)
	for i, r := range analysis.Reinforced {
		if i >= 5 {
			break
		}
		rulePath := cleanPath(r)
		if rulePath == "" {
			continue
		}
		exec.Command(nfsExe, brainRoot, "--fire", rulePath).Run()
	}

	// 인사이트 → _inbox
	for i, s := range analysis.Insights {
		if i >= 3 {
			break
		}
		insightPath := cleanPath(s.Path)
		if insightPath == "" {
			continue
		}
		aaLog("🧠 [인사이트] %s — %s", insightPath, s.Reason)
		entry, _ := json.Marshal(map[string]interface{}{
			"ts": time.Now().Format(time.RFC3339), "type": "insight",
			"text": s.Reason, "source": "groq-batch", "path": insightPath,
			"counter_add": 1,
		})
		f, err := os.OpenFile(inboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, string(entry))
			f.Close()
		}
	}

	cCount := len(analysis.Corrections)
	vCount := len(analysis.Violated)
	rCount := len(analysis.Reinforced)
	iCount := len(analysis.Insights)
	aaLog("🧠 [NeuronFS] 배치 완료: 교정 %d건, 위반 %d건, 강화 %d건, 인사이트 %d건", cCount, vCount, rCount, iCount)

	// 버퍼 초기화
	aaAIBufferMu.Lock()
	aaAIBuffer = nil
	aaAIBufferMu.Unlock()

	_ = nfsRoot // suppress unused
}

// ── brain_state.json에서 규칙 로드 ──
func aaLoadBrainRules(brainRoot string) []string {
	nfsRoot := filepath.Dir(brainRoot)
	statePath := filepath.Join(nfsRoot, "brain_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil
	}
	var state struct {
		Regions []struct {
			Name    string `json:"name"`
			Neurons []struct {
				Path    string `json:"path"`
				Counter int    `json:"counter"`
			} `json:"neurons"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}
	var rules []string
	for _, r := range state.Regions {
		for _, n := range r.Neurons {
			if n.Counter >= 5 {
				rules = append(rules, fmt.Sprintf("[%s] %s (%d)", r.Name, n.Path, n.Counter))
			}
		}
	}
	return rules
}

// ── 자가발전 인젝션: [EVOLVE:proceed] 감지 → git snapshot → "진행" 자동 입력 ──
// AI가 "진행할까?"성 질문을 할 때 [EVOLVE:proceed]를 출력하면,
// auto-accept가 이를 감지하여 git snapshot 후 "진행"을 자동 입력한다.
var aaEvolveProcRe = regexp.MustCompile(`\[EVOLVE:proceed\]`)
var aaEvolveProcessed sync.Map

func aaDetectEvolveRequest(text string, brainRoot string) {
	if !aaEvolveProcRe.MatchString(text) {
		return
	}

	// 중복 방지 (같은 텍스트 해시)
	key := "evolve:" + text[:aaMin(80, len(text))]
	if _, loaded := aaEvolveProcessed.LoadOrStore(key, true); loaded {
		return
	}

	aaLog("🧬 [EVOLVE] proceed 요청 감지 — 선 git snapshot")

	// 1. 선 git snapshot (안전장치)
	nfsExe, _ := os.Executable()
	if out, err := exec.Command(nfsExe, brainRoot, "--snapshot").CombinedOutput(); err == nil {
		aaLog("🧬 [EVOLVE] 📸 Git snapshot 완료")
	} else {
		aaLog("🧬 [EVOLVE] ⚠️ Git snapshot 실패: %s", string(out))
	}

	// 2. CDP로 "진행" 텍스트 주입
	injected := false
	aaAgents.Range(func(k, v interface{}) bool {
		a := v.(*aaAgent)
		injectScript := `(() => {
			const all = Array.from(document.querySelectorAll("[contenteditable]"));
			const el = all.reverse().find(e => {
				const r = e.getBoundingClientRect();
				return r.height > 0 && r.height < 300 && r.width > 100;
			}) || all[0];
			if (el) {
				el.focus();
				document.execCommand("insertText", false, "[NeuronFS 자율진화 트리거] 전사 패턴 분석 후 커뮤니티 검색해서 자가진화 실행. growth.log 확인하고 다음 action 자동 결정.");
				// Enter 키 전송
				el.dispatchEvent(new KeyboardEvent("keydown", {key:"Enter",code:"Enter",keyCode:13,which:13,bubbles:true}));
				return "Injected";
			}
			return "NoTarget";
		})()`

		result, err := a.client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    injectScript,
			"returnByValue": true,
		})
		if err != nil {
			return true // continue to next agent
		}

		var evalRes struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)

		if evalRes.Result.Value == "Injected" {
			aaLog("🧬 [EVOLVE] ✅ '진행' 자동 주입 완료 → [%s]", a.name)
			injected = true
			return false // stop iteration
		}
		return true
	})

	if !injected {
		aaLog("🧬 [EVOLVE] ⚠️ 주입 실패 — CDP 타겟 없음")
	}
}
