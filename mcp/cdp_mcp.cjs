const WebSocket = require('ws');
const ws = new WebSocket('ws://127.0.0.1:9000/devtools/page/90B738D1D7CA3744C1E72490614AEB21');
ws.on('open', () => {
  ws.send(JSON.stringify({
    id: 1, method: 'Runtime.evaluate',
    params: { expression: "try{require('vscode').commands.executeCommand('mcp.restartAllServers').then(()=>'OK')}catch(e){e.message}", returnByValue: true, awaitPromise: true }
  }));
});
ws.on('message', (d) => { console.log(d.toString().substring(0,300)); ws.close(); process.exit(0); });
setTimeout(() => { console.log('timeout'); process.exit(1); }, 5000);
