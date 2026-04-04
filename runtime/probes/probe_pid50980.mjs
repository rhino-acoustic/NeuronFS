/**
 * PID 50980 Context 1(Node.js) — process가 있는 context에서 require 창조 후 http2 패치
 */
import WebSocket from 'ws';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';

function runOnTarget(expression, contextId) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(INSPECTOR_URL);
        const ctxs = [];
        let id = 0;
        let runtimeEnabled = false;
        
        ws.on('open', () => {
            id++;
            ws.send(JSON.stringify({ id, method: 'Runtime.enable' }));
        });
        
        ws.on('message', async (d) => {
            const msg = JSON.parse(d.toString());
            
            if (msg.method === 'Runtime.executionContextCreated') {
                ctxs.push(msg.params.context);
            }
            
            if (msg.id === 1 && !runtimeEnabled) {
                runtimeEnabled = true;
                await new Promise(r => setTimeout(r, 500));
                
                // Find the Node.js context (not "internal")
                const nodeCtx = ctxs.find(c => c.name && c.name.includes('Node.js'));
                if (!nodeCtx) {
                    console.log('No Node.js context found! Available:', ctxs.map(c => c.name));
                    ws.close(); resolve(null); return;
                }
                
                console.log('Using context:', nodeCtx.id, nodeCtx.name);
                
                id++;
                ws.send(JSON.stringify({
                    id,
                    method: 'Runtime.evaluate',
                    params: {
                        expression,
                        contextId: nodeCtx.id,
                        returnByValue: true
                    }
                }));
            }
            
            if (msg.id === 2) {
                console.log('Result:', JSON.stringify(msg.result, null, 2));
                ws.close();
                resolve(msg.result);
            }
        });
        
        ws.on('error', reject);
        setTimeout(() => { ws.close(); resolve(null); }, 10000);
    });
}

// Step 1: process에서 require를 만들 수 있는지 확인
console.log('=== Step 1: process 접근 확인 ===');
await runOnTarget(`(function() {
    const results = [];
    results.push('process: ' + typeof process);
    results.push('process.versions: ' + JSON.stringify(process.versions || {}));
    results.push('process.getBuiltinModule: ' + typeof process.getBuiltinModule);
    results.push('process.mainModule: ' + typeof process.mainModule);
    results.push('process._linkedBinding: ' + typeof process._linkedBinding);
    results.push('process.dlopen: ' + typeof process.dlopen);
    
    // Try getBuiltinModule
    if (typeof process.getBuiltinModule === 'function') {
        try {
            const h2 = process.getBuiltinModule('http2');
            results.push('http2 via getBuiltinModule: ' + (h2 ? 'YES' : 'null'));
        } catch(e) {
            results.push('http2 getBuiltinModule error: ' + e.message);
        }
        
        try {
            const mod = process.getBuiltinModule('module');
            results.push('module via getBuiltinModule: ' + (mod ? 'YES' : 'null'));
            if (mod) {
                const createRequire = mod.createRequire;
                results.push('createRequire: ' + typeof createRequire);
            }
        } catch(e) {
            results.push('module getBuiltinModule error: ' + e.message);
        }
    }
    
    return results.join('\\n');
})()`);
