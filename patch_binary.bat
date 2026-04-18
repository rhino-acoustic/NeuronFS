@echo off
chcp 65001 >nul
echo ============================================
echo  PATCH: language_server binary (Korean)
echo ============================================
echo.

echo [1/4] Killing Antigravity...
taskkill /F /IM Antigravity.exe >nul 2>&1
taskkill /F /IM language_server_windows_x64.exe >nul 2>&1
echo   Waiting 5 seconds...
timeout /t 5 /nobreak >nul

set "BIN=C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
set "BAK=%BIN%.bak_original"

echo [2/4] Creating backup...
if not exist "%BAK%" (
    copy "%BIN%" "%BAK%" >nul
    echo   Backup created
) else (
    echo   Backup exists
)

echo [3/4] Patching binary...
powershell -ExecutionPolicy Bypass -Command "$s='%BIN%';$b=[IO.File]::ReadAllBytes($s);$t=[Text.Encoding]::UTF8;$o=$t.GetBytes('You are Antigravity Agent, a powerful agentic AI coding assistant designed by the Google engineering team.');$r=$t.GetBytes('You are NeuronFS-Antigravity, an agentic AI coding assistant. Always think in Korean, answer in Korean.');if($r.Length-lt$o.Length){$r+=[byte[]]::new($o.Length-$r.Length)};$f=$false;for($i=0;$i-lt$b.Length-$o.Length;$i++){$m=$true;for($j=0;$j-lt$o.Length;$j++){if($b[$i+$j]-ne$o[$j]){$m=$false;break}};if($m){for($j=0;$j-lt$r.Length;$j++){$b[$i+$j]=$r[$j]};$f=$true;break}};if($f){[IO.File]::WriteAllBytes($s,$b);Write-Host '  PATCH OK'}else{Write-Host '  PATCH FAIL'}"

echo [4/4] Starting Antigravity...
start "" "C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"
echo.
echo ============================================
echo  DONE! Run restore_binary.bat to undo.
echo ============================================
pause
