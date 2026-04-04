/**
 * PID 50980 — net.Socket.write + http2.connect 둘 다 패치
 * 14912 포트로 가는 모든 바이트를 캡처
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');

function injectSocketPatch() {
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
                
                const patchCode = `(function() {
                    const net = process.getBuiltinModule('net');
                    const fsMod = process.getBuiltinModule('fs');
                    const pathMod = process.getBuiltinModule('path');
                    const DUMP_DIR = '${DUMP_DIR_ESC}';
                    
                    if (net.__neuronfs_sock_patched_v15) return 'ALREADY_PATCHED_v15';
                    net.__neuronfs_sock_patched_v15 = true;
                    
                    let captureCount = 0;
                    
                    function dumpData(label, data) {
                        captureCount++;
                        const ts = Date.now();
                        const fname = 'raw_' + ts + '_' + captureCount;
                        const binPath = pathMod.join(DUMP_DIR, fname + '.bin');
                        const metaPath = binPath + '.meta.json';
                        try {
                            fsMod.writeFileSync(binPath, Buffer.isBuffer(data) ? data : Buffer.from(data));
                            fsMod.writeFileSync(metaPath, JSON.stringify({
                                ts: new Date(ts).toISOString(),
                                label: label,
                                size: data.length,
                                source: 'pid50980_socket_v15'
                            }));
                        } catch {}
                    }
                    
                    // net.Socket.prototype.write 패치
                    const origWrite = net.Socket.prototype.write;
                    net.Socket.prototype.write = function(data, enc, cb) {
                        try {
                            const rport = this.remotePort;
                            const lport = this.localPort;
                            // 14912(LSP/gRPC to language server) 포트만 캡처
                            if (rport === 14912 || lport === 14912 ||
                                rport === 14905 || lport === 14905 ||
                                rport === 14906 || lport === 14906) {
                                if (data && data.length > 10) {
                                    const port = rport || lport || '?';
                                    dumpData('sock_write:' + port + ':pid50980', 
                                        Buffer.isBuffer(data) ? data : Buffer.from(data));
                                }
                            }
                        } catch {}
                        return origWrite.call(this, data, enc, cb);
                    };
                    
                    // Readable의 push(data)도 패치 — 수신 방향
                    const stream = process.getBuiltinModule('stream');
                    const origPush = stream.Readable.prototype.push;
                    stream.Readable.prototype.push = function(data) {
                        try {
                            if (data && data.length > 50 && this.remotePort) {
                                const rport = this.remotePort;
                                if (rport === 14912 || rport === 14905 || rport === 14906) {
                                    dumpData('sock_read:' + rport + ':pid50980',
                                        Buffer.isBuffer(data) ? data : Buffer.from(data));
                                }
                            }
                        } catch {}
                        return origPush.call(this, data);
                    };
                    
                    return 'PATCHED_v15_socket_ok_captures=' + captureCount;
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

const result = await injectSocketPatch();
const val = result && result.result ? result.result.value : (result ? result.value : null);
console.log('Patch value:', val);

if (val && val.includes('PATCHED') || val && val.includes('patched')) {
    console.log('\n=== PID 50980 Socket Patch SUCCESS ===');
    console.log('14912/14905/14906 포트 TCP 트래픽을 raw_*.bin으로 캡처합니다');
    
    // 감시
    let lastCount = 0;
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('raw_') && !f.endsWith('.meta.json'));
            if (files.length > lastCount) {
                for (const f of files.slice(lastCount)) {
                    const mp = path.join(DUMP_DIR, f + '.meta.json');
                    if (fs.existsSync(mp)) {
                        const m = JSON.parse(fs.readFileSync(mp, 'utf8'));
                        console.log(`🔴 RAW #${files.indexOf(f)+1}: ${f} (${m.size}B) | ${m.label}`);
                    }
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
    
    console.log('\n대기 중...\n');
} else {
    console.log('패치 실패');
    process.exit(1);
}
