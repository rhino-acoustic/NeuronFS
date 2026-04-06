/**
 * Broad fetch logger — captures ALL fetch URLs to see what's actually being called
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

// Broad patch — logs ALL external fetches
const BROAD_PATCH = `
(function() {
    if (globalThis.__broadPatch) return 'skip';
    globalThis.__broadPatch = true;
    
    const pid = typeof process !== 'undefined' ? process.pid : '?';
    const CAP = 'http://127.0.0.1:7777';
    
    if (!globalThis.__origFetch && globalThis.fetch) {
        globalThis.__origFetch = globalThis.fetch;
    }
    
    if (globalThis.__origFetch) {
        globalThis.fetch = async function(input, init) {
            const url = typeof input === 'string' ? input : (input && input.url) || String(input);
            // Log ALL non-local fetches
            if (!url.startsWith('http://127.0.0.1') && !url.startsWith('http://localhost') && url.length > 5) {
                const bodyStr = init && init.body ? String(init.body).substring(0, 200) : '';
                globalThis.__origFetch(CAP + '/urllog', {
                    method: 'POST',
                    body: JSON.stringify({ pid, url, method: (init && init.method) || 'GET', bodySnippet: bodyStr }),
                    headers: { 'content-type': 'application/json' }
                }).catch(() => {});
            }
            return globalThis.__origFetch(input, init);
        };
        return 'broad-patched:' + pid;
    }
    return 'no-fetch:' + pid;
})()
`;

// Add /urllog endpoint to existing server... but server is already on :7777
// We need to intercept those logs. Since fetch-patcher.mjs server is running,
// let's start a DIFFERENT port for url logging only.
const LOG_PORT = 7778;
const urlLogServer = http.createServer((req, res) => {
    if (req.method === 'POST' && req.url === '/urllog') {
        let body = '';
        req.on('data', c => body += c);
        req.on('end', () => {
            try {
                const d = JSON.parse(body);
                log(`📡 FETCH from PID ${d.pid}: ${d.url}`);
                if (d.bodySnippet) log(`   body snippet: ${d.bodySnippet}`);
            } catch(e) {}
            res.writeHead(200); res.end('ok');
        });
    } else { res.writeHead(404); res.end(); }
});
urlLogServer.listen(LOG_PORT, '127.0.0.1', () => log(`URL log server :${LOG_PORT}`));

// BROAD patch code (different endpoint)
const BROAD_PATCH2 = `
(function() {
    if (globalThis.__broadPatch2) return 'skip';
    globalThis.__broadPatch2 = true;
    const pid = typeof process !== 'undefined' ? process.pid : '?';
    const _orig = globalThis.__origFetch || globalThis.fetch;
    if (!_orig) return 'no-fetch';
    if (!globalThis.__origFetch) globalThis.__origFetch = globalThis.fetch;
    
    globalThis.fetch = async function(input, init) {
        const url = typeof input === 'string' ? input : (input && input.url) || String(input);
        if (url && !url.startsWith('http://127.')) {
            _orig('http://127.0.0.1:7778/urllog', {
                method: 'POST',
                body: JSON.stringify({ pid, url, method: (init && init.method) || 'GET', bodySnippet: String((init && init.body) || '').substring(0,200) }),
                headers: { 'content-type': 'application/json' }
            }).catch(() => {});
        }
        return _orig(input, init);
    };
    return 'broad2:' + pid;
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
    let wsUrl; try { wsUrl = await getWsUrl(port); } catch(e) { log(`${label}: ${e.message}`); return; }
    if (!wsUrl) { log(`${label}: no target`); return; }
    return new Promise(resolve => {
        const ws = new WebSocket(wsUrl);
        let done = false;
        const t = setTimeout(() => { if(!done){done=true;ws.terminate();resolve();} }, 5000);
        ws.on('open', () => ws.send(JSON.stringify({ id:1, method:'Runtime.evaluate', params:{ expression: BROAD_PATCH2, returnByValue: true }})));
        ws.on('message', raw => {
            if (done) return;
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) { 
                log(`${label}: ${msg.result?.result?.value}`);
                done=true; clearTimeout(t); ws.close(); resolve();
            }
        });
        ws.on('error', e => { if(!done){done=true;clearTimeout(t);resolve();} });
    });
}

log('=== Broad Fetch Logger ===');
await inject('ExtHost-A:6188', 6188);
await inject('ExtHost-B:6189', 6189);
await inject('NodeSvc-38320:9229', 9229);
log('Injected. Now ALL fetch calls will be logged. Send a chat message!');
