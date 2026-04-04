# bat에서 recover → repair 로 수정
$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$lines = [IO.File]::ReadAllLines($batPath)
for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match 'antigravity_database_manager.*recover') {
        $lines[$i] = 'python "%NEURONFS_DIR%\tools\Antigravity-Database-Manager\antigravity_database_manager.py" --headless repair'
        Write-Host "Fixed line $($i+1)"
        break
    }
}
[IO.File]::WriteAllLines($batPath, $lines)
Write-Host "DONE"
