const WebSocket = require('ws');
const ws = new WebSocket('ws://127.0.0.1:9000/devtools/page/90B738D1D7CA3744C1E72490614AEB21');
ws.on('open', async () => {
  // Ctrl+Shift+P → "reload window" 키 디스패치
  ws.send(JSON.stringify({ id:1, method:'Input.dispatchKeyEvent', params:{type:'keyDown',modifiers:6,key:'p',code:'KeyP',windowsVirtualKeyCode:80} }));
  await new Promise(r=>setTimeout(r,800));
  ws.send(JSON.stringify({ id:2, method:'Input.insertText', params:{text:'reload window'} }));
  await new Promise(r=>setTimeout(r,800));
  ws.send(JSON.stringify({ id:3, method:'Input.dispatchKeyEvent', params:{type:'keyDown',key:'Enter',code:'Enter',windowsVirtualKeyCode:13} }));
  await new Promise(r=>setTimeout(r,500));
  console.log('reloadWindow dispatched');
  ws.close(); process.exit(0);
});
ws.on('message',()=>{});
setTimeout(()=>{console.log('timeout');process.exit(1);},6000);
