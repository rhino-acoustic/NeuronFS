/**
 * transcript-extractor.mjs
 * CDP로 Antigravity UI에서 대화 내용을 직접 추출하여 transcript_latest.jsonl 에 저장
 * bot-heartbeat처럼 주기적으로 실행됨
 */
import http from 'http';
import WebSocket from 'ws';
import { appendFileSync, mkdirSync, existsSync, readFileSync, writeFileSync } from 'fs';
import { join } from 'path';

const CDP_PORT = 9000;
const BRAIN = process.env.NEURONFS_BRAIN || 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const INBOX = join(BRAIN, '_agents', 'global_inbox');
const TRANSCRIPT = join(INBOX, 'transcript_latest.jsonl');
const INTERVAL_MS = 10000; // 10초마다 폴링

function getJson(url) {
    return new Promise((r, j) => http.get(url, res => {
        let d = ''; res.on('data', c => d += c);
        res.on('end', () => { try { r(JSON.parse(d)); } catch(e) { j(e); } });
    }).on('error', j));
}

async function extractFromTarget(target) {
    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });

    let id = 1;
    const pending = new Map();
    ws.on('message', m => {
        const d = JSON.parse(m);
        if (d.id && pending.has(d.id)) { pending.get(d.id)(d); pending.delete(d.id); }
    });
    const call = (method, params) => new Promise((resolve, reject) => {
        const i = id++;
        const t = setTimeout(() => { pending.delete(i); reject(new Error('timeout')); }, 8000);
        pending.set(i, d => { clearTimeout(t); resolve(d); });
        ws.send(JSON.stringify({ id: i, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 200));

    // Shadow DOM 전체 순회 후 AI 응답 메시지 블록 추출
    const script = `(() => {
        function walk(root) {
            const f = [];
            const w = n => {
                if (!n) return;
                if (n.shadowRoot) w(n.shadowRoot);
                for (const c of (n.children || [])) if (c.nodeType === 1) { f.push(c); w(c); }
            };
            w(root); return f;
        }
        const all = walk(document.body);

        // AI 응답 블록: data-turn-index 또는 메시지 컨테이너 패턴
        // 가장 바깥 메시지 래퍼만 선택 (자식 중복 방지)
        const msgBlocks = all.filter(el => {
            if (el.offsetParent === null) return false;
            const cls = (el.className || '').toString();
            // Antigravity 메시지 블록 패턴
            const isMsg = (cls.includes('leading-relaxed') && cls.includes('select-text') && !el.closest('[class*="leading-relaxed"]'))
                       || el.getAttribute('data-message-author-role') !== null
                       || el.getAttribute('data-turn-index') !== null;
            return isMsg;
        });

        // 사용자 입력 블록: role=textbox 위의 최근 사용자 메시지
        const userMsgs = all.filter(el => {
            const cls = (el.className || '').toString();
            return cls.includes('user-message') || 
                   (el.getAttribute && el.getAttribute('data-role') === 'user');
        });

        const results = [];

        // AI 블록에서 텍스트 추출
        for (const el of msgBlocks) {
            const text = (el.innerText || el.textContent || '').trim();
            if (text.length > 20) results.push({ role: 'assistant', text: text.substring(0, 3000) });
        }

        // 전체 innerText에서 USER_REQUEST 태그로도 시도
        const full = document.body.innerText || '';
        return { blocks: results, totalLen: full.length, blockCount: msgBlocks.length };
    })()`;

    const res = await call('Runtime.evaluate', { expression: script, returnByValue: true });
    ws.close();
    return res?.result?.result?.value || null;
}

// 마지막으로 저장한 블록 수 트래킹 (중복 저장 방지)
const lastCounts = {};

async function tick() {
    try {
        const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
        const targets = list.filter(t => t.url?.includes('workbench') && t.type === 'page');

        if (targets.length === 0) {
            console.log('[transcript] No workbench targets found');
            return;
        }

        mkdirSync(INBOX, { recursive: true });

        for (const target of targets) {
            const name = (target.title || 'unknown').toLowerCase().split(' ')[0];
            try {
                const data = await extractFromTarget(target);
                if (!data) continue;

                const prev = lastCounts[name] || 0;
                if (data.blockCount === prev) continue; // 변화 없으면 스킵

                lastCounts[name] = data.blockCount;

                for (const block of (data.blocks || [])) {
                    if (block.text && block.text.length > 20) {
                        const entry = JSON.stringify({
                            ts: new Date().toISOString(),
                            source: name,
                            role: block.role,
                            text: block.text
                        }) + '\n';
                        appendFileSync(TRANSCRIPT, entry);
                    }
                }
                console.log(`[transcript] ${name}: ${data.blockCount} blocks saved`);
            } catch (e) {
                // 개별 타겟 실패는 무시
            }
        }
    } catch (e) {
        console.log(`[transcript] tick error: ${e.message}`);
    }
}

console.log(`[transcript-extractor] Started. Writing to: ${TRANSCRIPT}`);
console.log(`[transcript-extractor] Polling every ${INTERVAL_MS / 1000}s via CDP:${CDP_PORT}`);

tick();
setInterval(() => tick().catch(() => {}), INTERVAL_MS);
