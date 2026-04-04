/**
 * NeuronFS v14 — PID 50980 (Antigravity Utility) gRPC 패치
 * process.getBuiltinModule()로 http2 접근하여 프로토타입 패치 주입
 * 목표: SendUserCascadeMessage, StartCascade, StreamAgentStateUpdates 캡처
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');

function injectPatch() {
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
                if (!nodeCtx) {
                    console.log('No Node.js context found!');
                    ws.close(); resolve(null); return;
                }
                
                console.log('Injecting into context:', nodeCtx.id, nodeCtx.name);
                
                const patchCode = `(function() {
                    const http2 = process.getBuiltinModule('http2');
                    const fsMod = process.getBuiltinModule('fs');
                    const pathMod = process.getBuiltinModule('path');
                    const DUMP_DIR = '${DUMP_DIR_ESC}';
                    
                    if (!fsMod.existsSync(DUMP_DIR)) {
                        try { fsMod.mkdirSync(DUMP_DIR, { recursive: true }); } catch {}
                    }
                    
                    let captureCount = 0;
                    
                    function dumpData(label, key, data) {
                        captureCount++;
                        const ts = Date.now();
                        const fname = 'chat_' + ts + '_' + captureCount;
                        const binPath = pathMod.join(DUMP_DIR, fname + '.bin');
                        const metaPath = binPath + '.meta.json';
                        try {
                            fsMod.writeFileSync(binPath, data);
                            fsMod.writeFileSync(metaPath, JSON.stringify({
                                ts: new Date(ts).toISOString(),
                                label: label,
                                size: data.length,
                                source: 'pid50980_chat_v14'
                            }));
                        } catch {}
                    }
                    
                    // === http2 프로토타입 패치 ===
                    try {
                        const tmpSession = http2.connect('https://127.0.0.1:1');
                        const h2Proto = Object.getPrototypeOf(tmpSession);
                        try { tmpSession.close(); } catch {}
                        try { tmpSession.destroy(); } catch {}
                        
                        if (h2Proto && typeof h2Proto.request === 'function' && !h2Proto.__neuronfs_chat_patched) {
                            h2Proto.__neuronfs_chat_patched = true;
                            const origRequest = h2Proto.request;
                            
                            h2Proto.request = function(headers, options) {
                                const stream = origRequest.call(this, headers, options);
                                const h2path = (headers && headers[':path']) || '/';
                                
                                // 송신 캡처
                                const origW = stream.write;
                                stream.write = function(data, enc, cb) {
                                    try {
                                        if (data && data.length > 5) {
                                            dumpData('h2_req:' + h2path, h2path,
                                                Buffer.isBuffer(data) ? data : Buffer.from(data));
                                        }
                                    } catch {}
                                    return origW.call(this, data, enc, cb);
                                };
                                
                                const origE = stream.end;
                                stream.end = function(data, enc, cb) {
                                    try {
                                        if (data && data.length > 5) {
                                            dumpData('h2_req:' + h2path, h2path,
                                                Buffer.isBuffer(data) ? data : Buffer.from(data));
                                        }
                                    } catch {}
                                    return origE.call(this, data, enc, cb);
                                };
                                
                                // 수신 캡처
                                stream.on('data', (chunk) => {
                                    try {
                                        if (chunk && chunk.length > 5) {
                                            dumpData('h2_res:' + h2path, h2path + ':res',
                                                Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                                        }
                                    } catch {}
                                });
                                
                                // 헤더 캡처
                                stream.on('response', (rh) => {
                                    try {
                                        dumpData('h2_hdrs:' + h2path, h2path + ':hdrs',
                                            Buffer.from(JSON.stringify(rh)));
                                    } catch {}
                                });
                                
                                return stream;
                            };
                            
                            return 'PATCHED_OK_v14_chat';
                        } else if (h2Proto && h2Proto.__neuronfs_chat_patched) {
                            return 'ALREADY_PATCHED_v14';
                        }
                    } catch(e) {
                        return 'PATCH_ERROR: ' + e.message;
                    }
                    return 'UNKNOWN_STATE';
                })()`;
                
                id++;
                ws.send(JSON.stringify({
                    id,
                    method: 'Runtime.evaluate',
                    params: {
                        expression: patchCode,
                        contextId: nodeCtx.id,
                        returnByValue: true
                    }
                }));
            }
            
            if (msg.id === 2) {
                console.log('Patch result:', JSON.stringify(msg.result));
                ws.close();
                resolve(msg.result);
            }
        });
        
        ws.on('error', reject);
        setTimeout(() => { ws.close(); resolve(null); }, 15000);
    });
}

const result = await injectPatch();

const val = result && result.result ? result.result.value : (result ? result.value : null);

if (val && (val.includes('PATCHED_OK') || val.includes('ALREADY_PATCHED'))) {
    console.log('\n====================================');
    console.log('  PID 50980 CHAT gRPC PATCHED!');
    console.log('  채팅 메시지가 chat_*.bin으로 캡처됩니다');
    console.log('====================================');
    
    // 캡처 감시
    let lastCount = 0;
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('chat_') && !f.endsWith('.meta.json'));
            if (files.length > lastCount) {
                const newFiles = files.slice(lastCount);
                for (const f of newFiles) {
                    const metaPath = path.join(DUMP_DIR, f + '.meta.json');
                    if (fs.existsSync(metaPath)) {
                        const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
                        console.log(`🔥 CHAT #${files.indexOf(f) + 1}: ${f} (${meta.size}B) | ${meta.label}`);
                    }
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
    
    console.log('\n🎯 대기 중... 채팅 메시지를 보내세요!\n');
} else {
    console.log('패치 실패:', JSON.stringify(result));
    process.exit(1);
}
