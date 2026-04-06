/**
 * Force-open Node.js inspector on Electron main process (PID 23136)
 * and inject the raw body dump patch directly into the running https module
 */
import WebSocket from 'ws';
import http from 'http';
import net from 'net';
import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
function log(msg) { const l = `[${new Date().toISOString()}] ${msg}`; console.log(l); fs.appendFileSync(HOOK_LOG, l+'\n'); }

// Read the updated v4-hook.cjs and extract just the req.end raw dump logic
// We inject this as a hot-patch to the already-running https module
const HOTPATCH = `
(function hotpatch() {
    if (globalThis.__v5hp) return 'already:v5hp';
    
    // Try all ways to get modules
    let https, fs, path, os;
    const tryGet = (mod) => {
        try { if (process.getBuiltinModule) return process.getBuiltinModule(mod); } catch(_) {}
        try { return require(mod); } catch(_) {}
        try { return process.mainModule && process.mainModule.require(mod); } catch(_) {}
        return null;
    };
    
    https = tryGet('https'); fs = tryGet('fs'); path = tryGet('path'); os = tryGet('os');
    if (!https || !fs) return 'no-modules:https=' + !!https + ',fs=' + !!fs;
    
    const INBOX = (os ? path.join(os.homedir(), 'NeuronFS', 'brain_v4', '_agents', 'global_inbox') : 
                        'C:\\\\Users\\\\BASEMENT_ADMIN\\\\NeuronFS\\\\brain_v4\\\\_agents\\\\global_inbox');
    const AI_HOSTS = ['cloudcode-pa.googleapis', 'generativelanguage', 'googleapis.com/v1internal'];
    
    // The hook already patched https.request — we need to re-patch it to add raw dump
    // The hook stored the original as https.request itself (it wraps it)
    // We wrap AGAIN on top
    if (!https.__v5hp) {
        https.__v5hp = true;
        const _prev = https.request.bind(https);
        https.request = function(...args) {
            const opts = typeof args[0] === 'object' ? args[0] : {};
            const host = opts?.hostname || opts?.host || (typeof args[0] === 'string' ? args[0] : '');
            const isAI = AI_HOSTS.some(p => host.includes(p));
            const req = _prev.apply(this, args);
            
            if (isAI && !req.__v5wrapped) {
                req.__v5wrapped = true;
                const chunks = [];
                const protoWrite = req.write.bind(req);
                const protoEnd = req.end.bind(req);
                
                req.write = function(chunk, enc, cb) {
                    if (chunk) {
                        const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
                        chunks.push(buf);
                        // Write chunk immediately to disk
                        try {
                            const ts = Date.now();
                            fs.writeFileSync(path.join(INBOX, 'raw_chunk_' + ts + '.bin'), buf);
                            fs.appendFileSync(path.join(INBOX, 'https_debug.log'),
                                '[' + new Date().toISOString() + '] V5HP_WRITE pid:' + process.pid + ' bytes:' + buf.length + ' host:' + host.substring(0,60) + '\\n');
                        } catch(e) {}
                    }
                    return protoWrite(chunk, enc, cb);
                };
                
                req.end = function(chunk) {
                    if (chunk) {
                        const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
                        chunks.push(buf);
                    }
                    if (chunks.length > 0) {
                        try {
                            const full = Buffer.concat(chunks);
                            const ts = Date.now();
                            fs.writeFileSync(path.join(INBOX, 'raw_req_' + ts + '.bin'), full);
                            fs.appendFileSync(path.join(INBOX, 'https_debug.log'),
                                '[' + new Date().toISOString() + '] V5HP_END pid:' + process.pid + ' total:' + full.length + ' host:' + host.substring(0,60) + '\\n');
                        } catch(e) {}
                    }
                    return protoEnd(chunk);
                };
                
                req.on('response', (res) => {
                    const rc = [];
                    res.on('data', c => rc.push(c));
                    res.on('end', () => {
                        try {
                            const buf = Buffer.concat(rc);
                            if (buf.length > 5) {
                                const ts = Date.now();
                                fs.writeFileSync(path.join(INBOX, 'raw_res_' + ts + '.bin'), buf);
                                fs.appendFileSync(path.join(INBOX, 'https_debug.log'),
                                    '[' + new Date().toISOString() + '] V5HP_RESP pid:' + process.pid + ' bytes:' + buf.length + '\\n');
                            }
                        } catch(e) {}
                    });
                });
            }
            return req;
        };
    }
    
    globalThis.__v5hp = true;
    try { fs.appendFileSync(path.join(INBOX, 'https_debug.log'),
        '[' + new Date().toISOString() + '] V5HP_ACTIVE pid:' + process.pid + '\\n'); } catch(_) {}
    return 'v5hp:pid' + process.pid;
})()
`;

async function findInspectorPort(pid) {
    // Signal the process
    try { process._debugProcess(pid); log(`signaled PID ${pid}`); } catch(e) { log(`signal fail: ${e.message}`); }
    await new Promise(r => setTimeout(r, 2000));
    
    // Scan for new listening ports by PID
    try {
        const out = execSync('netstat -ano', { encoding: 'utf8' });
        const lines = out.split('\n').filter(l => l.includes('LISTENING') && l.trim().endsWith(String(pid)));
        log(`Ports for PID ${pid}: ${lines.map(l => l.trim().split(/\s+/)[1]).join(', ')}`);
        for (const line of lines) {
            const parts = line.trim().split(/\s+/);
            const addr = parts[1]; // 127.0.0.1:PORT
            const port = parseInt(addr.split(':').pop());
            if (port !== 9000 && port !== 5143) {
                return port; // Found new inspector port
            }
        }
    } catch(e) {}
    return null;
}

async function tryInspect(port, label) {
    let targets;
    try {
        targets = await new Promise((resolve, reject) => {
            http.get({ host: '127.0.0.1', port, path: '/json' }, res => {
                let d = ''; res.on('data', c => d += c);
                res.on('end', () => { try { resolve(JSON.parse(d)); } catch(e) { reject(e); } });
            }).on('error', reject);
        });
    } catch(e) { log(`${label}: port ${port} not responding: ${e.message}`); return false; }
    
    if (!targets || targets.length === 0) { log(`${label}: no targets`); return false; }
    log(`${label}: ${targets.length} targets`);

    for (const target of targets) {
        const ws = new WebSocket(target.webSocketDebuggerUrl, { handshakeTimeout: 3000 });
        const result = await new Promise(resolve => {
            let done = false;
            const t = setTimeout(() => { if(!done){done=true;ws.terminate();resolve(null);} }, 5000);
            ws.on('open', () => ws.send(JSON.stringify({ id:1, method:'Runtime.evaluate', params: { expression: HOTPATCH, returnByValue: true } })));
            ws.on('message', raw => {
                if (done) return;
                const msg = JSON.parse(raw.toString());
                if (msg.id === 1) { done=true; clearTimeout(t); ws.close(); resolve(msg.result?.result?.value); }
            });
            ws.on('error', () => { if(!done){done=true;clearTimeout(t);resolve(null);} });
        });
        if (result) { log(`${label}: ${result}`); if (result.includes('v5hp') || result.includes('already')) return true; }
    }
    return false;
}

async function main() {
    log('=== V5 HotPatch — Main Process Inspector ===');
    
    const pid = 23136;
    const newPort = await findInspectorPort(pid);
    if (newPort) {
        log(`New inspector port: ${newPort}`);
        const ok = await tryInspect(newPort, `PID${pid}:${newPort}`);
        if (ok) { log('SUCCESS — hotpatch active in main process'); process.exit(0); }
    }
    
    // Fallback: try default inspector ports
    log('Trying default inspector ports...');
    for (const port of [9229, 5858, 9230]) {
        const ok = await tryInspect(port, `port${port}`);
        if (ok) { log(`SUCCESS on port ${port}`); process.exit(0); }
    }
    
    log('Main process inspector not reachable via standard ports.');
    log('RECOMMENDATION: Restart Antigravity to load the updated hook (raw body capture will work).');
    process.exit(1);
}

main().catch(e => log(`FATAL: ${e.message}`));
