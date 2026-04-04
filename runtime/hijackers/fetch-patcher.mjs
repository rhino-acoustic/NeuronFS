/**
 * NeuronFS Extension Host Fetch Patcher
 * 1. Starts local capture server on :7777
 * 2. Injects globalThis.fetch patch into Extension Host via V8 inspector
 * 3. Captures all AI requests without needing `require` or `fs`
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const CAPTURE_PORT = 7777;
const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
fs.mkdirSync(INBOX, { recursive: true });

function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(HOOK_LOG, line + '\n');
}

// ─── 1. Capture Server ───────────────────────────────────────────────────────
function startCaptureServer() {
    const server = http.createServer((req, res) => {
        if (req.method === 'POST' && req.url === '/capture') {
            let body = '';
            req.on('data', c => body += c);
            req.on('end', () => {
                try {
                    const data = JSON.parse(body);
                    const fname = `grpc_fetch_${Date.now()}.json`;
                    const fpath = path.join(INBOX, fname);
                    fs.writeFileSync(fpath, JSON.stringify(data, null, 2));
                    log(`🎯 CAPTURED AI REQUEST → ${fname} (${data.requestBody?.length || 0} bytes body)`);
                    log(`   URL: ${(data.url || '').substring(0, 100)}`);
                } catch (e) {
                    log(`capture parse error: ${e.message}`);
                }
                res.writeHead(200);
                res.end('ok');
            });
        } else if (req.method === 'POST' && req.url === '/ping') {
            log('probe ping from Extension Host ✅');
            res.writeHead(200);
            res.end('pong');
        } else {
            res.writeHead(404);
            res.end();
        }
    });
    server.listen(CAPTURE_PORT, '127.0.0.1', () => {
        log(`Capture server listening on :${CAPTURE_PORT}`);
    });
    return server;
}

// ─── 2. Fetch Patch Injection ─────────────────────────────────────────────────
const PATCH_CODE = `
(function() {
    if (globalThis.__neuronfs_v4) return 'already-patched';
    
    const _orig = globalThis.fetch;
    if (!_orig) return 'no-fetch';
    
    globalThis.__neuronfs_v4 = true;
    const _AI = ['cloudcode-pa', 'generativelanguage', 'anthropic', 'aiplatform'];
    
    globalThis.fetch = async function(input, init) {
        const url = typeof input === 'string' ? input : (input && input.url) || '';
        const isAI = _AI.some(p => url.includes(p));
        
        if (isAI) {
            try {
                const bodyStr = init && init.body ? String(init.body) : '';
                _orig('http://127.0.0.1:7777/capture', {
                    method: 'POST',
                    headers: { 'content-type': 'application/json' },
                    body: JSON.stringify({
                        ts: Date.now(),
                        pid: (typeof process !== 'undefined') ? process.pid : -1,
                        url: url,
                        method: (init && init.method) || 'POST',
                        requestHeaders: (init && init.headers) ? JSON.stringify(init.headers) : '',
                        requestBody: bodyStr.substring(0, 50000)
                    })
                }).catch(function() {});
            } catch(_) {}
        }
        return _orig(input, init);
    };
    
    // Send probe ping to confirm patch is live
    _orig('http://127.0.0.1:7777/ping', { method: 'POST', body: 'hello' }).catch(function() {});
    
    return 'patched:' + (typeof process !== 'undefined' ? process.pid : '?');
})()
`;

async function getInspectorUrl(port) {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port, path: '/json' }, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)[0]?.webSocketDebuggerUrl); }
                catch (e) { reject(e); }
            });
        }).on('error', reject);
    });
}

async function injectPatch(pid, port) {
    let wsUrl;
    try { wsUrl = await getInspectorUrl(port); }
    catch (e) { log(`PID ${pid}: cannot reach port ${port}: ${e.message}`); return false; }

    if (!wsUrl) { log(`PID ${pid}: no target at port ${port}`); return false; }

    return new Promise((resolve) => {
        const ws = new WebSocket(wsUrl);
        let done = false;
        const timer = setTimeout(() => { if (!done) { done = true; ws.terminate(); resolve(false); } }, 8000);

        ws.on('open', () => {
            log(`PID ${pid}: injecting fetch patch...`);
            ws.send(JSON.stringify({
                id: 1,
                method: 'Runtime.evaluate',
                params: { expression: PATCH_CODE, returnByValue: true, awaitPromise: true }
            }));
        });
        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) {
                const val = msg.result?.result?.value;
                log(`PID ${pid}: patch result → "${val}"`);
                done = true;
                clearTimeout(timer);
                ws.close();
                resolve(val && val.startsWith('patched'));
            }
        });
        ws.on('error', e => { log(`PID ${pid}: error: ${e.message}`); if (!done) { done = true; clearTimeout(timer); resolve(false); } });
    });
}

// ─── 3. Main ──────────────────────────────────────────────────────────────────
async function main() {
    log('=== NeuronFS Fetch Patcher v4 ===');

    startCaptureServer();

    // Signal inspector activation for both Extension Hosts
    const targets = [
        { pid: 39396, port: 6188 },
        { pid: 48240, port: 6189 },
    ];

    // Re-signal in case ports closed
    for (const { pid } of targets) {
        try { process._debugProcess(pid); log(`PID ${pid}: inspector signaled`); }
        catch (e) { log(`PID ${pid}: signal failed: ${e.message}`); }
    }

    await new Promise(r => setTimeout(r, 1500));

    // Inject patch into all Extension Hosts
    let patched = 0;
    for (const { pid, port } of targets) {
        const ok = await injectPatch(pid, port);
        if (ok) patched++;
    }

    log(`Injection complete: ${patched}/${targets.length} patched`);
    log('Waiting for AI requests... (send a chat message now)');

    // Keep server alive
    setInterval(() => {
        const files = fs.readdirSync(INBOX).filter(f => f.startsWith('grpc_fetch_'));
        if (files.length > 0) {
            log(`Inbox has ${files.length} captures so far`);
        }
    }, 30000);
}

main().catch(e => log(`FATAL: ${e.message}`));
