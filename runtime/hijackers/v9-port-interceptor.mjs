/**
 * NeuronFS v9 — Extension Server Port 1461 트래픽 인터셉터
 * 
 * Extension Host → language_server (port 1461) 사이의
 * HTTP/gRPC 통신을 가로채서 AI 컨텍스트를 추출.
 * 
 * CDP Runtime.evaluate로 fetch/http.request를 후킹하여
 * extension_server_port로 나가는 모든 요청의 body를 덤프.
 */
import http from 'http';
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';

const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\v9_captures';
const LOG_FILE = path.join(DUMP_DIR, 'v9_interceptor.log');
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

const DUMP_DIR_ESCAPED = DUMP_DIR.replace(/\\/g, '\\\\');

// 핵심 패치: net.Socket.prototype.write 후킹하되,
// 이번에는 포트 1461(extension_server_port)만 필터링!
// 이 포트의 트래픽 = Extension Host ↔ language_server 사이의 AI gRPC 통신
const MEMORY_PATCH_CODE = `
(() => {
    if (globalThis.__neuronfs_v9_port) return 'already_patched';
    globalThis.__neuronfs_v9_port = true;

    const fs = process.getBuiltinModule('fs');
    const path = process.getBuiltinModule('path');
    const net = process.getBuiltinModule('net');
    const stream = process.getBuiltinModule('stream');
    const DUMP = '${DUMP_DIR_ESCAPED}';
    let seq = 0;

    function dumpPayload(label, data) {
        seq++;
        const ts = Date.now();
        try {
            const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
            const fname = path.join(DUMP, 'ext_' + ts + '_' + seq);
            
            // 바이너리 + 텍스트 둘 다 저장
            fs.writeFileSync(fname + '.bin', buf);
            
            // UTF-8로도 저장 (읽기 쉽게)
            const text = buf.toString('utf8');
            fs.writeFileSync(fname + '.txt', text);
            
            const meta = { 
                label, 
                size: buf.length, 
                ts: new Date().toISOString(), 
                seq,
                preview: text.substring(0, 200).replace(/[\\x00-\\x1f]/g, '.')
            };
            fs.writeFileSync(fname + '.meta.json', JSON.stringify(meta, null, 2));
        } catch(e) {}
    }

    // ========== 전략 1: net.Socket.prototype.write — 포트 1461만 ==========
    const origWrite = net.Socket.prototype.write;
    net.Socket.prototype.write = function(data, encoding, cb) {
        try {
            const port = this.remotePort;
            // 1461 = extension_server_port (ExtHost ↔ language_server)
            if (port === 1461 || port === 2037) {
                const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
                if (buf.length > 100) {
                    dumpPayload('WRITE_PORT_' + port, buf);
                }
            }
        } catch(e) {}
        return origWrite.call(this, data, encoding, cb);
    };

    // ========== 전략 2: Readable.push — 포트 1461 응답 ==========
    const origPush = stream.Readable.prototype.push;
    stream.Readable.prototype.push = function(chunk, encoding) {
        try {
            if (chunk && chunk.length > 100) {
                const port = this.remotePort;
                if (port === 1461 || port === 2037) {
                    const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
                    dumpPayload('READ_PORT_' + port, buf);
                }
            }
        } catch(e) {}
        return origPush.call(this, chunk, encoding);
    };

    // ========== 전략 3: globalThis.fetch 후킹 ==========
    if (typeof globalThis.fetch === 'function') {
        const origFetch = globalThis.fetch;
        globalThis.fetch = async function(input, init) {
            try {
                const url = typeof input === 'string' ? input : input?.url || '';
                if (url.includes('localhost:1461') || url.includes('127.0.0.1:1461') ||
                    url.includes('localhost:2037') || url.includes('127.0.0.1:2037') ||
                    url.includes('cloudcode') || url.includes('googleapis')) {
                    const body = init?.body;
                    if (body && typeof body === 'string' && body.length > 100) {
                        dumpPayload('FETCH_' + url.substring(0, 80), body);
                    }
                }
            } catch(e) {}
            return origFetch.call(globalThis, input, init);
        };
    }

    // ========== 전략 4: 모든 소켓 connect 이벤트 감시 ==========
    let socketReport = 'SOCKET_SCAN:\\n';
    const origConnect = net.Socket.prototype.connect;
    net.Socket.prototype.connect = function(...args) {
        const result = origConnect.apply(this, args);
        try {
            this.once('connect', () => {
                const rPort = this.remotePort;
                const rAddr = this.remoteAddress;
                if (rPort === 1461 || rPort === 2037) {
                    // 이 소켓이 extension_server와 연결됨.
                    // 모든 데이터를 캡처하기 위해 'data' 이벤트 리스너 추가
                    this.on('data', (data) => {
                        if (data.length > 100) {
                            dumpPayload('SOCKET_DATA_' + rPort, data);
                        }
                    });
                }
            });
        } catch(e) {}
        return result;
    };

    return 'patched_ok_v9_port_interceptor';
})()
`;

async function injectPatch(wsUrl, label) {
    return new Promise((resolve, reject) => {
        log(`[${label}] WebSocket 연결 중: ${wsUrl}`);
        const ws = new WebSocket(wsUrl);
        let msgId = 1;
        const pending = new Map();

        ws.on('open', async () => {
            log(`[${label}] 연결 성공!`);

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

                log(`[${label}] v9 포트 인터셉터 주입 중...`);
                const result = await call('Runtime.evaluate', {
                    expression: MEMORY_PATCH_CODE,
                    returnByValue: true,
                    timeout: 8000
                });

                const value = result?.result?.value;
                const errDesc = result?.result?.description;
                log(`[${label}] 결과: ${JSON.stringify(value || errDesc)}`);

                if (value && value.startsWith('patched_ok')) {
                    log(`[${label}] ✅ v9 패치 성공!`);
                } else if (value === 'already_patched') {
                    log(`[${label}] ⚠️ 이미 패치됨`);
                } else {
                    log(`[${label}] ❌ 패치 실패: ${JSON.stringify(result?.result)}`);
                }

                ws.close();
                resolve(value);
            } catch(e) {
                log(`[${label}] 에러: ${e.message}`);
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

function watchDumps() {
    log(`📁 덤프 감시: ${DUMP_DIR}`);
    let lastCount = 0;

    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('ext_') && f.endsWith('.txt'));
            if (files.length > lastCount) {
                const newFiles = files.slice(lastCount);
                for (const f of newFiles) {
                    const full = path.join(DUMP_DIR, f);
                    const stat = fs.statSync(full);
                    log(`💎 캡처: ${f} (${stat.size} bytes)`);

                    // 메타 확인
                    const metaPath = full.replace('.txt', '.meta.json');
                    if (fs.existsSync(metaPath)) {
                        try {
                            const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
                            log(`   📌 ${meta.label}`);
                            log(`   📄 ${meta.preview}`);
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
    log('============================================================');
    log('🧠 NeuronFS v9 — Extension Server Port 인터셉터');
    log('============================================================');
    log('타겟: Extension Host → language_server (port 1461/2037)');
    log('전략 1: net.Socket.write → 포트 1461 필터');
    log('전략 2: Readable.push → 포트 1461 응답');
    log('전략 3: fetch() 후킹');
    log('전략 4: Socket connect 감시');
    log('============================================================');

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
                    if (v && (v.startsWith('patched_ok') || v === 'already_patched')) patchedCount++;
                } catch(e) {
                    log(`패치 실패: ${e.message}`);
                }
            }
        }
    }

    log('');
    log(`✅ 패치 완료: ${patchedCount}개 Extension Host`);
    log(`📁 덤프 디렉토리: ${DUMP_DIR}`);

    if (patchedCount > 0) {
        watchDumps();
        log('');
        log('🔥 대기 중... AI 채팅 시 포트 1461 트래픽이 자동 캡처됩니다!');
    } else {
        log('❌ 패치된 프로세스 없음 — 종료');
        process.exit(1);
    }
}

main().catch(e => { log(`치명적 에러: ${e.message}`); process.exit(1); });
