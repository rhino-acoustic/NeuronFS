# bat 파일에 인덱스 복원 단계 추가
# Antigravity 시작 직전에 rebuild_chat_index.ps1 호출
$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$lines = [IO.File]::ReadAllLines($batPath)

# Find "Launching Antigravity" line
$launchIdx = -1
for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match "Launching Antigravity") { $launchIdx = $i; break }
}

Write-Host "Found launch line at: $($launchIdx + 1)"

if ($launchIdx -gt 0) {
    $before = $lines[0..($launchIdx - 1)]
    $after = $lines[$launchIdx..($lines.Count - 1)]
    
    $insertion = @(
        "",
        ":: --- Chat History Index Rebuild (fix empty index bug) ---",
        'echo [INFO] Rebuilding chat history index from .pb files...',
        'powershell -NoProfile -ExecutionPolicy Bypass -File "%NEURONFS_DIR%\rebuild_chat_index.ps1"',
        ""
    )
    
    $newLines = $before + $insertion + $after
    [IO.File]::WriteAllLines($batPath, $newLines)
    Write-Host "DONE: Added index rebuild before Antigravity launch"
    Write-Host "Total lines: $($newLines.Count)"
} else {
    Write-Host "ERROR: Launch line not found"
}
