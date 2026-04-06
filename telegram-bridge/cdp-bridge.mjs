#!/usr/bin/env node
/**
 * NeuronFS Telegram Bridge (File Watcher Mode)
 * 
 * NO CDP interference — reads existing transcript files only.
 * 
 * Architecture:
 *   Existing CDP hijacker writes → _transcripts/*.txt
 *   This bridge watches those files → diff extracts new lines → sends to Telegram
 *   Telegram commands → NeuronFS REST API (:9090)
 *
 * Zero DOM injection. Zero interference.
 */

import TelegramBot from 'node-telegram-bot-api';
import fs from 'fs';
import path from 'path';
import crypto from 'crypto';
import http from 'http';

const AGENTS_DIR = path.resolve(import.meta.dirname, '..', 'brain_v4', '_agents');

// ═══ Configuration ═══
const BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN;
if (!BOT_TOKEN) { console.error('[FATAL] TELEGRAM_BOT_TOKEN 환경변수 필요'); process.exit(1); }
const BRAIN_ROOT = process.env.BRAIN_ROOT || path.resolve(import.meta.dirname, '..', 'brain_v4');
const TRANSCRIPTS_DIR = path.join(BRAIN_ROOT, '_transcripts');
const CHAT_ID_FILE = path.join(import.meta.dirname, '.chat_id');

// Chat ID 복원
let AUTHORIZED_CHAT_ID = process.env.TELEGRAM_CHAT_ID || null;
try { if (!AUTHORIZED_CHAT_ID && fs.existsSync(CHAT_ID_FILE)) AUTHORIZED_CHAT_ID = fs.readFileSync(CHAT_ID_FILE, 'utf8').trim(); } catch {}

// ═══ State ═══
const CDP_PORT = 9000;
let currentRoom = 'NeuronFS';
const fileOffsets = new Map();          // file → last read byte offset
const sentHashes = new Set();           // MD5 ring buffer (50)
const MAX_SENT = 50;
let lastSentLine = '';
let lastInboundHash = '';
let lastSentMsgId = null;

// ═══ Telegram Bot ═══
const bot = new TelegramBot(BOT_TOKEN, { polling: true });
console.log('[TG] 🤖 Bot polling started...');

bot.on('polling_error', (e) => console.error('[TG] Polling error:', e.message));

bot.on('message', async (msg) => {
  const chatId = msg.chat.id;
  
  if (!AUTHORIZED_CHAT_ID) {
    AUTHORIZED_CHAT_ID = String(chatId);
    try { fs.writeFileSync(CHAT_ID_FILE, AUTHORIZED_CHAT_ID, 'utf8'); } catch {}
    console.log(`[TG] 🔐 Registered & saved: ${chatId}`);
    await bot.sendMessage(chatId, [
      '✅ NeuronFS Bridge 연결',
      '',
      '📋 명령어:',
      '/status — 시스템 상태',
      '/brain — 뉴런 요약',
      '/inject — 강제 주입',
      '/neurons — 뉴런 수',
      '',
      '📡 AI 응답이 자동으로 여기에 전달됩니다.',
      '',
      '/mount [방] — 방 전환 (NeuronFS/bot1/entp/enfp)',
      '/rooms — 실제 열린 창 목록'
    ].join('\n'));
    return;
  }
  
  if (String(chatId) !== AUTHORIZED_CHAT_ID) return;
  
  const text = msg.text;
  if (!text) return;

  try {
    if (text === '/start') {
      await bot.sendMessage(chatId, `🧠 NeuronFS Bridge\n현재 방: 📌 ${currentRoom}`);
    } else if (text === '/status') {
      await bot.sendMessage(chatId, await getStatus(), { parse_mode: 'HTML' });
    } else if (text === '/brain') {
      await bot.sendMessage(chatId, await getBrain(), { parse_mode: 'HTML' });
    } else if (text === '/inject') {
      await apiCall('/api/inject', 'POST');
      await bot.sendMessage(chatId, '🔄 Inject 트리거 완료');
    } else if (text === '/neurons') {
      const state = await apiCall('/api/state');
      await bot.sendMessage(chatId, `🧬 뉴런: ${state?.totalNeurons || '?'} | 활성: ${state?.totalActivation || '?'}`);
    } else if (text === '/rooms') {
      // CDP에서 실제 열린 창 목록
      try {
        const windows = await getCDPWindows();
        const list = windows.map(w => `${w.name === currentRoom ? '📌' : '  '} ${w.name}`).join('\n');
        await bot.sendMessage(chatId, `🏠 열린 창:\n\n${list}\n\n/mount [이름] 으로 전환`, { parse_mode: 'HTML' });
      } catch { await bot.sendMessage(chatId, '❌ CDP 연결 안 됨'); }
    } else if (text.startsWith('/mount')) {
      const room = text.split(' ')[1]?.trim();
      if (!room) {
        await bot.sendMessage(chatId, `현재 방: 📌 ${currentRoom}`);
      } else {
        currentRoom = room;
        await bot.sendMessage(chatId, `✅ 방 전환: 📌 ${currentRoom}`);
        console.log(`[MOUNT] 📌 ${currentRoom}`);
      }
    } else {
      // 중복 방지 (텔레그램 재전송 대비)
      const msgHash = crypto.createHash('md5').update(text).digest('hex');
      if (msgHash === lastInboundHash) return;
      lastInboundHash = msgHash;
      
      // 일반 메시지 → 현재 마운트된 방의 inbox에 전달
      const inboxDir = path.join(AGENTS_DIR, currentRoom, 'inbox');
      if (!fs.existsSync(inboxDir)) fs.mkdirSync(inboxDir, { recursive: true });
      const ts = Date.now();
      const filename = `tg_${ts}.md`;
      const content = `# from: telegram\n# priority: normal\n\n${text}`;
      fs.writeFileSync(path.join(inboxDir, filename), content, 'utf8');
      await bot.sendMessage(chatId, `✅ [${currentRoom}] 전달됨`);
      console.log(`[TG→${currentRoom}] 📩 ${filename}`);
    }
  } catch (e) {
    await bot.sendMessage(chatId, `❌ ${e.message}`).catch(() => {});
  }
});

// ═══ File Watcher: transcript → Telegram ═══

function getCurrentTranscriptFile() {
  const now = new Date();
  const date = now.toISOString().split('T')[0];
  const hour = now.getHours().toString().padStart(2, '0');
  // Try NeuronFS_ pattern first, then global_
  const candidates = [
    `NeuronFS_${date}_${hour}h.txt`,
    `global_${date}_${hour}h.txt`
  ];
  for (const name of candidates) {
    const fp = path.join(TRANSCRIPTS_DIR, name);
    if (fs.existsSync(fp)) return fp;
  }
  return path.join(TRANSCRIPTS_DIR, candidates[0]); // default
}

function md5(s) {
  return crypto.createHash('md5').update(s).digest('hex');
}

function isDuplicate(text) {
  const h = md5(text);
  if (sentHashes.has(h)) return true;
  sentHashes.add(h);
  if (sentHashes.size > MAX_SENT) {
    const first = sentHashes.values().next().value;
    sentHashes.delete(first);
  }
  return false;
}

function extractNewContent(filePath) {
  try {
    const stat = fs.statSync(filePath);
    const offset = fileOffsets.get(filePath) || 0;
    
    if (stat.size <= offset) return null; // no new data
    
    const fd = fs.openSync(filePath, 'r');
    const buf = Buffer.alloc(stat.size - offset);
    fs.readSync(fd, buf, 0, buf.length, offset);
    fs.closeSync(fd);
    
    fileOffsets.set(filePath, stat.size);
    return buf.toString('utf8');
  } catch {
    return null;
  }
}

function parseTranscriptLines(raw) {
  // ALL line types forwarded with distinct icons
  const lines = raw.split('\n');
  const messages = [];
  let currentMsg = null;
  
  for (const line of lines) {
    // Match any [HH:MM:SS] TYPE: pattern
    const match = line.match(/^\[(\d{2}:\d{2}:\d{2})\]\s+(\w+)(@[\w@]+)?:\s*(.*)/);
    if (match) {
      if (currentMsg && currentMsg.text.trim()) messages.push(currentMsg);
      const [, time, role, , text] = match;
      let icon = '📋';
      if (role === 'AI' || role === 'AI_RESP') icon = '💬';
      else if (role === 'USER') icon = '👤';
      else if (role === 'THINK') icon = '🧠';
      else if (role === 'CMD') icon = '⚡';
      else if (role === 'PD') icon = '✏️';
      currentMsg = { icon, role, time, text: (text || '').trim() };
    } else if (currentMsg && line.trim()) {
      currentMsg.text += '\n' + line;
    }
  }
  
  if (currentMsg && currentMsg.text.trim()) messages.push(currentMsg);
  return messages;
}

async function pollTranscripts() {
  if (!AUTHORIZED_CHAT_ID) return;
  
  const filePath = getCurrentTranscriptFile();
  const newContent = extractNewContent(filePath);
  if (!newContent) return;
  
  console.log(`[POLL] ${path.basename(filePath)}: +${newContent.length} bytes`);
  const messages = parseTranscriptLines(newContent);
  if (messages.length) console.log(`[POLL] ${messages.length} messages parsed`);
  
  for (const msg of messages) {
    if (msg.text.length < 5) continue;
    
    // 텔레그램 에코, AI_RESP 시스템, JSON 응답 차단
    if (msg.text.includes('[telegram')) continue;
    if (msg.role === 'AI_RESP') continue;
    if (msg.text.startsWith('{"') && msg.text.endsWith('}')) continue;
    
    let text = msg.text.length > 3900 
      ? msg.text.substring(0, 3900) + '\n\n[...truncated]' 
      : msg.text;
    
    if (text === lastSentLine) continue;
    if (isDuplicate(text)) continue;
    
    // HTML 이스케이프
    const esc = (s) => s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
    
    // 역할별 포맷 분기
    let fullMsg;
    if (msg.role === 'CMD') {
      fullMsg = `⚡ <pre>${esc(text)}</pre>`;
    } else if (msg.role === 'THINK') {
      fullMsg = `🧠 <i>${esc(text)}</i>`;
    } else if (msg.role === 'USER') {
      fullMsg = `👤 ${esc(text)}`;
    } else {
      // AI: 먼저 이스케이프 후 코드 블록 변환
      const escaped = esc(text);
      fullMsg = `💬 ${escaped.replace(/```(\w*)\n([\s\S]*?)```/g, (_, lang, code) => 
        `<pre>${code.trim()}</pre>`
      ).replace(/`([^`]+)`/g, '<code>$1</code>')}`;
    }
    
    const opts = { parse_mode: 'HTML' };
    
    // Progressive edit
    if (lastSentMsgId && lastSentLine && text.startsWith(lastSentLine.substring(0, 50))) {
      try {
        await bot.editMessageText(fullMsg, { chat_id: AUTHORIZED_CHAT_ID, message_id: lastSentMsgId, ...opts });
        lastSentLine = text;
        continue;
      } catch {}
    }
    
    lastSentLine = text;
    
    try {
      const sent = await bot.sendMessage(AUTHORIZED_CHAT_ID, fullMsg, opts);
      lastSentMsgId = sent.message_id;
      console.log(`[→TG] ${msg.icon} ${text.substring(0, 60)}...`);
      await new Promise(r => setTimeout(r, 50));
    } catch (e) {
      console.error('[TG] Send error:', e.message);
    }
  }
}

// ═══ CDP Window Discovery ═══

function getCDPWindows() {
  return new Promise((resolve, reject) => {
    http.get(`http://127.0.0.1:${CDP_PORT}/json/list`, (res) => {
      let d = ''; res.on('data', c => d += c);
      res.on('end', () => {
        try {
          const targets = JSON.parse(d);
          const windows = targets
            .filter(t => t.type === 'page' && t.url?.includes('workbench.html'))
            .map(t => ({
              name: t.title?.split(' - ')[0]?.trim() || t.title,
              title: t.title
            }));
          resolve(windows);
        } catch(e) { reject(e); }
      });
    }).on('error', reject);
  });
}

// ═══ NeuronFS API Helper ═══

async function apiCall(endpoint, method = 'GET') {
  try {
    const res = await fetch(`http://127.0.0.1:9090${endpoint}`, { method, signal: AbortSignal.timeout(3000) });
    if (res.ok) return await res.json();
  } catch {}
  return null;
}

async function getStatus() {
  const data = await apiCall('/api/health');
  if (data) {
    return `✅ <b>NeuronFS ONLINE</b>\n\n🧠 API: 정상\n📡 Telegram: 연결됨\n📂 Brain: ${BRAIN_ROOT}`;
  }
  return `⚠️ <b>NeuronFS API 미응답</b>\n프로세스는 실행 중일 수 있음`;
}

async function getBrain() {
  const data = await apiCall('/api/state');
  if (data) {
    let msg = `🧠 <b>Brain State</b>\n뉴런: ${data.totalNeurons} | 활성: ${data.totalActivation}\n\n`;
    if (data.regions) {
      for (const r of data.regions) {
        msg += `  ${r.name}: ${r.neurons}\n`;
      }
    }
    return msg;
  }
  return '❌ Brain 조회 실패';
}

// ═══ Boot ═══

// Initialize file offsets to current position (don't replay history)
try {
  const files = fs.readdirSync(TRANSCRIPTS_DIR).filter(f => f.endsWith('.txt'));
  for (const f of files) {
    const fp = path.join(TRANSCRIPTS_DIR, f);
    const stat = fs.statSync(fp);
    fileOffsets.set(fp, stat.size); // start from end
  }
  console.log(`[BOOT] 📂 Initialized ${files.length} transcript files (skip history)`);
} catch {}

console.log('[BOOT] 🚀 NeuronFS Telegram Bridge (File Watcher)');
console.log(`[BOOT]   Bot: @jlootfs_bot`);
console.log(`[BOOT]   Transcripts: ${TRANSCRIPTS_DIR}`);

// Poll every 3 seconds
setInterval(pollTranscripts, 3000);

// Crash protection
process.on('uncaughtException', (e) => console.error('[CRASH]', e.message));
process.on('unhandledRejection', (e) => console.error('[CRASH]', e?.message || e));
