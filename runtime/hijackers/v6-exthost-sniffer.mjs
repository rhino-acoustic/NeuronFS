/**
 * NeuronFS v6 — Extension Host 메모리 직접 탈취
 * 
 * 1. Extension Host PID에 process._debugProcess() 시그널 전송 → 인스펙터 강제 활성화
 * 2. 열린 인스펙터 포트 탐색 (포트 스캔)
 * 3. 해당 포트의 CDP WebSocket에 연결
 * 4. Network.enable → requestWillBeSent에서 postData(XML 컨텍스트 엔벨로프) 탈취
 */
import http from 'http';
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const DUMP_DIR = path.join(INBOX, 'cdp_captures');
const OUTPUT = path.join(INBOX, 'latest_hijacked_context.md');
const LOG_FILE = path.join(INBOX, 'v6_exthost_sniffer.log');
fs.mkdirSync(DUMP_DIR, { recursive: true });

const AI_PATTERNS = ['cloudcode', 'generativelanguage', 'anthropic', 'cascade', 'aiplatform'];
let captureCount = 0;

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

// 1단계: Extension Host PID 찾기
function findExtHostPids() {
    try {
        const out = execSync(
            'Get-WmiObject Win32_Process | Where-Object { $_.CommandLine -like \'*experimental-network-inspection*\' } | ForEach-Object { $_.ProcessId }',
            { shell: 'powershell.exe', encoding: 'utf8' }
        );
        return out.trim().split(/\r?\n/).map(s => parseInt(s.trim())).filter(n => !isNaN(n));
    } catch(e) {
        log(`ExtHost PID 탐색 실패: ${e.message}`);
        return [];
    }
}

// 2단계: process._debugProcess(pid)로 인스펙터 강제 활성화
function activateInspector(pid) {
    try {
        log(`PID ${pid}에 _debugProcess 시그널 전송...`);
        execSync(`node -e "process._debugProcess(${pid})"`, { encoding: 'utf8', timeout: 5000 });
        log(`PID ${pid} 인스펙터 활성화 시그널 전송 완료`);
        return true;
    } catch(e) {
        log(`PID ${pid} 시그널 실패: ${e.message}`);
        return false;
    }
}

// 3단계: 열린 인스펙터 포트 찾기
async function findInspectorPort(pid) {
    // netstat로 해당 PID가 LISTENING하는 포트들 가져오기
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
        
        // 각 포트에 /json 쿼리해서 인스펙터인지 확인  
        for (const port of ports) {
            try {
                const data = await getJson(`http://127.0.0.1:${port}/json`);
                if (Array.isArray(data)) {
                    log(`PID ${pid}의 인스펙터 포트 발견: ${port} (타겟 ${data.length}개)`);
                    return { port, targets: data };
                }
            } catch {
                // 인스펙터 아님, 다음 포트
            }
        }
    } catch(e) {
        log(`PID ${pid} 포트 스캔 실패: ${e.message}`);
    }
    return null;
}

// 4단계: 인스펙터에 연결하여 Network 모니터링
function monitorExtHost(wsUrl, label) {
    log(`[${label}] WebSocket 연결 중: ${wsUrl}`);
    const ws = new WebSocket(wsUrl);
    let msgId = 1;

    ws.on('open', () => {
        log(`[${label}] 연결 성공! Network.enable 실행...`);
        ws.send(JSON.stringify({ id: msgId++, method: 'Network.enable', params: { maxPostDataSize: 10485760 } }));
    });

    ws.on('message', (raw) => {
        let msg;
        try { msg = JSON.parse(raw.toString()); } catch { return; }

        // 에러 체크
        if (msg.id && msg.error) {
            log(`[${label}] 도메인 에러: ${JSON.stringify(msg.error)}`);
        }
        if (msg.id && msg.result !== undefined && !msg.error) {
            log(`[${label}] 도메인 활성화 성공 (id:${msg.id})`);
        }

        const method = msg.method;
        if (!method) return;

        // 모든 네트워크 요청 캡처 (Heartbeat 제외, 나머지 전부 덤프)
        if (method === 'Network.requestWillBeSent') {
            const { requestId, request } = msg.params;
            const url = request?.url || '';
            const isHeartbeat = url.includes('/Heartbeat');
            
            if (!isHeartbeat) {
                captureCount++;
                log(`[${label}] 🎯 비-Heartbeat 요청 #${captureCount}: ${request.method} ${url.substring(0, 200)}`);

                const postData = request.postData || '';
                if (postData.length > 0) {
                    const dumpFile = path.join(DUMP_DIR, `grpc_${Date.now()}_${captureCount}.bin`);
                    fs.writeFileSync(dumpFile, postData);
                    log(`[${label}] 💎 탈취 성공! ${postData.length} bytes → ${path.basename(dumpFile)}`);

                    fs.writeFileSync(OUTPUT, [
                        `# 탈취된 컨텍스트 — ${new Date().toISOString()}`,
                        ``,
                        `**출처:** ${label}`,
                        `**URL:** ${url}`,
                        `**크기:** ${postData.length} bytes`,
                        ``,
                        '```',
                        postData.substring(0, 80000),
                        '```',
                        ''
                    ].join('\n'), 'utf8');
                } else {
                    // postData가 별도로 올 수 있으므로 getRequestPostData 시도
                    ws.send(JSON.stringify({
                        id: msgId++,
                        method: 'Network.getRequestPostData',
                        params: { requestId }
                    }));
                    log(`[${label}] ⏳ postData 별도 요청 (requestId: ${requestId})`);
                }
            }
        }

        // Network.getRequestPostData 응답 처리
        if (msg.id && msg.result?.postData) {
            captureCount++;
            const body = msg.result.postData;
            const dumpFile = path.join(DUMP_DIR, `deferred_${Date.now()}_${captureCount}.json`);
            fs.writeFileSync(dumpFile, body, 'utf8');
            log(`[${label}] 💎 지연 탈취 성공! ${body.length} bytes → ${path.basename(dumpFile)}`);
            fs.writeFileSync(OUTPUT, [
                `# 탈취된 컨텍스트 — ${new Date().toISOString()}`,
                ``,
                `**출처:** ${label} (지연)`,
                `**크기:** ${body.length} bytes`,
                ``,
                '```json',
                body.substring(0, 80000),
                '```',
                ''
            ].join('\n'), 'utf8');
        }

        // 응답 로깅
        if (method === 'Network.responseReceived') {
            const url = msg.params?.response?.url || '';
            if (AI_PATTERNS.some(p => url.includes(p))) {
                log(`[${label}] 📥 AI 응답: status:${msg.params.response.status} ${url.substring(0, 80)}`);
            }
        }
    });

    ws.on('error', (e) => log(`[${label}] WS 에러: ${e.message}`));
    ws.on('close', () => {
        log(`[${label}] WS 연결 종료. 5초 후 재연결...`);
        setTimeout(() => monitorExtHost(wsUrl, label), 5000);
    });
}

// 메인 실행
async function main() {
    log('');
    log('======================================================');
    log('🧠 NeuronFS v6 — Extension Host 메모리 직접 탈취');
    log('======================================================');
    
    // Extension Host PID 찾기
    const pids = findExtHostPids();
    if (pids.length === 0) {
        log('❌ Extension Host 프로세스를 찾을 수 없습니다.');
        process.exit(1);
    }
    log(`Extension Host PID 발견: ${pids.join(', ')}`);

    for (const pid of pids) {
        // 인스펙터 활성화 시그널 전송
        activateInspector(pid);
        await new Promise(r => setTimeout(r, 1000)); // 인스펙터 초기화 대기

        // 인스펙터 포트 탐색
        const result = await findInspectorPort(pid);
        if (!result) {
            log(`PID ${pid}: 인스펙터 포트 미발견 — 건너뜀`);
            continue;
        }

        const { port, targets } = result;
        for (const t of targets) {
            if (t.webSocketDebuggerUrl) {
                monitorExtHost(t.webSocketDebuggerUrl, `ExtHost-${pid}:${port}`);
            }
        }
    }

    log('');
    log('🔥 Extension Host 모니터링 활성화 완료!');
    log('   이제 AI 채팅을 보내면 XML 컨텍스트가 탈취됩니다.');
    log('======================================================');
}

main().catch(e => { log(`치명적 에러: ${e.message}`); process.exit(1); });
