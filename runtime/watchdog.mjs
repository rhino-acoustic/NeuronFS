import fs from 'fs';
import path from 'path';
import https from 'https';
import { exec, spawn } from 'child_process';

const TG_BRIDGE_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\telegram-bridge';
const FALLBACK_LOG = path.join(path.dirname(new URL(import.meta.url).pathname.slice(1)), 'watchdog_fallback.log');
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

function revive(target) {
    target.restartCount++;
    if (target.restartCount > 5) {
        fallbackLog(`⛔ [포기] ${target.name} ${target.restartCount}회 재시작 실패. 수동 개입 필요.`);
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
    exec('tasklist /FI "IMAGENAME eq node.exe" /FO CSV /NH', (err, stdout) => {
        if (err) return;

        // tasklist 결과에서 PID 목록 수집 후, 각 PID의 commandline 확인
        exec('powershell -NoProfile -Command "Get-CimInstance Win32_Process | Where-Object Name -eq \'node.exe\' | Select-Object -ExpandProperty CommandLine"', (err2, stdout2) => {
            if (err2) return;
            const out = (stdout2 || '').toLowerCase();

            for (const t of TARGETS) {
                if (out.includes(t.name.toLowerCase())) {
                    // 살아있음
                    if (t.alertSent) {
                        tgAlert(`✅ [복구 확인] ${t.name} 정상 가동 중.`);
                        t.alertSent = false;
                    }
                    t.restartCount = 0; // 성공하면 카운터 리셋
                } else {
                    // 죽었음 → 자동 재시작
                    if (!t.alertSent) {
                        t.alertSent = true;
                    }
                    revive(t);
                }
            }
        });
    });
}

setInterval(checkProcesses, 30000); // 30초 주기
console.log('👀 NeuronFS Watchdog v2 (Auto-Revive) 구동 시작...');
checkProcesses();
