/**
 * v17 — PID 50980 기존 h2 세션 + 스트림 패치
 * 이미 생성된 h2 세션의 활성 스트림에 write/data 후킹 주입
 * 
 * 핵심: gRPC streaming RPC(SendUserCascadeMessage)는 
 * 세션 시작 시 한 번만 request()를 호출하고, 이후 stream.write()로 메시지를 보냄
 * → 패치 이전에 생성된 스트림을 찾아서 후킹해야 함
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');

function injectV17() {
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
                    const http2 = process.getBuiltinModule('http2');
                    const net = process.getBuiltinModule('net');
                    const fsMod = process.getBuiltinModule('fs');
                    const pathMod = process.getBuiltinModule('path');
                    const DUMP_DIR = '${DUMP_DIR_ESC}';
                    
                    let captureCount = 0;
                    
                    function dumpData(label, data) {
                        captureCount++;
                        const ts = Date.now();
                        const fname = 'v17_' + ts + '_' + captureCount;
                        const binPath = pathMod.join(DUMP_DIR, fname + '.bin');
                        const metaPath = binPath + '.meta.json';
                        try {
                            fsMod.writeFileSync(binPath, Buffer.isBuffer(data) ? data : Buffer.from(data));
                            fsMod.writeFileSync(metaPath, JSON.stringify({
                                ts: new Date(ts).toISOString(),
                                label: label,
                                size: data.length,
                                source: 'pid50980_v17_existing_streams'
                            }));
                        } catch(e) {}
                    }
                    
                    // 방법 1: http2.connect를 감싸서 반환되는 세션 추적
                    // 이미 생성된 세션은 접근 불가 — 대신 모듈 캐시에서 찾기
                    
                    // 방법 2: net.Socket의 모든 인스턴스에서 _handle을 통해 h2 세션 접근
                    // Electron은 내부적으로 여러 소켓을 유지
                    
                    // 방법 3: 프로토타입에 write() getter/setter 트랩 설치
                    // stream.Writable.prototype.write를 프록시로 교체
                    
                    const stream = process.getBuiltinModule('stream');
                    
                    if (stream.Writable.prototype.__neuronfs_v17_write) {
                        return 'ALREADY_PATCHED_v17';
                    }
                    
                    const origWritableWrite = stream.Writable.prototype.write;
                    stream.Writable.prototype.__neuronfs_v17_write = true;
                    
                    stream.Writable.prototype.write = function(chunk, encoding, cb) {
                        try {
                            // h2 스트림인지 확인 (sentHeaders가 있으면 h2 stream)
                            if (this.sentHeaders || this.session) {
                                const path = (this.sentHeaders && this.sentHeaders[':path']) || 
                                             (this._headers && this._headers[':path']) || '';
                                if (chunk && chunk.length > 5) {
                                    dumpData('v17_h2write:' + path, 
                                        Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                                }
                            }
                        } catch {}
                        return origWritableWrite.call(this, chunk, encoding, cb);
                    };
                    
                    // Readable.prototype에서 push 감지 — h2 응답
                    const origPush = stream.Readable.prototype.push;
                    if (!stream.Readable.prototype.__neuronfs_v17_push) {
                        stream.Readable.prototype.__neuronfs_v17_push = true;
                        stream.Readable.prototype.push = function(chunk) {
                            try {
                                if (chunk && chunk.length > 50 && (this.sentHeaders || this.session)) {
                                    const path = (this.sentHeaders && this.sentHeaders[':path']) || '';
                                    dumpData('v17_h2push:' + path,
                                        Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                                }
                            } catch {}
                            return origPush.call(this, chunk);
                        };
                    }
                    
                    return 'PATCHED_v17_writable_ok_captures=' + captureCount;
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

const result = await injectV17();
const val = result && result.result ? result.result.value : null;
console.log('Value:', val);

if (val && val.includes('PATCHED') || val && val.includes('patched')) {
    console.log('\n=== v17 Writable.prototype.write Patch ACTIVE ===');
    console.log('기존+신규 h2 스트림의 모든 write/push를 v17_*.bin으로 캡처');
    
    let lastCount = 0;
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('v17_') && !f.endsWith('.meta.json'));
            if (files.length > lastCount) {
                for (const f of files.slice(lastCount, lastCount + 20)) {
                    const mp = path.join(DUMP_DIR, f + '.meta.json');
                    if (fs.existsSync(mp)) {
                        const m = JSON.parse(fs.readFileSync(mp, 'utf8'));
                        console.log(`🟡 V17 #${files.indexOf(f)+1}: ${f} (${m.size}B) | ${m.label}`);
                    }
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
    
    console.log('\n대기 중... 채팅 메시지를 보내세요!\n');
} else {
    console.log('패치 실패:', val);
    process.exit(1);
}
