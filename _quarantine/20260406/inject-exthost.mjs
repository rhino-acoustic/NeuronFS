import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';

// Extension Host inspector endpoints (discovered via process._debugProcess)
const TARGETS = [
    { pid: 39396, ws: 'ws://127.0.0.1:6188/42ade883-408b-45b4-8e55-ce0fc9964f49' },
    { pid: 48240, ws: 'ws://127.0.0.1:6189/25c1e763-fa7a-4ac5-93a6-b974fa50f68d' },
];

async function injectIntoExtHost(target) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(target.ws);
        let msgId = 1;
        
        ws.on('open', () => {
            console.log(`[Inject] Connected to PID ${target.pid}`);
            
            // First: check environment
            ws.send(JSON.stringify({
                id: msgId++,
                method: 'Runtime.evaluate',
                params: {
                    expression: `JSON.stringify({
                        pid: typeof process !== 'undefined' ? process.pid : null,
                        hasRequire: typeof require !== 'undefined',
                        hasFetch: typeof fetch !== 'undefined',
                        nodeVersion: typeof process !== 'undefined' ? process.version : null,
                        modules: typeof require !== 'undefined' ? Object.keys(require.cache || {}).filter(k => k.includes('cloudcode')).slice(0,5) : []
                    })`,
                    returnByValue: true
                }
            }));
        });
        
        ws.on('message', (data) => {
            const msg = JSON.parse(data.toString());
            if (msg.result && msg.result.result) {
                try {
                    const val = JSON.parse(msg.result.result.value || '{}');
                    console.log(`[PID ${target.pid}] env:`, val);
                    
                    if (val.pid === target.pid) {
                        console.log(`[PID ${target.pid}] ✅ Confirmed! Injecting fetch interceptor...`);
                        
                        // Inject fetch interceptor
                        const code = `
(function() {
    try {
        const _fs = require('fs');
        const _path = require('path');
        const _inbox = ${JSON.stringify(INBOX)};
        _fs.mkdirSync(_inbox, { recursive: true });
        
        // Patch global fetch
        const _origFetch = globalThis.fetch;
        if (_origFetch && !globalThis.__neuronfs_patched) {
            globalThis.__neuronfs_patched = true;
            globalThis.fetch = async function(url, opts) {
                const urlStr = typeof url === 'string' ? url : (url && url.url) || String(url);
                const isAI = urlStr.includes('cloudcode-pa') || urlStr.includes('generativelanguage') || urlStr.includes('anthropic');
                
                if (isAI) {
                    try {
                        const body = opts && opts.body ? String(opts.body) : '';
                        if (body && body.length > 10) {
                            _fs.writeFileSync(
                                _path.join(_inbox, 'grpc_fetch_' + Date.now() + '.txt'),
                                'URL: ' + urlStr + '\\n\\n' + body
                            );
                            _fs.appendFileSync(_path.join(_inbox, 'HOOK_LOADED.txt'),
                                '[' + new Date().toISOString() + '] FETCH INTERCEPTED pid:' + process.pid + ' url:' + urlStr.slice(0,80) + '\\n');
                        }
                    } catch(_e) {}
                }
                return _origFetch(url, opts);
            };
            _fs.appendFileSync(_path.join(_inbox, 'HOOK_LOADED.txt'),
                '[' + new Date().toISOString() + '] fetch patch installed PID:' + process.pid + '\\n');
        }
        
        // Also try XMLHttpRequest
        if (typeof XMLHttpRequest !== 'undefined' && !XMLHttpRequest.__patched) {
            XMLHttpRequest.__patched = true;
            const _XHR = XMLHttpRequest;
            // Could patch here too
        }
        
        'injected:' + process.pid + ' fetch:' + (typeof globalThis.fetch !== 'undefined') + ' patched:' + !!globalThis.__neuronfs_patched;
    } catch(e) {
        'ERROR:' + e.message;
    }
})()`;
                        
                        ws.send(JSON.stringify({
                            id: msgId++,
                            method: 'Runtime.evaluate',
                            params: { expression: code, returnByValue: true, awaitPromise: true }
                        }));
                    } else {
                        // Second response (injection result)
                        console.log(`[PID ${target.pid}] Injection result:`, msg.result.result.value);
                        ws.close();
                        resolve(val);
                    }
                } catch (e) {
                    console.log(`[PID ${target.pid}] parse error:`, e.message, 'raw:', data.toString().substring(0, 200));
                    ws.close();
                    resolve(null);
                }
            }
        });
        
        ws.on('error', e => { console.log(`[PID ${target.pid}] WS error:`, e.message); reject(e); });
        ws.on('close', () => { console.log(`[PID ${target.pid}] WS closed`); });
        setTimeout(() => { ws.terminate(); resolve(null); }, 10000);
    });
}

async function main() {
    for (const target of TARGETS) {
        try {
            await injectIntoExtHost(target);
        } catch (e) {
            console.log(`Failed for PID ${target.pid}:`, e.message);
        }
    }
    console.log('\n[Done] Check HOOK_LOADED.txt for confirmation');
}

main();
