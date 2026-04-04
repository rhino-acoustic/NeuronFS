import http from 'http';
import WebSocket from 'ws';
const CDP_PORT = 9000;

function getJson(url) {
    return new Promise((r, j) => http.get(url, res => { let d = ''; res.on('data', c => d += c); res.on('end', () => r(JSON.parse(d))); }).on('error', j));
}

async function checkBot(name) {
    const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
    const t = list.find(x => x.url?.includes('workbench') && x.title?.toLowerCase().startsWith(name));
    if (!t) { console.log(`${name}: NOT FOUND`); return; }
    const ws = new WebSocket(t.webSocketDebuggerUrl);
    await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });
    let id = 1; const pending = new Map();
    ws.on('message', m => { const d = JSON.parse(m); if (d.id && pending.has(d.id)) { pending.get(d.id)(d); pending.delete(d.id); } });
    const call = (method, params) => new Promise((resolve, reject) => {
        const i = id++; const tm = setTimeout(() => { pending.delete(i); reject('timeout'); }, 5000);
        pending.set(i, d => { clearTimeout(tm); resolve(d); }); ws.send(JSON.stringify({ id: i, method, params }));
    });
    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 300));

    const script = `(() => {
        function collectAll(root) {
            const f = [];
            const walk = n => {
                if (!n) return;
                if (n.shadowRoot) walk(n.shadowRoot);
                for (const c of (n.children || [])) if (c.nodeType === 1) { f.push(c); walk(c); }
            };
            walk(root); return f;
        }
        const all = collectAll(document.body);
        
        // Method 1: Stop/Cancel button = AI is generating
        const stopBtn = all.find(el => {
            const txt = (el.textContent || '').trim().toLowerCase();
            const aria = (el.getAttribute('aria-label') || '').toLowerCase();
            const isBtn = el.tagName === 'BUTTON' || el.getAttribute('role') === 'button';
            return isBtn && el.offsetParent !== null && (
                txt === 'stop' || txt === 'cancel' || 
                aria.includes('stop') || aria.includes('cancel generation')
            );
        });

        // Method 2: Loading dots / spinner animation
        const spinner = all.find(el => {
            const cls = (el.className || '').toString();
            return cls.includes('animate-spin') || cls.includes('animate-pulse') || cls.includes('loading');
        });

        // Method 3: Streaming text indicator (cursor blink)
        const streaming = all.find(el => {
            const cls = (el.className || '').toString();
            return cls.includes('cursor') && cls.includes('blink');
        });

        const working = !!(stopBtn || spinner || streaming);
        return { working, stop: !!stopBtn, spin: !!spinner, stream: !!streaming };
    })()`;

    const res = await call('Runtime.evaluate', { expression: script, returnByValue: true });
    const v = res?.result?.result?.value;
    const status = v?.working ? '🔄 WORKING' : '💤 IDLE';
    const details = [];
    if (v?.stop) details.push('stop-btn');
    if (v?.spin) details.push('spinner');
    if (v?.stream) details.push('streaming');
    console.log(`${name}: ${status} ${details.length ? '(' + details.join('+') + ')' : ''}`);
    ws.close();
    return v?.working;
}

(async () => {
    console.log('=== Real Bot Status ===');
    for (const b of ['bot1', 'entp', 'enfp']) {
        try { await checkBot(b); } catch (e) { console.log(`${b}: ERR ${e}`); }
    }
    process.exit(0);
})();
