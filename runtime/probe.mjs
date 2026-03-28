/**
 * probe.mjs — CDP로 IDE 채팅 출력 영역의 활동 여부를 체크
 * 
 * 출력: JSON {"active": true/false, "lastTextLength": N}
 * 원리: 2초 간격으로 채팅 출력 영역의 텍스트 길이를 비교
 *       — 변하면 active, 안 변하면 idle
 */
import http from 'http';
import WebSocket from 'ws';

const CDP_PORT = 9000;
const CHECK_INTERVAL = 2000; // 2초

function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => { let d=''; res.on('data',c=>d+=c); res.on('end',()=>resolve(JSON.parse(d))); }).on('error', reject);
    });
}

async function probe() {
    try {
        const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
        const target = list.find(t => t.url?.includes('workbench.html'));
        if (!target) { console.log(JSON.stringify({active:false,error:"no target"})); return; }

        const ws = new WebSocket(target.webSocketDebuggerUrl);
        await new Promise((resolve, reject) => { ws.on('open', resolve); ws.on('error', reject); });
        
        let id = 1;
        const pending = new Map();
        ws.on('message', msg => { const d=JSON.parse(msg); if(d.id&&pending.has(d.id)){pending.get(d.id)(d);pending.delete(d.id);} });
        const call = (m,p) => new Promise(res => { const i=id++; pending.set(i,res); ws.send(JSON.stringify({id:i,method:m,params:p})); });

        await call('Runtime.enable', {});
        await new Promise(r => setTimeout(r, 300));

        // 1차 측정: 채팅 출력 영역의 텍스트 길이
        const measureScript = `(() => {
            function collectAll(root) {
                const found = [];
                const walk = node => {
                    if (!node) return;
                    if (node.shadowRoot) walk(node.shadowRoot);
                    const children = node.children || [];
                    for (let i = 0; i < children.length; i++) { if (children[i].nodeType === 1) { found.push(children[i]); walk(children[i]); } }
                };
                walk(root);
                return found;
            }
            const allEls = collectAll(document.body);
            // 채팅 출력 영역: role="log" 또는 큰 scrollable 영역
            let chatOutput = allEls.find(el => el.getAttribute('role') === 'log');
            if (!chatOutput) {
                // 폴백: 가장 큰 textContent를 가진 scrollable 영역
                chatOutput = allEls.filter(el => el.scrollHeight > 500 && el.textContent.length > 100)
                    .sort((a,b) => b.textContent.length - a.textContent.length)[0];
            }
            return chatOutput ? chatOutput.textContent.length : -1;
        })()`;

        const r1 = await call('Runtime.evaluate', { expression: measureScript, returnByValue: true });
        const len1 = r1?.result?.result?.value ?? -1;

        // 2초 대기
        await new Promise(r => setTimeout(r, CHECK_INTERVAL));

        // 2차 측정
        const r2 = await call('Runtime.evaluate', { expression: measureScript, returnByValue: true });
        const len2 = r2?.result?.result?.value ?? -1;

        const active = len2 > len1; // 텍스트가 증가했으면 AI가 활동 중
        console.log(JSON.stringify({ active, len1, len2, delta: len2-len1 }));
        
        ws.close();
    } catch (e) {
        console.log(JSON.stringify({ active: false, error: e.message }));
    }
}
probe();
