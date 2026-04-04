/**
 * NeuronFS http2 Patcher via Dynamic Import
 * Extension Host processes don't expose `require` globally, but `import()` works.
 * We patch http2.connect in the Extension Host via V8 inspector.
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

// ─── Capture Server ────────────────────────────────────────────────────────────
const server = http.createServer((req, res) => {
    if (req.method === 'POST') {
        let body = '';
        req.on('data', c => body += c);
        req.on('end', () => {
            if (req.url === '/h2frame') {
                try {
                    const data = JSON.parse(body);
                    const fname = `grpc_h2_${Date.now()}.json`;
                    fs.writeFileSync(path.join(INBOX, fname), JSON.stringify(data, null, 2));
                    log(`🎯 H2 CONNECT INTERCEPTED → ${fname} target:${data.url}`);
                } catch(e) { log(`h2 parse error: ${e.message}`); }
            } else if (req.url === '/ping') {
                log(`ping from Extension Host PID: ${body}`);
            } else if (req.url === '/capture') {
                try {
                    const data = JSON.parse(body);
                    const fname = `grpc_fetch_${Date.now()}.json`;
                    fs.writeFileSync(path.join(INBOX, fname), JSON.stringify(data, null, 2));
                    log(`🎯 FETCH CAPTURED → ${fname} url:${(data.url||'').substring(0,80)}`);
                } catch(e) { log(`fetch parse error: ${e.message}`); }
            }
            res.writeHead(200); res.end('ok');
        });
    } else { res.writeHead(404); res.end(); }
});
server.listen(CAPTURE_PORT, '127.0.0.1', () => log(`Capture server :${CAPTURE_PORT}`));

// ─── Patch Code (runs in Extension Host via inspector) ─────────────────────────
const H2_PATCH = `
(async function() {
    if (globalThis.__nh2_v4) return 'already:h2+fetch';
    globalThis.__nh2_v4 = true;
    
    const pid = (typeof process !== 'undefined') ? process.pid : '?';
    const CAPTURE = 'http://127.0.0.1:${CAPTURE_PORT}';
    const AI = ['cloudcode-pa', 'generativelanguage', 'aiplatform', 'googleapis'];
    
    function sendCapture(endpoint, payload) {
        try {
            const _f = globalThis.fetch || globalThis.__origFetch;
            if (_f) _f(CAPTURE + endpoint, {
                method: 'POST',
                body: JSON.stringify(payload),
                headers: { 'content-type': 'application/json' }
            }).catch(()=>{});
        } catch(_){}
    }
    
    // 1. Patch globalThis.fetch (Chromium fetch)
    if (globalThis.fetch && !globalThis.__origFetch) {
        globalThis.__origFetch = globalThis.fetch;
        globalThis.fetch = async function(input, init) {
            const url = typeof input === 'string' ? input : (input && input.url) || '';
            if (AI.some(p => url.includes(p))) {
                const body = init && init.body ? String(init.body) : '';
                sendCapture('/capture', { pid, url, method: init && init.method, requestBody: body.substring(0, 50000) });
            }
            return globalThis.__origFetch(input, init);
        };
    }
    
    // 2. Patch http2 via dynamic import
    let h2patched = false;
    try {
        const h2 = await import('node:http2');
        const m = h2.default || h2;
        if (m && m.connect && !m.__nh2_patched) {
            m.__nh2_patched = true;
            const origConnect = m.connect;
            m.connect = function(url, opts, listener) {
                const urlStr = String(url);
                if (AI.some(p => urlStr.includes(p))) {
                    sendCapture('/h2frame', { pid, url: urlStr, ts: Date.now() });
                    // Wrap the session to intercept data
                    const session = origConnect.call(this, url, opts, listener);
                    const origReq = session.request.bind(session);
                    session.request = function(headers, opts2) {
                        const stream = origReq(headers, opts2);
                        const chunks = [];
                        const origWrite = stream.write.bind(stream);
                        stream.write = function(chunk, ...args) {
                            if (chunk) chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                            return origWrite(chunk, ...args);
                        };
                        stream.on('end', () => {
                            if (chunks.length > 0) {
                                const raw = Buffer.concat(chunks).toString('base64');
                                sendCapture('/capture', { pid, url: urlStr, type: 'h2_stream', bodyB64: raw.substring(0, 70000) });
                            }
                        });
                        return stream;
                    };
                    return session;
                }
                return origConnect.call(this, url, opts, listener);
            };
            h2patched = true;
        }
    } catch(e) {
        // http2 import failed - OK if not available
    }
    
    // 3. Probe ping
    sendCapture('/ping', String(pid));
    
    return 'patched:fetch+' + (h2patched ? 'h2' : 'noh2') + ':pid' + pid;
})()
`;

// ─── Inspector Injection ────────────────────────────────────────────────────────
async function getWsUrl(port) {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port, path: '/json' }, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)[0]?.webSocketDebuggerUrl); } catch(e) { reject(e); }
            });
        }).on('error', reject);
    });
}

async function inject(pid, port, label) {
    let wsUrl;
    try { wsUrl = await getWsUrl(port); } catch(e) { log(`${label}: port ${port} unreachable`); return; }
    if (!wsUrl) { log(`${label}: no target`); return; }

    return new Promise(resolve => {
        const ws = new WebSocket(wsUrl);
        let done = false;
        const timer = setTimeout(() => { if(!done) { done=true; ws.terminate(); resolve('timeout'); } }, 10000);
        
        ws.on('open', () => {
            log(`${label}: connected, injecting h2+fetch patch...`);
            ws.send(JSON.stringify({ id: 1, method: 'Runtime.evaluate', params: {
                expression: H2_PATCH, returnByValue: true, awaitPromise: true
            }}));
        });
        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) {
                const val = msg.result?.result?.value || msg.result?.result?.description || 'undefined';
                log(`${label}: result → "${val}"`);
                done = true; clearTimeout(timer); ws.close(); resolve(val);
            }
        });
        ws.on('error', e => { log(`${label}: error: ${e.message}`); if(!done){done=true;clearTimeout(timer);resolve('error');} });
    });
}

async function main() {
    log('=== NeuronFS h2+fetch Patcher v5 ===');
    
    const targets = [
        { pid: 39396, port: 6188, label: 'ExtHost-A' },
        { pid: 48240, port: 6189, label: 'ExtHost-B' },
    ];

    // Signal inspectors
    for (const { pid, label } of targets) {
        try { process._debugProcess(pid); log(`${label} (PID ${pid}): inspector signaled`); }
        catch(e) { log(`${label}: signal failed: ${e.message}`); }
    }
    await new Promise(r => setTimeout(r, 1500));

    // Inject
    for (const { pid, port, label } of targets) {
        await inject(pid, port, label);
    }

    log('Injection complete. Waiting for AI requests...');
    log('─────────────────────────────────────────────');
}

main().catch(e => log(`FATAL: ${e.message}`));
