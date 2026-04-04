import WebSocket from 'ws';
import http from 'http';

const WORKERS = [
    'C92EB0B0CDE37A67284DD9A794790DEB',
    '978E6441A4A3CF57B8C7D6F4E503B4CA',
    '97C93AD783E9670625E789C8CFCA08F1',
    'D431CEFCCD0936F1A0E99E0934D3E782',
    '324E33BA3F4EE2C6316EBD3E3ADE0DDB',
    // page targets too
    '6B29E321EBAAA7CD7DB9FC0E607D15DF',
    '292B481F46A5B1BA757B2A34CD017A65',
    'DE280005D00F3A1C1E583FF519024F66',
];

async function probe(id) {
    return new Promise((resolve) => {
        const ws = new WebSocket(`ws://127.0.0.1:9000/devtools/page/${id}`);
        const timeout = setTimeout(() => { ws.terminate(); resolve({ id, error: 'timeout' }); }, 3000);
        let msgId = 1;
        ws.on('open', () => {
            // Evaluate process.pid and require availability
            ws.send(JSON.stringify({
                id: msgId++,
                method: 'Runtime.evaluate',
                params: {
                    expression: `JSON.stringify({ pid: typeof process !== 'undefined' ? process.pid : null, hasRequire: typeof require !== 'undefined', hasFs: typeof require !== 'undefined' ? (function(){try{require('fs');return true}catch(e){return e.message}})() : false, title: typeof document !== 'undefined' ? document.title : 'worker' })`,
                    returnByValue: true
                }
            }));
        });
        ws.on('message', (data) => {
            clearTimeout(timeout);
            try {
                const msg = JSON.parse(data);
                if (msg.result && msg.result.result) {
                    const val = JSON.parse(msg.result.result.value || '{}');
                    resolve({ id, ...val });
                }
            } catch (e) {
                resolve({ id, parseError: e.message });
            }
            ws.close();
        });
        ws.on('error', (e) => { clearTimeout(timeout); resolve({ id, error: e.message }); ws.terminate(); });
    });
}

async function main() {
    // First get fresh target list
    const targets = await new Promise((resolve) => {
        http.get('http://127.0.0.1:9000/json', (res) => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => resolve(JSON.parse(d)));
        });
    });
    
    console.log('Probing all targets...');
    for (const t of targets) {
        const id = t.webSocketDebuggerUrl.split('/').pop();
        const result = await probe(id);
        console.log(`[${t.type}] id=${id.substring(0,8)} title="${t.title.substring(0,30)}" →`, JSON.stringify(result).substring(0, 150));
    }
}

main().catch(console.error);
