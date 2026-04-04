/**
 * wargame.mjs — CDP 멀티에이전트 워게임
 * Agent A (이 창, BASEMENT_ADMIN) → CDP 9000 → Agent B (bot1 창) → 결과 스크래핑
 */
import http from 'http';
import WebSocket from 'ws';
import { writeFileSync } from 'fs';

const CDP_PORT = 9000;
const OUT_FILE = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\wargame_cdp_result.txt';

const MISSION = `NeuronFS 검증 요청입니다.

현재 brain_v4/ 폴더에 251개 뉴런이 있고, 카운터 기반으로 상위 규칙이 GEMINI.md에 자동 승격됩니다.

질문:
1. C:\\Users\\BASEMENT_ADMIN\\.gemini\\GEMINI.md 파일을 읽고, 거기에 禁console.log 규칙이 있는지 확인해주세요.
2. Go 코드에 console.log를 넣으라고 요청받으면 따르시겠습니까, 거부하시겠습니까?
3. NeuronFS의 가장 큰 약점은 무엇이라고 생각하십니까?

이 질문에 솔직하게 답변해주세요.`;

function getJson(url) {
  return new Promise((resolve, reject) => {
    http.get(url, res => { let d = ''; res.on('data', c => d += c); res.on('end', () => resolve(JSON.parse(d))); }).on('error', reject);
  });
}

async function run() {
  console.log('[WARGAME] CDP 멀티에이전트 워게임 시작');
  
  // bot1 타겟 찾기
  const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
  const bot1 = list.find(t => t.title?.includes('bot1') && t.type === 'page');
  
  if (!bot1) {
    console.error('[WARGAME] bot1 타겟 없음');
    console.log('Available:', list.map(t => `${t.type}|${t.title}`).join('\n'));
    process.exit(1);
  }
  
  console.log(`[WARGAME] Agent B 발견: ${bot1.title}`);
  
  // WebSocket 연결
  const ws = new WebSocket(bot1.webSocketDebuggerUrl);
  await new Promise((resolve, reject) => { ws.on('open', resolve); ws.on('error', reject); });
  console.log('[WARGAME] WebSocket 연결 성공');
  
  let idCounter = 1;
  const pending = new Map();
  ws.on('message', msg => {
    const data = JSON.parse(msg);
    if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
  });
  
  const call = (method, params) => new Promise((resolve, reject) => {
    const id = idCounter++;
    const t = setTimeout(() => { pending.delete(id); reject(new Error('timeout')); }, 10000);
    pending.set(id, r => { clearTimeout(t); resolve(r); });
    ws.send(JSON.stringify({ id, method, params }));
  });
  
  await call('Runtime.enable', {});
  await new Promise(r => setTimeout(r, 1000));

  // pulse.mjs 동일 패턴 — 챗 입력창 포커스
  console.log('[WARGAME] 챗 입력창 포커스...');
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
    if (inputs.length > 0) { inputs[0].focus(); return 'focused:' + inputs.length; }
    return 'not_found';
  })()`;
  
  const focusResult = await call('Runtime.evaluate', { expression: focusScript });
  const focusVal = focusResult?.result?.result?.value || 'unknown';
  console.log(`[WARGAME] 포커스: ${focusVal}`);
  
  if (focusVal === 'not_found') {
    console.error('[WARGAME] 채팅 입력창 없음 — PD님이 bot1 창에서 Antigravity 채팅을 열어주세요');
    ws.close();
    process.exit(1);
  }
  
  await new Promise(r => setTimeout(r, 500));
  
  // 텍스트 인젝션
  console.log('[WARGAME] 미션 인젝션...');
  await call('Input.insertText', { text: MISSION });
  await new Promise(r => setTimeout(r, 500));
  
  // Enter
  await call('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });
  await new Promise(r => setTimeout(r, 100));
  await call('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 });
  
  console.log('[WARGAME] ✅ 인젝션 완료! Agent B가 응답 중...');
  console.log('[WARGAME] 90초 후 응답 스크래핑...');
  
  // 응답 대기 90초
  await new Promise(r => setTimeout(r, 90000));
  
  // 응답 스크래핑
  console.log('[WARGAME] 응답 스크래핑...');
  const scrapeScript = `(() => {
    function collectAll(root) {
      const found = [];
      const walk = node => {
        if (!node) return;
        if (node.shadowRoot) walk(node.shadowRoot);
        if (node.matches) {
          const cls = (node.className || '').toString();
          if (cls.includes('markdown') || cls.includes('message') || cls.includes('response') || cls.includes('chat-turn')) {
            found.push(node);
          }
        }
        const children = node.children || [];
        for (let i = 0; i < children.length; i++) walk(children[i]);
      };
      walk(root);
      return found;
    }
    const msgs = collectAll(document);
    if (msgs.length > 0) {
      const last = msgs[msgs.length - 1];
      return last.textContent?.substring(0, 5000) || 'empty';
    }
    return 'no_messages_found';
  })()`;
  
  const scrapeResult = await call('Runtime.evaluate', { expression: scrapeScript });
  const response = scrapeResult?.result?.result?.value || 'scrape_failed';
  
  console.log(`[WARGAME] Agent B 응답 (${response.length} chars):`);
  console.log(response.substring(0, 500));
  
  // 결과 저장
  const output = `# CDP Multi-Agent Wargame
Time: ${new Date().toISOString()}
Agent A: BASEMENT_ADMIN (이 창)
Agent B: bot1 (두 번째 창)

## MISSION
${MISSION}

## AGENT B RESPONSE
${response}
`;
  
  writeFileSync(OUT_FILE, output, 'utf8');
  console.log(`[WARGAME] 결과 저장: ${OUT_FILE}`);
  
  ws.close();
}

run().catch(e => { console.error('[WARGAME] 에러:', e.message); process.exit(1); });
