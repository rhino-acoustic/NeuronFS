# NeuronFS Watchdog — 프로세스 자동재시작
# 매 30초마다 neuronfs.exe가 살아있는지 확인하고, 죽었으면 재시작
# 사용: .\watchdog.ps1 (백그라운드에서 실행)

$exe = "C:\Users\BASEMENT_ADMIN\NeuronFS\neuronfs.exe"
$args = @("C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4", "--api")
$checkInterval = 30  # 초

Write-Host "[WATCHDOG] 🐕 NeuronFS 감시 시작 (매 ${checkInterval}초)"

while ($true) {
    $proc = Get-Process neuronfs -ErrorAction SilentlyContinue
    if (-not $proc) {
        $ts = Get-Date -Format "HH:mm:ss"
        Write-Host "[$ts] [WATCHDOG] ⚠️ neuronfs 프로세스 사망 감지 — 재시작 중..."
        Start-Process -FilePath $exe -ArgumentList $args -NoNewWindow
        Start-Sleep -Seconds 3
        $newProc = Get-Process neuronfs -ErrorAction SilentlyContinue
        if ($newProc) {
            Write-Host "[$ts] [WATCHDOG] ✅ 재시작 성공 (PID: $($newProc.Id))"
        } else {
            Write-Host "[$ts] [WATCHDOG] ❌ 재시작 실패"
        }
    }
    Start-Sleep -Seconds $checkInterval
}
