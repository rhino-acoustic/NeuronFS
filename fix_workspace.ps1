# bat launch line: add workspace folder
$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$lines = [IO.File]::ReadAllLines($batPath)

for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match 'start "" "%ANTIGRAVITY%" --remote-debugging-port=%CDP_PORT%$') {
        $lines[$i] = 'start "" "%ANTIGRAVITY%" "%NEURONFS_DIR%" --remote-debugging-port=%CDP_PORT%'
        Write-Host "Fixed line $($i+1): $($lines[$i])"
        break
    }
}

[IO.File]::WriteAllLines($batPath, $lines)
Write-Host "DONE"
