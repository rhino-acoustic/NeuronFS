# 라인 번호로 직접 교체 (53-71번 라인)
$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$lines = [IO.File]::ReadAllLines($batPath)

Write-Host "Before: $($lines.Count) lines"
Write-Host "L53: $($lines[52])"
Write-Host "L54: $($lines[53])"
Write-Host "L71: $($lines[70])"

# 라인 53-71을 교체 (0-indexed: 52-70)
$before = $lines[0..51]
$after = $lines[70..($lines.Count-1)]  # :AG_DONE 포함

$replacement = @(
    ":: --- Antigravity Graceful Shutdown (CDP + CloseMainWindow) ---",
    'echo [INFO] Antigravity graceful shutdown (session flush)...',
    'powershell -NoProfile -ExecutionPolicy Bypass -File "%NEURONFS_DIR%\graceful_shutdown.ps1"'
)

$newLines = $before + $replacement + $after
[IO.File]::WriteAllLines($batPath, $newLines)
Write-Host "DONE: $($lines.Count) -> $($newLines.Count) lines"
