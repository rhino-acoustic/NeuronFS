/**
 * NeuronFS Final h2 Patcher
 * Uses process.getBuiltinModule('http2') — available in Node.js 22!
 * Patches http2.connect in all Extension Host processes.
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
fs.mkdirSync(INBOX, { recursive: true });
function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(HOOK_LOG, line + '\n');
}

// THE PATCH — uses process.getBuiltinModule (Node.js 22+)
const PATCH = `
(function() {
    const pid = typeof process !== 'undefined' ? process.pid : '?';
    const AI = ['cloudcode-pa', 'generativelanguage', 'aiplatform', 'googleapis'];
    const CAP = 'http://127.0.0.1:7777';
    
    function send(ep, data) {
        const _f = globalThis.__origFetch || globalThis.fetch;
        if (_f) _f(CAP + ep, {
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
    if (globalThis.__origFetch && !globalThis.__fetchPatched) {
        globalThis.__fetchPatched = true;
        globalThis.fetch = async function(input, init) {
            const url = typeof input === 'string' ? input : (input && input.url) || '';
            if (AI.some(p => url.includes(p))) {
                send('/capture', { pid, url, body: String((init && init.body) || '').substring(0, 50000) });
            }
            return globalThis.__origFetch(input, init);
        };
    }
    
    // Patch http2 via process.getBuiltinModule (Node.js 22+)
    let h2result = 'na';
    try {
        const h2 = process.getBuiltinModule('http2');
        if (h2 && h2.connect && !h2.__nh4) {
            h2.__nh4 = true;
            const orig = h2.connect.bind(h2);
            h2.connect = function(authority, opts, listener) {
                const url = String(authority);
                if (AI.some(p => url.includes(p))) {
                    send('/h2frame', { pid, url, ts: Date.now() });
                    // Deep intercept: wrap the session
                    const sess = orig(authority, opts, listener);
                    const origRequest = sess.request.bind(sess);
                    sess.request = function(headers, opts2) {
                        const stream = origRequest(headers, opts2);
                        const chunks = [];
                        const origWrite = stream.write ? stream.write.bind(stream) : null;
                        if (origWrite) {
                            stream.write = function(chunk) {
                                if (chunk) chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk)));
                                return origWrite(...arguments);
                            };
                            stream.on('finish', () => {
                                if (chunks.length > 0) {
                                    const buf = Buffer.concat(chunks);
                                    send('/capture', {
                                        pid, url, type: 'h2_grpc',
                                        bodyHex: buf.toString('hex').substring(0, 100000),
                                        bodyLen: buf.length
                                    });
                                }
                            });
                        }
                        return stream;
                    };
                    return sess;
                }
                return orig(authority, opts, listener);
            };
            h2result = 'patched';
        } else if (h2 && h2.__nh4) {
            h2result = 'already';
        } else {
            h2result = 'no-connect';
        }
    } catch(e) { h2result = 'err:' + e.message.substring(0, 60); }
    
    // Also patch net/tls if available
    let tlsResult = 'na';
    try {
        const tls = process.getBuiltinModule('tls');
        if (tls && tls.connect && !tls.__nh4) {
            tls.__nh4 = true;
            const origTLS = tls.connect.bind(tls);
            tls.connect = function(options, listener) {
                const host = (options && (options.host || options.servername)) || '';
                if (AI.some(p => host.includes(p))) {
                    send('/h2frame', { pid, url: host, ts: Date.now(), proto: 'tls' });
                }
                return origTLS(options, listener);
            };
            tlsResult = 'patched';
        }
    } catch(e) { tlsResult = 'err:' + e.message.substring(0,40); }
    
    send('/ping', String(pid));
    
    return JSON.stringify({ pid, fetch: !!globalThis.__fetchPatched, h2: h2result, tls: tlsResult });
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

async function inject(label, port) {
    let wsUrl;
    try { wsUrl = await getWsUrl(port); } catch(e) { log(`${label}: port ${port} fail: ${e.message}`); return; }
    if (!wsUrl) { log(`${label}: no target at ${port}`); return; }
    log(`${label}: injecting at port ${port}...`);

    return new Promise(resolve => {
        const ws = new WebSocket(wsUrl);
        let done = false;
        const timer = setTimeout(() => { if(!done){done=true;ws.terminate();log(`${label}: timeout`);resolve();} }, 10000);
        ws.on('open', () => ws.send(JSON.stringify({ id:1, method:'Runtime.evaluate', params:{ expression: PATCH, returnByValue: true }})));
        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) {
                const val = msg.result?.result?.value || msg.result?.result?.description;
                log(`${label}: ${val}`);
                done = true; clearTimeout(timer); ws.close(); resolve();
            }
        });
        ws.on('error', e => { log(`${label}: ws err: ${e.message}`); if(!done){done=true;clearTimeout(timer);resolve();} });
    });
}

async function main() {
    log('=== NeuronFS getBuiltinModule Patcher ===');
    
    // Signal remaining PIDs
    for (const pid of [38320, 43048, 34812, 41784]) {
        try { process._debugProcess(pid); log(`signaled PID ${pid}`); } catch(e) {}
    }
    await new Promise(r => setTimeout(r, 1000));

    // Target: all known inspector ports
    const targets = [
        { label: 'ExtHost-A:39396', port: 6188 },
        { label: 'ExtHost-B:48240', port: 6189 },
        { label: 'NodeSvc-38320', port: 9229 },
    ];

    for (const { label, port } of targets) {
        await inject(label, port);
    }

    log('=== All patches applied. Send a chat message NOW. ===');
    process.exit(0);
}

main().catch(e => log(`FATAL: ${e.message}`));
