/**
 * NeuronFS Unified Hijack Launcher
 * IDE 재시작 후에도 자동으로 모든 포트/PID/Inspector를 탐지하여 패치
 * 
 * 동적 요소:
 *   - PID: 매 실행마다 변경
 *   - Inspector port: _debugProcess 후 랜덤 할당
 *   - Inspector target ID: 세션마다 변경
 *   - Language Server ports (14905 등): 매번 변경
 *   - Named Pipe path: 매번 변경
 * 
 * 고정 요소:
 *   - --remote-debugging-port=9000 (Antigravity 설정)
 *   - gRPC service path: /exa.language_server_pb.LanguageServerService/*
 *   - 덤프 디렉토리 경로
 */
import { execSync } from 'child_process';
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const CDP_PORT = 9000;
const BRAIN_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const TRANSCRIPT_DIR = path.join(BRAIN_DIR, '_transcripts');
const SESSION_NEURON_DIR = path.join(BRAIN_DIR, 'hippocampus', 'session_log');
const SENSOR_DIR = path.join(BRAIN_DIR, 'sensors', 'tools');
const GLOBAL_INBOX_DIR = path.join(BRAIN_DIR, '_agents', 'global_inbox');

function getKST() {
    return new Date(Date.now() + 32400000).toISOString().replace('T', ' ').substring(11, 19);
}

function log(msg) {
    console.log(`[${getKST()}] ${msg}`);
}

// ============================================================
// NeuronFS 뉴런 발화 엔진
// ============================================================
function fireNeuron(regionPath, content) {
    // NeuronFS: 파일 = 발화 기록, 폴더 = 뉴런
    const dir = path.join(BRAIN_DIR, regionPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    
    const ts = new Date().toISOString().replace(/[:.]/g, '-');
    const fname = `${ts}.md`;
    fs.writeFileSync(path.join(dir, fname), content, 'utf8');
}

function appendTranscript(entry) {
    if (!fs.existsSync(TRANSCRIPT_DIR)) fs.mkdirSync(TRANSCRIPT_DIR, { recursive: true });
    // KST 기준 날짜+시간(시) 조합 -> 예: 2026-04-03_18h.txt
    const d = new Date(Date.now() + 32400000);
    const timeKey = d.toISOString().replace('T', '_').substring(0, 13) + 'h';
    const file = path.join(TRANSCRIPT_DIR, `${timeKey}.txt`);
    fs.appendFileSync(file, entry + '\n', 'utf8');
}

// ============================================================
// Session Transcript: 세션 복원용 rolling buffer (최근 20턴)
// v4-hook.cjs의 [Last Session Memory] 시스템이 이 파일을 읽음
// ============================================================
const TRANSCRIPT_JSONL = path.join(BRAIN_DIR, '_agents', 'global_inbox', 'transcript_latest.jsonl');
const MAX_TRANSCRIPT_LINES = 20;

function updateSessionTranscript(role, text, cascadeId) {
    try {
        const entry = JSON.stringify({
            ts: new Date().toISOString(),
            role: role,
            text: (text || '').substring(0, 2000),
            cascade: (cascadeId || '').substring(0, 12)
        });

        // Append
        fs.appendFileSync(TRANSCRIPT_JSONL, entry + '\n', 'utf8');

        // Rolling: 20행 초과 시 오래된 것 삭제
        const lines = fs.readFileSync(TRANSCRIPT_JSONL, 'utf8').trim().split('\n').filter(l => l.trim());
        if (lines.length > MAX_TRANSCRIPT_LINES) {
            const trimmed = lines.slice(-MAX_TRANSCRIPT_LINES).join('\n') + '\n';
            fs.writeFileSync(TRANSCRIPT_JSONL, trimmed, 'utf8');
        }
    } catch (e) {
        // 쓰기 실패 무시
    }
}

let sessionMessageCount = 0;
let sessionStartTime = new Date().toISOString();

// ============================================================
// Groq 청크 처리 파이프라인 — 대화 축적 → 뉴런 자동 추출
// ============================================================
const GROQ_API_KEY = process.env.GROQ_API_KEY || '';
const GROQ_MODEL = 'llama-3.3-70b-versatile';
const CHUNK_SIZE = 25; // 25개 메시지마다 Groq 호출 (맥락 풍부 → 추출 품질↑)
let messageBuffer = [];

async function groqExtractNeurons(messages) {
    const prompt = `당신은 NeuronFS 뉴런 추출기입니다.
다음 AI 채팅 대화에서 **재사용 가능한 규칙/패턴/교훈**을 추출하세요.

NeuronFS 규칙:
- 폴더 이름 = 규칙 자체 (한국어, 짧게)
- 禁 접두사 = 금지 규칙 (예: 禁하드코딩)
- 推 접두사 = 권장 규칙 (예: 推로컬깃활용)
- 접두사 없음 = 사실/지식

규칙은 반드시 brain_v4의 적절한 영역에 배치:
- cortex/neuronfs/runtime/ — 런타임/기술 규칙
- cortex/dev/ — 개발 규칙
- cortex/methodology/ — 방법론
- hippocampus/에러_패턴/ — 에러 패턴
- brainstem/ — 핵심 불변 규칙

출력 형식 (JSON 배열만, 설명 없이):
[{"path": "cortex/neuronfs/runtime/禁예시규칙", "record": "발화 근거 한 줄"}]

규칙이 없으면 빈 배열 [] 반환.

=== 대화 ===
${messages.join('\n')}`;

    try {
        const body = JSON.stringify({
            model: GROQ_MODEL,
            messages: [{ role: 'user', content: prompt }],
            temperature: 0.1,
            max_tokens: 1000
        });

        const res = await fetch('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${GROQ_API_KEY}`,
                'Content-Type': 'application/json'
            },
            body
        });

        if (!res.ok) {
            log(`⚠️ Groq API error: ${res.status}`);
            return [];
        }

        const data = await res.json();
        const text = data.choices?.[0]?.message?.content || '[]';
        
        // JSON 추출 (텍스트에 마크다운 블록이 있을 수 있음)
        const jsonMatch = text.match(/\[[\s\S]*\]/);
        if (!jsonMatch) return [];
        
        return JSON.parse(jsonMatch[0]);
    } catch (e) {
        log(`⚠️ Groq parse error: ${e.message}`);
        return [];
    }
}

async function processChunk() {
    if (messageBuffer.length < CHUNK_SIZE) return;
    
    const chunk = messageBuffer.splice(0, CHUNK_SIZE);
    log(`🧠 Groq 청크 처리: ${chunk.length}개 메시지 → 뉴런 추출 중...`);
    
    const neurons = await groqExtractNeurons(chunk);
    
    if (neurons.length === 0) {
        log('  → 추출된 규칙 없음');
        return;
    }
    
    for (const n of neurons) {
        if (!n.path || !n.record) continue;
        
        // NeuronFS grow API 사용 (중복 체크 + 메타데이터 자동)
        try {
            const neuronPath = n.path.replace(/\//g, '\\\\');
            const fullPath = path.join(BRAIN_DIR, ...n.path.split('/'));
            if (!fs.existsSync(fullPath)) fs.mkdirSync(fullPath, { recursive: true });
            
            // 뉴런 파일에 생성 이유 기록
            const neuronFiles = fs.readdirSync(fullPath).filter(f => f.endsWith('.neuron'));
            let counter = 1;
            if (neuronFiles.length > 0) {
                const maxN = Math.max(...neuronFiles.map(f => parseInt(f) || 0));
                counter = maxN + 1;
                // 기존 파일 삭제 (카운터 증가)
                for (const nf of neuronFiles) fs.unlinkSync(path.join(fullPath, nf));
            }
            
            // 뉴런 파일에 생성 이유 JSON 기록
            const meta = JSON.stringify({
                created: new Date().toISOString(),
                reason: n.record,
                source: 'groq_auto_extract',
                chunk_size: chunk.length
            });
            fs.writeFileSync(path.join(fullPath, `${counter}.neuron`), meta, 'utf8');
            
            log(`  🔥 뉴런 생성: ${n.path} (counter: ${counter})`);
            appendTranscript(`[${new Date().toISOString()}] NEURON_GROW: ${n.path} — ${n.record}`);
        } catch (e) {
            log(`  ⚠️ 뉴런 생성 실패: ${n.path} - ${e.message}`);
        }
    }
    
    log(`  ✅ ${neurons.length}개 뉴런 생성 완료`);
}
// ============================================================
// 하네스 강화 사이클 — 50메시지마다 실패/성공 패턴 추출
// reco 원칙: "프롬프트를 고치지 말고 하네스를 고쳐라"
// Attention Residual: 관련 뉴런 간 .axon 생성 → 선택적 참조
// ============================================================

// 기존 禁/推 뉴런 목록 수집 (중복 생성 방지)
function getExistingRuleNeurons() {
    const rules = [];
    const scanDir = (dir, prefix) => {
        try {
            const entries = fs.readdirSync(dir, { withFileTypes: true });
            for (const e of entries) {
                if (e.isDirectory() && (e.name.startsWith('禁') || e.name.startsWith('推'))) {
                    rules.push(prefix + '/' + e.name);
                }
                if (e.isDirectory() && !e.name.startsWith('_') && !e.name.startsWith('.')) {
                    scanDir(path.join(dir, e.name), prefix + '/' + e.name);
                }
            }
        } catch {}
    };
    
    const regions = ['brainstem', 'cortex', 'ego', 'prefrontal'];
    for (const r of regions) {
        scanDir(path.join(BRAIN_DIR, r), r);
    }
    return rules;
}

// .axon 파일 생성 — 뉴런 간 연결 (Attention Residual)
function createAxon(sourcePath, targetRegion, reason) {
    const sourceDir = path.join(BRAIN_DIR, ...sourcePath.split('/'));
    if (!fs.existsSync(sourceDir)) return;
    
    const axonName = `connect_${targetRegion}.axon`;
    const axonPath = path.join(sourceDir, axonName);
    
    // 이미 있으면 스킵
    if (fs.existsSync(axonPath)) return;
    
    fs.writeFileSync(axonPath, JSON.stringify({
        target: targetRegion,
        reason: reason,
        created: new Date().toISOString(),
        source: 'harness_cycle'
    }), 'utf8');
    
    log(`  🔗 [Axon] ${sourcePath} → ${targetRegion}: ${reason}`);
}

async function harnessCycle() {
    log(`🔧 [Harness Cycle] 시작 — 세션 #${sessionMessageCount}`);
    
    // ── 1. 최근 전사 수집 ──
    const d = new Date(Date.now() + 32400000);
    const timeKey = d.toISOString().replace('T', '_').substring(0, 13) + 'h';
    const transcriptFile = path.join(TRANSCRIPT_DIR, `${timeKey}.txt`);
    
    let recentLines = [];
    try {
        const content = fs.readFileSync(transcriptFile, 'utf8');
        const lines = content.split('\n').filter(l => l.trim());
        recentLines = lines.slice(-150); // 최근 150줄 (50메시지 × 약 3줄)
    } catch { return; }
    
    if (recentLines.length < 10) return;
    
    // ── 2. corrections.jsonl에서 최근 교정도 수집 ──
    let corrections = [];
    try {
        const corrPath = path.join(BRAIN_DIR, '_inbox', 'corrections.jsonl');
        const corrData = fs.readFileSync(corrPath, 'utf8');
        corrections = corrData.split('\n').filter(l => l.trim()).slice(-10);
    } catch {}
    
    // ── 3. 기존 禁/推 뉴런 목록 (중복 방지) ──
    const existingRules = getExistingRuleNeurons();
    
    // ── 4. Groq 호출 — 고도화된 프롬프트 ──
    const harnessPrompt = `당신은 NeuronFS 하네스 강화 분석기입니다.

## 작업
다음 AI 채팅 전사 로그와 교정 이력에서 **재사용 가능한 규칙**을 추출하세요.

## 추출 기준
1. **禁(금지)**: AI가 반복 실수하거나 사용자가 교정한 패턴
2. **推(권장)**: 효과적이었던 패턴, 사용자가 칭찬한 접근법
3. **axon(연결)**: 추출한 규칙이 기존 어떤 영역과 관련되는지

## NeuronFS 경로 규칙
- cortex/ = 기술/지식: cortex/dev/, cortex/neuronfs/, cortex/methodology/
- brainstem/ = 핵심 불변 규칙 (극히 중요한 것만)
- ego/ = 소통/톤 관련
- hippocampus/ = 에러 패턴, 기억

## 기존 규칙 (중복 생성 금지!)
${existingRules.slice(0, 30).join('\n')}

## 출력 형식 (JSON 배열만, 설명 없이)
[{
  "path": "cortex/영역/禁또는推_규칙명",
  "record": "근거 한 줄",
  "type": "ban|recommend",
  "related_regions": ["cortex", "ego"]
}]

규칙이 없거나 기존과 중복이면 빈 배열 [] 반환.

=== 교정 이력 ===
${corrections.join('\n')}

=== 전사 로그 (최근 150줄) ===
${recentLines.join('\n')}`;

    try {
        const body = JSON.stringify({
            model: GROQ_MODEL,
            messages: [{ role: 'user', content: harnessPrompt }],
            temperature: 0.1,
            max_tokens: 2000
        });

        const res = await fetch('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${GROQ_API_KEY}`,
                'Content-Type': 'application/json'
            },
            body
        });

        if (!res.ok) {
            log(`⚠️ Harness Groq error: ${res.status}`);
            return;
        }

        const data = await res.json();
        const text = data.choices?.[0]?.message?.content || '[]';
        const jsonMatch = text.match(/\[[\s\S]*\]/);
        if (!jsonMatch) return;
        
        const rules = JSON.parse(jsonMatch[0]);
        if (rules.length === 0) {
            log(`  🔧 [Harness] 추출된 패턴 없음 (기존과 중복이거나 규칙 미발견)`);
            return;
        }

        let created = 0;
        for (const r of rules) {
            if (!r.path || !r.record) continue;
            
            // 기존 뉴런 중복 체크
            if (existingRules.some(e => e.includes(r.path.split('/').pop()))) {
                log(`  ⏭️ [Harness] 중복 스킵: ${r.path}`);
                continue;
            }
            
            const fullPath = path.join(BRAIN_DIR, ...r.path.split('/'));
            if (!fs.existsSync(fullPath)) fs.mkdirSync(fullPath, { recursive: true });
            
            // 뉴런 카운터 설정
            const neuronFiles = fs.readdirSync(fullPath).filter(f => f.endsWith('.neuron'));
            let counter = 1;
            if (neuronFiles.length > 0) {
                const maxN = Math.max(...neuronFiles.map(f => parseInt(f) || 0));
                counter = maxN + 1;
                for (const nf of neuronFiles) fs.unlinkSync(path.join(fullPath, nf));
            }
            
            // 생성 이유 + 소스 기록
            const meta = JSON.stringify({
                created: new Date().toISOString(),
                reason: r.record,
                type: r.type || 'unknown',
                source: 'harness_cycle_50',
                session_count: sessionMessageCount,
                related_regions: r.related_regions || []
            });
            fs.writeFileSync(path.join(fullPath, `${counter}.neuron`), meta, 'utf8');
            
            // ── Axon 연결 생성 (Attention Residual) ──
            if (r.related_regions && Array.isArray(r.related_regions)) {
                for (const region of r.related_regions) {
                    createAxon(r.path, region, r.record);
                }
            }
            
            const icon = r.type === 'ban' ? '⛔' : '✨';
            log(`  ${icon} [Harness] ${r.path}: ${r.record}`);
            appendTranscript(`[${new Date().toISOString()}] HARNESS_${r.type?.toUpperCase()}: ${r.path} — ${r.record}`);
            created++;
        }

        if (created === 0) {
            log(`  🔧 [Harness] 중복 필터링 후 신규 규칙 없음`);
            return;
        }

        log(`  🔧 [Harness] ${created}개 규칙 생성 → emit 트리거`);
        
        // ── 5. neuronfs emit 트리거 (GEMINI.md 즉시 갱신) ──
        try {
            execSync('C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\neuronfs.exe emit', { timeout: 15000 });
            log(`  ✅ [Harness] GEMINI.md 갱신 완료 — 하네스 강화됨`);
        } catch (e) {
            log(`  ⚠️ [Harness] emit 실패: ${e.message}`);
        }
        
        appendTranscript(`[${new Date().toISOString()}] HARNESS_CYCLE_COMPLETE: ${created} rules, session #${sessionMessageCount}`);
    } catch (e) {
        log(`⚠️ Harness cycle error: ${e.message}`);
    }
}

// ============================================================
// Phase 1: 프로세스 탐지
// ============================================================
function discoverProcesses() {
    log('Phase 1: 프로세스 탐지');
    const result = {};
    
    try {
        const tmpPs = path.join(DUMP_DIR, '_proc.ps1');
        fs.writeFileSync(tmpPs, "Get-CimInstance Win32_Process | Where-Object { \$_.Name -like '*Antigravity*' -or \$_.Name -like '*language_server*' } | ForEach-Object { Write-Output \"\$( \$_.ProcessId)|\$( \$_.ParentProcessId)|\$( \$_.CommandLine)\" }");
        const output = execSync('powershell -NoProfile -ExecutionPolicy Bypass -File "' + tmpPs + '"', { encoding: 'utf8', timeout: 15000 });
        
        const lines = output.split('\n').filter(l => l.trim());
        
        for (const line of lines) {
            const [pidStr, ppidStr, ...cmdParts] = line.trim().split('|');
            const pid = parseInt(pidStr);
            const parentPid = parseInt(ppidStr);
            const cmd = cmdParts.join('|');
            if (isNaN(pid)) continue;
            
            let type = 'unknown';
            if (cmd.includes('language_server')) type = 'language_server';
            else if (cmd.includes('--type=renderer')) type = 'renderer';
            else if (cmd.includes('--type=utility') && cmd.includes('--inspect-port')) type = 'exthost_or_chathost';
            else if (cmd.includes('--type=utility')) type = 'utility';
            else if (cmd.includes('--type=gpu')) type = 'gpu';
            else if (cmd.includes('--type=crashpad')) type = 'crashpad';
            else if (!cmd.includes('--type=')) type = 'main_or_child';
            
            result[pid] = { pid, parentPid, type, cmd: cmd.substring(0, 200) };
        }
    } catch (e) {
        log('WMIC failed: ' + e.message);
    }
    
    // ExtHost 식별: --inspect-port를 가진 utility
    // ChatHost 식별: language_server를 자식으로 가지면서 inspect-port를 가진 utility
    const langServers = Object.values(result).filter(p => p.type === 'language_server');
    for (const ls of langServers) {
        const parent = result[ls.parentPid];
        if (parent && parent.type === 'exthost_or_chathost') {
            // 이 parent의 포트를 확인
            try {
                const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${parent.pid}"`, { encoding: 'utf8', timeout: 5000 });
                const ports = [...netstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
                parent.ports = ports;
                ls.ports = [];
                
                const lsNetstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${ls.pid}"`, { encoding: 'utf8', timeout: 5000 });
                ls.ports = [...lsNetstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
            } catch {}
        }
    }
    
    return result;
}

// ============================================================
// Phase 2: Inspector 포트 탐지 + 활성화
// ============================================================
async function findInspectorPort(pid) {
    // _debugProcess 시그널 전송
    try {
        execSync(`node -e "process._debugProcess(${pid})"`, { timeout: 5000 });
    } catch {}
    
    await new Promise(r => setTimeout(r, 1500));
    
    // 포트 스캔
    try {
        const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${pid}"`, { encoding: 'utf8', timeout: 5000 });
        const ports = [...netstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));
        
        for (const port of ports) {
            try {
                const data = await fetchJson(`http://127.0.0.1:${port}/json`);
                if (Array.isArray(data) && data.length > 0) {
                    return { port, targets: data };
                }
            } catch {}
        }
    } catch {}
    return null;
}

function fetchJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, { timeout: 2000 }, (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => { try { resolve(JSON.parse(data)); } catch { reject(new Error('parse')); } });
        }).on('error', reject);
    });
}

// ============================================================
// Phase 3: CDP 렌더러 Network 모니터 (채팅 캡처)
// ============================================================
function startCDPMonitor() {
    log('Phase 3: CDP Network Monitor on port ' + CDP_PORT);
    
    return fetchJson(`http://127.0.0.1:${CDP_PORT}/json`).then(targets => {
        log(`  ${targets.length} targets found`);
        
        for (const t of targets) {
            if (t.webSocketDebuggerUrl) {
                const label = `${t.type}:${(t.title || 'worker').substring(0, 20)}`;
                attachCDPNetwork(t.webSocketDebuggerUrl, label);
            }
        }
    });
}

function attachCDPNetwork(wsUrl, label) {
    const ws = new WebSocket(wsUrl);
    let id = 0;
    const pendingRequests = {}; // requestId → {rpc, ts}
    const pendingBodyFetches = {}; // msg.id → {rpc, requestId}
    
    ws.on('open', () => {
        log(`  [${label}] connected`);
        ws.send(JSON.stringify({ id: ++id, method: 'Network.enable', params: { maxTotalBufferSize: 50000000 } }));
    });
    
    ws.on('message', (d) => {
        try {
            const msg = JSON.parse(d.toString());
            
            // ===== 요청 캡처 =====
            if (msg.method === 'Network.requestWillBeSent') {
                const r = msg.params;
                const url = r.request ? r.request.url : '';
                
                if (url.includes('LanguageServerService') && 
                    !url.includes('Heartbeat') && !url.includes('GetUnleash') && !url.includes('GetStatus') &&
                    !url.includes('LogTelemetry') && !url.includes('UpdateActiveFile') && !url.includes('LogProcessInfo')) {
                    
                    const rpc = url.split('/').pop();
                    const reqId = r.requestId;
                    pendingRequests[reqId] = { rpc, ts: Date.now() };
                    
                    log(`🎯 [${label}] ${rpc}`);
                    
                    if (r.request.postData && r.request.postData.length > 0) {
                        const ts = Date.now();
                        // [OPTIMIZE] 바이너리 스니핑 덤프 저장 비활성화 (쓰레기 방지)
                        // const fname = `chat_net_${ts}`;
                        // fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin'), r.request.postData);
                        // fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin.meta.json')...

                        // 채팅 텍스트 미리보기 + NeuronFS 발화
                        try {
                            const json = JSON.parse(r.request.postData);
                            const cascadeId = json.cascadeId || 'unknown';
                            
                            if (json.items) {
                                for (const item of json.items) {
                                    if (item.text) {
                                        sessionMessageCount++;
                                        log(`  📝 [${label}] "${item.text}"`);
                                        appendTranscript(`[${getKST()}] [${label}] USER@${cascadeId.substring(0,8)}: ${item.text}`);
                                        updateSessionTranscript('user', item.text, cascadeId);
                                        // [REMOVED] fireNeuron session_log — 매 메시지마다 .md 생성하여 197K+ 폭발 유발
                                        
                                        // ━━━ 하네스 강화 사이클 (50메시지마다) ━━━
                                        // reco 원칙: "프롬프트를 고치지 말고 하네스를 고쳐라"
                                        if (sessionMessageCount > 0 && sessionMessageCount % 50 === 0) {
                                            log(`🔧 [Harness Cycle] 50사이클 → 하네스 강화 시작`);
                                            harnessCycle().catch(e => log(`⚠️ Harness cycle error: ${e.message}`));
                                        }

                                        // Groq 청크 버퍼에 추가
                                        messageBuffer.push(`[USER] ${item.text}`);
                                        processChunk().catch(() => {});
                                    }
                                }
                            }
                            if (json.interaction && json.interaction.runCommand) {
                                const cmd = json.interaction.runCommand.proposedCommandLine?.substring(0, 120) || '';
                                log(`  🖥️ [${label}] CMD confirm: "${cmd}"`);
                                appendTranscript(`[${getKST()}] [${label}] CMD@${cascadeId.substring(0,8)}: ${cmd}`);
                                messageBuffer.push(`[CMD] ${cmd}`);
                                processChunk().catch(() => {});
                            }
                        } catch {}
                    }
                }
            }
            
            // ===== 응답 완료 → 본문 요청 =====
            if (msg.method === 'Network.loadingFinished') {
                const reqId = msg.params.requestId;
                const detail = pendingRequests[reqId];
                if (detail) {
                    const elapsed = Date.now() - detail.ts;
                    log(`📩 [${label}] ${detail.rpc} response complete (${elapsed}ms, ${msg.params.encodedDataLength || '?'}B)`);
                    
                    // 응답 본문 가져오기
                    const fetchId = ++id;
                    pendingBodyFetches[fetchId] = { rpc: detail.rpc, requestId: reqId };
                    ws.send(JSON.stringify({
                        id: fetchId,
                        method: 'Network.getResponseBody',
                        params: { requestId: reqId }
                    }));
                    
                    delete pendingRequests[reqId];
                }
            }
            
            // ===== 응답 본문 수신 =====
            if (msg.id && pendingBodyFetches[msg.id]) {
                const detail = pendingBodyFetches[msg.id];
                delete pendingBodyFetches[msg.id];
                
                if (msg.result && msg.result.body) {
                    const body = msg.result.base64Encoded 
                        ? Buffer.from(msg.result.body, 'base64').toString('utf8')
                        : msg.result.body;
                    
                    const ts = Date.now();

                    log(`  📦 AI Response (${detail.rpc}): ${body.length}B`);
                    
                    // 트랜스크립트에 AI 응답 기록 (처음 2000자)
                    const preview = body.substring(0, 2000).replace(/\n/g, ' ');
                    appendTranscript(`[${getKST()}] [${label}] AI_RESP@${detail.rpc}: ${preview}`);
                    updateSessionTranscript('assistant', preview, '');
                    
                    // === Groq buffer에 AI 응답도 추가 (맥락 완성) ===
                    // thinking, tool calls, 코드 변경 등을 파싱하여 추가
                    try {
                        // thinking 추출
                        const thinkMatch = body.match(/thinking["\s:]+([^"]{20,500})/i);
                        if (thinkMatch) {
                            const think = thinkMatch[1].substring(0, 300);
                            appendTranscript(`[${getKST()}] [${label}] THINK: ${think}`);
                            messageBuffer.push(`[THINK] ${think}`);
                        }
                        
                        // tool calls 추출 (ran command, edited file, view_file 등)
                        const toolPatterns = [
                            /run_command[^}]*CommandLine["\s:]+([^"]{10,200})/gi,
                            /replace_file_content[^}]*TargetFile["\s:]+([^"]{10,200})/gi,
                            /view_file[^}]*AbsolutePath["\s:]+([^"]{10,200})/gi,
                            /write_to_file[^}]*TargetFile["\s:]+([^"]{10,200})/gi,
                        ];
                        for (const pat of toolPatterns) {
                            const matches = [...body.matchAll(pat)];
                            for (const m of matches) {
                                const toolInfo = m[1].substring(0, 150);
                                const toolName = pat.source.split('[')[0];
                                appendTranscript(`[${getKST()}] [${label}] TOOL@${toolName}: ${toolInfo}`);
                                messageBuffer.push(`[TOOL:${toolName}] ${toolInfo}`);
                            }
                        }
                        
                        // AI 텍스트 응답 (앞 500자)
                        const textPreview = body.substring(0, 500).replace(/\n/g, ' ');
                        messageBuffer.push(`[AI] ${textPreview}`);
                        
                        processChunk().catch(() => {});
                    } catch {}
                }
            }
            
        } catch {}
    });
    
    ws.on('error', () => {});
    ws.on('close', () => {
        log(`  [${label}] disconnected — reconnecting in 5s`);
        setTimeout(() => attachCDPNetwork(wsUrl, label), 5000);
    });
}

// ============================================================
// Phase 4: ExtHost h2 프로토타입 패치 (코드 캡처)
// ============================================================
async function patchExtHost(inspectorPort, targets) {
    const DUMP_DIR_ESC = DUMP_DIR.replace(/\\/g, '\\\\');
    
    for (const t of targets) {
        if (!t.webSocketDebuggerUrl) continue;
        
        log(`Phase 4: ExtHost 패치 → ${t.webSocketDebuggerUrl.substring(0, 60)}`);
        
        try {
            await new Promise((resolve, reject) => {
                const ws = new WebSocket(t.webSocketDebuggerUrl);
                
                ws.on('open', () => {
                    // 이미 v7-memory-hijack.mjs와 동일한 패치 코드 주입
                    // 간소화: 프로토타입 반 패치
                    const patchCode = `(function() {
                        const http2 = typeof require !== 'undefined' ? require('http2') : 
                                      (typeof process !== 'undefined' && process.getBuiltinModule ? process.getBuiltinModule('http2') : null);
                        if (!http2) return 'no_http2';
                        const fs = typeof require !== 'undefined' ? require('fs') : process.getBuiltinModule('fs');
                        const path = typeof require !== 'undefined' ? require('path') : process.getBuiltinModule('path');
                        const DUMP_DIR = '${DUMP_DIR_ESC}';
                        if (!fs.existsSync(DUMP_DIR)) try { fs.mkdirSync(DUMP_DIR, {recursive:true}); } catch {}
                        
                        let cc = 0;
                        function dump(label, data) {
                            cc++;
                            const ts = Date.now();
                            const f = 'mem_' + ts + '_' + cc;
                            try {
                                fs.writeFileSync(path.join(DUMP_DIR, f + '.bin'), Buffer.isBuffer(data) ? data : Buffer.from(data));
                                fs.writeFileSync(path.join(DUMP_DIR, f + '.bin.meta.json'), JSON.stringify({
                                    ts: new Date(ts).toISOString(), label, size: data.length, source: 'unified_h2'
                                }));
                            } catch {}
                        }
                        
                        try {
                            const tmp = http2.connect('https://127.0.0.1:1');
                            const proto = Object.getPrototypeOf(tmp);
                            try { tmp.close(); } catch {} try { tmp.destroy(); } catch {}
                            
                            if (proto && typeof proto.request === 'function' && !proto.__nfs_uni) {
                                proto.__nfs_uni = true;
                                const orig = proto.request;
                                proto.request = function(h, o) {
                                    const s = orig.call(this, h, o);
                                    const p = (h && h[':path']) || '/';
                                    const ow = s.write; s.write = function(d,e,c) { try { if(d&&d.length>10) dump('h2_req:'+p, Buffer.isBuffer(d)?d:Buffer.from(d)); } catch {} return ow.call(this,d,e,c); };
                                    const oe = s.end; s.end = function(d,e,c) { try { if(d&&d.length>10) dump('h2_req:'+p, Buffer.isBuffer(d)?d:Buffer.from(d)); } catch {} return oe.call(this,d,e,c); };
                                    s.on('data', (c) => { try { if(c&&c.length>10) dump('h2_res:'+p, Buffer.isBuffer(c)?c:Buffer.from(c)); } catch {} });
                                    s.on('response', (rh) => { try { dump('h2_hdrs:'+p, Buffer.from(JSON.stringify(rh))); } catch {} });
                                    return s;
                                };
                                return 'patched_ok';
                            }
                            return proto.__nfs_uni ? 'already_patched' : 'no_request';
                        } catch(e) { return 'err:' + e.message; }
                    })()`;
                    
                    ws.send(JSON.stringify({
                        id: 1,
                        method: 'Runtime.evaluate',
                        params: { expression: patchCode, returnByValue: true }
                    }));
                });
                
                ws.on('message', (d) => {
                    const r = JSON.parse(d.toString());
                    if (r.id === 1) {
                        const val = r.result?.result?.value || r.result?.value || 'unknown';
                        log(`  Result: ${val}`);
                        ws.close();
                        resolve(val);
                    }
                });
                
                ws.on('error', (e) => { reject(e); });
                setTimeout(() => { ws.close(); resolve('timeout'); }, 10000);
            });
        } catch (e) {
            log(`  Patch error: ${e.message}`);
        }
    }
}

// ============================================================
// Main
// ============================================================
async function main() {
    log('');
    log('═══════════════════════════════════════════════');
    log('  NeuronFS Unified Hijack Launcher');
    log('  동적 포트/PID 자동 탐지 + 전체 채널 패치');
    log('═══════════════════════════════════════════════');
    log('');
    
    if (!fs.existsSync(DUMP_DIR)) fs.mkdirSync(DUMP_DIR, { recursive: true });
    
    // 1. 프로세스 탐지
    const procs = discoverProcesses();
    const extHosts = Object.values(procs).filter(p => p.type === 'exthost_or_chathost');
    const langServers = Object.values(procs).filter(p => p.type === 'language_server');
    
    log(`ExtHost/ChatHost: ${extHosts.map(p => p.pid).join(', ') || 'none'}`);
    log(`Language Servers: ${langServers.map(p => `${p.pid}(ports:${p.ports?.join(',') || '?'})`).join(', ') || 'none'}`);
    
    // 2. ExtHost Inspector 활성화 + h2 패치
    for (const eh of extHosts) {
        log(`\nActivating inspector for PID ${eh.pid}...`);
        const inspector = await findInspectorPort(eh.pid);
        if (inspector) {
            log(`  Inspector: port ${inspector.port}, ${inspector.targets.length} targets`);
            await patchExtHost(inspector.port, inspector.targets);
        } else {
            log(`  No inspector found for PID ${eh.pid}`);
        }
    }
    
    // 3. CDP 렌더러 모니터 (채팅 캡처)
    log('');
    try {
        await startCDPMonitor();
    } catch (e) {
        log(`CDP monitor failed: ${e.message}`);
        log(`  (--remote-debugging-port=${CDP_PORT} 가 설정되어 있는지 확인)`);
    }
    
    // 4. 세션 시작 기록 (트랜스크립트만 — 뉴런 폭발 방지)
    appendTranscript(`[${sessionStartTime}] HIJACK_START: ExtHosts=${extHosts.map(p=>p.pid).join(',')} LS=${langServers.map(p=>p.pid).join(',')}`);
    
    // 5. 캡처 데이터 TTL 정리 (1시간마다)
    setInterval(() => {
        try {
            const files = fs.readdirSync(DUMP_DIR).filter(f => !f.startsWith('_'));
            const now = Date.now();
            let cleaned = 0;
            for (const f of files) {
                const fp = path.join(DUMP_DIR, f);
                const stat = fs.statSync(fp);
                // 2시간 이상 된 Heartbeat/GetUnleash 파일만 정리
                if (now - stat.mtimeMs > 7200000) {
                    try {
                        const meta = f.endsWith('.meta.json') ? JSON.parse(fs.readFileSync(fp)) : null;
                        if (meta && (meta.label?.includes('Heartbeat') || meta.label?.includes('GetUnleash') || meta.label?.includes('GetStatus'))) {
                            const binFile = f.replace('.meta.json', '');
                            fs.unlinkSync(fp);
                            const binPath = path.join(DUMP_DIR, binFile);
                            if (fs.existsSync(binPath)) fs.unlinkSync(binPath);
                            cleaned++;
                        }
                    } catch {}
                }
            }
            if (cleaned > 0) log(`🗑️ TTL cleanup: ${cleaned} stale files removed`);
        } catch {}
    }, 3600000);
    
    log('');
    log('═══════════════════════════════════════════════');
    log('  ✅ 모든 채널 활성화 + NeuronFS 뉴런 발화');
    log(`  📁 ${DUMP_DIR}`);
    log(`  🧠 ${TRANSCRIPT_DIR}`);
    log('  채팅→뉴런: hippocampus/session_log');
    log('  트랜스크립트: _transcripts/{date}.txt');
    log('═══════════════════════════════════════════════');
    log('');
}

main().catch(e => { log('Fatal: ' + e.message); process.exit(1); });
