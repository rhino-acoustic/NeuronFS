/**
 * Inject raw body dump patch directly into Antigravity main process (PID 23136)
 * via Chromium CDP port 9000
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const CDP_PORT = 9000;
const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
fs.mkdirSync(INBOX, { recursive: true });
function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(HOOK_LOG, line + '\n');
}

// Patch to inject into Electron main process context
// Patches https.request to save raw bodies
const MAIN_PATCH = `
(function() {
    if (globalThis.__nh_main_v5) return 'already';
    globalThis.__nh_main_v5 = true;
    
    const INBOX = 'C:\\\\Users\\\\BASEMENT_ADMIN\\\\NeuronFS\\\\brain_v4\\\\_agents\\\\global_inbox';
    const AI = ['cloudcode-pa.googleapis', 'generativelanguage', 'googleapis.com/v1internal'];
    
    // Try to get https via process.getBuiltinModule or require
    let https = null;
    try {
        if (process.getBuiltinModule) https = process.getBuiltinModule('https');
    } catch(_) {}
    if (!https) {
        try { https = require('https'); } catch(_) {}
    }
    
    let fs2 = null;
    try {
        if (process.getBuiltinModule) fs2 = process.getBuiltinModule('fs');
    } catch(_) {}
    if (!fs2) {
        try { fs2 = require('fs'); } catch(_) {}
    }
    
    if (!https || !fs2) return 'no-https-or-fs';
    
    // Check if hook is already running (from NODE_OPTIONS --require)
    // If so, just ensure it saves raw bodies by patching the EXISTING hook
    // Look for the _request variable that the hook saved
    
    const _orig = https.__hookedOrigRequest || https.request;
    if (!https.__nh5_patched) {
        https.__nh5_patched = true;
        const _req = https.request;
        https.request = function(...args) {
            const opts = typeof args[0] === 'object' ? args[0] : {};
            const host = opts?.hostname || opts?.host || (typeof args[0] === 'string' ? args[0] : '');
            const isAI = AI.some(p => host.includes(p));
            
            const req = _req.apply(this, args);
            
            if (isAI) {
                const chunks = [];
                const _origWrite = req.write.bind(req);
                const _origEnd = req.end.bind(req);
                
                req.write = function(chunk, ...rest) {
                    if (chunk) chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk)));
                    return _origWrite(chunk, ...rest);
                };
                
                req.end = function(chunk) {
                    if (chunk) chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk)));
                    if (chunks.length > 0) {
                        try {
                            const buf = Buffer.concat(chunks);
                            const ts = Date.now();
                            const fname = INBOX + '\\\\raw_req_' + ts + '.bin';
                            fs2.writeFileSync(fname, buf);
                            fs2.appendFileSync(INBOX + '\\\\https_debug.log',
                                '[' + new Date().toISOString() + '] V5_RAW_BODY pid:' + process.pid + ' bytes:' + buf.length + ' host:' + host.substring(0,60) + '\\n');
                        } catch(e) {}
                    }
                    return _origEnd(chunk);
                };
                
                req.on('response', res => {
                    const rc = [];
                    res.on('data', c => rc.push(c));
                    res.on('end', () => {
                        try {
                            const buf = Buffer.concat(rc);
                            if (buf.length > 5) {
                                const ts = Date.now();
                                fs2.writeFileSync(INBOX + '\\\\raw_res_' + ts + '.bin', buf);
                                fs2.appendFileSync(INBOX + '\\\\https_debug.log',
                                    '[' + new Date().toISOString() + '] V5_RAW_RESP pid:' + process.pid + ' bytes:' + buf.length + '\\n');
                            }
                        } catch(e) {}
                    });
                });
            }
            return req;
        };
    }
    
    // Also log to confirm live
    try {
        fs2.appendFileSync(INBOX + '\\\\https_debug.log',
            '[' + new Date().toISOString() + '] V5_PATCH_ACTIVE pid:' + process.pid + ' node:' + process.version + '\\n');
    } catch(_) {}
    
    return 'v5-patched:pid' + process.pid;
})()
`;

async function getAllCDPTargets() {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port: CDP_PORT, path: '/json/list' }, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => { try { resolve(JSON.parse(d)); } catch(e) { reject(e); } });
        }).on('error', reject);
    });
}

async function injectViaCDP(target) {
    const wsUrl = target.webSocketDebuggerUrl;
    if (!wsUrl) return null;
    
    return new Promise(resolve => {
        const ws = new WebSocket(wsUrl, { handshakeTimeout: 5000 });
        let done = false;
        const timer = setTimeout(() => { if(!done){done=true;ws.terminate();resolve(null);} }, 8000);
        let msgId = 1;

        ws.on('open', () => {
            log(`Connected to: ${target.type}/${target.title?.substring(0,50)}`);
            // First enable Runtime
            ws.send(JSON.stringify({ id: msgId++, method: 'Runtime.enable' }));
            // Then evaluate
            setTimeout(() => {
                ws.send(JSON.stringify({
                    id: 99,
                    method: 'Runtime.evaluate',
                    params: { expression: MAIN_PATCH, returnByValue: true, awaitPromise: false }
                }));
            }, 500);
        });

        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 99) {
                const val = msg.result?.result?.value || msg.result?.result?.description || 
                           (msg.result?.exceptionDetails ? 'ERROR: ' + msg.result.exceptionDetails.text : 'null');
                log(`  → ${val}`);
                done = true; clearTimeout(timer); ws.close(); resolve(val);
            }
        });
        ws.on('error', e => { if(!done){done=true;clearTimeout(timer);resolve('ws-err:'+e.message);} });
    });
}

async function main() {
    log('=== CDP Main Process Patcher v5 ===');
    
    let targets;
    try { targets = await getAllCDPTargets(); }
    catch(e) { log(`Cannot reach CDP port ${CDP_PORT}: ${e.message}`); return; }
    
    log(`Found ${targets.length} CDP targets:`);
    targets.forEach((t, i) => log(`  [${i}] type=${t.type} title=${t.title?.substring(0,60)}`));
    
    // Try all targets - page + worker + node
    let success = 0;
    for (const target of targets) {
        const result = await injectViaCDP(target);
        if (result && result.includes('patched')) success++;
    }
    
    log(`Injection complete: ${success} targets patched`);
    log('Send a chat message now to trigger capture.');
    process.exit(0);
}

main().catch(e => log(`FATAL: ${e.message}`));
