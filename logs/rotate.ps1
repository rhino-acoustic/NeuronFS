# NeuronFS 로그 로테이션 — 매일 자정 실행 또는 watchdog에서 호출
$logsDir = "C:\Users\BASEMENT_ADMIN\NeuronFS\logs"
$maxDays = 7

Get-ChildItem $logsDir -File -Filter "*.log" | ForEach-Object {
    $size = $_.Length / 1MB
    if ($size -gt 5) {
        $date = Get-Date -Format "yyyyMMdd"
        $archived = Join-Path "$logsDir\archive" "$($_.BaseName)_$date$($_.Extension)"
        Move-Item $_.FullName $archived -Force
        New-Item $_.FullName -ItemType File -Force | Out-Null
        Write-Output "로테이션: $($_.Name) → archive/ (${size}MB)"
    }
}

# 7일 이상 아카이브 삭제
Get-ChildItem "$logsDir\archive" -File | Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$maxDays) } | Remove-Item -Force
