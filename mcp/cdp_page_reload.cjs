const WebSocket = require('ws');
const ws = new WebSocket('ws://127.0.0.1:9000/devtools/page/90B738D1D7CA3744C1E72490614AEB21');
ws.on('open', () => {
  ws.send(JSON.stringify({ id: 1, method: 'Page.reload', params: {} }));
  setTimeout(() => { console.log('Page.reload sent'); ws.close(); process.exit(0); }, 1000);
});
ws.on('error', (e) => { console.log('err:', e.message); process.exit(1); });
setTimeout(() => process.exit(1), 5000);
