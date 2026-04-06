import http from 'http';
import WebSocket from 'ws';

const CDP_PORT = 9000;
const AGENT_TARGETS = ['bot1', 'entp', 'enfp'];

const message = `[pm 주입] **우선순위: 파일 시스템 우회 긴급 렌더링**
현재 타 에이전트와 소통하지 말고 당신만의 고유 룰(Role)과 논리력으로 해답을 찾아 즉각 답변하라.

[문제 3] 5x5 그리드 토러스(Torus) (0은 흰색, 1은 검은색)
(0,0), (2,2), (4,4)를 누를 때마다(누른 곳과 십자형 상하좌우 총 5칸 색 반전), 3번 누른 뒤 남은 1(검정 표면)은 총 몇 칸인가?`;

function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => resolve(JSON.parse(d)));
        }).on('error', reject);
    });
}

async function injectToAgent(agentId, text) {
    try {
        const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
        const target = list.find(t =>
            t.url?.includes('workbench.html') &&
            t.title?.toLowerCase().startsWith(agentId.toLowerCase()) &&
            t.type === 'page'
        );
        if (!target) { console.log(`[PM] X ${agentId} window not found`); return false; }

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
            const t = setTimeout(() => { pending.delete(myId); reject(new Error('timeout')); }, 5000);
            pending.set(myId, r => { clearTimeout(t); resolve(r); });
            ws.send(JSON.stringify({ id: myId, method, params }));
        });

        await call('Runtime.enable', {});
        await new Promise(r => setTimeout(r, 200));

        const focusExpr = `(() => {
            function collectAll(root) {
                const found = [];
                const walk = n => { if(!n)return; if(n.shadowRoot)walk(n.shadowRoot); if(n.matches&&n.matches('[aria-label="Message input"][role="textbox"]'))found.push(n); for(const c of(n.children||[]))walk(c); };
                walk(root); return found;
            }
            const inputs = collectAll(document);
            if (inputs.length > 0) { inputs[0].textContent=''; inputs[0].focus(); return 'ok'; }
            let ci = collectAll(document.body).find(el => el.getAttribute('contenteditable')==='true' && el.getAttribute('role')==='textbox' && el.offsetParent!==null);
            if(ci){ ci.textContent=''; ci.focus(); return 'ok'; }
            return 'not_found';
        })()`;

        const focusRes = await call('Runtime.evaluate', { expression: focusExpr });
        if (focusRes?.result?.result?.value !== 'ok') {
            console.log(`[PM] X ${agentId} chat input not found or inactive`);
            ws.close(); return false;
        }

        await new Promise(r => setTimeout(r, 200));
        await call('Input.insertText', { text: text });
        await new Promise(r => setTimeout(r, 200));
        await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13});
        await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13});

        console.log(`[PM] O ${agentId} Force-Injection Success.`);
        ws.close();
        return true;
    } catch (e) {
        console.log(`[PM] X Error on ${agentId}: ${e.message}`);
        return false;
    }
}

async function run() {
    console.log('[PM] Executing forceful injection across all bots...');
    for (const agent of AGENT_TARGETS) {
        await injectToAgent(agent, message);
        await new Promise(r => setTimeout(r, 1000));
    }
    console.log('[PM] Injection complete.');
    process.exit(0);
}
run();
