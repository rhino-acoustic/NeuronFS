/**
 * NeuronFS Extension Host Network Inspector
 * Uses Node.js 22 --experimental-network-inspection via V8 inspector protocol
 * to capture ALL fetch/XHR requests from Extension Host (PIDs 39396, 48240)
 * NO CODE INJECTION NEEDED — built-in network domain
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import http from 'http';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const HOOK_LOG = path.join(INBOX, 'HOOK_LOADED.txt');
fs.mkdirSync(INBOX, { recursive: true });

const INSPECTOR_ENDPOINTS = [
    { pid: 39396, port: 6188 },
    { pid: 48240, port: 6189 },
];

const AI_PATTERNS = ['cloudcode-pa', 'generativelanguage', 'anthropic', 'cascade'];

function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(HOOK_LOG, line + '\n');
}

async function getTargetUrl(port) {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port, path: '/json' }, (res) => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try {
                    const targets = JSON.parse(d);
                    resolve(targets[0]?.webSocketDebuggerUrl);
                } catch (e) { reject(e); }
            });
        }).on('error', reject);
    });
}

async function monitorProcess({ pid, port }) {
    let wsUrl;
    try {
        wsUrl = await getTargetUrl(port);
    } catch (e) {
        log(`PID ${pid}: cannot get target URL (port ${port}): ${e.message}`);
        return;
    }

    if (!wsUrl) {
        log(`PID ${pid}: no inspector target found`);
        return;
    }

    log(`PID ${pid}: connecting to ${wsUrl}`);
    const ws = new WebSocket(wsUrl);
    let msgId = 1;
    const pendingRequests = new Map(); // requestId → {url, method}

    ws.on('open', () => {
        log(`PID ${pid}: inspector connected`);
        
        // Enable Runtime domain
        ws.send(JSON.stringify({ id: msgId++, method: 'Runtime.enable' }));
        
        // Enable Network domain (requires --experimental-network-inspection in Node 22)
        ws.send(JSON.stringify({ id: msgId++, method: 'Network.enable' }));

        // Also try to evaluate to confirm we're in the right context
        ws.send(JSON.stringify({
            id: msgId++,
            method: 'Runtime.evaluate',
            params: { expression: `'pid:' + process.pid + ' node:' + process.version`, returnByValue: true }
        }));
    });

    ws.on('message', (raw) => {
        let msg;
        try { msg = JSON.parse(raw.toString()); } catch { return; }

        // Runtime.evaluate response
        if (msg.result?.result?.value) {
            log(`PID ${pid}: context confirmed: ${msg.result.result.value}`);
        }

        // Network domain events
        const method = msg.method;
        if (!method) return;

        if (method === 'Network.requestWillBeSent') {
            const { requestId, request } = msg.params;
            const url = request?.url || '';
            const isAI = AI_PATTERNS.some(p => url.includes(p));
            
            if (isAI) {
                log(`PID ${pid}: 🎯 AI REQUEST DETECTED: ${url.substring(0, 100)}`);
                pendingRequests.set(requestId, { url, method: request.method, body: request.postData });
                
                // Write full request immediately
                const dump = {
                    ts: new Date().toISOString(),
                    pid,
                    url,
                    method: request.method,
                    headers: request.headers,
                    body: request.postData || '',
                };
                const dumpFile = path.join(INBOX, `grpc_fetch_${Date.now()}.json`);
                fs.writeFileSync(dumpFile, JSON.stringify(dump, null, 2));
                log(`PID ${pid}: dump written → ${path.basename(dumpFile)} (${dump.body.length} bytes)`);
            }
        }

        if (method === 'Network.requestServedFromCache') {
            // ignore
        }

        if (method === 'Network.responseReceived') {
            const { requestId, response } = msg.params;
            if (pendingRequests.has(requestId)) {
                const req = pendingRequests.get(requestId);
                log(`PID ${pid}: response for ${req.url.substring(0, 60)} → status:${response.status}`);
            }
        }

        if (method === 'Network.loadingFinished') {
            const { requestId } = msg.params;
            if (pendingRequests.has(requestId)) {
                // Get response body
                ws.send(JSON.stringify({
                    id: msgId++,
                    method: 'Network.getResponseBody',
                    params: { requestId }
                }));
                pendingRequests.delete(requestId);
            }
        }

        // Response body
        if (msg.result?.body) {
            const body = msg.result.body;
            if (body && body.length > 0) {
                const bodyFile = path.join(INBOX, `grpc_response_${Date.now()}.txt`);
                fs.writeFileSync(bodyFile, body);
                log(`PID ${pid}: response body captured → ${path.basename(bodyFile)} (${body.length} bytes)`);
            }
        }
    });

    ws.on('error', e => log(`PID ${pid}: WS error: ${e.message}`));
    ws.on('close', () => log(`PID ${pid}: WS closed — attempting reconnect in 5s`));
}

async function main() {
    log('=== NeuronFS Network Inspector starting ===');
    log(`Watching PIDs: ${INSPECTOR_ENDPOINTS.map(e => e.pid).join(', ')}`);
    
    for (const endpoint of INSPECTOR_ENDPOINTS) {
        monitorProcess(endpoint); // non-blocking, runs in background
    }
    
    // Keep alive
    log('Now send a message in the chat to trigger AI request...');
}

main();
