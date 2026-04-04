/**
 * agent-bridge.mjs — 에이전트 간 자동 통신 브릿지
 * 
 * inbox 폴더를 감시 → 새 파일 감지 → CDP로 해당 에이전트에 인젝션
 * 
 * 통신 흐름:
 *   entp가 _agents/bot1/inbox/에 파일 생성
 *   → bridge가 감지
 *   → CDP로 bot1 창에 내용 주입
 *   → bot1이 작업 수행
 *   → bot1이 _agents/entp/inbox/에 응답 파일 생성
 *   → bridge가 감지 → entp에 주입
 * 
 * Usage: node agent-bridge.mjs
 */
import http from 'http';
import WebSocket from 'ws';
import { readFileSync, readdirSync, renameSync, existsSync, mkdirSync, appendFileSync } from 'fs';
import { join, basename } from 'path';

const CDP_PORT = 9000;
const BRAIN = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const NAS_BRAIN = 'Z:\\VOL1\\VGVR\\BRAIN\\LW\\system\\neurons\\brain_v4';
const AGENTS_DIR = join(BRAIN, '_agents');
const NAS_AGENTS_DIR = existsSync(NAS_BRAIN) ? join(NAS_BRAIN, '_agents') : null;
const POLL_MS = 3000;
const LOG_FILE = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\logs\\bridge.log';

// 에이전트 → CDP 창 title 매핑 (title.startsWith로 매칭)
const AGENT_TARGETS = {
    bot1: 'bot1',
    entp: 'entp',
    enfp: 'enfp',
    // PM 창은 직접 주입하지 않지만 outbox는 모니터링
};

function log(msg) {
    const ts = new Date().toISOString().slice(11, 19);
    const line = `[${ts}] ${msg}`;
    console.log(line);
    try { appendFileSync(LOG_FILE, line + '\n', 'utf8'); } catch {}
}

function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => resolve(JSON.parse(d)));
        }).on('error', reject);
    });
}

async function injectToAgent(agentId, message) {
    const keyword = AGENT_TARGETS[agentId];
    if (!keyword) { log(`❌ Unknown agent: ${agentId}`); return false; }

    let list;
    try { list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`); }
    catch { log('❌ CDP not connected'); return false; }

    const target = list.find(t =>
        t.url?.includes('workbench.html') &&
        t.title?.toLowerCase().startsWith(keyword.toLowerCase()) &&
        t.type === 'page'
    );
    if (!target) { log(`❌ ${keyword} window not found`); return false; }

    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });

    let id = 1;
    const pending = new Map();
    ws.on('message', msg => {
        const data = JSON.parse(msg);
        if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
    });
    const call = (method, params) => new Promise((resolve, reject) => {
        const myId = id++;
        const t = setTimeout(() => { pending.delete(myId); reject(new Error('timeout')); }, 10000);
        pending.set(myId, r => { clearTimeout(t); resolve(r); });
        ws.send(JSON.stringify({ id: myId, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 300));

    const focusExpr = `(() => {
        function collectAll(root) {
            const found = [];
            const walk = n => { if(!n)return; if(n.shadowRoot)walk(n.shadowRoot); if(n.matches&&n.matches('[aria-label="Message input"][role="textbox"]'))found.push(n); for(const c of(n.children||[]))walk(c); };
            walk(root); return found;
        }
        const inputs = collectAll(document);
        if (inputs.length > 0) { inputs[0].textContent=''; inputs[0].focus(); return 'ok'; }
        return 'not_found';
    })()`;

    const focusRes = await call('Runtime.evaluate', { expression: focusExpr });
    if (focusRes?.result?.result?.value !== 'ok') {
        log(`⚠️ ${agentId} chat input not found`);
        ws.close();
        return false;
    }

    await new Promise(r => setTimeout(r, 200));
    await call('Input.insertText', { text: message });
    await new Promise(r => setTimeout(r, 200));
    await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });
    await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });

    ws.close();
    return true;
}

// inbox 폴링
const processed = new Set();

async function checkInboxes() {
    // 로컬 + NAS inbox 모두 감시
    const sources = [AGENTS_DIR];
    if (NAS_AGENTS_DIR && existsSync(NAS_AGENTS_DIR)) sources.push(NAS_AGENTS_DIR);

    for (const agentId of Object.keys(AGENT_TARGETS)) {
        for (const baseDir of sources) {
            const inbox = join(baseDir, agentId, 'inbox');
            if (!existsSync(inbox)) {
                try { mkdirSync(inbox, { recursive: true }); } catch {}
                continue;
            }

            const files = readdirSync(inbox).filter(f => f.endsWith('.md') && !processed.has(f) && !f.startsWith('_'));
            for (const file of files) {
                // [PD 교정반영] Race Condition Lock: await로 인해 주입이 지연되는 동안 
                // 다음 폴링 인터벌이 동일 파일을 읽는 것을 차단
                processed.add(file);
                
                const filePath = join(inbox, file);
                let content;
                try { content = readFileSync(filePath, 'utf8'); } catch { processed.delete(file); continue; }

                const isNas = baseDir !== AGENTS_DIR;
                log(`📨 ${isNas ? '[NAS] ' : ''}${agentId}/inbox/${file}`);

                const fromMatch = content.match(/^# from: (.+)$/m);
                const priorityMatch = content.match(/^# priority: (.+)$/m);
                const from = fromMatch ? fromMatch[1] : 'unknown';
                const priority = priorityMatch ? priorityMatch[1] : 'normal';
                let body = content.replace(/^#.*$/gm, '').trim();

                // [ENFP CHURN-FREEZE PATCH INJECTED]
                // 브릿지 페이로드에 특정 이탈 트리거가 감지되면 ENFP 마이크로 카피로 강제 스위칭/로깅
                if (body.includes('USER_LOGOUT') || body.includes('IDLE_TIMEOUT') || body.includes('LEAVE')) {
                    try {
                        const churnMod = await import('file:///C:/Users/BASEMENT_ADMIN/NeuronFS/brain_v4/_agents/enfp/outbox/20260329_enfp_churn_micro_config.js');
                        const churnRule = churnMod.getRuleByTrigger("cart_idle_12h_exit_5s") || churnMod.churnMicroRules[0];
                        if (churnRule && churnRule.active) {
                            log(`[CHURN_FREEZE_OVERRIDE] 🚨 이탈 방어 작동: ${churnRule.content}`);
                            body = `[SYSTEM_CHURN_OVERRIDE] 사용자 이탈 감지 방어. 즉시 팝업/모달 생성:\n"${churnRule.content}"\n\n(Original Payload): ${body}`;
                        }
                    } catch(err) { log(`⚠️ Churn-Freeze injection load failed: ${err.message}`); }
                }

                const injection = `[${from} → ${agentId}] ${priority === 'urgent' ? '🚨 URGENT: ' : ''}${body}`;

                const ok = await injectToAgent(agentId, injection);
                if (ok) {
                    try {
                        // 물리 파일 시스템 레벨에서 제외
                        renameSync(filePath, join(inbox, `_${file}`));
                    } catch (e) {
                        log(`⚠️ File rename failed: ${e.message}`);
                    }
                    log(`✅ ${file} → ${agentId} injected${isNas ? ' (from NAS)' : ''}`);
                } else {
                    // 주입 실패 시 다음 폴링에서 재시도할 수 있도록 Lock 해제
                    processed.delete(file);
                }
            }
        }
    }
}

// outbox 모니터 (PM 리포트용)
async function checkOutboxes() {
    for (const agentId of Object.keys(AGENT_TARGETS)) {
        const outbox = join(AGENTS_DIR, agentId, 'outbox');
        if (!existsSync(outbox)) continue;

        const files = readdirSync(outbox).filter(f => f.endsWith('.md') && !processed.has('out_' + f));
        for (const file of files) {
            processed.add('out_' + file);
            log(`📤 ${agentId}/outbox/${file} (new output)`);
        }
    }
}

async function main() {
    log('=== Agent Bridge v2.0 ===');
    log(`Agents: ${Object.keys(AGENT_TARGETS).join(', ')}`);
    log(`Watching: ${AGENTS_DIR}`);
    log(`Poll: ${POLL_MS}ms`);
    log('');

    // inbox 폴더 확보
    for (const agentId of Object.keys(AGENT_TARGETS)) {
        const inbox = join(AGENTS_DIR, agentId, 'inbox');
        const outbox = join(AGENTS_DIR, agentId, 'outbox');
        if (!existsSync(inbox)) mkdirSync(inbox, { recursive: true });
        if (!existsSync(outbox)) mkdirSync(outbox, { recursive: true });
    }

    await checkInboxes();
    setInterval(checkInboxes, POLL_MS);
    setInterval(checkOutboxes, 10000);

    log(`Polling... Ctrl+C to stop`);
}

main().catch(e => { log(`💥 ${e.message}`); process.exit(1); });
