import { readFileSync, readdirSync, existsSync, statSync } from 'fs';
import { join } from 'path';
import http from 'http';
import WebSocket from 'ws';

function getJson(url) {
    return new Promise((r, j) => http.get(url, res => {
        let d = ''; res.on('data', c => d += c);
        res.on('end', () => r(JSON.parse(d)));
    }).on('error', j));
}

async function kickBot1() {
    console.log('Fetching logs...');
    let recentLogs = "로그 없음";
    const brainDir = 'C:\\\\Users\\\\BASEMENT_ADMIN\\\\.gemini\\\\antigravity\\\\brain';
    if (existsSync(brainDir)) {
        let latestFile = ''; let latestTime = 0;
        const sessions = readdirSync(brainDir);
        for(const s of sessions) {
            const op = join(brainDir, s, '.system_generated', 'logs', 'overview.txt');
            if(existsSync(op)) {
                const st = statSync(op);
                if(st.mtimeMs > latestTime) { latestTime = st.mtimeMs; latestFile = op; }
            }
        }
        if(latestFile) {
            const lines = readFileSync(latestFile, 'utf8').split('\n');
            recentLogs = lines.slice(-40).join('\n');
        }
    }

    const kickMsg = [MEMORY_OBSERVER_TASK] 아래는 최근 시스템 대화 로그의 스냅샷이다.\n---\n\\n---\n이전 대화에서 뉴런화되지 않은 중요한 아키텍처 결정(암묵적 룰, 해결책)이 발견되면, [Folder-as-Neuron] 온톨로지에 맞춰 즉시 'brain_v4' 하위에 적절한 개념어 폴더를 '새로 만들고(mkdir)', 그 내부에 '1.neuron'이라는 빈 파일을 생성(touch)하라.\n(절대 경고: '개념.1.neuron' 형태의 단일 텍스트 파일을 만들지 마라).\n새로운 아키텍처 룰이 없다면 단답하지 말고, "최근 대화 로그 스캔 완료: [로그의 핵심 내용 요약] -> 하지만 새로 폴더 뉴런으로 추출할 만한 시스템/아키텍처 규칙은 없었습니다." 라고 1~2줄로 리포트하여 네가 로그를 정상적으로 읽었음을 증명하라.;

    const list = await getJson('http://127.0.0.1:9000/json/list');
    const target = list.find(t => t.url?.includes('workbench') && t.title?.toLowerCase().includes('bot1'));
    if (!target) { console.log('bot1 탭을 찾을 수 없습니다!'); process.exit(1); }

    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise(r => ws.on('open', r));

    let id = 1; const pending = new Map();
    ws.on('message', m => { const d = JSON.parse(m); if (d.id && pending.has(d.id)) { pending.get(d.id)(d); pending.delete(d.id); } });
    const call = (method, params) => new Promise(resolve => {
        const i = id++; pending.set(i, resolve); ws.send(JSON.stringify({ id: i, method, params }));
    });

    await call('Runtime.enable', {});
    const focusScript = (() => {
        const allEls = []; const walk = n => { if(!n)return; if(n.shadowRoot)walk(n.shadowRoot); for(const c of(n.children||[]))if(c.nodeType===1){allEls.push(c);walk(c);} };
        walk(document.body);
        let ci = allEls.find(el => (el.getAttribute('aria-label')||'').toLowerCase()==='message input' && el.getAttribute('role')==='textbox');
        if(!ci) return {ok:false};
        ci.textContent=''; ci.focus(); ci.click(); return {ok:true};
    })();
    
    await call('Runtime.evaluate', { expression: focusScript, returnByValue: true });
    await new Promise(r => setTimeout(r, 200));
    await call('Input.insertText', { text: kickMsg });
    await new Promise(r => setTimeout(r, 200));
    await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13 });
    await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13 });
    
    ws.close();
    console.log('bot1에게 강제 투척 완료!');
}
kickBot1().catch(console.error);
