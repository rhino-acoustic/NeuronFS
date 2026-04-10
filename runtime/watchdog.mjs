import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';
import { exec, spawn, execSync } from 'child_process';

const TG_BRIDGE_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\telegram-bridge';
const LOCK_FILE = path.join(path.dirname(new URL(import.meta.url).pathname.slice(1)), 'watchdog.lock');
const FALLBACK_LOG = path.join(path.dirname(new URL(import.meta.url).pathname.slice(1)), 'watchdog_fallback.log');

// ── Singleton Guard: 이미 와치독이 돌고 있으면 즉시 종료 ──
try {
    if (fs.existsSync(LOCK_FILE)) {
        const lockPid = parseInt(fs.readFileSync(LOCK_FILE, 'utf8').trim());
        try {
            process.kill(lockPid, 0); // 살아있는지 확인
            console.log(`[WATCHDOG] Already running (PID=${lockPid}). Exiting.`);
            process.exit(0);
        } catch {
            // PID 죽어있음 → stale lock → 넘어감
        }
    }
    fs.writeFileSync(LOCK_FILE, String(process.pid), 'utf8');
} catch {}

// 종료 시 lockfile 삭제
function cleanup() {
    try { fs.unlinkSync(LOCK_FILE); } catch {}
}
process.on('exit', cleanup);
process.on('SIGINT', () => { cleanup(); process.exit(0); });
process.on('SIGTERM', () => { cleanup(); process.exit(0); });

let TG_TOKEN = process.env.TELEGRAM_BOT_TOKEN || '';
if (!TG_TOKEN) { try { TG_TOKEN = fs.readFileSync(path.join(TG_BRIDGE_DIR, '.token'), 'utf8').trim(); } catch {} }
let TG_CHAT_ID = '';
try { TG_CHAT_ID = fs.readFileSync(path.join(TG_BRIDGE_DIR, '.chat_id'), 'utf8').trim(); } catch {}

function fallbackLog(msg) {
    const ts = new Date().toISOString();
    const line = `[${ts}] ${msg}\n`;
    try { fs.appendFileSync(FALLBACK_LOG, line, 'utf8'); } catch (e) { console.error('[FALLBACK_LOG_ERR]', e.message); }
    console.log(`[FALLBACK] ${msg}`);
}

function tgAlert(msg) {
    if (!TG_TOKEN || !TG_CHAT_ID) return;
    const body = JSON.stringify({ chat_id: TG_CHAT_ID, text: `🚨 [NeuronFS 파수꾼]\n${msg}` });
    const req = https.request({
        hostname: 'api.telegram.org', path: `/bot${TG_TOKEN}/sendMessage`, method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) }
    });
    req.on('error', () => {});
    req.write(body); req.end();
}

// 감시 대상: 이름 + 재시작 명령어
const TARGETS = [
    {
        name: 'auto-accept.mjs',
        script: 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\_tmp_aa_v2\\auto-accept.mjs',
        alertSent: false, restartCount: 0, isExe: false, zombieCount: 0,
        health: []
    },
    {
        name: 'headless-executor.mjs',
        script: 'C:\\Users\\BASEMENT_ADMIN\\_architecture_hijack_v4\\headless-executor.mjs',
        alertSent: false, restartCount: 0, isExe: false, zombieCount: 0,
        health: []
    },
    {
        name: 'hijack-launcher.mjs',
        script: 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\hijackers\\hijack-launcher.mjs',
        alertSent: false, restartCount: 0, isExe: false, zombieCount: 0,
        health: [
            { type: 'port', port: 9000, desc: 'CDP port' },
            { type: 'log', file: 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\logs\\tg_debug.log', maxAge: 300, desc: 'TG log 5분' }
        ]
    },
    {
        name: 'neuronfs.exe',
        script: 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\neuronfs.exe',
        args: ['C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4', '--api'],
        alertSent: false, restartCount: 0, isExe: true, zombieCount: 0,
        health: [
            { type: 'http', url: 'http://127.0.0.1:9090/api/health', desc: 'API health' },
            { type: 'port', port: 9090, desc: 'API port' }
        ]
    }
];

// 동기적으로 프로세스 이름별 카운트 가져오기 (race condition 방지)
function getProcessCounts() {
    try {
        const raw = execSync(
            'powershell -NoProfile -Command "Get-CimInstance Win32_Process -Filter \\"name=\'node.exe\' OR name=\'neuronfs.exe\'\\" | Select-Object ProcessId, CommandLine, Name | ConvertTo-Json -Compress"',
            { encoding: 'utf8', timeout: 15000 }
        );
        let procs = JSON.parse(raw || '[]');
        if (!Array.isArray(procs)) procs = [procs];
        return procs;
    } catch {
        return [];
    }
}

function revive(target) {
    target.restartCount++;
    if (target.restartCount > 5) {
        fallbackLog(`⛔ [포기] ${target.name} ${target.restartCount}회 재시작 실패. 수동 개입 필요.`);
        return;
    }

    // ★ 이중 확인: revive 직전에 다시 한번 프로세스 확인 (race condition 방지)
    const procs = getProcessCounts();
    const alive = procs.filter(p => (p.CommandLine || '').toLowerCase().includes(target.name.toLowerCase()));
    if (alive.length > 0) {
        console.log(`[REVIVE_SKIP] ${target.name} already running (${alive.length}). Skip.`);
        target.restartCount = 0;
        return;
    }

    tgAlert(`🔄 [자동 복구] ${target.name} 사망 감지 → 재시작 시도 #${target.restartCount}`);
    let child;
    if (target.isExe) {
        child = spawn(target.script, target.args || [], {
            detached: true, stdio: 'ignore', windowsHide: true,
            cwd: path.dirname(target.script)
        });
    } else {
        child = spawn('node', [target.script], {
            detached: true, stdio: 'ignore', windowsHide: true
        });
    }
    child.unref();
    console.log(`[REVIVE] ${target.name} PID=${child.pid}`);
}

function checkProcesses() {
    const procs = getProcessCounts();

    for (const t of TARGETS) {
        const matching = procs.filter(p => {
            if (t.isExe) return (p.Name || '').toLowerCase() === t.name.toLowerCase();
            return (p.CommandLine || '').toLowerCase().includes(t.name.toLowerCase());
        });

        if (matching.length === 0) {
            // L1: 프로세스 없음 → 즉시 재시작
            if (!t.alertSent) t.alertSent = true;
            t.zombieCount = 0;
            revive(t);
        } else {
            // L1 통과 → L2~L4 deep health check
            if (t.alertSent) {
                tgAlert(`✅ [복구 확인] ${t.name} 정상 가동 중.`);
                t.alertSent = false;
            }
            t.restartCount = 0;

            // Deep health check
            if (t.health && t.health.length > 0) {
                deepHealthCheck(t, matching);
            } else {
                t.zombieCount = 0;
            }

            // 중복 실행 감지 → 가장 오래된 것만 남기고 나머지 종료
            if (matching.length > 1) {
                const dupes = matching.slice(1);
                for (const d of dupes) {
                    try { process.kill(d.ProcessId); } catch {}
                    console.log(`[DEDUP] ${t.name} killed duplicate PID=${d.ProcessId} (${matching.length} → 1)`);
                }
                fallbackLog(`⚠️ [중복 정리] ${t.name}: ${matching.length}개 → 1개`);
            }
        }
    }
}

// ── Deep Health Check: 프로세스 살아있지만 좀비인지 다방향 검증 ──
function deepHealthCheck(target, matching) {
    let failed = [];

    for (const h of target.health) {
        switch (h.type) {
            case 'port': {
                // L2: 포트 리스닝 확인
                try {
                    const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr "127.0.0.1:${h.port}"`, { encoding: 'utf8', timeout: 5000 });
                    if (!netstat.trim()) failed.push(h.desc);
                } catch { failed.push(h.desc); }
                break;
            }
            case 'http': {
                // L3: HTTP 헬스체크 (비동기이므로 결과는 다음 사이클에 반영)
                const req = http.get(h.url, { timeout: 3000 }, (res) => {
                    if (res.statusCode !== 200) {
                        target.zombieCount++;
                        fallbackLog(`⚠️ [ZOMBIE] ${target.name} ${h.desc}: HTTP ${res.statusCode}`);
                    } else {
                        // HTTP 성공 → zombieCount 감소
                        if (target.zombieCount > 0) target.zombieCount--;
                    }
                    res.resume();
                });
                req.on('error', () => {
                    target.zombieCount++;
                    fallbackLog(`⚠️ [ZOMBIE] ${target.name} ${h.desc}: 연결 실패`);
                });
                req.on('timeout', () => {
                    req.destroy();
                    target.zombieCount++;
                    fallbackLog(`⚠️ [ZOMBIE] ${target.name} ${h.desc}: 타임아웃`);
                });
                break;
            }
            case 'log': {
                // L4: 로그 신선도 확인
                try {
                    const stat = fs.statSync(h.file);
                    const ageSec = (Date.now() - stat.mtimeMs) / 1000;
                    if (ageSec > (h.maxAge || 300)) failed.push(`${h.desc} (${Math.round(ageSec)}s stale)`);
                } catch { failed.push(h.desc); }
                break;
            }
        }
    }

    // 동기 체크 실패
    if (failed.length > 0) {
        target.zombieCount++;
        fallbackLog(`⚠️ [ZOMBIE] ${target.name} 동기체크 실패: ${failed.join(', ')} (${target.zombieCount}/3)`);
    } else if (target.zombieCount > 0 && target.health.every(h => h.type !== 'http')) {
        target.zombieCount = 0; // 동기 체크만 있고 전부 통과하면 리셋
    }

    // 3회 연속 좀비 → 강제 종료 + 재시작
    if (target.zombieCount >= 3) {
        fallbackLog(`🔴 [ZOMBIE_KILL] ${target.name} 3회 연속 좀비 → 강제 재시작`);
        tgAlert(`🔴 [좀비 감지] ${target.name} 응답 없음 3회 → 강제 재시작`);
        for (const m of matching) {
            try { process.kill(m.ProcessId, 'SIGKILL'); } catch {}
        }
        target.zombieCount = 0;
        setTimeout(() => revive(target), 3000);
    }
}

// ── 엔터프라이즈 메트릭 ──
const BOOT_TIME = Date.now();
const metrics = {
    checkCount: 0,
    totalCheckMs: 0,
    restarts: {},    // { name: [{ts, reason}] }
    lastSeen: {},    // { name: timestamp }
    downSince: {},   // { name: timestamp | null }
};
for (const t of TARGETS) {
    metrics.restarts[t.name] = [];
    metrics.lastSeen[t.name] = Date.now();
    metrics.downSince[t.name] = null;
}

// revive 원래 함수를 감싸서 메트릭 기록
const _origRevive = revive;
revive = function(target) {
    metrics.restarts[target.name].push({ ts: Date.now(), reason: target.zombieCount > 0 ? 'zombie' : 'dead' });
    if (!metrics.downSince[target.name]) metrics.downSince[target.name] = Date.now();
    _origRevive(target);
};

// checkProcesses를 감싸서 소요시간 측정
const _origCheck = checkProcesses;
checkProcesses = function() {
    const t0 = Date.now();
    _origCheck();
    const elapsed = Date.now() - t0;
    metrics.checkCount++;
    metrics.totalCheckMs += elapsed;
    // 체크가 느리면 경고
    if (elapsed > 10000) fallbackLog(`⚠️ [SLOW_CHECK] ${elapsed}ms`);
};

// alive 감지 시 lastSeen/downSince 갱신 (checkProcesses 내부에서 호출)
const _origCheckInner = checkProcesses;

function updateMetrics(targetName, alive) {
    if (alive) {
        metrics.lastSeen[targetName] = Date.now();
        metrics.downSince[targetName] = null;
    }
}

function calcUptime(targetName) {
    const total = Date.now() - BOOT_TIME;
    const downMs = metrics.restarts[targetName].reduce((sum, r, i, arr) => {
        const downStart = r.ts;
        const downEnd = (i + 1 < arr.length) ? arr[i + 1].ts : (metrics.downSince[targetName] || Date.now());
        // 대략 30초 다운타임으로 추정 (revive 후 복구까지)
        return sum + Math.min(30000, downEnd - downStart);
    }, 0);
    return total > 0 ? ((total - downMs) / total * 100).toFixed(2) : '100.00';
}

const METRICS_FILE = path.join(path.dirname(new URL(import.meta.url).pathname.slice(1)), '..', 'logs', 'watchdog_metrics.json');

function writeMetricsFile() {
    const data = {
        ts: new Date().toISOString(),
        bootTime: new Date(BOOT_TIME).toISOString(),
        uptimeMs: Date.now() - BOOT_TIME,
        checkCount: metrics.checkCount,
        avgCheckMs: metrics.checkCount > 0 ? Math.round(metrics.totalCheckMs / metrics.checkCount) : 0,
        services: TARGETS.map(t => ({
            name: t.name,
            status: metrics.downSince[t.name] ? 'DOWN' : 'UP',
            sla: parseFloat(calcUptime(t.name)),
            restarts: metrics.restarts[t.name].length,
            lastRestart: metrics.restarts[t.name].length > 0 ? new Date(metrics.restarts[t.name].slice(-1)[0].ts).toISOString() : null,
            zombieCount: t.zombieCount,
            lastSeen: new Date(metrics.lastSeen[t.name]).toISOString(),
            healthChecks: (t.health || []).map(h => h.desc)
        }))
    };
    try { fs.writeFileSync(METRICS_FILE, JSON.stringify(data, null, 2), 'utf8'); } catch {}
}

function statusReport() {
    const uptimeMin = ((Date.now() - BOOT_TIME) / 60000).toFixed(1);
    const avgCheckMs = metrics.checkCount > 0 ? (metrics.totalCheckMs / metrics.checkCount).toFixed(0) : 0;
    let report = `📊 [NeuronFS 상태 보고]\n`;
    report += `⏱️ Watchdog 가동: ${uptimeMin}분 | 체크: ${metrics.checkCount}회 | 평균: ${avgCheckMs}ms\n`;
    for (const t of TARGETS) {
        const sla = calcUptime(t.name);
        const restartCount = metrics.restarts[t.name].length;
        const status = metrics.downSince[t.name] ? '🔴 DOWN' : '🟢 UP';
        const zombie = t.zombieCount > 0 ? ` zombie=${t.zombieCount}` : '';
        report += `${status} ${t.name}: SLA=${sla}% | restarts=${restartCount}${zombie}\n`;
    }
    fallbackLog(report);
    tgAlert(report);
    writeMetricsFile();
}

// 매 체크 후 메트릭 파일 갱신
const _wrapCheck = checkProcesses;
checkProcesses = function() { _wrapCheck(); writeMetricsFile(); };

setInterval(checkProcesses, 30000);
setInterval(statusReport, 30 * 60 * 1000);
console.log(`👀 NeuronFS Watchdog v4 (Enterprise) PID=${process.pid}`);
setTimeout(checkProcesses, 5000);
setTimeout(statusReport, 60000);
