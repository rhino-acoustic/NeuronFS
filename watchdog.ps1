# ═══════════════════════════════════════════════════════════════════
# NeuronFS Watchdog v2 — 통합 감시 데몬
# ═══════════════════════════════════════════════════════════════════
#
# 기능:
#   1) neuronfs.exe 생존 감시 (30초)
#   2) agent-bridge.mjs 생존 감시 (30초)
#   3) 주기적 harness 실행 (10분)
#   4) heartbeat 로그 (1분)
#   5) inbox 신규 메시지 알림
#
# 사용: .\watchdog.ps1
#       .\watchdog.ps1 -Duration 120  (2시간 후 자동 종료)
# ═══════════════════════════════════════════════════════════════════

param(
    [int]$Duration = 0  # 분 단위, 0이면 무한
)

$ErrorActionPreference = "SilentlyContinue"

# ── 경로 ──
$nfsRoot = "C:\Users\BASEMENT_ADMIN\NeuronFS"
$nfsExe = "$nfsRoot\neuronfs.exe"
if (-not (Test-Path $nfsExe)) { $nfsExe = "$nfsRoot\runtime\neuronfs.exe" }
$nfsArgs = @("$nfsRoot\brain_v4", "--api")
$bridgeScript = "$nfsRoot\runtime\agent-bridge.mjs"
$bridgeDir = "$nfsRoot\runtime"
$harnessScript = "$nfsRoot\harness.ps1"
$brain = "$nfsRoot\brain_v4"
$logFile = "$nfsRoot\logs\watchdog.log"
$agents = @("bot1", "entp", "enfp", "pm")

# ── 타이머 ──
$startTime = Get-Date
$endTime = if ($Duration -gt 0) { $startTime.AddMinutes($Duration) } else { $null }
$loopInterval = 30    # 초
$harnessEvery = 600   # 초 (10분)
$heartbeatEvery = 60  # 초
$lastHarness = [datetime]::MinValue
$lastHeartbeat = [datetime]::MinValue
$loopCount = 0

# ── 로그 ──
if (-not (Test-Path (Split-Path $logFile))) { New-Item (Split-Path $logFile) -ItemType Directory -Force | Out-Null }

function Log($msg) {
    $ts = Get-Date -Format "HH:mm:ss"
    $line = "[$ts] $msg"
    Write-Host $line
    Add-Content $logFile $line -Encoding UTF8
}

function LogColor($msg, $color) {
    $ts = Get-Date -Format "HH:mm:ss"
    $line = "[$ts] $msg"
    Write-Host $line -ForegroundColor $color
    Add-Content $logFile $line -Encoding UTF8
}

# ── 배너 ──
Write-Host ""
Write-Host "╔══════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  NeuronFS Watchdog v2 — 통합 감시 데몬           ║" -ForegroundColor Cyan
Write-Host "║  neuronfs + bridge + harness + heartbeat         ║" -ForegroundColor Cyan
if ($endTime) {
    Write-Host "║  종료: $($endTime.ToString('HH:mm:ss'))                               ║" -ForegroundColor Cyan
} else {
    Write-Host "║  종료: 수동 (Ctrl+C)                             ║" -ForegroundColor Cyan
}
Write-Host "╚══════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Log "🐕 시작 (neuronfs + bridge + harness 통합)"

# ── 메인 루프 ──
while ($true) {
    $now = Get-Date
    $loopCount++

    # ── 종료 체크 ──
    if ($endTime -and $now -gt $endTime) {
        LogColor "⏰ $Duration 분 경과 — 자동 종료" "Yellow"
        break
    }

    # ── 1) neuronfs.exe 감시 ──
    $nfsProc = Get-Process neuronfs -ErrorAction SilentlyContinue
    if (-not $nfsProc) {
        LogColor "⚠️ neuronfs 사망 → 재시작" "Red"
        Start-Process -FilePath $nfsExe -ArgumentList $nfsArgs -NoNewWindow
        Start-Sleep 3
        $nfsProc = Get-Process neuronfs -ErrorAction SilentlyContinue
        if ($nfsProc) { LogColor "✅ neuronfs 재시작 (PID:$($nfsProc.Id))" "Green" }
        else { LogColor "❌ neuronfs 재시작 실패" "Red" }
    }

    # ── 2) agent-bridge.mjs 감시 ──
    $bridgeAlive = $false
    $nodeProcs = Get-Process node -ErrorAction SilentlyContinue
    if ($nodeProcs) {
        foreach ($np in $nodeProcs) {
            try {
                $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId=$($np.Id)").CommandLine
                if ($cmd -and $cmd -match "agent-bridge") { $bridgeAlive = $true; break }
            } catch {}
        }
    }
    if (-not $bridgeAlive) {
        LogColor "⚠️ agent-bridge 사망 → 재시작" "Yellow"
        $logOut = "C:\Users\BASEMENT_ADMIN\NeuronFS\logs\bridge.log"
        $logErr = "C:\Users\BASEMENT_ADMIN\NeuronFS\logs\bridge_err.log"
        Start-Process node -ArgumentList $bridgeScript -WorkingDirectory $bridgeDir -NoNewWindow -RedirectStandardOutput $logOut -RedirectStandardError $logErr
        Start-Sleep 2
        LogColor "✅ agent-bridge 재시작" "Green"
    }

    # ── 3) 주기적 harness (10분마다) ──
    if (($now - $lastHarness).TotalSeconds -ge $harnessEvery) {
        Log "🔍 harness 실행 (10분 주기)"
        try {
            $harnessOutput = & powershell -File $harnessScript 2>&1 | Out-String
            if ($harnessOutput -match "FAIL:\s*0") {
                LogColor "✅ harness PASS" "Green"
            } else {
                LogColor "⚠️ harness 위반 감지" "Red"
                # 위반 시 bot1(ANCHOR)에게 알림
                $alertDir = "$brain\_agents\bot1\inbox"
                if (-not (Test-Path $alertDir)) { New-Item $alertDir -ItemType Directory -Force | Out-Null }
                $alertFile = "$alertDir\$(Get-Date -Format 'yyyyMMdd_HHmmss')_watchdog_harness_fail.md"
                @"
# from: watchdog
# to: bot1
# time: $(Get-Date -Format "yyyy-MM-ddTHH:mm:ss")
# priority: urgent
# type: alert

harness 위반 감지됨. 즉시 코드 검증 필요.

$harnessOutput
"@ | Out-File $alertFile -Encoding UTF8
            }
        } catch {
            LogColor "⚠️ harness 실행 에러: $($_.Exception.Message)" "Yellow"
        }
        $lastHarness = $now
    }

    # ── 4) heartbeat (1분마다) ──
    if (($now - $lastHeartbeat).TotalSeconds -ge $heartbeatEvery) {
        $elapsed = [math]::Round(($now - $startTime).TotalMinutes)
        $remaining = if ($endTime) { [math]::Round(($endTime - $now).TotalMinutes) } else { "∞" }
        
        # 뉴런 수
        $neuronCount = (Get-ChildItem $brain -Recurse -Filter "*.neuron" -ErrorAction SilentlyContinue).Count
        
        # 전체 에이전트 inbox 대기
        $inboxStatus = ""
        foreach ($ag in $agents) {
            $agInbox = (Get-ChildItem "$brain\_agents\$ag\inbox" -File -ErrorAction SilentlyContinue).Count
            if ($agInbox -gt 0) { $inboxStatus += " | 📨${ag}:$agInbox" }
        }
        
        # API 상태
        $apiOk = $false
        $apiNeurons = 0
        try { $s = Invoke-RestMethod "http://localhost:9090/api/state" -TimeoutSec 2; $apiOk = $true; $apiNeurons = $s.totalNeurons } catch {}
        
        # heartbeat 로그 — 봇하트비트 확인
        $hbAlive = $false
        Get-CimInstance Win32_Process -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -match "bot-heartbeat" } | ForEach-Object { $hbAlive = $true }
        
        $status = "💓 ${elapsed}m"
        if ($remaining -ne "∞") { $status += " | 남은:${remaining}m" }
        $status += " | nfs:$(if($nfsProc){'✅'}else{'❌'})"
        $status += " | bridge:$(if($bridgeAlive){'✅'}else{'❌'})"
        $status += " | hb:$(if($hbAlive){'✅'}else{'❌'})"
        $status += " | api:$(if($apiOk){"✅($apiNeurons)"}else{'❌'})"
        $status += " | neurons:$neuronCount"
        $status += $inboxStatus
        
        Log $status
        $lastHeartbeat = $now
    }

    # ── 5) 전체 에이전트 inbox 신규 메시지 확인 ──
    foreach ($ag in $agents) {
        $agInboxDir = "$brain\_agents\$ag\inbox"
        if (Test-Path $agInboxDir) {
            $newMsgs = Get-ChildItem $agInboxDir -File -ErrorAction SilentlyContinue | Where-Object { $_.LastWriteTime -gt $now.AddSeconds(-$loopInterval) }
            if ($newMsgs) {
                foreach ($m in $newMsgs) {
                    LogColor "📨 새 메시지: $ag/inbox/$($m.Name)" "Magenta"
                }
            }
        }
    }

    # ── 6) bot-heartbeat.mjs 감시 ──
    $hbAlive = $false
    Get-CimInstance Win32_Process -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -match "bot-heartbeat" } | ForEach-Object { $hbAlive = $true }
    if (-not $hbAlive) {
        LogColor "⚠️ bot-heartbeat 사망 → 재시작" "Yellow"
        $hbScript = "$nfsRoot\runtime\bot-heartbeat.mjs"
        Start-Process node -ArgumentList $hbScript -WorkingDirectory "$nfsRoot\runtime" -NoNewWindow -RedirectStandardOutput "$nfsRoot\logs\heartbeat.log" -RedirectStandardError NUL
        Start-Sleep 2
        LogColor "✅ bot-heartbeat 재시작" "Green"
    }

    Start-Sleep $loopInterval
}

Log "🐕 Watchdog 종료"
