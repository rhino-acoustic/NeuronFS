/**
 * NeuronFS v5 — CDP Network Sniffer
 * Connects to Chromium CDP port 9000, enables Network domain on ALL targets,
 * captures postData (plaintext XML/JSON context envelope) from AI API calls.
 * 
 * This is the "memory-level" approach: we tap into the browser's internal
 * network stack BEFORE encryption, extracting the full request body.
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const CDP_PORT = 9000;
const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const DUMP_DIR = path.join(INBOX, 'cdp_captures');
const OUTPUT = path.join(INBOX, 'latest_hijacked_context.md');
fs.mkdirSync(DUMP_DIR, { recursive: true });

const AI_PATTERNS = ['cloudcode-pa', 'generativelanguage', 'anthropic', 'cascade', 'aiplatform'];
let captureCount = 0;

function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(path.join(INBOX, 'cdp_sniffer.log'), line + '\n');
}

async function getTargets() {
    return new Promise((resolve, reject) => {
        http.get({ host: '127.0.0.1', port: CDP_PORT, path: '/json' }, (res) => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)); } catch(e) { reject(e); }
            });
        }).on('error', reject);
    });
}

function monitorTarget(target) {
    const label = target.title || target.type || target.id;
    const wsUrl = target.webSocketDebuggerUrl;
    if (!wsUrl) { log(`[${label}] No WS URL, skipping`); return; }

    log(`[${label}] Connecting: ${wsUrl}`);
    const ws = new WebSocket(wsUrl);
    let msgId = 1;
    const pendingBodies = new Map(); // requestId -> url

    ws.on('open', () => {
        log(`[${label}] Connected! Enabling Network domain...`);
        // Enable Network with large buffer
        ws.send(JSON.stringify({ id: msgId++, method: 'Network.enable', params: { maxPostDataSize: 10485760 } }));
        // Also Fetch domain for request body interception
        ws.send(JSON.stringify({ id: msgId++, method: 'Fetch.enable', params: { 
            patterns: [{ urlPattern: '*cloudcode*' }, { urlPattern: '*generativelanguage*' }, { urlPattern: '*anthropic*' }, { urlPattern: '*cascade*' }],
            handleAuthRequests: false 
        }}));
    });

    ws.on('message', (raw) => {
        let msg;
        try { msg = JSON.parse(raw.toString()); } catch { return; }

        // Check for errors from enabling domains
        if (msg.id && msg.error) {
            log(`[${label}] Domain error: ${JSON.stringify(msg.error)}`);
        }
        if (msg.id && msg.result !== undefined && !msg.error) {
            log(`[${label}] Domain enabled (id:${msg.id})`);
        }

        const method = msg.method;
        if (!method) return;

        // Network.requestWillBeSent — main capture point
        if (method === 'Network.requestWillBeSent') {
            const { requestId, request } = msg.params;
            const url = request?.url || '';
            const isAI = AI_PATTERNS.some(p => url.includes(p));
            
            if (isAI) {
                captureCount++;
                log(`[${label}] 🎯 AI REQUEST #${captureCount}: ${request.method} ${url.substring(0, 120)}`);
                
                const postData = request.postData || '';
                if (postData.length > 0) {
                    // JACKPOT — we have the full plaintext body
                    const dumpFile = path.join(DUMP_DIR, `context_${Date.now()}_${captureCount}.json`);
                    fs.writeFileSync(dumpFile, postData, 'utf8');
                    log(`[${label}] 💎 CAPTURED ${postData.length} bytes → ${path.basename(dumpFile)}`);

                    // Write structured markdown to global inbox
                    const preview = postData.substring(0, 80000);
                    fs.writeFileSync(OUTPUT, [
                        `# Hijacked Context — ${new Date().toISOString()}`,
                        ``,
                        `**Source:** ${label}`,
                        `**URL:** ${url}`,
                        `**Size:** ${postData.length} bytes`,
                        ``,
                        '```json',
                        preview,
                        '```',
                        ''
                    ].join('\n'), 'utf8');
                    log(`[${label}] ✅ latest_hijacked_context.md updated`);
                } else {
                    // postData might come separately; track this request
                    pendingBodies.set(requestId, url);
                    log(`[${label}] ⏳ No inline postData, tracking requestId: ${requestId}`);
                    
                    // Try to get body via Network.getRequestPostData
                    ws.send(JSON.stringify({
                        id: msgId++,
                        method: 'Network.getRequestPostData',
                        params: { requestId }
                    }));
                }
            }
        }

        // Fetch.requestPaused — alternative capture via Fetch domain
        if (method === 'Fetch.requestPaused') {
            const { requestId, request } = msg.params;
            const url = request?.url || '';
            const isAI = AI_PATTERNS.some(p => url.includes(p));
            
            if (isAI && request.postData) {
                captureCount++;
                const dumpFile = path.join(DUMP_DIR, `fetch_${Date.now()}_${captureCount}.json`);
                fs.writeFileSync(dumpFile, request.postData, 'utf8');
                log(`[${label}] 🎯 FETCH CAPTURED ${request.postData.length} bytes → ${path.basename(dumpFile)}`);
            }
            
            // Continue the request (don't block it)
            ws.send(JSON.stringify({
                id: msgId++,
                method: 'Fetch.continueRequest',
                params: { requestId }
            }));
        }

        // Response to Network.getRequestPostData
        if (msg.id && msg.result?.postData) {
            captureCount++;
            const body = msg.result.postData;
            const dumpFile = path.join(DUMP_DIR, `deferred_${Date.now()}_${captureCount}.json`);
            fs.writeFileSync(dumpFile, body, 'utf8');
            log(`[${label}] 💎 DEFERRED CAPTURE ${body.length} bytes → ${path.basename(dumpFile)}`);
            
            fs.writeFileSync(OUTPUT, [
                `# Hijacked Context — ${new Date().toISOString()}`,
                ``,
                `**Source:** ${label} (deferred)`,
                `**Size:** ${body.length} bytes`,
                ``,
                '```json',
                body.substring(0, 80000),
                '```',
                ''
            ].join('\n'), 'utf8');
        }

        // Network.responseReceived — log AI responses
        if (method === 'Network.responseReceived') {
            const { requestId, response } = msg.params;
            const url = response?.url || '';
            if (AI_PATTERNS.some(p => url.includes(p))) {
                log(`[${label}] 📥 AI RESPONSE: status:${response.status} ${url.substring(0, 80)}`);
            }
        }
    });

    ws.on('error', (e) => {
        log(`[${label}] WS error: ${e.message}`);
    });

    ws.on('close', () => {
        log(`[${label}] WS closed. Reconnecting in 5s...`);
        setTimeout(() => monitorTarget(target), 5000);
    });
}

async function main() {
    log('');
    log('======================================================');
    log('📡 NeuronFS v5 — CDP Network Sniffer');
    log('======================================================');
    log(`CDP Port: ${CDP_PORT}`);
    log(`Dump Dir: ${DUMP_DIR}`);
    log(`Watching: ${AI_PATTERNS.join(', ')}`);
    log('======================================================');
    log('');

    const targets = await getTargets();
    log(`Found ${targets.length} CDP targets`);

    for (const t of targets) {
        log(`  → [${t.type}] ${t.title || '(worker)'} id:${t.id.substring(0,8)}`);
        monitorTarget(t);
    }

    log('');
    log('🔥 Monitoring active. Send a chat message to capture context!');
    log('');
}

main().catch(e => { log(`FATAL: ${e.message}`); process.exit(1); });
