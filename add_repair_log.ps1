# 간단한 방식: repair 출력을 log 파일로 리다이렉트
$batPath = 'C:\Users\BASEMENT_ADMIN\NeuronFS\start_v4_swarm.bat'
$lines = [IO.File]::ReadAllLines($batPath)

for ($i = 0; $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match 'antigravity_database_manager.*repair') {
        # repair 출력을 로그로 저장하고, Antigravity 프로세스가 없는지 확인
        $newLines = $lines[0..($i-1)]
        $newLines += ':: Verify Antigravity is dead before repair'
        $newLines += 'tasklist /FI "IMAGENAME eq Antigravity.exe" 2>nul | find /I "Antigravity.exe" >nul'
        $newLines += 'if %errorlevel% equ 0 ('
        $newLines += '    echo [WARN] Antigravity still running! Force killing...'
        $newLines += '    taskkill /F /IM "Antigravity.exe" >nul 2>&1'
        $newLines += '    timeout /t 3 /nobreak >nul'
        $newLines += ')'
        $newLines += 'echo [INFO] Running chat history repair...'
        $newLines += 'python "%NEURONFS_DIR%\tools\Antigravity-Database-Manager\antigravity_database_manager.py" --headless repair > "%NEURONFS_DIR%\repair_output.log" 2>&1'
        $newLines += 'echo [OK] Repair exit code: %errorlevel%'
        $newLines += $lines[($i+1)..($lines.Count-1)]
        
        [IO.File]::WriteAllLines($batPath, $newLines)
        Write-Host "Added force-kill + repair logging"
        break
    }
}
Write-Host "DONE"
