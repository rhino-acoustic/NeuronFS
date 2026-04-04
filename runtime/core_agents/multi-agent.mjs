/**
 * multi-agent.mjs — CDP 멀티에이전트 통신 엔진
 * 
 * Agent A(현재 창) ↔ Agent B(bot1 창) 간 CDP 통신
 * brain_v4/_agents/ 를 통한 뉴런 기반 소통
 * 
 * Usage: node multi-agent.mjs "검증 요청 메시지"
 *        node multi-agent.mjs --scrape  (Agent B 응답만 읽기)
 */
import http from 'http';
import WebSocket from 'ws';
import { writeFileSync, readFileSync, existsSync, mkdirSync } from 'fs';
import { join } from 'path';

const CDP_PORT = 9000;
const BRAIN = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const AGENTS_DIR = join(BRAIN, '_agents');
const RESULTS_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS';

const MESSAGE = process.argv[2] || '';
const SCRAPE_ONLY = MESSAGE === '--scrape';

// ─── CDP 헬퍼 ────────────────────────────────
function getJson(url) {
  return new Promise((resolve, reject) => {
    http.get(url, res => {
      let d = ''; res.on('data', c => d += c);
      res.on('end', () => resolve(JSON.parse(d)));
    }).on('error', reject);
  });
}

async function connectToBot1() {
  const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
  const bot1 = list.find(t => t.title?.includes('bot1') && t.type === 'page');
  if (!bot1) throw new Error('bot1 target not found');
  
  const ws = new WebSocket(bot1.webSocketDebuggerUrl);
  await new Promise((r, j) => { ws.on('open', r); ws.on('error', j); });
  
  let id = 1;
  const pending = new Map();
  ws.on('message', msg => {
    const data = JSON.parse(msg);
    if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
  });
  
  const call = (method, params) => new Promise((resolve, reject) => {
    const myId = id++;
    const t = setTimeout(() => { pending.delete(myId); reject(new Error('timeout')); }, 15000);
    pending.set(myId, r => { clearTimeout(t); resolve(r); });
    ws.send(JSON.stringify({ id: myId, method, params }));
  });
  
  await call('Runtime.enable', {});
  return { ws, call };
}

// ─── 인젝션 ──────────────────────────────────
async function injectMessage(call, message) {
  const focusScript = `(() => {
    function collectAll(root) {
      const found = [];
      const walk = node => {
        if (!node) return;
        if (node.shadowRoot) walk(node.shadowRoot);
        if (node.matches && node.matches('[aria-label="Message input"][role="textbox"]')) found.push(node);
        const children = node.children || [];
        for (let i = 0; i < children.length; i++) walk(children[i]);
      };
      walk(root);
      return found;
    }
    const inputs = collectAll(document);
    if (inputs.length > 0) { inputs[0].focus(); return 'ok'; }
    return 'not_found';
  })()`;
  
  const res = await call('Runtime.evaluate', { expression: focusScript });
  if (res?.result?.result?.value !== 'ok') throw new Error('Chat input not found — open Antigravity chat in bot1 window');
  
  await new Promise(r => setTimeout(r, 300));
  await call('Input.insertText', { text: message });
  await new Promise(r => setTimeout(r, 300));
  await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });
  await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });
}

// ─── 스크래핑 ────────────────────────────────
async function scrapeResponse(call) {
  const expr = `(() => {
    const panel = document.querySelector('.antigravity-agent-side-panel');
    if (panel) return panel.innerText;
    return 'panel_not_found';
  })()`;
  
  const res = await call('Runtime.evaluate', { expression: expr });
  return res?.result?.result?.value || 'scrape_failed';
}

// ─── 에이전트 통신 기록 ──────────────────────
function recordToNeuron(agentId, content) {
  const dir = join(AGENTS_DIR, agentId);
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  writeFileSync(join(dir, `msg_${timestamp}.md`), content, 'utf8');
  console.log(`[MULTI] ${agentId} 메시지 기록: msg_${timestamp}.md`);
}

// ─── 메인 ────────────────────────────────────
async function main() {
  console.log('[MULTI-AGENT] CDP 멀티에이전트 엔진 시작');
  
  const { ws, call } = await connectToBot1();
  console.log('[MULTI-AGENT] Agent B(bot1) 연결 성공');
  
  if (SCRAPE_ONLY) {
    console.log('[MULTI-AGENT] 스크래핑 모드...');
    const response = await scrapeResponse(call);
    
    recordToNeuron('agent_b', response);
    writeFileSync(join(RESULTS_DIR, 'agent_b_latest.txt'), response, 'utf8');
    console.log(`[MULTI-AGENT] Agent B 응답 저장 (${response.length} chars)`);
    
    ws.close();
    return;
  }
  
  if (!MESSAGE) {
    console.log('Usage: node multi-agent.mjs "메시지"');
    console.log('       node multi-agent.mjs --scrape');
    ws.close();
    return;
  }
  
  // Agent A → Agent B 메시지 전송
  console.log(`[MULTI-AGENT] Agent A → B: "${MESSAGE.substring(0, 100)}..."`);
  recordToNeuron('agent_a', `# Agent A → B\n${new Date().toISOString()}\n\n${MESSAGE}`);
  
  await injectMessage(call, MESSAGE);
  console.log('[MULTI-AGENT] 인젝션 완료 — Agent B 응답 대기...');
  console.log('[MULTI-AGENT] 응답이 끝나면: node multi-agent.mjs --scrape');
  
  ws.close();
}

main().catch(e => { console.error('[MULTI-AGENT] 에러:', e.message); process.exit(1); });
