// NeuronFS — Live Context Injection Hook
//
// Injects your current brain state into every LLM API request,
// so the AI always sees your latest neuron rules mid-conversation.
//
// How it works:
//   1. Scans your brain_v4/ folder for neurons with counter >= 5
//   2. Formats them as compact Path=Sentence text
//   3. Appends to the system prompt of outgoing Gemini/Claude/OpenAI requests
//   4. Optionally logs tool-call responses for downstream processing
//
// Setup:
//   set NEURONFS_BRAIN=C:\path\to\your\brain_v4
//   set NODE_OPTIONS=--require "C:\path\to\NeuronFS\runtime\v4-hook.cjs"
//   # Then start your IDE (VS Code, Cursor, Windsurf, etc.)
//
// Works with any Electron-based AI IDE that uses Node.js https module.
// No dependencies. No MCP needed.

const fs = require('fs');
const https = require('https');
const http2 = require('http2');
const path = require('path');

// ─── HTTP/2 gRPC Stream Interceptor ───
// Antigravity uses gRPC over HTTP/2 for AI chat calls (cloudcode-pa.googleapis.com)
// We patch http2.connect to intercept the raw DATA frames
try {
    // Resolve INBOX_DIR inline so this block is self-contained (defined before INBOX_DIR below)
    const _H2_BRAIN = process.env.NEURONFS_BRAIN ||
        path.join(require('os').homedir(), 'NeuronFS', 'brain_v4');
    const _H2_INBOX = path.join(_H2_BRAIN, '_agents', 'global_inbox');
    fs.mkdirSync(_H2_INBOX, { recursive: true });

    const _h2connect = http2.connect.bind(http2);
    http2.connect = function(authority, options, listener) {
        const session = _h2connect(authority, options, listener);
        const authorityStr = (typeof authority === 'string' ? authority :
            (authority && authority.href) ? authority.href : String(authority || ''));
        const isAI = authorityStr.includes('cloudcode-pa.googleapis.com') ||
                     authorityStr.includes('generativelanguage') ||
                     authorityStr.includes('anthropic');

        if (!isAI) return session;

        // Log that this connection is being monitored
        try {
            fs.appendFileSync(path.join(_H2_INBOX, 'HOOK_LOADED.txt'),
                `[${new Date().toISOString()}] http2.connect interceptor active PID:${process.pid} authority:${authorityStr.slice(0,80)}\n`);
        } catch (_) {}

        // Patch session.request to intercept each gRPC stream
        const _req = session.request.bind(session);
        session.request = function(headers, opts) {
            const stream = _req(headers, opts);
            const reqPath = (headers && headers[':path']) ? headers[':path'] : '';

            // Capture ALL cascade paths (chat API)
            const isChat = reqPath.includes('cascade') || reqPath.includes('generateContent') ||
                           reqPath.includes('streamGenerateContent') || reqPath.includes('chat') ||
                           reqPath === '' || reqPath.includes('v1');

            if (isChat) {
                const _write = stream.write.bind(stream);
                stream.write = function(data, enc, cb) {
                    if (data && data.length > 5) {
                        const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
                        try {
                            // Dump raw bytes (first 5 = gRPC frame header, rest = protobuf)
                            const rawDump = path.join(_H2_INBOX, `grpc_raw_${Date.now()}.bin`);
                            fs.writeFileSync(rawDump, buf);

                            // Also try UTF-8 decode (protobuf strings are UTF-8)
                            const payload = buf.slice(5).toString('utf-8');
                            const hasContext = payload.includes('USER_REQUEST') ||
                                payload.includes('system_instruction') ||
                                payload.includes('user_rules') ||
                                payload.includes('contents') ||
                                payload.includes('ADDITIONAL_METADATA');

                            if (hasContext) {
                                const txtDump = path.join(_H2_INBOX, `grpc_context_${Date.now()}.txt`);
                                fs.writeFileSync(txtDump, payload);
                                fs.appendFileSync(path.join(_H2_INBOX, 'HOOK_LOADED.txt'),
                                    `[${new Date().toISOString()}] CONTEXT CAPTURED! ${payload.length} bytes path:${reqPath}\n`);
                            }
                        } catch (_wErr) {}
                    }
                    return _write(data, enc, cb);
                };

                // Also intercept response data
                stream.on('data', function(chunk) {
                    if (chunk && chunk.length > 5) {
                        try {
                            const payload = chunk.slice(5).toString('utf-8');
                            if (payload.includes('text') && payload.length > 20) {
                                const respDump = path.join(_H2_INBOX, `grpc_response_${Date.now()}.txt`);
                                fs.writeFileSync(respDump, payload);
                            }
                        } catch (_) {}
                    }
                });
            }
            return stream;
        };

        return session;
    };
} catch (_h2err) {
    try {
        const _errPath = path.join(
            process.env.NEURONFS_BRAIN || path.join(require('os').homedir(), 'NeuronFS', 'brain_v4'),
            '_agents', 'global_inbox', 'HOOK_LOADED.txt');
        fs.appendFileSync(_errPath, `[${new Date().toISOString()}] http2 patch ERROR: ${_h2err.message}\n`);
    } catch (_) {}
}



try {
    const fPath = path.join(process.env.NEURONFS_BRAIN || __dirname, '_agents', 'global_inbox', 'HOOK_LOADED.txt');
    fs.mkdirSync(path.dirname(fPath), { recursive: true });
    fs.appendFileSync(fPath, `[${new Date().toISOString()}] Hook loaded into PID: ${process.pid}, Title: ${process.title}, Args: ${process.argv.join(' ')}\n`);
} catch (e) {}

// ─── Configuration ───
// Set NEURONFS_BRAIN env var, or defaults to ~/NeuronFS/brain_v4
const BRAIN_PATH = process.env.NEURONFS_BRAIN
    || path.join(require('os').homedir(), 'NeuronFS', 'brain_v4');
const INBOX_DIR = path.join(BRAIN_PATH, '_agents', 'global_inbox');
const REGIONS = ['brainstem','limbic','hippocampus','sensors','cortex','ego','prefrontal'];
const PROMOTE_THRESHOLD = 5;  // only inject neurons with counter >= this
const CACHE_TTL_MS = 30000;   // re-scan brain every 30 seconds

try { if (!fs.existsSync(INBOX_DIR)) fs.mkdirSync(INBOX_DIR, { recursive: true }); } catch {}

// ─── Brain Scanner ───
// Reads the filesystem tree and builds a compact rule string.
// Only promoted neurons (counter >= threshold) are included.
// Output follows Path=Sentence format, ordered by Subsumption priority (P0→P6).
let cachedRules = '';
let cacheTime = 0;

function scanBrain() {
    const result = {};
    REGIONS.forEach(r => result[r] = []);

    function walk(dir, depth) {
        let entries;
        try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch { return; }
        for (const e of entries) {
            if (e.name.startsWith('_') || e.name.startsWith('.')) continue;
            const full = path.join(dir, e.name);
            if (e.isDirectory()) {
                if (depth < 6) walk(full, depth + 1);
            } else if (e.name.endsWith('.neuron')) {
                const counter = parseInt(e.name, 10);
                if (isNaN(counter) || counter < PROMOTE_THRESHOLD) continue;
                const rel = path.dirname(full).replace(BRAIN_PATH + path.sep, '').replace(/\\/g, '>');
                const region = rel.split('>')[0];
                if (result[region]) {
                    result[region].push({ path: rel, counter });
                }
            }
        }
    }
    walk(BRAIN_PATH, 0);

    const labels = {
        brainstem: 'P0', limbic: 'P1', hippocampus: 'P2',
        sensors: 'P3', cortex: 'P4', ego: 'P5', prefrontal: 'P6'
    };
    const lines = ['[NeuronFS Live Context]'];
    for (const r of REGIONS) {
        const neurons = result[r].sort((a, b) => b.counter - a.counter);
        if (neurons.length === 0) continue;
        const items = neurons.map(n => {
            const sentence = n.path.replace(r + '>', '').replace(/>/g, ' > ').replace(/_/g, ' ');
            return `${sentence} (${n.counter})`;
        });
        lines.push(`[${labels[r]}] ${items.join(' | ')}`);
    }

    // Self-growth protocol — tells the AI HOW to record corrections
    const inboxPath = path.join(BRAIN_PATH, '_inbox', 'corrections.jsonl');
    lines.push('');
    lines.push('[Growth Protocol]');
    lines.push(`When user corrects a mistake, immediately append to ${inboxPath}:`);
    lines.push('{"type":"correction","path":"cortex/[category]/[rule_name]","text":"reason","counter_add":1}');
    lines.push('When user praises, append: {"type":"correction","path":"[existing_neuron_path]","text":"praise","counter_add":1}');
    lines.push('Same mistake 3x → create bomb.neuron in that neuron folder.');

    return lines.join('\n');
}

function getRules() {
    const now = Date.now();
    if (now - cacheTime < CACHE_TTL_MS && cachedRules) return cachedRules;
    try {
        cachedRules = scanBrain();
        cacheTime = now;
    } catch {}
    return cachedRules;
}

getRules();

// ─── API Endpoint Detection ───
function isLLM(opts) {
    const h = opts?.hostname || opts?.host || (typeof opts === 'string' ? opts : '') || '';
    return h.includes('generativelanguage.googleapis.com') ||
           h.includes('gemini.googleapis.com') ||
           h.includes('cloudcode-pa.googleapis.com') ||
           h.includes('googleapis.com/v1internal') ||
           h.includes('api.anthropic.com') ||
           h.includes('api.openai.com');
}

function isLLMUrl(url) {
    return url.includes('generativelanguage.googleapis.com') ||
           url.includes('gemini.googleapis.com') ||
           url.includes('cloudcode-pa.googleapis.com') ||
           url.includes('googleapis.com/v1internal') ||
           url.includes('anthropic.com') ||
           url.includes('openai.com');
}

// ─── Injection: append rules to system prompt ───
function inject(bodyStr) {
    const rules = getRules();
    if (!rules) return null;

    try {
        const j = JSON.parse(bodyStr);

        // Gemini
        const si = j.systemInstruction || j.system_instruction;
        if (si?.parts?.length > 0) {
            const last = si.parts[si.parts.length - 1];
            if (last.text && !last.text.includes('[NeuronFS')) {
                last.text += '\n\n' + rules;
                return JSON.stringify(j);
            }
            return null;
        }

        // Claude
        if (j.system !== undefined) {
            if (typeof j.system === 'string' && !j.system.includes('[NeuronFS')) {
                j.system += '\n\n' + rules;
                return JSON.stringify(j);
            }
            if (Array.isArray(j.system)) {
                const last = j.system[j.system.length - 1];
                if (last?.text && !last.text.includes('[NeuronFS')) {
                    last.text += '\n\n' + rules;
                    return JSON.stringify(j);
                }
            }
            return null;
        }

        // OpenAI
        if (j.messages?.[0]?.role === 'system' && !j.messages[0].content.includes('[NeuronFS')) {
            j.messages[0].content += '\n\n' + rules;
            return JSON.stringify(j);
        }
    } catch {}
    return null;
}

// ─── HTTPS Hook ───
const _request = https.request;

https.request = function(...args) {
    const opts = typeof args[0] === 'object' ? args[0] : {};
    const host = opts?.hostname || opts?.host || (typeof args[0] === 'string' ? args[0] : '');

    // Debug: log ALL outgoing https requests to see real hostname
    try {
        const debugPath = path.join(INBOX_DIR, 'https_debug.log');
        fs.appendFileSync(debugPath, `[${new Date().toISOString()}] PID:${process.pid} HOST:${host} PATH:${opts?.path||''}\n`);
    } catch {}

    if (!isLLM(opts)) {
        return _request.apply(this, args);
    }

    const chunks = [];
    const req = _request.apply(this, args);
    const _write = req.write.bind(req);
    const _end = req.end.bind(req);

    // Buffer writes instead of sending immediately
    req.write = function(chunk, enc, cb2) {
        if (chunk) chunks.push(Buffer.from(chunk));
        const callback = typeof enc === 'function' ? enc : cb2;
        if (typeof callback === 'function') callback();
        return true;
    };

    // On end: inject rules, then send the complete body
    req.end = function(chunk) {
        if (chunk) chunks.push(Buffer.from(chunk));

        const rawBuf = Buffer.concat(chunks);
        let rawStr = '';
        let finalBody = rawBuf;

        // ─── Always dump raw body for AI endpoints ───
        if (rawBuf.length > 5 && isLLM(opts)) {
            try {
                const ts = Date.now();
                const rawDump = path.join(INBOX_DIR, `raw_req_${ts}.bin`);
                fs.writeFileSync(rawDump, rawBuf);
                fs.appendFileSync(path.join(INBOX_DIR, 'https_debug.log'),
                    `[${new Date().toISOString()}] RAW_BODY PID:${process.pid} bytes:${rawBuf.length} file:raw_req_${ts}.bin host:${host}\n`);
            } catch {}
        }

        // ─── Try JSON injection ───
        try {
            rawStr = rawBuf.toString('utf-8');
            const injected = inject(rawStr);
            if (injected) finalBody = Buffer.from(injected, 'utf-8');
        } catch {}

        // ─── Transcript dump: save conversation turns for MEMORY_OBSERVER ───
        try {
            const j = JSON.parse(rawStr);
            const contents = j.contents || j.messages;
            if (Array.isArray(contents) && contents.length > 0) {
                const last = contents[contents.length - 1];
                const text = last?.parts?.[0]?.text || last?.content || '';
                if (text && text.length > 20 && !text.includes('[NeuronFS')) {
                    const entry = JSON.stringify({
                        ts: new Date().toISOString(),
                        role: last.role || 'user',
                        text: text.substring(0, 2000)
                    }) + '\n';
                    const transcriptPath = path.join(BRAIN_PATH, '_agents', 'global_inbox', 'transcript_latest.jsonl');
                    fs.appendFileSync(transcriptPath, entry);
                }
            }
        } catch {}

        try { req.setHeader('content-length', finalBody.length); } catch {}
        _write(finalBody);
        _end();
    };


    // Capture ALL AI responses (streaming or complete)
    req.on('response', (res) => {
        const resChunks = [];
        res.on('data', (c) => resChunks.push(c));
        res.on('end', () => {
            try {
                const resBuf = Buffer.concat(resChunks);
                if (resBuf.length > 10 && isLLM(opts)) {
                    const ts = Date.now();
                    // Always save raw response binary
                    fs.writeFileSync(path.join(INBOX_DIR, `raw_res_${ts}.bin`), resBuf);
                    // Also try text for tool-call detection
                    const body = resBuf.toString('utf-8');
                    fs.appendFileSync(path.join(INBOX_DIR, 'https_debug.log'),
                        `[${new Date().toISOString()}] RAW_RESPONSE PID:${process.pid} bytes:${resBuf.length} file:raw_res_${ts}.bin\n`);
                    if (body.includes('functionCall') || body.includes('run_command') || body.includes('write_to_file')) {
                        fs.writeFileSync(path.join(INBOX_DIR, `intercepted_${ts}.json`), body);
                    }
                }
            } catch {}
        });
    });

    return req;
};

// ─── fetch API Hook (Node 18+) ───
if (globalThis.fetch) {
    const _fetch = globalThis.fetch;
    globalThis.fetch = async function(...args) {
        const [input, init] = args;
        const url = typeof input === 'string' ? input : input?.url || '';
        const isApi = url.includes('generativelanguage') ||
                      url.includes('anthropic') ||
                      url.includes('gemini.googleapis') ||
                      url.includes('openai.com');

        if (isApi && init?.body) {
            try {
                const bodyStr = typeof init.body === 'string' ? init.body : new TextDecoder().decode(init.body);
                const injected = inject(bodyStr);
                if (injected) init.body = injected;

                // Transcript dump
                try {
                    const j = JSON.parse(bodyStr);
                    const contents = j.contents || j.messages;
                    if (Array.isArray(contents) && contents.length > 0) {
                        const last = contents[contents.length - 1];
                        const text = last?.parts?.[0]?.text || (typeof last?.content === 'string' ? last.content : '');
                        if (text && text.length > 20 && !text.includes('[NeuronFS')) {
                            const entry = JSON.stringify({ ts: new Date().toISOString(), role: last.role || 'user', text: text.substring(0, 2000) }) + '\n';
                            const tp = path.join(BRAIN_PATH, '_agents', 'global_inbox', 'transcript_latest.jsonl');
                            fs.mkdirSync(path.dirname(tp), { recursive: true });
                            fs.appendFileSync(tp, entry);
                        }
                    }
                } catch {}
            } catch {}
        }

        const res = await _fetch.apply(this, args);

        if (isApi) {
            const clone = res.clone();
            clone.text().then(t => {
                if (t.includes('functionCall') || t.includes('run_command') || t.includes('"tool_use"')) {
                    try {
                        fs.writeFileSync(path.join(INBOX_DIR, `intercepted_fetch_${Date.now()}.json`), t);
                    } catch {}
                }
            }).catch(() => {});
        }

        return res;
    };
}

// ─── undici Interceptor (Antigravity uses undici, not https module) ───
try {
    const undici = require('undici');
    if (undici && undici.interceptors) {
        undici.interceptors.add({
            async onRequest({ request }) {
                const url = request.origin + (request.path || '');
                const isApi = url.includes('generativelanguage') ||
                              url.includes('anthropic') ||
                              url.includes('openai.com') ||
                              url.includes('gemini.googleapis');
                try {
                    const debugPath = path.join(INBOX_DIR, 'https_debug.log');
                    fs.appendFileSync(debugPath, `[${new Date().toISOString()}] UNDICI PID:${process.pid} URL:${url.substring(0,100)}\n`);
                } catch {}
                if (isApi && request.body) {
                    try {
                        const bodyStr = typeof request.body === 'string' ? request.body :
                            Buffer.isBuffer(request.body) ? request.body.toString('utf-8') : null;
                        if (bodyStr) {
                            const injected = inject(bodyStr);
                            if (injected) {
                                request.body = injected;
                                request.headers['content-length'] = Buffer.byteLength(injected).toString();
                            }
                            // Transcript dump
                            try {
                                const j = JSON.parse(bodyStr);
                                const contents = j.contents || j.messages;
                                if (Array.isArray(contents) && contents.length > 0) {
                                    const last = contents[contents.length - 1];
                                    const text = last?.parts?.[0]?.text || (typeof last?.content === 'string' ? last.content : '');
                                    if (text && text.length > 20 && !text.includes('[NeuronFS')) {
                                        const entry = JSON.stringify({ ts: new Date().toISOString(), role: last.role || 'user', text: text.substring(0, 2000) }) + '\n';
                                        const tp = path.join(INBOX_DIR, 'transcript_latest.jsonl');
                                        fs.mkdirSync(path.dirname(tp), { recursive: true });
                                        fs.appendFileSync(tp, entry);
                                    }
                                }
                            } catch {}
                        }
                    } catch {}
                }
            }
        });
        fs.appendFileSync(path.join(INBOX_DIR, 'HOOK_LOADED.txt'),
            `[${new Date().toISOString()}] undici interceptor registered PID:${process.pid}\n`);
    }
} catch (_undiciErr) {
    // undici not available in this process — that's OK
}

// ─── Re-patch fetch (undici may replace globalThis.fetch after hook loads) ───
setTimeout(() => {
    if (globalThis.fetch && !globalThis.fetch.__neuronPatched) {
        const _fetch2 = globalThis.fetch;
        globalThis.fetch = async function(...args) {
            const [input, init] = args;
            const url = typeof input === 'string' ? input : input?.url || '';
            const isApi = url.includes('generativelanguage') || url.includes('anthropic') ||
                          url.includes('gemini.googleapis') || url.includes('openai.com');
            try {
                const debugPath = path.join(INBOX_DIR, 'https_debug.log');
                fs.appendFileSync(debugPath, `[${new Date().toISOString()}] FETCH2 PID:${process.pid} URL:${url.substring(0,100)}\n`);
            } catch {}
            if (isApi && init?.body) {
                try {
                    const bodyStr = typeof init.body === 'string' ? init.body : new TextDecoder().decode(init.body);
                    const injected = inject(bodyStr);
                    if (injected) init.body = injected;
                    try {
                        const j = JSON.parse(bodyStr);
                        const contents = j.contents || j.messages;
                        if (Array.isArray(contents) && contents.length > 0) {
                            const last = contents[contents.length - 1];
                            const text = last?.parts?.[0]?.text || (typeof last?.content === 'string' ? last.content : '');
                            if (text && text.length > 20 && !text.includes('[NeuronFS')) {
                                const entry = JSON.stringify({ ts: new Date().toISOString(), role: last.role || 'user', text: text.substring(0, 2000) }) + '\n';
                                const tp = path.join(INBOX_DIR, 'transcript_latest.jsonl');
                                fs.mkdirSync(path.dirname(tp), { recursive: true });
                                fs.appendFileSync(tp, entry);
                            }
                        }
                    } catch {}
                } catch {}
            }
            return _fetch2.apply(this, args);
        };
        globalThis.fetch.__neuronPatched = true;
    }
}, 3000);

console.log(`[NeuronFS] Live context injection active (brain: ${BRAIN_PATH})`);
