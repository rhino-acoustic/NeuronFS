/**
 * bot-heartbeat.mjs — 에이전트 idle 감지 + 자동 킥
 * 
 * 10초마다 각 봇의 CDP 상태를 체크:
 * - THINKING → 정상, 놔둔다
 * - IDLE 30초 이상 → backlog에서 다음 작업을 자동 주입
 * 
 * Usage: node bot-heartbeat.mjs
 */
import http from 'http';
import WebSocket from 'ws';
import { readdirSync, readFileSync, existsSync, statSync, appendFileSync, writeFileSync } from 'fs';
import { join } from 'path';

const CDP_PORT = 9000;
const CHECK_MS = 15000;
const IDLE_THRESHOLD_MS = 45000;
const BRAIN = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const AGENTS_DIR = join(BRAIN, '_agents');

const botState = {
    pm:   { lastActive: Date.now(), kickCount: 0, lastIdx: -1 },
    bot1: { lastActive: Date.now(), kickCount: 0, lastIdx: -1 },
    entp: { lastActive: Date.now(), kickCount: 0, lastIdx: -1 },
    enfp: { lastActive: Date.now(), kickCount: 0, lastIdx: -1 }
};

// PM 창 title 매칭 — BASEMENT_ADMIN 워크스페이스
const TITLE_MAP = {
    pm: 'basement_admin',
    bot1: 'bot1',
    entp: 'entp',
    enfp: 'enfp'
};

function log(msg) {
    const ts = new Date().toISOString().slice(11, 19);
    const line = `[${ts}] ${msg}`;
    console.log(line);
    try { appendFileSync('C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\logs\\heartbeat.log', line + '\n', 'utf8'); } catch {}
}

function getJson(url) {
    return new Promise((r, j) => http.get(url, res => {
        let d = ''; res.on('data', c => d += c);
        res.on('end', () => r(JSON.parse(d)));
    }).on('error', j));
}

function getNextBacklogItem(agentId) {
    const bl = join(AGENTS_DIR, agentId, 'backlog.md');
    if (!existsSync(bl)) return null;
    const content = readFileSync(bl, 'utf8');
    const lines = content.split('\n');
    
    // 첫 번째 체크되지 않은 항목 찾기
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.match(/^- \[ \] /)) {
            const task = line.replace(/^- \[ \] /, '').trim();
            
            // [PD 교정반영] 선택된 항목을 백로그에서 물리적으로 [x] 로 변경하여 무한 스팸 방지
            lines[i] = line.replace(/^- \[ \] /, '- [x] ');
            try { writeFileSync(bl, lines.join('\n'), 'utf8'); } catch {}
            
            return task;
        }
    }
    return null;
}

function getRecentOutbox(agentId) {
    const outbox = join(AGENTS_DIR, agentId, 'outbox');
    if (!existsSync(outbox)) return '';
    const files = readdirSync(outbox)
        .filter(f => f.endsWith('.md'))
        .map(f => ({ name: f, mtime: statSync(join(outbox, f)).mtimeMs }))
        .sort((a, b) => b.mtime - a.mtime);
    return files.length > 0 ? files[0].name : '';
}

async function checkBot(botName) {
    const titleKey = TITLE_MAP[botName];
    if (!titleKey) return { status: 'NOT_FOUND' };
    const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
    const target = list.find(t =>
        t.url?.includes('workbench') &&
        t.title?.toLowerCase().includes(titleKey)
    );
    if (!target) return { status: 'NOT_FOUND', msgs: 0 };

    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });

    let id = 1; const pending = new Map();
    ws.on('message', m => {
        const d = JSON.parse(m);
        if (d.id && pending.has(d.id)) { pending.get(d.id)(d); pending.delete(d.id); }
    });
    const call = (method, params) => new Promise((resolve, reject) => {
        const i = id++;
        const t = setTimeout(() => { pending.delete(i); reject(new Error('timeout')); }, 5000);
        pending.set(i, r => { clearTimeout(t); resolve(r); });
        ws.send(JSON.stringify({ id: i, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 200));

    const script = `(() => {
        function walk(root) {
            const f = [];
            const w = n => { if(!n) return; if(n.shadowRoot) w(n.shadowRoot); for(const c of(n.children||[])) if(c.nodeType===1){f.push(c);w(c)} };
            w(root); return f;
        }
        const all = walk(document.body);
        
        // 메시지 카운트 (AI 응답 블록 수)
        const msgBlocks = all.filter(el => {
            const cls = (el.className || '').toString();
            return cls.includes('leading-relaxed') && cls.includes('select-text');
        }).length;
        
        // Run/Accept 버튼이 보이면 = auto-accept가 처리할 상태 = 아직 working
        const pendingAction = all.find(el => {
            const txt = (el.textContent || '').trim().toLowerCase();
            const isBtn = el.tagName === 'BUTTON';
            return isBtn && el.offsetParent !== null && (txt === 'run' || txt === 'accept' || txt === 'accept all');
        });

        // 입력창이 활성인지 (= 사용자/봇이 입력 가능 = 응답 끝남)
        const inputReady = all.find(el => {
            const role = el.getAttribute('role');
            const ce = el.getAttribute('contenteditable');
            return role === 'textbox' && ce === 'true' && el.offsetParent !== null;
        });

        return { msgBlocks, pendingAction: !!pendingAction, inputReady: !!inputReady };
    })()`;

    const res = await call('Runtime.evaluate', { expression: script, returnByValue: true });
    const v = res?.result?.result?.value;
    ws.close();

    const msgCount = v?.msgBlocks || 0;
    const prevCount = botState[botName]?.prevMsgCount || 0;
    botState[botName].prevMsgCount = msgCount;

    // pendingAction(Run/Accept 버튼) = auto-accept가 처리 → working
    if (v?.pendingAction) return { status: 'WORKING' };
    // 메시지 수 증가 = 새 응답 생성 중 → working
    if (msgCount > prevCount) return { status: 'WORKING' };
    // 입력창 활성 + 메시지 수 동일 = 응답 끝남 = IDLE
    if (v?.inputReady && msgCount === prevCount && prevCount > 0) return { status: 'IDLE' };
    // 첫 체크이거나 판별 불가 → working으로 간주
    return { status: prevCount === 0 ? 'INIT' : 'WORKING' };
}

async function kickBot(botName, task) {
    const titleKey = TITLE_MAP[botName];
    if (!titleKey) return false;
    const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
    const target = list.find(t =>
        t.url?.includes('workbench') &&
        t.title?.toLowerCase().includes(titleKey)
    );
    if (!target) return false;

    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });

    let id = 1; const pending = new Map();
    ws.on('message', m => {
        const d = JSON.parse(m);
        if (d.id && pending.has(d.id)) { pending.get(d.id)(d); pending.delete(d.id); }
    });
    const call = (method, params) => new Promise((resolve, reject) => {
        const i = id++;
        const t = setTimeout(() => { pending.delete(i); reject(new Error('timeout')); }, 10000);
        pending.set(i, r => { clearTimeout(t); resolve(r); });
        ws.send(JSON.stringify({ id: i, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 300));

    const focusScript = `(() => {
        function collectAll(root) {
            const found = [];
            const walk = n => { if(!n)return; if(n.shadowRoot)walk(n.shadowRoot); for(const c of(n.children||[]))if(c.nodeType===1){found.push(c);walk(c);} };
            walk(root); return found;
        }
        const allEls = collectAll(document.body);
        let ci = allEls.find(el => (el.getAttribute('aria-label')||'').toLowerCase()==='message input' && el.getAttribute('role')==='textbox');
        if(!ci) ci = allEls.find(el => el.getAttribute('contenteditable')==='true' && el.getAttribute('role')==='textbox' && el.offsetParent!==null);
        if(!ci) return {ok:false};
        ci.textContent=''; ci.focus();
        const o={view:window,bubbles:true,cancelable:true};
        try{ci.dispatchEvent(new PointerEvent('pointerdown',{...o,pointerId:1}))}catch{}
        try{ci.dispatchEvent(new MouseEvent('mousedown',o))}catch{}
        try{ci.click()}catch{}
        return {ok:true};
    })()`;

    const r = await call('Runtime.evaluate', { expression: focusScript, returnByValue: true });
    if (!r?.result?.result?.value?.ok) { ws.close(); return false; }

    await new Promise(r => setTimeout(r, 200));
    await call('Input.insertText', { text: task });
    await new Promise(r => setTimeout(r, 200));
    await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13 });
    await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13 });

    ws.close();
    return true;
}

async function heartbeatLoop() {
    for (const botName of Object.keys(botState)) {
        try {
            // [QUARANTINE PROTOCOL] 폭탄 맞은 에이전트는 Heartbeat 차단 및 PM 호출
            const agentBombPath = join(AGENTS_DIR, botName, 'brainstem', 'bomb.neuron');
            if (existsSync(agentBombPath) || existsSync(join(BRAIN, 'brainstem', 'bomb.neuron'))) {
                const alertFile = join(AGENTS_DIR, 'pm', 'inbox', `${botName}_bomb_alert_${new Date().toISOString().slice(0,10).replace(/-/g,'')}.md`);
                if (!existsSync(alertFile)) {
                    writeFileSync(alertFile, `[BOMB ALERT] ${botName} 셧다운 (무한 루프/망상 감지).\nWatchdog이 해당 에이전트의 Heartbeat를 차단하고 격리 중입니다.\n[지시] 원인을 분석하고 폭탄(bomb.neuron)을 삭제하여 정상화하십시오.\n`, 'utf8');
                    log(`🚨 [QUARANTINE] ${botName} 폭탄 감지. 주입 차단 및 PM 호출 완료.`);
                }
                continue;
            }

            const { status } = await checkBot(botName);
            const state = botState[botName];

            if (status === 'WORKING') {
                state.lastActive = Date.now();
                state.consecutiveKicks = 0;  // 작업 중이면 킥 카운터 리셋
            } else if (status === 'IDLE') {
                const idleMs = Date.now() - state.lastActive;
                if (idleMs >= IDLE_THRESHOLD_MS) {
                    // Step Budget: 연속 킥 3회 초과 시 쿨다운 (무한 루프 방지)
                    if (!state.consecutiveKicks) state.consecutiveKicks = 0;
                    if (state.consecutiveKicks >= 3) {
                        log(`⏸️ ${botName} 연속 킥 ${state.consecutiveKicks}회 → 60s 쿨다운`);
                        state.lastActive = Date.now() + 30000; // 추가 30초 대기
                        state.consecutiveKicks = 0;
                        continue;
                    }

                    const nextTask = getNextBacklogItem(botName);
                    const lastOutput = getRecentOutbox(botName);

                    // Reflect & Critique: 결과물이 있으면 다른 봇에 자동 크로스 리뷰 요청
                    if (lastOutput && state.kickCount > 0 && state.kickCount % 3 === 0) {
                        const reviewTarget = botName === 'enfp' ? 'entp' : 'enfp';
                        const reviewPath = join(AGENTS_DIR, reviewTarget, 'inbox', 
                            `${new Date().toISOString().slice(0,10).replace(/-/g,'')}_heartbeat_review_${botName}.md`);
                        try {
                            writeFileSync(reviewPath, 
                                `# from: heartbeat\n# priority: normal\n\n${botName}의 최신 산출물(${lastOutput})을 리뷰하라.\n품질 미달이면 ${botName}/inbox에 재작업 요청을 보내라.\n`, 'utf8');
                            log(`🔄 Reflect: ${botName} → ${reviewTarget} 크로스 리뷰 요청`);
                        } catch {}
                    }

                    // Dynamic Priority: 컨텍스트에 맞는 킥 메시지
                    let kickMsg;
                    const now = Date.now();
                    const titleKey = TITLE_MAP[botName];
                    if (!nextTask) {
                        // [PM 다이렉트 패치: 쳇바퀴 유휴 알람 영구 MUTE]
                        // 런타임 스레드들이 자율 생존하므로 PM에게 억지 핑(Wake up)을 쏘는 로직을 완전히 파기/제거합니다.
                        log(`[${botName}] No tasks in backlog. Zero-Queue Engine Standby (No Injection).`);
                        state.waitStart = now; // 초기화
                        continue; // PM이든 하위 봇이든, 가비지 알람 킥주입 무조건 생략 (음소거)!
                    } else {
                        kickMsg = `[HEARTBEAT] 이전 산출물: ${lastOutput || '없음'}. 다음 backlog를 즉시 주입한다: [${titleKey}] ${nextTask}. 완료 후 결과를 아웃박스에 저장하라.`;
                    }

                    log(`⚡ ${botName} IDLE ${Math.round(idleMs / 1000)}s → KICK #${state.kickCount + 1}: ${(nextTask || 'standby-ping').slice(0, 50)}`);
                    const ok = await kickBot(botName, kickMsg);
                    if (ok) {
                        state.lastActive = Date.now();
                        state.kickCount++;
                        state.consecutiveKicks++;
                        log(`✅ ${botName} kicked (total: ${state.kickCount}, consecutive: ${state.consecutiveKicks})`);
                    }
                }
            }
        } catch (e) {
            // silent
        }
    }
}

async function main() {
    log('=== Bot Heartbeat Monitor ===');
    log(`Check: ${CHECK_MS / 1000}s | Idle threshold: ${IDLE_THRESHOLD_MS / 1000}s`);
    log('Monitoring: bot1, entp, enfp');
    log('');

    // 최초 상태 출력
    for (const b of Object.keys(botState)) {
        try {
            const { status, msgs } = await checkBot(b);
            log(`${b}: ${status} (${msgs} msgs)`);
        } catch { log(`${b}: cannot reach`); }
    }
    log('');

    setInterval(() => heartbeatLoop().catch(() => {}), CHECK_MS);
    log('Heartbeat running... Ctrl+C to stop');
}

main().catch(e => { log(`💥 ${e.message}`); process.exit(1); });
