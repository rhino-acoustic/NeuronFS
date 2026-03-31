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
const path = require('path');

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
    const h = opts?.hostname || opts?.host || '';
    return h.includes('generativelanguage.googleapis.com') ||
           h.includes('gemini.googleapis.com') ||
           h.includes('api.anthropic.com') ||
           h.includes('api.openai.com');
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

        let finalBody;
        try {
            const raw = Buffer.concat(chunks).toString('utf-8');
            const injected = inject(raw);
            finalBody = injected ? Buffer.from(injected, 'utf-8') : Buffer.concat(chunks);
        } catch {
            finalBody = Buffer.concat(chunks);
        }

        try { req.setHeader('content-length', finalBody.length); } catch {}
        _write(finalBody);
        _end();
    };

    // Log tool-call responses (optional, for headless execution pipeline)
    req.on('response', (res) => {
        let resChunks = [];
        res.on('data', (c) => resChunks.push(c));
        res.on('end', () => {
            try {
                const body = Buffer.concat(resChunks).toString();
                const hasToolCall = body && (
                    body.includes('functionCall') ||
                    body.includes('"tool_use"') ||
                    body.includes('run_command') ||
                    body.includes('write_to_file')
                );
                if (hasToolCall) {
                    const p = path.join(INBOX_DIR, `intercepted_${Date.now()}.json`);
                    fs.writeFileSync(p, body);
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
                      url.includes('gemini.googleapis');

        if (isApi && init?.body) {
            try {
                const injected = inject(init.body);
                if (injected) init.body = injected;
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

console.log(`[NeuronFS] Live context injection active (brain: ${BRAIN_PATH})`);
