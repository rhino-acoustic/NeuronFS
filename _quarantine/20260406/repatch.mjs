/**
 * Add http2.connect patch to already-running Extension Host
 * Reuses existing capture server on :7777
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(HOOK_LOG, line + '\n');
}

// Patch: add h2 + also check what ALL globals look like
const PATCH = `
(async function() {
    const pid = (typeof process !== 'undefined') ? process.pid : '?';
    const already = !!globalThis.__nh2_v4;
    
    // Reset flags to re-patch
    globalThis.__nh2_v4 = false;
    globalThis.__neuronfs_patched = false;
    
    const CAPTURE = 'http://127.0.0.1:7777';
    const AI = ['cloudcode-pa', 'generativelanguage', 'aiplatform', 'googleapis', 'cascade'];
    
    function send(ep, data) {
        const _f = globalThis.__origFetch || globalThis.fetch;
        if (_f) _f(CAPTURE + ep, {
            method: 'POST',
            body: typeof data === 'string' ? data : JSON.stringify(data),
            headers: { 'content-type': 'application/json' }
        }).catch(() => {});
    }
    
    // Save original fetch
    if (!globalThis.__origFetch && globalThis.fetch) {
        globalThis.__origFetch = globalThis.fetch;
    }
    
    // Patch fetch
    if (globalThis.__origFetch) {
        globalThis.__neuronfs_patched = true;
        globalThis.fetch = async function(input, init) {
            const url = typeof input === 'string' ? input : (input && input.url) || '';
            if (AI.some(p => url.includes(p))) {
                const bodyStr = init && init.body ? String(init.body) : '';
                send('/capture', { pid, url, method: (init && init.method) || 'GET', requestBody: bodyStr.substring(0, 50000) });
            }
            return globalThis.__origFetch(input, init);
        };
    }
    
    // Try http2 dynamic import
    let h2result = 'skipped';
    try {
        const mod = await import('node:http2');
        const m = mod.default || mod;
        if (m && m.connect) {
            if (!m.__nh2_patched) {
                m.__nh2_patched = true;
                const orig = m.connect.bind(m);
                m.connect = function(authority, opts, listener) {
                    const url = String(authority);
                    if (AI.some(p => url.includes(p))) {
                        send('/h2frame', { pid, url, ts: Date.now() });
                    }
                    return orig(authority, opts, listener);
                };
            }
            h2result = 'ok';
        }
    } catch(e) { h2result = 'err:' + e.message.substring(0, 50); }
    
    // Also try require() via process.mainModule trick
    let req2 = null;
    try {
        if (typeof __webpack_require__ !== 'undefined') {
            req2 = 'has webpack_require';
        }
    } catch(_) {}
    
    // Probe ping
    send('/ping', String(pid));
    globalThis.__nh2_v4 = true;
    
    return JSON.stringify({ pid, fetchPatched: !!globalThis.__neuronfs_patched, h2: h2result, wasPrev: already, webpack: req2 });
})()
`;

async function getWsUrl(port) {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port, path: '/json' }, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => { try { resolve(JSON.parse(d)[0]?.webSocketDebuggerUrl); } catch(e) { reject(e); } });
        }).on('error', reject);
    });
}

async function inject(pid, port, label) {
    let wsUrl;
    try { wsUrl = await getWsUrl(port); } catch(e) { log(`${label}: port ${port} unreachable: ${e.message}`); return null; }
    if (!wsUrl) { log(`${label}: no target at port ${port}`); return null; }

    return new Promise(resolve => {
        const ws = new WebSocket(wsUrl);
        let done = false;
        const timer = setTimeout(() => { if(!done){done=true;ws.terminate();resolve(null);} }, 10000);
        ws.on('open', () => {
            ws.send(JSON.stringify({ id:1, method:'Runtime.evaluate', params:{ expression: PATCH, returnByValue:true, awaitPromise:true }}));
        });
        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) {
                const val = msg.result?.result?.value;
                let parsed = {};
                try { parsed = JSON.parse(val || '{}'); } catch(_) {}
                log(`${label}: ${JSON.stringify(parsed)}`);
                done = true; clearTimeout(timer); ws.close(); resolve(parsed);
            }
        });
        ws.on('error', e => { log(`${label}: error: ${e.message}`); if(!done){done=true;clearTimeout(timer);resolve(null);} });
    });
}

async function main() {
    log('=== h2+fetch re-inject (reusing :7777) ===');
    
    const targets = [
        { pid: 39396, port: 6188, label: 'ExtHost-A' },
        { pid: 48240, port: 6189, label: 'ExtHost-B' },
    ];

    for (const { pid } of targets) {
        try { process._debugProcess(pid); } catch(e) { log(`signal ${pid} failed`); }
    }
    await new Promise(r => setTimeout(r, 1500));

    for (const target of targets) {
        await inject(target.pid, target.port, target.label);
    }

    log('Done. Now send a chat message in Antigravity AI panel.');
    process.exit(0);
}

main().catch(e => log(`FATAL: ${e.message}`));
