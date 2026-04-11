const WebSocket = require('ws');
const ws = new WebSocket('ws://127.0.0.1:9000/devtools/page/90B738D1D7CA3744C1E72490614AEB21');
ws.on('open', () => {
  // VS Code 내부 API로 MCP 서버 리스타트 명령 실행
  ws.send(JSON.stringify({
    id: 1,
    method: 'Runtime.evaluate',
    params: {
      expression: "globalThis.require && globalThis.require('vscode').commands.executeCommand('mcp.restartAllServers').then(()=>'MCP restarted').catch(e=>e.message)",
      returnByValue: true,
      awaitPromise: true
    }
  }));
});
ws.on('message', (data) => {
  console.log(data.toString().substring(0, 200));
  ws.close();
  process.exit(0);
});
setTimeout(() => { console.log('timeout'); process.exit(1); }, 5000);
