# bat에서 rebuild_chat_index.ps1 → antigravity_database_manager.py 교체
$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$lines = [IO.File]::ReadAllLines($batPath)

for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match 'rebuild_chat_index\.ps1') {
        $lines[$i-1] = ':: --- Chat History Index Recovery (community tool) ---'
        $lines[$i] = 'python "%NEURONFS_DIR%\tools\Antigravity-Database-Manager\antigravity_database_manager.py" --headless recover --force'
        Write-Host "Fixed line $($i+1): $($lines[$i])"
        break
    }
}

[IO.File]::WriteAllLines($batPath, $lines)
Write-Host "DONE"
