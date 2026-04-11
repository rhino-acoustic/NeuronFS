const WebSocket = require('ws');
const ws = new WebSocket('ws://127.0.0.1:9000/devtools/page/90B738D1D7CA3744C1E72490614AEB21');
ws.on('open', async () => {
  // 1. Ctrl+Shift+P로 명령 팔레트 열기
  ws.send(JSON.stringify({
    id: 1, method: 'Input.dispatchKeyEvent',
    params: { type: 'keyDown', modifiers: 6, key: 'p', code: 'KeyP', windowsVirtualKeyCode: 80 }
  }));
  
  await new Promise(r => setTimeout(r, 500));
  
  // 2. "restart mcp" 타이핑
  const text = 'restart mcp servers';
  ws.send(JSON.stringify({
    id: 2, method: 'Input.insertText',
    params: { text: text }
  }));
  
  await new Promise(r => setTimeout(r, 500));
  
  // 3. Enter
  ws.send(JSON.stringify({
    id: 3, method: 'Input.dispatchKeyEvent',
    params: { type: 'keyDown', key: 'Enter', code: 'Enter', windowsVirtualKeyCode: 13 }
  }));
  
  await new Promise(r => setTimeout(r, 1000));
  console.log('MCP restart command dispatched');
  ws.close();
  process.exit(0);
});
ws.on('message', (d) => { /* silent */ });
setTimeout(() => { console.log('timeout'); process.exit(1); }, 8000);
