/**
 * NeuronFS v9 — 완전 양방향 메모리 탈취
 * 
 * Extension Host V8 인스펙터에 연결 → Runtime.evaluate로
 * net.Socket.write + Readable.push + http2.connect + tls + http를 후킹.
 * 
 * v8과의 차이: gRPC-Web (HTTPS/H2) 트래픽도 캡처.
 * Language Server 포트 14905(gRPC HTTPS)의 plaintext protobuf 바이너리 확보.
 */
import http from 'http';
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const DUMP_DIR = path.join(INBOX, 'cdp_captures');
const LOG_FILE = path.join(INBOX, 'v9_memory_hijack.log');
fs.mkdirSync(DUMP_DIR, { recursive: true });

function log(msg) {
    const line = `[${new Date().toISOString()}] ${msg}`;
    console.log(line);
    fs.appendFileSync(LOG_FILE, line + '\n');
}

function getJson(url) {
    return new Promise((resolve, reject) => {
        const req = http.get(url, { timeout: 2000 }, (res) => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)); } catch(e) { reject(e); }
            });
        });
        req.on('error', reject);
        req.on('timeout', () => { req.destroy(); reject(new Error('timeout')); });
    });
}

function findExtHostPids() {
    try {
        const out = execSync(
            "Get-WmiObject Win32_Process | Where-Object { $_.CommandLine -like '*experimental-network-inspection*' } | ForEach-Object { $_.ProcessId }",
            { shell: 'powershell.exe', encoding: 'utf8' }
        );
        return out.trim().split(/\r?\n/).map(s => parseInt(s.trim())).filter(n => !isNaN(n));
    } catch(e) {
        log(`ExtHost PID 탐색 실패: ${e.message}`);
        return [];
    }
}

function activateInspector(pid) {
    try {
        log(`PID ${pid}에 _debugProcess 시그널 전송...`);
        execSync(`node -e "process._debugProcess(${pid})"`, { encoding: 'utf8', timeout: 5000 });
        log(`PID ${pid} 인스펙터 활성화 완료`);
        return true;
    } catch(e) {
        log(`PID ${pid} 시그널 실패: ${e.message}`);
        return false;
    }
}

async function findInspectorPort(pid) {
    try {
        const out = execSync(
            `netstat -ano | findstr "LISTENING" | findstr "${pid}"`,
            { encoding: 'utf8', timeout: 5000 }
        );
        const ports = [];
        for (const line of out.split('\n')) {
            const match = line.match(/127\.0\.0\.1:(\d+)\s+0\.0\.0\.0/);
            if (match) ports.push(parseInt(match[1]));
        }
        for (const port of ports) {
            try {
                const data = await getJson(`http://127.0.0.1:${port}/json`);
                if (Array.isArray(data)) {
                    log(`PID ${pid} 인스펙터 포트: ${port}`);
                    return { port, targets: data };
                }
            } catch {}
        }
    } catch(e) {
        log(`PID ${pid} 포트 스캔 실패: ${e.message}`);
    }
    return null;
}

// ====================================================================
// v9 메모리 패치: net.Socket + Readable.push + http2.connect 후킹
// gRPC-Web (HTTPS port 14905) 트래픽의 plaintext protobuf 까지 캡처
// ====================================================================
const DUMP_DIR_ESCAPED = DUMP_DIR.replace(/\\/g, '\\\\');

const MEMORY_PATCH_CODE = `
(() => {
    if (globalThis.__neuronfs_v12_patched) return 'already_patched_v12';
    globalThis.__neuronfs_v12_patched = true;

    // 이전 v8 패치 정리
    if (globalThis.__neuronfs_v8_patched) {
        try {
            const net0 = process.getBuiltinModule('net');
            const stream0 = process.getBuiltinModule('stream');
            if (globalThis.__neuronfs_orig_socket_write) {
                net0.Socket.prototype.write = globalThis.__neuronfs_orig_socket_write;
            }
            if (globalThis.__neuronfs_orig_readable_push) {
                stream0.Readable.prototype.push = globalThis.__neuronfs_orig_readable_push;
            }
        } catch(e) {}
        delete globalThis.__neuronfs_v8_patched;
    }

    const fs = process.getBuiltinModule('fs');
    const pathMod = process.getBuiltinModule('path');
    const net = process.getBuiltinModule('net');
    const http2 = process.getBuiltinModule('http2');
    const { Readable } = process.getBuiltinModule('stream');
    const DUMP = '${DUMP_DIR_ESCAPED}';
    let seq = 0;

    function dumpData(label, url, chunk) {
        seq++;
        const ts = Date.now();
        try {
            const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
            const fname = pathMod.join(DUMP, 'mem_' + ts + '_' + seq + '.bin');
            fs.writeFileSync(fname, buf);
            const meta = { label, url, size: buf.length, ts: new Date().toISOString(), seq };
            fs.writeFileSync(fname + '.meta.json', JSON.stringify(meta));
        } catch(e) {}
    }

    function isNoisy(buf) {
        try {
            const preview = buf.toString('utf8', 0, Math.min(500, buf.length));
            if (preview.includes('Heartbeat') || preview.includes('didChangeWatched')) return true;
            return false;
        } catch { return false; }
    }

    function isAiRelevant(buf) {
        try {
            const preview = buf.toString('utf8', 0, Math.min(1000, buf.length));
            return preview.includes('cascade') || preview.includes('Cascade') ||
                   preview.includes('chat') || preview.includes('Chat') ||
                   preview.includes('user_information') || preview.includes('conversation') ||
                   preview.includes('SendUser') || preview.includes('trajectory') ||
                   preview.includes('completion') || preview.includes('getChat') ||
                   preview.includes('cortex') || preview.includes('Cortex') ||
                   preview.includes('model') || preview.includes('prompt') ||
                   preview.includes('streaming') || preview.includes('"text"');
        } catch { return false; }
    }

    // ========== 1. net.Socket.write — LSP JSON-RPC (포트 14912) ==========
    const origWrite = net.Socket.prototype.write;
    globalThis.__neuronfs_orig_socket_write = origWrite;
    net.Socket.prototype.write = function(data, encoding, callback) {
        try {
            if (data && data.length > 100) {
                const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
                if (!isNoisy(buf)) {
                    // 큰 데이터 또는 AI 관련 키워드가 포함된 데이터만 캡처
                    if (data.length > 5000 || isAiRelevant(buf)) {
                        dumpData(
                            'sock_write:' + (this.remoteAddress||'?') + ':' + (this.remotePort||0),
                            (this.remoteAddress||'') + ':' + (this.remotePort||0),
                            buf
                        );
                    }
                }
            }
        } catch(e) {}
        return origWrite.call(this, data, encoding, callback);
    };

    // ========== 2. Readable.push — 인바운드 (모든 소켓) ==========
    const origPush = Readable.prototype.push;
    globalThis.__neuronfs_orig_readable_push = origPush;
    Readable.prototype.push = function(chunk, encoding) {
        try {
            if (chunk && chunk.length > 1000 && this.remoteAddress) {
                const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
                if (!isNoisy(buf)) {
                    dumpData(
                        'sock_read:' + (this.remoteAddress||'?') + ':' + (this.remotePort||0),
                        (this.remoteAddress||'') + ':' + (this.remotePort||0) + ':in',
                        buf
                    );
                }
            }
        } catch(e) {}
        return origPush.call(this, chunk, encoding);
    };

    // ========== 3. ClientHttp2Session.prototype.request — 프로토타입 패치 ==========
    // 모든 기존+미래 h2 세션에 적용됨 (프로토타입 공유)
    try {
        // 임시 세션으로 프로토타입 접근 후 즉시 닫기
        const tmpSession = http2.connect('https://127.0.0.1:1');
        const h2Proto = Object.getPrototypeOf(tmpSession);
        try { tmpSession.close(); } catch {}
        try { tmpSession.destroy(); } catch {}
        
        if (h2Proto && typeof h2Proto.request === 'function' && !h2Proto.__neuronfs_req_patched) {
            h2Proto.__neuronfs_req_patched = true;
            const origRequest = h2Proto.request;
            
            h2Proto.request = function(headers, options) {
                const stream = origRequest.call(this, headers, options);
                const h2path = (headers && headers[':path']) || '/';
                const authority = (this.originSet && this.originSet[0]) || 'h2';
                
                // 송신: write
                const origSW = stream.write;
                stream.write = function(data, enc, cb) {
                    try {
                        if (data && data.length > 10) {
                            dumpData('h2_req:' + h2path, authority + h2path,
                                Buffer.isBuffer(data) ? data : Buffer.from(data));
                        }
                    } catch(e) {}
                    return origSW.call(this, data, enc, cb);
                };
                
                // 송신: end(data)
                const origSE = stream.end;
                stream.end = function(data, enc, cb) {
                    try {
                        if (data && data.length > 10) {
                            dumpData('h2_req_end:' + h2path, authority + h2path,
                                Buffer.isBuffer(data) ? data : Buffer.from(data));
                        }
                    } catch(e) {}
                    return origSE.call(this, data, enc, cb);
                };
                
                // 수신: data
                stream.on('data', (chunk) => {
                    try {
                        if (chunk && chunk.length > 10) {
                            dumpData('h2_res:' + h2path, authority + h2path + ':res',
                                Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
                        }
                    } catch(e) {}
                });
                
                // 수신: response 헤더
                stream.on('response', (rh) => {
                    try {
                        dumpData('h2_hdrs:' + h2path, authority + h2path + ':hdrs',
                            Buffer.from(JSON.stringify(rh)));
                    } catch(e) {}
                });
                
                return stream;
            };
        }
    } catch(e) {}

    // ========== 4. http/https request 후킹 — gRPC-Web (HTTP/1.1) ==========
    // Antigravity가 grpc-web을 HTTP/1.1 POST로 보낼 수도 있음
    const httpMod = process.getBuiltinModule('http');
    const httpsMod = process.getBuiltinModule('https');

    function wrapHttpRequest(mod, modName) {
        const origRequest = mod.request;
        mod.request = function(urlOrOpts, opts, cb) {
            const req = origRequest.call(this, urlOrOpts, opts, cb);
            const reqPath = (typeof urlOrOpts === 'string') ? urlOrOpts : (urlOrOpts && urlOrOpts.path ? urlOrOpts.path : '/');
            const reqHost = (typeof urlOrOpts === 'object' && urlOrOpts.hostname) ? urlOrOpts.hostname : '';

            // 요청 바디 캡처
            const origReqWrite = req.write.bind(req);
            req.write = function(data, enc, cb2) {
                try {
                    if (data && data.length > 10) {
                        const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
                        dumpData(modName + '_req:' + reqPath, reqHost + reqPath, buf);
                    }
                } catch(e) {}
                return origReqWrite(data, enc, cb2);
            };

            const origReqEnd = req.end.bind(req);
            req.end = function(data, enc, cb2) {
                try {
                    if (data && data.length > 10) {
                        const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
                        dumpData(modName + '_req_end:' + reqPath, reqHost + reqPath, buf);
                    }
                } catch(e) {}
                return origReqEnd(data, enc, cb2);
            };

            // 응답 바디 캡처
            req.on('response', (res) => {
                res.on('data', (chunk) => {
                    try {
                        if (chunk && chunk.length > 10) {
                            const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
                            dumpData(modName + '_res:' + reqPath, reqHost + reqPath + ':res', buf);
                        }
                    } catch(e) {}
                });
            });

            return req;
        };
    }

    wrapHttpRequest(httpMod, 'http');
    wrapHttpRequest(httpsMod, 'https');

    return 'patched_ok_v11_prototype_h2';
})()
`;

async function injectPatch(wsUrl, label) {
    return new Promise((resolve, reject) => {
        log(`[${label}] WebSocket 연결 중: ${wsUrl}`);
        const ws = new WebSocket(wsUrl);
        let msgId = 1;
        const pending = new Map();

        ws.on('open', async () => {
            log(`[${label}] 연결 성공! Runtime.enable 실행...`);
            
            const call = (method, params) => new Promise((res, rej) => {
                const id = msgId++;
                const t = setTimeout(() => { pending.delete(id); rej(new Error('timeout')); }, 10000);
                pending.set(id, { resolve: r => { clearTimeout(t); res(r); }, reject: e => { clearTimeout(t); rej(e); } });
                ws.send(JSON.stringify({ id, method, params }));
            });

            ws.on('message', (raw) => {
                try {
                    const data = JSON.parse(raw.toString());
                    if (data.id !== undefined && pending.has(data.id)) {
                        const { resolve: res, reject: rej } = pending.get(data.id);
                        pending.delete(data.id);
                        if (data.error) rej(data.error);
                        else res(data.result || data);
                    }
                } catch {}
            });

            try {
                await call('Runtime.enable', {});
                await new Promise(r => setTimeout(r, 300));

                log(`[${label}] v9 메모리 패치 주입 중...`);
                const result = await call('Runtime.evaluate', {
                    expression: MEMORY_PATCH_CODE,
                    returnByValue: true
                });

                const value = result?.result?.value;
                log(`[${label}] 패치 결과: ${JSON.stringify(value)}`);

                if (value && value.startsWith('patched_ok')) {
                    log(`[${label}] ✅ v9 메모리 패치 성공! net.Socket + http2 + http/https 후킹 완료`);
                } else if (value === 'already_patched_v9') {
                    log(`[${label}] ⚠️ v9 이미 패치됨 — 스킵`);
                } else {
                    log(`[${label}] ❌ 패치 실패: ${JSON.stringify(result)}`);
                }

                ws.close();
                resolve(value);
            } catch(e) {
                log(`[${label}] 패치 에러: ${e.message}`);
                ws.close();
                reject(e);
            }
        });

        ws.on('error', (e) => {
            log(`[${label}] WS 에러: ${e.message}`);
            reject(e);
        });
    });
}

// 파일 감시: 덤프 디렉토리에 새 파일이 생기면 로그
function watchDumps() {
    log(`📁 덤프 디렉토리 감시 시작: ${DUMP_DIR}`);
    let lastCount = 0;
    let h2Count = 0;
    
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('mem_'));
            if (files.length > lastCount) {
                const newFiles = files.slice(lastCount);
                for (const f of newFiles) {
                    if (f.endsWith('.meta.json')) continue;
                    const full = path.join(DUMP_DIR, f);
                    const stat = fs.statSync(full);
                    
                    // meta 파일이 있으면 내용 표시
                    const metaPath = full + '.meta.json';
                    if (fs.existsSync(metaPath)) {
                        try {
                            const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
                            const isH2 = meta.label.startsWith('h2_') || meta.label.startsWith('http');
                            if (isH2) {
                                h2Count++;
                                log(`🔥 gRPC/HTTP 캡처 #${h2Count}: ${f} (${stat.size}B) | ${meta.label}`);
                            } else {
                                log(`💎 캡처: ${f} (${stat.size}B) | ${meta.label}`);
                            }
                        } catch {}
                    }
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
}

async function main() {
    log('');
    log('======================================================');
    log('🧠 NeuronFS v9 — 완전 양방향 메모리 탈취');
    log('   net.Socket + Readable.push + http2 + http/https');
    log('======================================================');
    
    const pids = findExtHostPids();
    if (pids.length === 0) {
        log('❌ Extension Host 프로세스 미발견');
        process.exit(1);
    }
    log(`Extension Host PID: ${pids.join(', ')}`);

    let patchedCount = 0;

    for (const pid of pids) {
        activateInspector(pid);
        await new Promise(r => setTimeout(r, 1500));

        const result = await findInspectorPort(pid);
        if (!result) {
            log(`PID ${pid}: 인스펙터 포트 미발견 — 건너뜀`);
            continue;
        }

        const { port, targets } = result;
        for (const t of targets) {
            if (t.webSocketDebuggerUrl) {
                try {
                    const v = await injectPatch(t.webSocketDebuggerUrl, `ExtHost-${pid}:${port}`);
                    if (v && (v.startsWith('patched_ok') || v === 'already_patched_v9')) patchedCount++;
                } catch(e) {
                    log(`패치 실패: ${e.message}`);
                }
            }
        }
    }

    log('');
    log(`======================================================`);
    log(`✅ v9 패치 완료: ${patchedCount}개 Extension Host`);
    log(`📁 덤프 디렉토리: ${DUMP_DIR}`);
    log(`   🔥 h2_req/h2_res = gRPC-Web protobuf`);
    log(`   💎 sock_write/sock_read = LSP JSON-RPC`);
    log(`======================================================`);

    if (patchedCount > 0) {
        watchDumps();
        log('');
        log('🔥 대기 중... AI 채팅을 보내면 h2_*.bin (gRPC protobuf) 파일이 생성됩니다!');
    } else {
        log('❌ 패치된 프로세스 없음 — 종료');
        process.exit(1);
    }
}

main().catch(e => { log(`치명적 에러: ${e.message}`); process.exit(1); });
