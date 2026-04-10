import fs from 'fs';
import path from 'path';
import https from 'https';
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
        alertSent: false, restartCount: 0
    },
    {
        name: 'headless-executor.mjs',
        script: 'C:\\Users\\BASEMENT_ADMIN\\_architecture_hijack_v4\\headless-executor.mjs',
        alertSent: false, restartCount: 0
    },
    {
        name: 'hijack-launcher.mjs',
        script: 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\hijackers\\hijack-launcher.mjs',
        alertSent: false, restartCount: 0
    }
];

// 동기적으로 프로세스 이름별 카운트 가져오기 (race condition 방지)
function getProcessCounts() {
    try {
        const raw = execSync(
            'powershell -NoProfile -Command "Get-CimInstance Win32_Process -Filter \\"name=\'node.exe\'\\" | Select-Object ProcessId, CommandLine | ConvertTo-Json -Compress"',
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
    const child = spawn('node', [target.script], {
        detached: true, stdio: 'ignore', windowsHide: true
    });
    child.unref();
    console.log(`[REVIVE] ${target.name} PID=${child.pid}`);
}

function checkProcesses() {
    const procs = getProcessCounts();

    for (const t of TARGETS) {
        const matching = procs.filter(p => (p.CommandLine || '').toLowerCase().includes(t.name.toLowerCase()));

        if (matching.length === 0) {
            // 죽었음 → 자동 재시작
            if (!t.alertSent) t.alertSent = true;
            revive(t);
        } else {
            // 살아있음
            if (t.alertSent) {
                tgAlert(`✅ [복구 확인] ${t.name} 정상 가동 중.`);
                t.alertSent = false;
            }
            t.restartCount = 0;

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

setInterval(checkProcesses, 30000); // 30초 주기
console.log(`👀 NeuronFS Watchdog v3 (Singleton + Dedup) PID=${process.pid}`);
// 첫 체크는 5초 후 (bat의 start /B 들이 뜰 시간 확보)
setTimeout(checkProcesses, 5000);
