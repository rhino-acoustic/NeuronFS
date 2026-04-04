/**
 * NeuronFS v8 — AI 컨텍스트 엔벨로프 메모리 탈취
 * 
 * 전략: Extension Host의 require() 캐시를 탐색하여
 * AI 컨텍스트를 조립하는 서비스 함수를 찾고 후킹.
 * 소켓 레벨이 아닌 APPLICATION 레벨에서 XML 탈취.
 * 
 * 동시에 JSON.stringify를 후킹하여 10KB 이상의 
 * JSON 직렬화를 모두 캡처 — gRPC 직렬화 직전 단계.
 */
import http from 'http';
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';

// 덤프를 Temp로 보내서 워크스페이스 감시 재귀 차단
const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\v8_captures';
const LOG_FILE = path.join(DUMP_DIR, 'v8_extractor.log');
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
// v8 핵심: JSON.stringify 후킹 + require 캐시 탐색
// AI 컨텍스트는 JS 객체 → JSON.stringify → protobuf serialize 순서로 처리.
// JSON.stringify 단계에서 잡으면 평문 XML/JSON을 확보할 수 있음.
// ====================================================================
const DUMP_DIR_ESCAPED = DUMP_DIR.replace(/\\/g, '\\\\');

const MEMORY_PATCH_CODE = `
(() => {
    // 이전 깨진 패치 플래그 초기화
    delete globalThis.__neuronfs_v8_ctx;
    
    if (globalThis.__neuronfs_v8c_ctx) return 'already_patched';
    globalThis.__neuronfs_v8c_ctx = true;

    const fs = process.getBuiltinModule('fs');
    const path = process.getBuiltinModule('path');
    const DUMP = '${DUMP_DIR_ESCAPED}';
    let seq = 0;

    function dumpText(label, content) {
        seq++;
        const ts = Date.now();
        try {
            const fname = path.join(DUMP, 'ctx_' + ts + '_' + seq + '.txt');
            fs.writeFileSync(fname, content);
            const meta = { label, size: content.length, ts: new Date().toISOString(), seq };
            fs.writeFileSync(fname + '.meta.json', JSON.stringify(meta, null, 2));
        } catch(e) {}
    }

    // ========== 전략 A: JSON.stringify 후킹 ==========
    // gRPC protobuf 직렬화 직전에 JSON.stringify 호출될 수 있음
    // 10KB 이상이고 XML 태그 포함하면 캡처
    const origStringify = JSON.stringify;
    JSON.stringify = function(value, replacer, space) {
        const result = origStringify.call(JSON, value, replacer, space);
        try {
            if (result && result.length > 10000) {
                const hasXml = result.includes('<identity>') || 
                               result.includes('<user_information>') ||
                               result.includes('<user_rules>') ||
                               result.includes('system_instruction') ||
                               result.includes('antml:') ||
                               result.includes('EPHEMERAL_MESSAGE');
                if (hasXml) {
                    dumpText('JSON_STRINGIFY_XML', result);
                }
            }
        } catch(e) {}
        return result;
    };

    // ========== 전략 B: 깊은 객체 탐색 (ESM 호환) ==========
    // require.cache 대신 globalThis 트리를 재귀 탐색하여 AI 관련 문자열 속성 검색
    let deepReport = 'DEEP_OBJECT_SCAN:\\n';
    const visited = new WeakSet();
    
    function scanObj(obj, prefix, depth) {
        if (depth > 3 || !obj || visited.has(obj)) return;
        try { visited.add(obj); } catch(e) { return; }
        try {
            const keys = Object.getOwnPropertyNames(obj);
            for (const k of keys) {
                try {
                    const val = obj[k];
                    if (typeof val === 'string' && val.length > 100) {
                        if (val.includes('<identity>') || val.includes('<user_information>') || 
                            val.includes('system_instruction') || val.includes('antml:')) {
                            deepReport += 'FOUND_XML at ' + prefix + '.' + k + ' (len=' + val.length + ')\\n';
                            dumpText('DEEP_SCAN_XML_' + k, val);
                        }
                    }
                    if (typeof val === 'object' && val !== null && depth < 2) {
                        scanObj(val, prefix + '.' + k, depth + 1);
                    }
                } catch(e) {}
            }
        } catch(e) {}
    }
    
    scanObj(globalThis, 'global', 0);
    dumpText('DEEP_SCAN', deepReport);

    // ========== 전략 C: global/process 객체에서 AI 서비스 탐색 ==========
    let globalReport = 'GLOBAL_SCAN:\\n';
    
    // globalThis의 모든 키 검색
    for (const key of Object.getOwnPropertyNames(globalThis)) {
        try {
            const val = globalThis[key];
            if (val && typeof val === 'object' && key.length > 3) {
                const subKeys = Object.keys(val);
                const aiKeys = subKeys.filter(k => 
                    k.toLowerCase().includes('cascade') ||
                    k.toLowerCase().includes('chat') ||
                    k.toLowerCase().includes('context') ||
                    k.toLowerCase().includes('prompt')
                );
                if (aiKeys.length > 0) {
                    globalReport += 'globalThis.' + key + ': ' + aiKeys.join(', ') + '\\n';
                }
            }
        } catch(e) {}
    }
    
    dumpText('GLOBAL_SCAN', globalReport);

    // ========== 전략 D: TextEncoder.encode 후킹 ==========
    // gRPC는 내부적으로 TextEncoder로 문자열 → Uint8Array 변환
    const origEncode = TextEncoder.prototype.encode;
    TextEncoder.prototype.encode = function(input) {
        try {
            if (typeof input === 'string' && input.length > 5000) {
                const hasXml = input.includes('<identity>') ||
                               input.includes('<user_information>') ||
                               input.includes('system_instruction') ||
                               input.includes('antml:');
                if (hasXml) {
                    dumpText('TEXTENCODER_XML', input);
                }
            }
        } catch(e) {}
        return origEncode.call(this, input);
    };

    // ========== 전략 E: Buffer.from 후킹 (문자열 → Buffer 변환) ==========
    const origBufferFrom = Buffer.from;
    Buffer.from = function(input, encoding) {
        try {
            if (typeof input === 'string' && input.length > 5000) {
                const hasXml = input.includes('<identity>') ||
                               input.includes('<user_information>') ||
                               input.includes('system_instruction') ||
                               input.includes('antml:');
                if (hasXml) {
                    dumpText('BUFFER_FROM_XML', input);
                }
            }
        } catch(e) {}
        return origBufferFrom.call(Buffer, input, encoding);
    };

    return 'patched_ok_v8_context_extractor';
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

                log(`[${label}] v8 컨텍스트 추출 패치 주입 중...`);
                const result = await call('Runtime.evaluate', {
                    expression: MEMORY_PATCH_CODE,
                    returnByValue: true
                });

                const value = result?.result?.value;
                log(`[${label}] 패치 결과: ${JSON.stringify(value)}`);

                if (value && value.startsWith('patched_ok')) {
                    log(`[${label}] ✅ v8 패치 성공!`);
                } else if (value === 'already_patched') {
                    log(`[${label}] ⚠️ 이미 패치됨`);
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

function watchDumps() {
    log(`📁 덤프 감시: ${DUMP_DIR}`);
    let lastCount = 0;

    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => f.startsWith('ctx_') && !f.endsWith('.meta.json'));
            if (files.length > lastCount) {
                const newFiles = files.slice(lastCount);
                for (const f of newFiles) {
                    const full = path.join(DUMP_DIR, f);
                    const stat = fs.statSync(full);
                    log(`💎 캡처: ${f} (${stat.size} bytes)`);

                    const metaPath = full + '.meta.json';
                    if (fs.existsSync(metaPath)) {
                        try {
                            const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
                            log(`   📌 ${meta.label}`);
                        } catch {}
                    }
                    
                    // 텍스트 파일이니까 처음 500자 미리보기
                    try {
                        const preview = fs.readFileSync(full, 'utf8').substring(0, 500);
                        log(`   📄 ${preview.replace(/\n/g, '\\n')}`);
                    } catch {}
                }
                lastCount = files.length;
            }
        } catch {}
    }, 2000);
}

async function main() {
    log('');
    log('======================================================');
    log('🧠 NeuronFS v8 — AI 컨텍스트 엔벨로프 추출기');
    log('======================================================');
    log('전략 A: JSON.stringify 후킹 (XML 태그 감지)');
    log('전략 B: require 캐시 모듈 스캔 (AI 서비스 탐색)');
    log('전략 C: globalThis 스캔 (서비스 객체 탐색)');
    log('전략 D: TextEncoder.encode 후킹');
    log('전략 E: Buffer.from 후킹');
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
        log('🔥 대기 중... AI 채팅을 보내면 ctx_*.txt 파일이 생성됩니다!');
        log('   XML 태그 (<identity>, <user_information>, system_instruction) 감지 시 자동 덤프');
    } else {
        log('❌ 패치된 프로세스 없음 — 종료');
        process.exit(1);
    }
}

main().catch(e => { log(`치명적 에러: ${e.message}`); process.exit(1); });
