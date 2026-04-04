/**
 * PID 50980 — Named Pipe 트래픽 캡처
 * \\.\pipe\server_572e21ccf60c9137 으로 가는 모든 write를 가로챕니다
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');

function injectPipePatch() {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(INSPECTOR_URL);
        const ctxs = [];
        let id = 0;
        let enabled = false;
        
        ws.on('open', () => {
            id++;
            ws.send(JSON.stringify({ id, method: 'Runtime.enable' }));
        });
        
        ws.on('message', async (d) => {
            const msg = JSON.parse(d.toString());
            
            if (msg.method === 'Runtime.executionContextCreated') {
                ctxs.push(msg.params.context);
            }
            
            if (msg.id === 1 && !enabled) {
                enabled = true;
                await new Promise(r => setTimeout(r, 500));
                
                const nodeCtx = ctxs.find(c => c.name && c.name.includes('Node.js'));
                if (!nodeCtx) { ws.close(); resolve(null); return; }
                
                console.log('Context:', nodeCtx.id, nodeCtx.name);
                
                // Named Pipe 연결 찾기 + write 패치
                const patchCode = `(function() {
                    const net = process.getBuiltinModule('net');
                    const fsMod = process.getBuiltinModule('fs');
                    const pathMod = process.getBuiltinModule('path');
                    const DUMP_DIR = '${DUMP_DIR_ESC}';
                    
                    if (net.__neuronfs_pipe_patched_v16) return 'ALREADY_PATCHED_v16';
                    net.__neuronfs_pipe_patched_v16 = true;
                    
                    let captureCount = 0;
                    
                    function dumpData(label, data) {
                        captureCount++;
                        const ts = Date.now();
                        const fname = 'pipe_' + ts + '_' + captureCount;
                        const binPath = pathMod.join(DUMP_DIR, fname + '.bin');
                        const metaPath = binPath + '.meta.json';
                        try {
                            fsMod.writeFileSync(binPath, Buffer.isBuffer(data) ? data : Buffer.from(data));
                            fsMod.writeFileSync(metaPath, JSON.stringify({
                                ts: new Date(ts).toISOString(),
                                label: label,
                                size: data.length,
                                source: 'pid50980_pipe_v16'
                            }));
                        } catch(e) {}
                    }
                    
                    // net.Socket.prototype.write 패치 — 모든 IPC/pipe 트래픽 캡처
                    const origWrite = net.Socket.prototype.write;
                    net.Socket.prototype.write = function(data, enc, cb) {
                        try {
                            // pipe 연결인 경우 캡처 (remotePort === undefined = IPC pipe)
                            if (this.remotePort === undefined && data && data.length > 10) {
                                dumpData('pipe_write:ipc', 
                                    Buffer.isBuffer(data) ? data : Buffer.from(data));
                            }
                        } catch {}
                        return origWrite.call(this, data, enc, cb);
                    };
                    
                    // net.connect 패치 — 새 pipe 연결도 캡처
                    const origConnect = net.connect;
                    net.connect = function(...args) {
                        const sock = origConnect.apply(this, args);
                        sock.on('data', (chunk) => {
                            try {
                                if (chunk && chunk.length > 10) {
                                    dumpData('pipe_read:ipc',
                                        Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                                }
                            } catch {}
                        });
                        return sock;
                    };
                    
                    // 기존 연결에서의 읽기도 패치
                    const stream = process.getBuiltinModule('stream');
                    const origPush = stream.Readable.prototype.push;
                    if (!stream.Readable.prototype.__neuronfs_push_patched_v16) {
                        stream.Readable.prototype.__neuronfs_push_patched_v16 = true;
                        stream.Readable.prototype.push = function(data) {
                            try {
                                // pipe (no remotePort) + 큰 데이터만
                                if (data && data.length > 50 && 
                                    this.remotePort === undefined && 
                                    typeof this._handle !== 'undefined') {
                                    dumpData('pipe_push:ipc',
                                        Buffer.isBuffer(data) ? data : Buffer.from(data));
                                }
                            } catch {}
                            return origPush.call(this, data);
                        };
                    }
                    
                    return 'PATCHED_v16_pipe_ok';
                })()`;
                
                id++;
                ws.send(JSON.stringify({
                    id,
                    method: 'Runtime.evaluate',
                    params: { expression: patchCode, contextId: nodeCtx.id, returnByValue: true }
                }));
            }
            
            if (msg.id === 2) {
                console.log('Result:', JSON.stringify(msg.result));
                ws.close();
                resolve(msg.result);
            }
        });
        
        ws.on('error', reject);
        setTimeout(() => { ws.close(); resolve(null); }, 15000);
    });
}

const result = await injectPipePatch();
const val = result && result.result ? result.result.value : null;
console.log('Value:', val);

if (val && val.includes('PATCHED')) {
    console.log('\n=== Named Pipe Capture ACTIVE ===');
    console.log('pipe_*.bin = PID 50980 <-> Language Server IPC');
    
    let lastCount = 0;
    setInterval(() => {
        try {
            const dir = DUMP_DIR;
            const files = fs.readdirSync(dir).filter(f => f.startsWith('pipe_') && !f.endsWith('.meta.json'));
            if (files.length > lastCount) {
                for (const f of files.slice(lastCount)) {
                    const mp = path.join(dir, f + '.meta.json');
                    if (fs.existsSync(mp)) {
                        const m = JSON.parse(fs.readFileSync(mp, 'utf8'));
                        console.log(`🟣 PIPE #${files.indexOf(f)+1}: ${f} (${m.size}B) | ${m.label}`);
                    }
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
    
    console.log('\n채팅 메시지를 보내면 pipe_*.bin에 캡처됩니다!\n');
} else {
    console.log('패치 실패');
    process.exit(1);
}
