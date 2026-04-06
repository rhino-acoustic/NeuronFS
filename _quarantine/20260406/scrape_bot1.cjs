const http = require('http');
const WebSocket = require('ws');
const fs = require('fs');

async function scrape() {
  const list = await new Promise((resolve, reject) => {
    http.get('http://127.0.0.1:9000/json/list', res => {
      let d = ''; res.on('data', c => d+=c); res.on('end', ()=>resolve(JSON.parse(d)));
    }).on('error', reject);
  });
  
  const bot1 = list.find(t => t.title?.includes('bot1') && t.type === 'page');
  const ws = new WebSocket(bot1.webSocketDebuggerUrl);
  await new Promise(r => { ws.on('open', r); });
  
  let id = 1;
  const pending = new Map();
  ws.on('message', msg => {
    const data = JSON.parse(msg);
    if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
  });
  const call = (method, params) => new Promise(resolve => {
    const myId = id++;
    pending.set(myId, resolve);
    ws.send(JSON.stringify({ id: myId, method, params }));
  });
  
  await call('Runtime.enable', {});
  await new Promise(r => setTimeout(r, 300));
  
  // antigravity-agent-side-panel의 전체 텍스트 가져오기
  const expr = `(() => {
    const panel = document.querySelector('.antigravity-agent-side-panel');
    if (panel) return panel.innerText;
    return 'panel_not_found';
  })()`;
  
  const res = await call('Runtime.evaluate', { expression: expr });
  const val = res?.result?.result?.value || 'failed';
  fs.writeFileSync('C:\\\\Users\\\\BASEMENT_ADMIN\\\\NeuronFS\\\\bot1_response.txt', val, 'utf8');
  console.log('saved: ' + val.length + ' chars');
  ws.close();
  setTimeout(() => process.exit(0), 500);
}
scrape().catch(e => { console.error(e.message); process.exit(1); });
