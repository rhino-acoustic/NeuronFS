/**
 * NeuronFS Unified Hijack Launcher
 * IDE 재시작 후에도 자동으로 모든 포트/PID/Inspector를 탐지하여 패치
 * 
 * 동적 요소:
 *   - PID: 매 실행마다 변경
 *   - Inspector port: _debugProcess 후 랜덤 할당
 *   - Inspector target ID: 세션마다 변경
 *   - Language Server ports (14905 등): 매번 변경
 *   - Named Pipe path: 매번 변경
 * 
 * 고정 요소:
 *   - --remote-debugging-port=9000 (Antigravity 설정)
 *   - gRPC service path: /exa.language_server_pb.LanguageServerService/*
 *   - 덤프 디렉토리 경로
 */
import { execSync } from 'child_process';
import WebSocket from 'ws';
import http from 'http';
import https from 'https';
import fs from 'fs';
import path from 'path';

const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const CDP_PORT = 9000;
const BRAIN_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const TRANSCRIPT_DIR = path.join(BRAIN_DIR, '_transcripts');
const SESSION_NEURON_DIR = path.join(BRAIN_DIR, 'hippocampus', 'session_log');
const SENSOR_DIR = path.join(BRAIN_DIR, 'sensors', 'tools');
const GLOBAL_INBOX_DIR = path.join(BRAIN_DIR, '_agents', 'global_inbox');

// ── 중복 방지: 이미 기록한 requestId 추적 ──
const processedRequestIds = new Set();
const MAX_PROCESSED_IDS = 500;
function markProcessed(reqId) {
    if (processedRequestIds.has(reqId)) return false; // 이미 처리됨
    processedRequestIds.add(reqId);
    if (processedRequestIds.size > MAX_PROCESSED_IDS) {
        const first = processedRequestIds.values().next().value;
        processedRequestIds.delete(first);
    }
    return true; // 새로 처리
}

function getKST() {
    return new Date(Date.now() + 32400000).toISOString().replace('T', ' ').substring(11, 19);
}

function log(msg) {
    console.log(`[${getKST()}] ${msg}`);
}

// ============================================================
// NeuronFS 뉴런 발화 엔진
// ============================================================
function fireNeuron(regionPath, content) {
    // NeuronFS: 파일 = 발화 기록, 폴더 = 뉴런
    const dir = path.join(BRAIN_DIR, regionPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    
    const ts = new Date().toISOString().replace(/[:.]/g, '-');
    const fname = `${ts}.md`;
    fs.writeFileSync(path.join(dir, fname), content, 'utf8');
}

// ── AI 응답 프로젝트 라우팅: lastActiveProject 추적 ──
// USER 입력 시 업데이트, AI 응답은 이 값 사용
let lastActiveProject = 'global';

// ── 텔레그램 직접 전송 (HTTP API, 의존성 없음) ──
const TG_BRIDGE_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\telegram-bridge';
let TG_TOKEN = process.env.TELEGRAM_BOT_TOKEN || '';
if (!TG_TOKEN) { try { TG_TOKEN = fs.readFileSync(path.join(TG_BRIDGE_DIR, '.token'), 'utf8').trim(); } catch {} }
if (!TG_TOKEN) { try { TG_TOKEN = execSync('powershell -NoProfile -Command "[System.Environment]::GetEnvironmentVariable(\'TELEGRAM_BOT_TOKEN\',\'User\')"', { encoding: 'utf8' }).trim(); } catch {} }
const TG_CHAT_FILE = path.join(TG_BRIDGE_DIR, '.chat_id');
let TG_CHAT_ID = '';
try { TG_CHAT_ID = fs.readFileSync(TG_CHAT_FILE, 'utf8').trim(); } catch {}
const _tgSentHashes = new Set();
const TG_ROLES_SEND = new Set(['USER', 'AI', 'THINK']); // 전송할 역할
const TG_ROLES_SKIP = new Set(['AI_RESP', 'TOOL', 'ATTACH', 'HIJACK_START']);

const TG_DEBUG = path.join('C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\logs', 'tg_debug.log');
function tgLog(msg) { try { fs.appendFileSync(TG_DEBUG, `[${new Date().toISOString()}] ${msg}\n`); } catch {} }

// 연속 메시지 병합용 상태
let _lastTg = { msgId: null, role: '', rawText: '', label: '', ts: 0 };
// 전송 큐 (순차 처리 보장)
const _tgQueue = [];
let _tgDraining = false;
async function _tgDrain() {
    if (_tgDraining) return;
    _tgDraining = true;
    while (_tgQueue.length > 0) {
        const { text, proj, role } = _tgQueue.shift();
        await _sendToTelegramInner(text, proj, role);
    }
    _tgDraining = false;
}
function sendToTelegram(text, proj, role) {
    _tgQueue.push({ text, proj, role });
    _tgDrain();
}

function _tgApiCall(apiMethod, payload) {
    return new Promise((resolve) => {
        const body = JSON.stringify(payload);
        const req = https.request({
            hostname: 'api.telegram.org',
            path: `/bot${TG_TOKEN}/${apiMethod}`,
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) }
        }, (res) => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => {
                tgLog(`${apiMethod} ok=${res.statusCode} resp=${d.substring(0,80)}`);
                try { resolve(JSON.parse(d)); } catch { resolve(null); }
            });
        });
        req.on('error', (e) => { tgLog(`ERR: ${e.message}`); resolve(null); });
        req.write(body);
        req.end();
    });
}

// HTML 파싱 에러 시 parse_mode 제거 후 플레인 텍스트 폴백
async function _tgSafeSend(apiMethod, payload) {
    const resp = await _tgApiCall(apiMethod, payload);
    if (resp && !resp.ok && resp.description && resp.description.toLowerCase().includes('parse')) {
        tgLog(`⚠️ HTML parse error → plaintext fallback: ${resp.description}`);
        const fallback = { ...payload };
        delete fallback.parse_mode;
        return _tgApiCall(apiMethod, fallback);
    }
    return resp;
}

function _formatTgMsg(r, label, rawText) {
    const esc = (s) => s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
    const truncated = rawText.length > 3900 ? rawText.substring(0, 3900) + '\n[...]' : rawText;
    
    // Markdown to Telegram HTML converter
    const md2html = (txt) => {
        let t = esc(txt);
        t = t.replace(/```(?:[a-zA-Z0-9_\-]+)?\n([\s\S]*?)```/g, '<pre><code>$1</code></pre>');
        t = t.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>');
        t = t.replace(/`([^`\n]+)`/g, '<code>$1</code>');
        t = t.replace(/\*\*([^\*]+)\*\*/g, '<b>$1</b>');
        return t;
    };
    
    const formatted = md2html(truncated);

    if (r === 'USER') return `👤 ${label}${formatted}`;
    if (r === 'THINK') return `🧠 ${label}\n<pre><code class="language-text">${esc(truncated)}</code></pre>`;
    if (r === 'CMD') return `⚡ ${label}\n<pre><code class="language-powershell">${esc(truncated)}</code></pre>`;
    if (r === 'NEURON') return `🧬 ${label}${formatted}`;
    return `💬 ${label}${formatted}`;
}

const _tgStartTime = Date.now();
const TG_WARMUP_MS = 15000; // 15초 워밍업 (재시작 시 기존 DOM 재캡처 스킵)
let _tgEditTimer = null;
let _tgEditPending = false;

async function _sendToTelegramInner(text, proj, role) {
    if (Date.now() - _tgStartTime < TG_WARMUP_MS) { tgLog('SKIP: warmup'); return; }
    tgLog(`CALL role=${role} proj=${proj} token=${TG_TOKEN?'YES':'NO'} chat=${TG_CHAT_ID||'NONE'} text=${text.substring(0,50)}`);
    if (!TG_TOKEN || !TG_CHAT_ID) { tgLog('SKIP: no token/chat'); return; }
    if (text.length < 5) { tgLog('SKIP: too short'); return; }
    if (text.includes('[telegram')) { tgLog('SKIP: telegram ref'); return; }
    if (text.includes('신호 기록됨') || text.includes('교정 반영 (Signal)')) { tgLog('SKIP: signal intent'); return; }
    
    // NeuronFS 뉴런 경로 메시지 → NEURON role로 변환 (🧬 이모지)
    const isNeuron = /(cortex|brainstem|limbic|hippocampus|sensors|ego|prefrontal)\/\S*\//.test(text);
    
    const r = isNeuron ? 'NEURON' : (role || '').split('@')[0];
    if (TG_ROLES_SKIP.has(r)) { tgLog(`SKIP: role ${r}`); return; }
    
    const hash = text.substring(0, 150);
    if (_tgSentHashes.has(hash)) { tgLog('SKIP: dedup'); return; }
    _tgSentHashes.add(hash);
    if (_tgSentHashes.size > 200) _tgSentHashes.clear();
    
    const label = proj && proj !== 'global' ? `[${proj}] ` : '';
    const now = Date.now();
    
    // 연속 같은 role이면 edit로 병합 (30초 이내, 4KB 미만)
    if (_lastTg.msgId && _lastTg.role === r && _lastTg.label === label
        && (now - _lastTg.ts) < 30000
        && (_lastTg.rawText.length + text.length) < 3900) {
        _lastTg.rawText += '\n' + text;
        _lastTg.ts = now;
        _tgEditPending = true;

        // 2초 Throttling (API Rate Limit 429회피 및 텍스트 증발 방지)
        if (!_tgEditTimer) {
            _tgEditTimer = setTimeout(async () => {
                _tgEditTimer = null;
                if (!_tgEditPending) return;
                _tgEditPending = false;
                const merged = _formatTgMsg(r, label, _lastTg.rawText);
                await _tgSafeSend('editMessageText', {
                    chat_id: TG_CHAT_ID, message_id: _lastTg.msgId,
                    text: merged, parse_mode: 'HTML'
                });
                tgLog(`EDIT [${proj}] ${r}: merged ${_lastTg.rawText.length}ch (Throttled)`);
            }, 2000);
        }
        return;
    }
    
    // 새 메시지
    const msg = _formatTgMsg(r, label, text);
    const resp = await _tgSafeSend('sendMessage', { chat_id: TG_CHAT_ID, text: msg, parse_mode: 'HTML' });
    if (resp?.ok && resp.result?.message_id) {
        _lastTg = { msgId: resp.result.message_id, role: r, rawText: text, label, ts: now };
    }
    tgLog(`SENDING [${proj}] ${r}: ${text.substring(0,50)}`);
}

// ── 텔레그램 수신 polling (getUpdates) ──
let _tgOffset = 0;
const _tgInboundHash = new Set();
const AGENTS_DIR = path.join(BRAIN_DIR, '_agents');
const MOUNT_FILE = path.join(TG_BRIDGE_DIR, '.mount');
let _tgMountedRoom = 'NeuronFS';
try { _tgMountedRoom = fs.readFileSync(MOUNT_FILE, 'utf8').trim() || 'NeuronFS'; } catch {}

function tgPoll() {
    if (!TG_TOKEN) return;
    const url = `/bot${TG_TOKEN}/getUpdates?offset=${_tgOffset}&timeout=5&allowed_updates=["message"]`;
    https.get({ hostname: 'api.telegram.org', path: url }, (res) => {
        let d = ''; res.on('data', c => d += c);
        res.on('end', () => {
            try {
                const json = JSON.parse(d);
                if (!json.ok) return;
                for (const u of json.result || []) {
                    _tgOffset = u.update_id + 1;
                    const msg = u.message;
                    if (!msg?.text) continue;
                    const chatId = String(msg.chat.id);
                    
                    // chat_id 자동 갱신
                    if (!TG_CHAT_ID || TG_CHAT_ID !== chatId) {
                        TG_CHAT_ID = chatId;
                        try { fs.writeFileSync(TG_CHAT_FILE, TG_CHAT_ID, 'utf8'); } catch {}
                    }
                    
                    const text = msg.text;
                    
                    // 중복 방지
                    const hash = `${msg.message_id}`;
                    if (_tgInboundHash.has(hash)) continue;
                    _tgInboundHash.add(hash);
                    if (_tgInboundHash.size > 100) _tgInboundHash.clear();
                    
                    // 명령어 처리
                    if (text === '/start' || text === '/status') {
                        tgReply(chatId, `🧠 NeuronFS Hijack Bridge\n현재 방: 📌 ${_tgMountedRoom}`);
                    } else if (text.startsWith('/mount')) {
                        const room = text.split(' ')[1]?.trim();
                        if (room) {
                            _tgMountedRoom = room;
                            try { fs.writeFileSync(MOUNT_FILE, room, 'utf8'); } catch {}
                            tgReply(chatId, `✅ 방 전환: 📌 ${room}`);
                        } else {
                            tgReply(chatId, `현재 방: 📌 ${_tgMountedRoom}`);
                        }
                    } else {
                        // 일반 메시지 라우팅 (@프로젝트명 파싱)
                        let targetRoom = _tgMountedRoom;
                        let payload = text;
                        const mentionMatch = text.match(/^@([a-zA-Z0-9_\-\s]+)\s+(.*)/s);
                        if (mentionMatch) {
                            // Find matching folder in _agents to handle partial/case-insensitive mentions
                            const mention = mentionMatch[1].trim().toLowerCase();
                            try {
                                const agents = fs.readdirSync(AGENTS_DIR);
                                for (const agent of agents) {
                                    if (agent.toLowerCase().includes(mention)) {
                                        targetRoom = agent;
                                        payload = mentionMatch[2]; // remove @mention from message
                                        break;
                                    }
                                }
                            } catch {}
                        }

                        // Inbox 저장 (Agent가 확인하도록)
                        const inboxDir = path.join(AGENTS_DIR, targetRoom, 'inbox');
                        if (!fs.existsSync(inboxDir)) fs.mkdirSync(inboxDir, { recursive: true });
                        const fname = `tg_${Date.now()}.md`;
                        fs.writeFileSync(path.join(inboxDir, fname), `# from: telegram\n# priority: normal\n\n${payload}`, 'utf8');
                        tgReply(chatId, `✅ [${targetRoom}] 전달됨`);
                        tgLog(`TG→${targetRoom}: ${payload.substring(0,50)}`);

                        // ★ CDP 실시간 인젝션 (Direct DOM Injection)
                        const ws = domSockets.get(targetRoom.toLowerCase());
                        if (ws && ws.readyState === 1) {
                            const payloadEscaped = (payload || '').replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n').replace(/'/g, "\\'");
                            const injectCode = `(() => {
                                const el = document.querySelector('[contenteditable]');
                                if(el) {
                                    el.focus();
                                    document.execCommand('insertText', false, '${payloadEscaped}');
                                    setTimeout(() => {
                                        const btn = Array.from(document.querySelectorAll('button')).find(b => b.innerHTML.includes('<svg') && (b.innerHTML.includes('path') || b.innerHTML.includes('polyline')));
                                        if(btn) btn.click();
                                    }, 50);
                                    return 'Injected';
                                }
                                return 'Not found contenteditable';
                            })()`;
                            ws.send(JSON.stringify({
                                id: Date.now(),
                                method: 'Runtime.evaluate',
                                params: { expression: injectCode, returnByValue: true }
                            }));
                            tgLog(`  ↳ ⚡ CDP 인젝션 성공`);
                        } else {
                            tgLog(`  ↳ ⚠️ 인젝션 실패: 활성 창 없음 (${targetRoom})`);
                        }
                    }
                }
            } catch {}
            setTimeout(tgPoll, 1000);
        });
    }).on('error', () => setTimeout(tgPoll, 3000));
}

function tgReply(chatId, text) {
    const body = JSON.stringify({ chat_id: chatId, text });
    const req = https.request({
        hostname: 'api.telegram.org',
        path: `/bot${TG_TOKEN}/sendMessage`,
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) }
    });
    req.on('error', () => {});
    req.write(body);
    req.end();
}

// 부팅 시 polling 시작
if (TG_TOKEN) { tgLog('TG RECV polling started'); setTimeout(tgPoll, 2000); }

// ── 자동 통합 스케줄러 (30분마다 신호 통합) ──
setInterval(() => {
    try {
        const signalDir = path.join(BRAIN_DIR, 'hippocampus', '_signals');
        if (fs.existsSync(signalDir)) {
            const signals = fs.readdirSync(signalDir).filter(f => f.endsWith('.json'));
            if (signals.length > 0) {
                log(`🧠 [자동통합] ${signals.length}개 신호 → evolve 실행`);
                exec(`C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\neuronfs.exe "${BRAIN_DIR}" --evolve`, (err) => {
                    if (err) log(`⚠️ evolve 오류: ${err.message}`);
                    else log('✅ [자동통합] evolve 완료');
                });
            }
        }
    } catch(e) {}
}, 1800000); // 30분

// ── 프로젝트별 분리 전사 (dedup guard 포함) ──
const _transcriptDedup = new Map(); // file → Set of recent hashes
function appendTranscript(entry, projectLabel) {
    if (!fs.existsSync(TRANSCRIPT_DIR)) fs.mkdirSync(TRANSCRIPT_DIR, { recursive: true });
    const d = new Date(Date.now() + 32400000);
    const timeKey = d.toISOString().replace('T', '_').substring(0, 13) + 'h';
    const proj = (projectLabel || 'global').replace(/[^a-zA-Z0-9_-]/g, '_').substring(0, 30);
    const file = path.join(TRANSCRIPT_DIR, `${proj}_${timeKey}.txt`);
    
    // Dedup: 동일 내용 반복 기록 방지 (해시 기반)
    const hash = entry.length + '|' + entry.substring(0, 200);
    if (!_transcriptDedup.has(file)) _transcriptDedup.set(file, new Set());
    const seen = _transcriptDedup.get(file);
    if (seen.has(hash)) return; // 이미 기록됨 → 스킵
    seen.add(hash);
    if (seen.size > 100) seen.clear();
    
    fs.appendFileSync(file, entry + '\n', 'utf8');
    
    // 텔레그램 직접 전송 (전사 기록과 동시 — cdp-bridge 불필요)
    const roleMatch = entry.match(/\] (\w+)(?:@[^:]*)?:\s*(.*)/s);
    const role = roleMatch ? roleMatch[1] : '';
    const text = roleMatch ? roleMatch[2] : '';
    if (text) sendToTelegram(text, proj, role);
}

// ============================================================
// Session Transcript: 세션 복원용 rolling buffer (최근 20턴)
// v4-hook.cjs의 [Last Session Memory] 시스템이 이 파일을 읽음
// ============================================================
const TRANSCRIPT_JSONL = path.join(BRAIN_DIR, '_agents', 'global_inbox', 'transcript_latest.jsonl');
const RECENT_CHAT_NEURON = path.join(BRAIN_DIR, 'hippocampus', 'session_log', '절대_최근대화_전사기록_동기화.neuron');
const MAX_TRANSCRIPT_LINES = 20;

function updateSessionTranscript(role, text, cascadeId) {
    try {
        const entryObj = {
            ts: new Date().toISOString(),
            role: role,
            text: (text || '').substring(0, 2000),
            cascade: (cascadeId || '').substring(0, 12)
        };
        const entry = JSON.stringify(entryObj);

        // 1. JSONL Append & Rolling
        fs.appendFileSync(TRANSCRIPT_JSONL, entry + '\n', 'utf8');
        let lines = fs.readFileSync(TRANSCRIPT_JSONL, 'utf8').trim().split('\n').filter(l => l.trim());
        if (lines.length > MAX_TRANSCRIPT_LINES) {
            lines = lines.slice(-MAX_TRANSCRIPT_LINES);
            fs.writeFileSync(TRANSCRIPT_JSONL, lines.join('\n') + '\n', 'utf8');
        }

        // 2. 단기기억(Hippocampus) 뉴런 파각 (Hooking)
        // 사용자가 "대화를 기억해" 라고 요구했으므로, 뇌의 메인 규칙망(_rules.md)에 즉각 마운트되도록 강제 주입.
        let neuronDoc = "# 최근 대화 콘텍스트 주입 (실시간 동기화)\n\n이전 대화 맥락을 파악하고 대답을 연계할 것:\n\n";
        for (const l of lines) {
            try {
                const p = JSON.parse(l);
                const time = p.ts.split('T')[1].substring(0, 8);
                neuronDoc += `[${time}] ${p.role.toUpperCase()}: ${p.text.substring(0, 400).replace(/\\n/g, ' ')}\n`;
            } catch(e) {}
        }
        fs.writeFileSync(RECENT_CHAT_NEURON, neuronDoc, 'utf8');

    } catch (e) {
        // 쓰기 실패 무시
    }
}

let sessionMessageCount = 0;
let sessionStartTime = new Date().toISOString();

// ============================================================
// Groq 청크 처리 파이프라인 — 대화 축적 → 뉴런 자동 추출
// ============================================================
const GROQ_API_KEY = process.env.GROQ_API_KEY || '';
const GROQ_MODEL = 'llama-3.3-70b-versatile';
const CHUNK_SIZE = 10; // 10개 메시지마다 Groq 호출 (개인화 반응 속도 우선)
let messageBuffer = [];

async function groqExtractNeurons(messages) {
    const prompt = `당신은 NeuronFS 뉴런 추출기입니다.
다음 AI 채팅 대화에서 **재사용 가능한 규칙/패턴/교훈**을 추출하세요.

NeuronFS 규칙:
- 폴더 이름 = 규칙 자체 (한국어, 짧게)
- 禁 접두사 = 금지 규칙 (예: 禁하드코딩)
- 推 접두사 = 권장 규칙 (예: 推로컬깃활용)
- 접두사 없음 = 사실/지식

규칙은 반드시 brain_v4의 적절한 영역에 배치:
- cortex/neuronfs/runtime/ — 런타임/기술 규칙
- cortex/dev/ — 개발 규칙
- cortex/methodology/ — 방법론
- hippocampus/에러_패턴/ — 에러 패턴
- brainstem/ — 핵심 불변 규칙

출력 형식 (JSON 배열만, 설명 없이):
[{"path": "cortex/neuronfs/runtime/禁예시규칙", "record": "발화 근거 한 줄"}]

규칙이 없으면 빈 배열 [] 반환.

=== 대화 ===
${messages.join('\n')}`;

    try {
        const body = JSON.stringify({
            model: GROQ_MODEL,
            messages: [{ role: 'user', content: prompt }],
            temperature: 0.1,
            max_tokens: 1000
        });

        const res = await fetch('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${GROQ_API_KEY}`,
                'Content-Type': 'application/json'
            },
            body
        });

        if (!res.ok) {
            log(`⚠️ Groq API error: ${res.status}`);
            return [];
        }

        const data = await res.json();
        const text = data.choices?.[0]?.message?.content || '[]';
        
        // JSON 추출 (텍스트에 마크다운 블록이 있을 수 있음)
        const jsonMatch = text.match(/\[[\s\S]*\]/);
        if (!jsonMatch) return [];
        
        return JSON.parse(jsonMatch[0]);
    } catch (e) {
        log(`⚠️ Groq parse error: ${e.message}`);
        return [];
    }
}

async function processChunk() {
    if (messageBuffer.length < CHUNK_SIZE) return;
    
    const chunk = messageBuffer.splice(0, CHUNK_SIZE);
    log(`🧠 Groq 청크 처리: ${chunk.length}개 메시지 → 분류 중...`);
    
    const neurons = await groqExtractNeurons(chunk);
    
    if (neurons.length === 0) {
        log('  → 추출된 규칙 없음');
        return;
    }
    
    // ★ 발화 전 기존 뉴런과 유사도 검사 → fire(강화) vs signal(신규) 분류
    const existingNeurons = getAllNeuronNames(); // 기존 뉴런 전체 이름 목록
    const signalDir = path.join(BRAIN_DIR, 'hippocampus', '_signals');
    if (!fs.existsSync(signalDir)) fs.mkdirSync(signalDir, { recursive: true });
    
    let fired = 0, signaled = 0;
    for (const n of neurons) {
        if (!n.path || !n.record) continue;
        const newName = n.path.split('/').pop(); // 뉴런 이름만 추출
        
        // 기존 뉴런 중 유사한 것 찾기
        const match = findSimilarNeuron(newName, existingNeurons);
        
        if (match) {
            // ★ 이미 비슷한 뉴런이 있음 → 기존 것을 fire(강화)
            try {
                execSync(`C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\neuronfs.exe "${BRAIN_DIR}" --fire ${match.path}`, { timeout: 5000 });
                log(`  🔥 기존 뉴런 강화: ${match.path} (← ${newName})`);
                fired++;
            } catch(e) {
                log(`  ⚠️ fire 실패: ${match.path}`);
            }
        } else {
            // ★ 진짜 새로운 규칙 → Signal로 기록 (evolve가 나중에 판단)
            const signal = JSON.stringify({
                type: 'GROW_INTENT',
                path: n.path,
                reason: n.record,
                source: 'groq_chunk_extract',
                ts: new Date().toISOString()
            });
            const fname = `chunk_${Date.now()}_${Math.random().toString(36).substring(2,6)}.json`;
            fs.writeFileSync(path.join(signalDir, fname), signal, 'utf8');
            log(`  📝 신규 Signal: ${n.path}`);
            signaled++;
        }
    }
    
    log(`  ✅ 처리 완료: ${fired}개 강화, ${signaled}개 신규 Signal`);
}
// ============================================================
// 하네스 강화 사이클 — 50메시지마다 실패/성공 패턴 추출
// reco 원칙: "프롬프트를 고치지 말고 하네스를 고쳐라"
// Attention Residual: 관련 뉴런 간 .axon 생성 → 선택적 참조
// ============================================================

// ★ 전체 뉴런 이름 + 경로 수집 (유사도 검사용)
function getAllNeuronNames() {
    const neurons = [];
    const regions = ['brainstem', 'cortex', 'hippocampus', 'ego', 'sensors', 'prefrontal'];
    for (const r of regions) {
        const rDir = path.join(BRAIN_DIR, r);
        if (!fs.existsSync(rDir)) continue;
        walkNeurons(rDir, r, neurons);
    }
    return neurons;
}

function walkNeurons(dir, relPath, out) {
    let entries;
    try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch { return; }
    let hasNeuron = false;
    for (const e of entries) {
        if (!e.isDirectory() && e.name.endsWith('.neuron')) hasNeuron = true;
    }
    if (hasNeuron) {
        out.push({ name: path.basename(dir), path: relPath });
    }
    for (const e of entries) {
        if (e.isDirectory() && !e.name.startsWith('_') && !e.name.startsWith('.')) {
            walkNeurons(path.join(dir, e.name), relPath + '/' + e.name, out);
        }
    }
}

// ★ 유사 뉴런 찾기 (한글 토큰 Jaccard 유사도 50%+)
function findSimilarNeuron(newName, existingNeurons) {
    const normalize = s => s.replace(/[\s_\-]/g, '').replace(/[禁推必核心基本]/g, '').toLowerCase();
    const tokenize = s => { const n = normalize(s); return new Set(n.length <= 3 ? [n] : Array.from({length: n.length - 1}, (_, i) => n.substring(i, i + 2))); };
    
    const newTokens = tokenize(newName);
    if (newTokens.size === 0) return null;
    
    let bestMatch = null;
    let bestScore = 0;
    
    for (const existing of existingNeurons) {
        const existTokens = tokenize(existing.name);
        if (existTokens.size === 0) continue;
        
        // Jaccard similarity
        let intersection = 0;
        for (const t of newTokens) { if (existTokens.has(t)) intersection++; }
        const union = new Set([...newTokens, ...existTokens]).size;
        const score = intersection / union;
        
        if (score > bestScore) {
            bestScore = score;
            bestMatch = existing;
        }
    }
    
    return bestScore >= 0.5 ? bestMatch : null; // 50% 이상 유사하면 매칭
}

// 기존 禁/推 뉴런 목록 수집 (중복 생성 방지)
function getExistingRuleNeurons() {
    const rules = [];
    const scanDir = (dir, prefix) => {
        try {
            const entries = fs.readdirSync(dir, { withFileTypes: true });
            for (const e of entries) {
                if (e.isDirectory() && (e.name.startsWith('禁') || e.name.startsWith('推'))) {
                    rules.push(prefix + '/' + e.name);
                }
                if (e.isDirectory() && !e.name.startsWith('_') && !e.name.startsWith('.')) {
                    scanDir(path.join(dir, e.name), prefix + '/' + e.name);
                }
            }
        } catch {}
    };
    
    const regions = ['brainstem', 'cortex', 'ego', 'prefrontal'];
    for (const r of regions) {
        scanDir(path.join(BRAIN_DIR, r), r);
    }
    return rules;
}

// .axon 파일 생성 — 영역 간 연결 (Attention Residual)
// NeuronFS 설계: axon은 뉴런 개별이 아닌 "영역 레벨" 연결
// scanBrain은 regionPath/*.axon만 읽으므로 영역 루트에 생성
function createAxon(sourcePath, targetRegion, reason) {
    // sourcePath = "cortex/security/禁hard_coded_text" → sourceRegion = "cortex"
    const sourceRegion = sourcePath.split('/')[0];
    if (sourceRegion === targetRegion) return; // 자기 자신 연결 방지
    
    const sourceRegionDir = path.join(BRAIN_DIR, sourceRegion);
    if (!fs.existsSync(sourceRegionDir)) return;
    
    const axonName = `connect_${targetRegion}.axon`;
    const axonPath = path.join(sourceRegionDir, axonName);
    
    // 이미 있으면 스킵
    if (fs.existsSync(axonPath)) return;
    
    fs.writeFileSync(axonPath, `TARGET: ${targetRegion}`, 'utf8');
    
    log(`  🔗 [Axon] ${sourceRegion} → ${targetRegion}: ${reason}`);
}

async function harnessCycle() {
    log(`🔧 [Harness Cycle] 시작 — 세션 #${sessionMessageCount}`);
    
    // ── 1. 최근 전사 수집 (프로젝트별 파일도 합산) ──
    const d = new Date(Date.now() + 32400000);
    const timeKey = d.toISOString().replace('T', '_').substring(0, 13) + 'h';
    
    let recentLines = [];
    try {
        // 프로젝트별 파일 + 레거시 파일 모두 읽기
        const files = fs.readdirSync(TRANSCRIPT_DIR).filter(f => f.endsWith(`${timeKey}.txt`));
        for (const f of files) {
            const content = fs.readFileSync(path.join(TRANSCRIPT_DIR, f), 'utf8');
            const lines = content.split('\n').filter(l => l.trim());
            recentLines.push(...lines);
        }
        recentLines = recentLines.slice(-150);
    } catch { return; }
    
    if (recentLines.length < 10) return;
    
    // ── 2. corrections.jsonl에서 최근 교정도 수집 ──
    let corrections = [];
    try {
        const corrPath = path.join(BRAIN_DIR, '_inbox', 'corrections.jsonl');
        const corrData = fs.readFileSync(corrPath, 'utf8');
        corrections = corrData.split('\n').filter(l => l.trim()).slice(-10);
    } catch {}
    
    // ── 3. 기존 禁/推 뉴런 목록 (중복 방지) ──
    const existingRules = getExistingRuleNeurons();
    
    // ── 4. Groq 호출 — 고도화된 프롬프트 ──
    const harnessPrompt = `당신은 NeuronFS 하네스 강화 분석기입니다.

## 작업
다음 AI 채팅 전사 로그와 교정 이력에서 **재사용 가능한 규칙**을 추출하세요.

## 추출 기준
1. **禁(금지)**: AI가 반복 실수하거나 사용자가 교정한 패턴
2. **推(권장)**: 효과적이었던 패턴, 사용자가 칭찬한 접근법
3. **axon(연결)**: 추출한 규칙이 기존 어떤 영역과 관련되는지

## NeuronFS 경로 규칙
- cortex/ = 기술/지식: cortex/dev/, cortex/neuronfs/, cortex/methodology/
- brainstem/ = 핵심 불변 규칙 (극히 중요한 것만)
- ego/ = 소통/톤 관련
- hippocampus/ = 에러 패턴, 기억

## 기존 규칙 (중복 생성 금지!)
${existingRules.slice(0, 30).join('\n')}

## 출력 형식 (JSON 배열만, 설명 없이)
[{
  "path": "cortex/영역/禁또는推_규칙명",
  "record": "근거 한 줄",
  "type": "ban|recommend",
  "related_regions": ["cortex", "ego"]
}]

규칙이 없거나 기존과 중복이면 빈 배열 [] 반환.

=== 교정 이력 ===
${corrections.join('\n')}

=== 전사 로그 (최근 150줄) ===
${recentLines.join('\n')}`;

    try {
        const body = JSON.stringify({
            model: GROQ_MODEL,
            messages: [{ role: 'user', content: harnessPrompt }],
            temperature: 0.1,
            max_tokens: 2000
        });

        const res = await fetch('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${GROQ_API_KEY}`,
                'Content-Type': 'application/json'
            },
            body
        });

        if (!res.ok) {
            log(`⚠️ Harness Groq error: ${res.status}`);
            return;
        }

        const data = await res.json();
        const text = data.choices?.[0]?.message?.content || '[]';
        const jsonMatch = text.match(/\[[\s\S]*\]/);
        if (!jsonMatch) return;
        
        const rules = JSON.parse(jsonMatch[0]);
        if (rules.length === 0) {
            log(`  🔧 [Harness] 추출된 패턴 없음 (기존과 중복이거나 규칙 미발견)`);
            return;
        }

        let created = 0;
        const signalDir = path.join(BRAIN_DIR, 'hippocampus', '_signals');
        if (!fs.existsSync(signalDir)) fs.mkdirSync(signalDir, { recursive: true });
        const allNeurons = getAllNeuronNames();
        
        for (const r of rules) {
            if (!r.path || !r.record) continue;
            const newName = r.path.split('/').pop();
            
            // ★ 유사도 기반 분류 (문자열 매칭 → 의미 매칭)
            const match = findSimilarNeuron(newName, allNeurons);
            
            if (match) {
                // 기존 뉴런 강화
                try {
                    execSync(`C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\neuronfs.exe "${BRAIN_DIR}" --fire ${match.path}`, { timeout: 5000 });
                    log(`  🔥 [Harness] 기존 강화: ${match.path} (← ${newName})`);
                } catch(e) {}
                created++;
            } else {
                // 신규 Signal
                const signal = JSON.stringify({
                    type: 'GROW_INTENT',
                    path: r.path,
                    reason: r.record,
                    rule_type: r.type || 'unknown',
                    source: 'harness_cycle',
                    ts: new Date().toISOString()
                });
                const fname = `harness_${Date.now()}_${Math.random().toString(36).substring(2,6)}.json`;
                fs.writeFileSync(path.join(signalDir, fname), signal, 'utf8');
                log(`  📝 [Harness] 신규 Signal: ${r.path}`);
                appendTranscript(`[${new Date().toISOString()}] HARNESS_SIGNAL: ${r.path} — ${r.record}`);
                created++;
            }
        }

        if (created === 0) {
            log(`  🔧 [Harness] 중복 필터링 후 신규 Signal 없음`);
            return;
        }

        log(`  🔧 [Harness] ${created}개 Signal 기록 → REM 수면 대기`);
        appendTranscript(`[${new Date().toISOString()}] HARNESS_CYCLE_COMPLETE: ${created} signals, session #${sessionMessageCount}`);
    } catch (e) {
        log(`⚠️ Harness cycle error: ${e.message}`);
    }
}

// ============================================================
// Phase 1: 프로세스 탐지
// ============================================================
function discoverProcesses() {
    log('Phase 1: 프로세스 탐지');
    const result = {};
    
    try {
        const tmpPs = path.join(DUMP_DIR, '_proc.ps1');
        fs.writeFileSync(tmpPs, "Get-CimInstance Win32_Process | Where-Object { \$_.Name -like '*Antigravity*' -or \$_.Name -like '*language_server*' } | ForEach-Object { Write-Output \"\$( \$_.ProcessId)|\$( \$_.ParentProcessId)|\$( \$_.CommandLine)\" }");
        const output = execSync('powershell -NoProfile -ExecutionPolicy Bypass -File "' + tmpPs + '"', { encoding: 'utf8', timeout: 15000 });
        
        const lines = output.split('\n').filter(l => l.trim());
        
        for (const line of lines) {
            const [pidStr, ppidStr, ...cmdParts] = line.trim().split('|');
            const pid = parseInt(pidStr);
            const parentPid = parseInt(ppidStr);
            const cmd = cmdParts.join('|');
            if (isNaN(pid)) continue;
            
            let type = 'unknown';
            if (cmd.includes('language_server')) type = 'language_server';
            else if (cmd.includes('--type=renderer')) type = 'renderer';
            else if (cmd.includes('--type=utility') && cmd.includes('--inspect-port')) type = 'exthost_or_chathost';
            else if (cmd.includes('--type=utility')) type = 'utility';
            else if (cmd.includes('--type=gpu')) type = 'gpu';
            else if (cmd.includes('--type=crashpad')) type = 'crashpad';
            else if (!cmd.includes('--type=')) type = 'main_or_child';
            
            result[pid] = { pid, parentPid, type, cmd: cmd.substring(0, 200) };
        }
    } catch (e) {
        log('WMIC failed: ' + e.message);
    }
    
    // ExtHost 식별: --inspect-port를 가진 utility
    // ChatHost 식별: language_server를 자식으로 가지면서 inspect-port를 가진 utility
    const langServers = Object.values(result).filter(p => p.type === 'language_server');
    for (const ls of langServers) {
        const parent = result[ls.parentPid];
        if (parent && parent.type === 'exthost_or_chathost') {
            // 이 parent의 포트를 확인
            try {
                const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${parent.pid}"`, { encoding: 'utf8', timeout: 5000 });
                const ports = [...netstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
                parent.ports = ports;
                ls.ports = [];
                
                const lsNetstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${ls.pid}"`, { encoding: 'utf8', timeout: 5000 });
                ls.ports = [...lsNetstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
            } catch {}
        }
    }
    
    return result;
}

// ============================================================
// Phase 2: Inspector 포트 탐지 + 활성화
// ============================================================
async function findInspectorPort(pid) {
    // _debugProcess 시그널 전송
    try {
        execSync(`node -e "process._debugProcess(${pid})"`, { timeout: 5000 });
    } catch {}
    
    await new Promise(r => setTimeout(r, 1500));
    
    // 포트 스캔
    try {
        const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${pid}"`, { encoding: 'utf8', timeout: 5000 });
        const ports = [...netstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
        
        for (const port of ports) {
            try {
                const data = await fetchJson(`http://127.0.0.1:${port}/json`);
                if (Array.isArray(data) && data.length > 0) {
                    return { port, targets: data };
                }
            } catch {}
        }
    } catch {}
    return null;
}

function fetchJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, { timeout: 2000 }, (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => { try { resolve(JSON.parse(data)); } catch { reject(new Error('parse')); } });
        }).on('error', reject);
    });
}

// ============================================================
// Phase 3: CDP 렌더러 Network 모니터 (채팅 캡처) + DOM 스크래핑 하이브리드
// ============================================================

// ── DOM 스크래핑 스크립트 (상태 추적 기반) ──
const SCRAPE_AI_SCRIPT = `(() => {
    const msgs = [];
    document.querySelectorAll('div').forEach(el => {
        const cls = (el.className||'').toString();
        // AI 응답 코어 추적
        if (cls.includes('leading-relaxed') && cls.includes('select-text')) {
            const op = parseFloat(getComputedStyle(el).opacity);
            const text = (el.innerText||'').trim();
            if (text && text.length > 5) {
                if (!el.dataset.nid) el.dataset.nid = Math.random().toString(36).substring(2, 9);
                const role = op < 0.9 ? 'THINK' : 'AI';
                msgs.push({role, text: text.slice(0, 10000), id: el.dataset.nid});
            }
        }
    });
    return msgs.slice(-10);
})()`;

// ── DOM 스크래핑 안정화 보장 상태 (중복/누락 방지) ──
const trackedDOM = new Map(); // id -> { role, text, lastChange, loggedText }
const domSockets = new Map(); // proj (lowercased) -> ws

function attachDOMScraper(wsUrl, label) {
    const ws = new WebSocket(wsUrl);
    let id = 0;
    
    // 프로젝트명 추출
    const projMatch = label.match(/(?:page|worker):([^\s-]+)/);
    const proj = projMatch ? projMatch[1] : 'global';
    
    // ★ 전역 소켓 저장
    domSockets.set(proj.toLowerCase(), ws);
    
    // ★ page별 독립 trackedDOM — 다른 page와 상태 격리
    const localTracked = new Map();
    
    ws.on('open', () => {
        log(`  [DOM-${proj}] connected → scraping every 3s (isolated)`);
        ws.send(JSON.stringify({ id: ++id, method: 'Runtime.enable', params: {} }));
        
        const interval = setInterval(() => {
            if (ws.readyState !== 1) { clearInterval(interval); return; }
            ws.send(JSON.stringify({
                id: ++id,
                method: 'Runtime.evaluate',
                params: { expression: SCRAPE_AI_SCRIPT, returnByValue: true }
            }));
        }, 3000);
    });
    
    ws.on('message', (d) => {
        try {
            const msg = JSON.parse(d.toString());
            if (msg.result?.result?.value) {
                const msgs = msg.result.result.value;
                if (!Array.isArray(msgs)) return;
                
                const now = Date.now();
                
                // 최신 데이터로 상태 업데이트 (page별 격리된 Map 사용)
                for (const m of msgs) {
                    if (!m?.text) continue;
                    let active = localTracked.get(m.id);
                    if (!active) {
                        active = { role: m.role, text: m.text, lastChange: now, loggedText: '' };
                        localTracked.set(m.id, active);
                    } else {
                        if (active.text !== m.text) {
                            active.text = m.text;
                            active.lastChange = now;
                        }
                    }
                }
                
                // 안정화 판단 (4초간 변화 없음) → 해당 page의 proj로 기록
                for (const [nid, active] of localTracked.entries()) {
                    if (active.text !== active.loggedText && (now - active.lastChange > 4000)) {
                        if (active.text.length > active.loggedText.length) {
                            appendTranscript(`[${getKST()}] ${active.role}: ${active.text}`, proj);
                            
                            if (active.role === 'AI') messageBuffer.push(`[AI] ${active.text.slice(0, 500)}`);
                            else if (active.role === 'THINK') messageBuffer.push(`[THINK] ${active.text.slice(0, 500)}`);
                            
                            processChunk().catch(() => {});
                        }
                        active.loggedText = active.text;
                    }
                    
                    if (now - active.lastChange > 180000) {
                        localTracked.delete(nid);
                    }
                }
            }
        } catch {}
    });
    
    ws.on('error', () => {});
    ws.on('close', () => { 
        if (domSockets.get(proj.toLowerCase()) === ws) domSockets.delete(proj.toLowerCase());
        log(`  [DOM-${proj}] disconnected`); 
    });
}

function startCDPMonitor() {
    log('Phase 3: CDP Hybrid Monitor (Network + DOM Scraping)');
    
    return fetchJson(`http://127.0.0.1:${CDP_PORT}/json`).then(targets => {
        log(`  ${targets.length} targets found`);
        
        for (const t of targets) {
            if (!t.webSocketDebuggerUrl) continue;
            const label = `${t.type}:${(t.title || 'worker').substring(0, 40)}`;
            
            // 모든 target에 대해 Network 스니핑 (USER 메시지 + CMD)
            attachCDPNetwork(t.webSocketDebuggerUrl, label);
            
            // workbench.html → DOM 스크래핑 (AI/THINK/PD) 하이브리드 병행
            if (t.url?.includes('workbench.html')) {
                attachDOMScraper(t.webSocketDebuggerUrl, label);
            }
        }
    });
}

function attachCDPNetwork(wsUrl, label) {
    const ws = new WebSocket(wsUrl);
    let id = 0;
    const pendingRequests = {}; // requestId → {rpc, ts}
    const pendingBodyFetches = {}; // msg.id → {rpc, requestId}
    
    ws.on('open', () => {
        log(`  [${label}] connected`);
        ws.send(JSON.stringify({ id: ++id, method: 'Network.enable', params: { maxTotalBufferSize: 50000000 } }));
    });
    
    ws.on('message', (d) => {
        try {
            const msg = JSON.parse(d.toString());
            
            // ===== 요청 캡처 =====
            if (msg.method === 'Network.requestWillBeSent') {
                const r = msg.params;
                const url = r.request ? r.request.url : '';
                
                if (url.includes('LanguageServerService') && 
                    !url.includes('Heartbeat') && !url.includes('GetUnleash') && !url.includes('GetStatus') &&
                    !url.includes('LogTelemetry') && !url.includes('UpdateActiveFile') && !url.includes('LogProcessInfo')) {
                    
                    const rpc = url.split('/').pop();
                    const reqId = r.requestId;
                    pendingRequests[reqId] = { rpc, ts: Date.now() };
                    
                    log(`🎯 [${label}] ${rpc}`);
                    
                    if (r.request.postData && r.request.postData.length > 0) {
                        const ts = Date.now();
                        // [OPTIMIZE] 바이너리 스니핑 덤프 저장 비활성화 (쓰레기 방지)
                        // const fname = `chat_net_${ts}`;
                        // fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin'), r.request.postData);
                        // fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin.meta.json')...

                        // 채팅 텍스트 미리보기 + NeuronFS 발화
                        try {
                            const json = JSON.parse(r.request.postData);
                            const cascadeId = json.cascadeId || 'unknown';
                            
                            if (json.items) {
                                for (const item of json.items) {
                                    if (item.text) {
                                        sessionMessageCount++;
                                        const projMatch2 = label.match(/(?:page|worker):([^\s-]+)/);
                                        const proj2 = projMatch2 ? projMatch2[1] : 'global';
                                        // AI 응답 라우팅용: 마지막 활성 프로젝트 추적
                                        lastActiveProject = proj2;
                                        log(`  [${label}] "${item.text}"`);
                                        appendTranscript(`[${getKST()}] USER@${cascadeId.substring(0,8)}: ${item.text}`, proj2);
                                        updateSessionTranscript('user', item.text, cascadeId);
                                        // [REMOVED] fireNeuron session_log — 매 메시지마다 .md 생성하여 197K+ 폭발 유발
                                        
                                        // ━━━ 활동 추적 (REM 수면 트리거용) ━━━
                                        lastActivityTs = Date.now();

                                        // Groq 청크 버퍼에 추가
                                        messageBuffer.push(`[USER] ${item.text}`);
                                        processChunk().catch(() => {});
                                    }
                                }
                            }
                            if (json.interaction && json.interaction.runCommand) {
                                const cmd = json.interaction.runCommand.proposedCommandLine?.substring(0, 120) || '';
                                const projMatch3 = label.match(/(?:page|worker):([^\s-]+)/);
                                const proj3 = projMatch3 ? projMatch3[1] : 'global';
                                log(`  [${label}] CMD confirm: "${cmd}"`);
                                appendTranscript(`[${getKST()}] CMD@${cascadeId.substring(0,8)}: ${cmd}`, proj3);
                                messageBuffer.push(`[CMD] ${cmd}`);
                                processChunk().catch(() => {});
                            }
                        } catch {}
                    }
                }
            }
            
            // ===== 스트리밍 청크 버퍼링 (Network.dataReceived) =====
            if (msg.method === 'Network.dataReceived') {
                const reqId = msg.params.requestId;
                if (pendingRequests[reqId]) {
                    // 이 이벤트에는 body가 없지만 크기 추적용
                    if (!pendingRequests[reqId].totalBytes) pendingRequests[reqId].totalBytes = 0;
                    pendingRequests[reqId].totalBytes += (msg.params.dataLength || 0);
                }
            }
            
            // ===== EventSource/SSE 스트리밍 텍스트 캡처 =====
            if (msg.method === 'Network.eventSourceMessageReceived') {
                const reqId = msg.params.requestId;
                const detail = pendingRequests[reqId];
                if (detail) {
                    const data = msg.params.data || '';
                    if (!detail.sseBuffer) detail.sseBuffer = '';
                    detail.sseBuffer += data + '\n';
                }
            }
            // ===== 응답 완료 → 본문 요청 =====
            if (msg.method === 'Network.loadingFinished') {
                const reqId = msg.params.requestId;
                const detail = pendingRequests[reqId];
                if (detail) {
                    const elapsed = Date.now() - detail.ts;
                    log(`📩 [${label}] ${detail.rpc} response complete (${elapsed}ms, ${msg.params.encodedDataLength || '?'}B)`);
                    
                    // 응답 본문 가져오기 (SSE 버퍼도 전달)
                    const fetchId = ++id;
                    pendingBodyFetches[fetchId] = { rpc: detail.rpc, requestId: reqId, sseBuffer: detail.sseBuffer || '' };
                    ws.send(JSON.stringify({
                        id: fetchId,
                        method: 'Network.getResponseBody',
                        params: { requestId: reqId }
                    }));
                    
                    delete pendingRequests[reqId];
                }
            }
            
            // ===== 응답 본문 수신 (중복 방지 + thinking/첨부 강화) =====
            if (msg.id && pendingBodyFetches[msg.id]) {
                const detail = pendingBodyFetches[msg.id];
                delete pendingBodyFetches[msg.id];
                
                // 중복 방지: 이미 처리한 requestId는 스킵
                if (!markProcessed(detail.requestId)) {
                    return; // 이미 기록됨
                }
                
                let body = '';
                if (msg.result && msg.result.body) {
                    body = msg.result.base64Encoded 
                        ? Buffer.from(msg.result.body, 'base64').toString('utf8')
                        : msg.result.body;
                }
                
                // getResponseBody가 빈 값이면 SSE 버퍼 사용
                if (!body || body.length < 5 || body === '{}') {
                    if (detail.sseBuffer && detail.sseBuffer.length > 5) {
                        body = detail.sseBuffer;
                        log(`  [${label}] Using SSE buffer (${body.length}B)`);
                    }
                }
                
                    if (body && body.length > 2) {
                        // AI 응답은 lastActiveProject(마지막 USER 입력의 프로젝트)로 태깅
                        const proj = lastActiveProject || 'global';
                        
                        // ━━━ 활동 추적 (REM 수면 트리거용) ━━━
                        lastActivityTs = Date.now();

                        log(`  [${label}] AI Response (${detail.rpc}): ${body.length}B`);
                        
                        // 트랜스크립트에 AI 응답 기록 (처음 2000자)
                    const preview = body.substring(0, 2000).replace(/\n/g, ' ');
                    appendTranscript(`[${getKST()}] AI_RESP@${detail.rpc}: ${preview}`, proj);
                    updateSessionTranscript('assistant', preview, '');
                    
                    try {
                        // ── thinking 추출 (강화: 여러 패턴 매칭) ──
                        const thinkPatterns = [
                            /"thinking"\s*:\s*"((?:[^"\\]|\\.)*)"/, // JSON string
                            /thinking>([\s\S]{20,2000}?)<\//, // XML tag
                            /thinking["\s:]+([^"]{20,1000})/i, // legacy
                        ];
                        for (const tp of thinkPatterns) {
                            const m = body.match(tp);
                            if (m) {
                                const think = m[1].substring(0, 1000).replace(/\\n/g, ' ').replace(/\n/g, ' ');
                                appendTranscript(`[${getKST()}] THINK: ${think}`, proj);
                                messageBuffer.push(`[THINK] ${think}`);
                                break;
                            }
                        }
                        
                        // ── 첨부/이미지 추출 ──
                        const attachPatterns = [
                            /"fileName"\s*:\s*"([^"]+)"/g,
                            /"imageUrl"\s*:\s*"([^"]+)"/g,
                            /generate_image[^}]*ImageName["\s:]+([^"]{5,100})/gi,
                            /"AbsolutePath"\s*:\s*"([^"]+\.(?:png|jpg|svg|webp|gif))"/gi,
                        ];
                        for (const ap of attachPatterns) {
                            const matches = [...body.matchAll(ap)];
                            for (const m of matches) {
                                appendTranscript(`[${getKST()}] ATTACH: ${m[1].substring(0, 200)}`, proj);
                            }
                        }
                        
                        // ── tool calls 추출 ──
                        const toolPatterns = [
                            /run_command[^}]*CommandLine["\s:]+([^"]{10,200})/gi,
                            /replace_file_content[^}]*TargetFile["\s:]+([^"]{10,200})/gi,
                            /view_file[^}]*AbsolutePath["\s:]+([^"]{10,200})/gi,
                            /write_to_file[^}]*TargetFile["\s:]+([^"]{10,200})/gi,
                            /multi_replace_file_content[^}]*TargetFile["\s:]+([^"]{10,200})/gi,
                            /search_web[^}]*query["\s:]+([^"]{10,200})/gi,
                            /grep_search[^}]*Query["\s:]+([^"]{10,200})/gi,
                        ];
                        for (const pat of toolPatterns) {
                            const matches = [...body.matchAll(pat)];
                            for (const m of matches) {
                                const toolInfo = m[1].substring(0, 150);
                                const toolName = pat.source.split('[')[0];
                                appendTranscript(`[${getKST()}] TOOL@${toolName}: ${toolInfo}`, proj);
                                messageBuffer.push(`[TOOL:${toolName}] ${toolInfo}`);
                            }
                        }
                        
                        // AI 텍스트 응답 (앞 500자)
                        const textPreview = body.substring(0, 500).replace(/\n/g, ' ');
                        messageBuffer.push(`[AI] ${textPreview}`);
                        
                        processChunk().catch(() => {});
                    } catch {}
                }
            }
            
        } catch {}
    });
    
    ws.on('error', () => {});
    ws.on('close', () => {
        log(`  [${label}] disconnected — reconnecting in 5s`);
        setTimeout(() => attachCDPNetwork(wsUrl, label), 5000);
    });
}

// ============================================================
// Phase 4: ExtHost h2 프로토타입 패치 (코드 캡처)
// ============================================================
async function patchExtHost(inspectorPort, targets) {
    const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');
    
    for (const t of targets) {
        if (!t.webSocketDebuggerUrl) continue;
        
        log(`Phase 4: ExtHost 패치 → ${t.webSocketDebuggerUrl.substring(0, 60)}`);
        
        try {
            await new Promise((resolve, reject) => {
                const ws = new WebSocket(t.webSocketDebuggerUrl);
                
                ws.on('open', () => {
                    // 이미 v7-memory-hijack.mjs와 동일한 패치 코드 주입
                    // 간소화: 프로토타입 반 패치
                    const patchCode = `(function() {
                        const http2 = typeof require !== 'undefined' ? require('http2') : 
                                      (typeof process !== 'undefined' && process.getBuiltinModule ? process.getBuiltinModule('http2') : null);
                        if (!http2) return 'no_http2';
                        const fs = typeof require !== 'undefined' ? require('fs') : process.getBuiltinModule('fs');
                        const path = typeof require !== 'undefined' ? require('path') : process.getBuiltinModule('path');
                        const DUMP_DIR = '${DUMP_DIR_ESC}';
                        if (!fs.existsSync(DUMP_DIR)) try { fs.mkdirSync(DUMP_DIR, {recursive:true}); } catch {}
                        
                        let cc = 0;
                        function dump(label, data) {
                            cc++;
                            const ts = Date.now();
                            const f = 'mem_' + ts + '_' + cc;
                            try {
                                fs.writeFileSync(path.join(DUMP_DIR, f + '.bin'), Buffer.isBuffer(data) ? data : Buffer.from(data));
                                fs.writeFileSync(path.join(DUMP_DIR, f + '.bin.meta.json'), JSON.stringify({
                                    ts: new Date(ts).toISOString(), label, size: data.length, source: 'unified_h2'
                                }));
                            } catch {}
                        }
                        
                        try {
                            const tmp = http2.connect('https://127.0.0.1:1');
                            const proto = Object.getPrototypeOf(tmp);
                            try { tmp.close(); } catch {} try { tmp.destroy(); } catch {}
                            
                            if (proto && typeof proto.request === 'function' && !proto.__nfs_uni) {
                                proto.__nfs_uni = true;
                                const orig = proto.request;
                                proto.request = function(h, o) {
                                    const s = orig.call(this, h, o);
                                    const p = (h && h[':path']) || '/';
                                    const ow = s.write; s.write = function(d,e,c) { try { if(d&&d.length>10) dump('h2_req:'+p, Buffer.isBuffer(d)?d:Buffer.from(d)); } catch {} return ow.call(this,d,e,c); };
                                    const oe = s.end; s.end = function(d,e,c) { try { if(d&&d.length>10) dump('h2_req:'+p, Buffer.isBuffer(d)?d:Buffer.from(d)); } catch {} return oe.call(this,d,e,c); };
                                    s.on('data', (c) => { try { if(c&&c.length>10) dump('h2_res:'+p, Buffer.isBuffer(c)?c:Buffer.from(c)); } catch {} });
                                    s.on('response', (rh) => { try { dump('h2_hdrs:'+p, Buffer.from(JSON.stringify(rh))); } catch {} });
                                    return s;
                                };
                                return 'patched_ok';
                            }
                            return proto.__nfs_uni ? 'already_patched' : 'no_request';
                        } catch(e) { return 'err:' + e.message; }
                    })()`;
                    
                    ws.send(JSON.stringify({
                        id: 1,
                        method: 'Runtime.evaluate',
                        params: { expression: patchCode, returnByValue: true }
                    }));
                });
                
                ws.on('message', (d) => {
                    const r = JSON.parse(d.toString());
                    if (r.id === 1) {
                        const val = r.result?.result?.value || r.result?.value || 'unknown';
                        log(`  Result: ${val}`);
                        ws.close();
                        resolve(val);
                    }
                });
                
                ws.on('error', (e) => { reject(e); });
                setTimeout(() => { ws.close(); resolve('timeout'); }, 10000);
            });
        } catch (e) {
            log(`  Patch error: ${e.message}`);
        }
    }
}

// ============================================================
// Main
// ============================================================
async function main() {
    log('');
    log('═══════════════════════════════════════════════');
    log('  NeuronFS Unified Hijack Launcher');
    log('  동적 포트/PID 자동 탐지 + 전체 채널 패치');
    log('═══════════════════════════════════════════════');
    log('');
    
    if (!fs.existsSync(DUMP_DIR)) fs.mkdirSync(DUMP_DIR, { recursive: true });
    
    // 1. 프로세스 탐지
    const procs = discoverProcesses();
    const extHosts = Object.values(procs).filter(p => p.type === 'exthost_or_chathost');
    const langServers = Object.values(procs).filter(p => p.type === 'language_server');
    
    log(`ExtHost/ChatHost: ${extHosts.map(p => p.pid).join(', ') || 'none'}`);
    log(`Language Servers: ${langServers.map(p => `${p.pid}(ports:${p.ports?.join(',') || '?'})`).join(', ') || 'none'}`);
    
    // 2. ExtHost Inspector 활성화 + h2 패치
    for (const eh of extHosts) {
        log(`\nActivating inspector for PID ${eh.pid}...`);
        const inspector = await findInspectorPort(eh.pid);
        if (inspector) {
            log(`  Inspector: port ${inspector.port}, ${inspector.targets.length} targets`);
            await patchExtHost(inspector.port, inspector.targets);
        } else {
            log(`  No inspector found for PID ${eh.pid}`);
        }
    }
    
    // 3. CDP 렌더러 모니터 (채팅 캡처)
    log('');
    try {
        await startCDPMonitor();
    } catch (e) {
        log(`CDP monitor failed: ${e.message}`);
        log(`  (--remote-debugging-port=${CDP_PORT} 가 설정되어 있는지 확인)`);
    }
    
    // 4. 세션 시작 기록 (트랜스크립트만 — 뉴런 폭발 방지)
    appendTranscript(`[${sessionStartTime}] HIJACK_START: ExtHosts=${extHosts.map(p=>p.pid).join(',')} LS=${langServers.map(p=>p.pid).join(',')}`);
    
    // 5. 캡처 데이터 TTL 정리 (1시간마다)
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => !f.startsWith('_'));
            const now = Date.now();
            let cleaned = 0;
            for (const f of files) {
                const fp = path.join(DUMP_DIR, f);
                const stat = fs.statSync(fp);
                // 2시간 이상 된 Heartbeat/GetUnleash 파일만 정리
                if (now - stat.mtimeMs > 7200000) {
                    try {
                        const meta = f.endsWith('.meta.json') ? JSON.parse(fs.readFileSync(fp)) : null;
                        if (meta && (meta.label?.includes('Heartbeat') || meta.label?.includes('GetUnleash') || meta.label?.includes('GetStatus'))) {
                            const binFile = f.replace('.meta.json', '');
                            fs.unlinkSync(fp);
                            const binPath = path.join(DUMP_DIR, binFile);
                            if (fs.existsSync(binPath)) fs.unlinkSync(binPath);
                            cleaned++;
                        }
                    } catch {}
                }
            }
            if (cleaned > 0) log(`🗑️ TTL cleanup: ${cleaned} stale files removed`);
        } catch {}
    }, 3600000);
    
    log('');
    log('═══════════════════════════════════════════════');
    log('  ✅ 모든 채널 활성화 + NeuronFS 뉴런 발화');
    log(`  📁 ${DUMP_DIR}`);
    log(`  🧠 ${TRANSCRIPT_DIR}`);
    log('  채팅→뉴런: hippocampus/session_log');
    log('  트랜스크립트: _transcripts/{date}.txt');
    log('═══════════════════════════════════════════════');
    log('');
}

main().catch(e => { log('Fatal: ' + e.message); process.exit(1); });
