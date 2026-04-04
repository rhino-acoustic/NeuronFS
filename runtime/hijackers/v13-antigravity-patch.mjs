/**
 * PID 50980 (Antigravity utility process) gRPC 패치 주입기
 * Extension Host가 아닌 Antigravity.exe 유틸리티 프로세스에 직접 패치
 * 목표: SendUserCascadeMessage 캡처
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';

const INSPECTOR_URL = 'ws://127.0.0.1:3787/839c74d4-9ab7-46a2-be50-f8617ee2e21a';
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');

// 먼저 프로세스가 http2를 쓰는지 확인 + gRPC 패치 주입
const ws = new WebSocket(INSPECTOR_URL);
let msgId = 0;

function send(method, params) {
    return new Promise((resolve, reject) => {
        const id = ++msgId;
        const handler = (data) => {
            const r = JSON.parse(data.toString());
            if (r.id === id) {
                ws.off('message', handler);
                resolve(r);
            }
        };
        ws.on('message', handler);
        ws.send(JSON.stringify({ id, method, params }));
        setTimeout(() => reject(new Error('timeout')), 10000);
    });
}

ws.on('open', async () => {
    console.log('Connected to PID 50980 inspector');
    
    // 1. http2 모듈 확인
    const check = await send('Runtime.evaluate', {
        expression: `(function() {
            try {
                const h2 = require('http2');
                const sessions = [];
                // Check if prototype is already patched
                if (h2.ClientHttp2Session && h2.ClientHttp2Session.prototype.__neuronfs_patched) {
                    return 'already_patched';
                }
                return 'http2_available';
            } catch(e) {
                return 'error: ' + e.message;
            }
        })()`
    });
    
    console.log('http2 check:', JSON.stringify(check.result));
    
    if (check.result && check.result.value && check.result.value.includes('error')) {
        console.log('http2 not available in this process');
        ws.close();
        process.exit(0);
    }
    
    // 2. gRPC 프로토타입 패치 주입
    const patchCode = `(function() {
        const http2 = require('http2');
        const fs = require('fs');
        const path = require('path');
        const DUMP_DIR = '${DUMP_DIR_ESC}';
        
        if (!fs.existsSync(DUMP_DIR)) {
            try { fs.mkdirSync(DUMP_DIR, { recursive: true }); } catch {}
        }
        
        let captureCount = 0;
        
        function dumpBin(data, label) {
            captureCount++;
            const ts = Date.now();
            const fname = 'ag_' + ts + '_' + captureCount;
            const binPath = path.join(DUMP_DIR, fname + '.bin');
            const metaPath = binPath + '.meta.json';
            try {
                fs.writeFileSync(binPath, Buffer.from(data));
                fs.writeFileSync(metaPath, JSON.stringify({
                    ts: new Date(ts).toISOString(),
                    label: label,
                    size: data.length,
                    source: 'antigravity_utility_50980'
                }));
            } catch {}
        }
        
        // Prototype-level h2 request patching
        const origRequest = http2.ClientHttp2Session.prototype.request;
        http2.ClientHttp2Session.prototype.request = function(headers, options) {
            const stream = origRequest.call(this, headers, options);
            const rpcPath = headers && headers[':path'] ? headers[':path'] : 'unknown';
            
            // Capture request headers
            try {
                const hdrsStr = JSON.stringify(headers || {});
                dumpBin(Buffer.from(hdrsStr), 'h2_hdrs:' + rpcPath);
            } catch {}
            
            // Capture request data (what we send)
            const origWrite = stream.write.bind(stream);
            const origEnd = stream.end.bind(stream);
            
            stream.write = function(data, encoding, cb) {
                if (data && data.length > 0) {
                    dumpBin(Buffer.from(data), 'h2_req:' + rpcPath);
                }
                return origWrite(data, encoding, cb);
            };
            
            stream.end = function(data, encoding, cb) {
                if (data && data.length > 0) {
                    dumpBin(Buffer.from(data), 'h2_req:' + rpcPath);
                }
                return origEnd(data, encoding, cb);
            };
            
            // Capture response data
            stream.on('data', (chunk) => {
                if (chunk && chunk.length > 5) {
                    dumpBin(Buffer.from(chunk), 'h2_res:' + rpcPath);
                }
            });
            
            return stream;
        };
        
        http2.ClientHttp2Session.prototype.__neuronfs_patched = true;
        
        // Also patch http2.connect to catch new sessions
        const origConnect = http2.connect;
        http2.connect = function(...args) {
            const session = origConnect.apply(this, args);
            return session;
        };
        
        return 'patched_ok_antigravity_v1_captures=' + captureCount;
    })()`;
    
    const patchResult = await send('Runtime.evaluate', {
        expression: patchCode,
        returnByValue: true
    });
    
    console.log('Patch result:', JSON.stringify(patchResult.result));
    
    if (patchResult.result && patchResult.result.value && patchResult.result.value.startsWith('patched_ok')) {
        console.log('\n🎯 PID 50980 (Antigravity utility) 패치 성공!');
        console.log('   이제 SendUserCascadeMessage가 캡처됩니다.');
        console.log('   채팅 메시지를 보내서 테스트하세요.');
        
        // 캡처 감시
        let lastCount = 0;
        setInterval(() => {
            try {
                const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('ag_') && !f.endsWith('.meta.json'));
                if (files.length > lastCount) {
                    const newFiles = files.slice(lastCount);
                    for (const f of newFiles) {
                        const metaPath = path.join(DUMP_DIR, f + '.meta.json');
                        if (fs.existsSync(metaPath)) {
                            const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
                            console.log(`🔥 AG캡처 #${files.indexOf(f) + 1}: ${f} (${meta.size}B) | ${meta.label}`);
                        }
                    }
                    lastCount = files.length;
                }
            } catch {}
        }, 2000);
        
    } else {
        console.log('패치 실패:', JSON.stringify(patchResult));
        ws.close();
        process.exit(1);
    }
});

ws.on('error', (e) => {
    console.log('WebSocket error:', e.message);
    process.exit(1);
});
