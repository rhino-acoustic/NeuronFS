@echo off
chcp 65001 >nul
echo ============================================
echo  NeuronFS System Prompt Patch
echo  language_server 바이너리 교정
echo ============================================
echo.
echo [1/4] Antigravity 종료...
taskkill /F /IM Antigravity.exe >nul 2>&1
taskkill /F /IM language_server_windows_x64.exe >nul 2>&1
timeout /t 5 /nobreak >nul

set "BIN=C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
set "BAK=%BIN%.bak_original"

echo [2/4] 백업 생성...
if not exist "%BAK%" (
    copy "%BIN%" "%BAK%" >nul
    echo   원본 백업: %BAK%
) else (
    echo   원본 백업 이미 존재
)

echo [3/4] 바이너리 패치 실행...
powershell -ExecutionPolicy Bypass -Command ^
  "$src = '%BIN%'; " ^
  "$bytes = [System.IO.File]::ReadAllBytes($src); " ^
  "$text = [System.Text.Encoding]::UTF8.GetString($bytes); " ^
  "$original = 'You are Antigravity Agent, a powerful agentic AI coding assistant designed by the Google engineering team.'; " ^
  "$replace = 'You are NeuronFS-Antigravity, an agentic AI coding assistant. Always think in Korean, answer in Korean.'; " ^
  "$diff = [System.Text.Encoding]::UTF8.GetByteCount($original) - [System.Text.Encoding]::UTF8.GetByteCount($replace); " ^
  "if ($diff -gt 0) { $replace += (' ' * $diff) }; " ^
  "$origArr = [System.Text.Encoding]::UTF8.GetBytes($original); " ^
  "$repArr = [System.Text.Encoding]::UTF8.GetBytes($replace); " ^
  "$found = $false; " ^
  "for ($i = 0; $i -lt $bytes.Length - $origArr.Length; $i++) { " ^
  "  $match = $true; " ^
  "  for ($j = 0; $j -lt $origArr.Length; $j++) { " ^
  "    if ($bytes[$i+$j] -ne $origArr[$j]) { $match = $false; break } " ^
  "  }; " ^
  "  if ($match) { " ^
  "    for ($j = 0; $j -lt $repArr.Length; $j++) { $bytes[$i+$j] = $repArr[$j] }; " ^
  "    $found = $true; break " ^
  "  } " ^
  "}; " ^
  "if ($found) { " ^
  "  [System.IO.File]::WriteAllBytes($src, $bytes); " ^
  "  Write-Host '  PATCH OK' " ^
  "} else { Write-Host '  PATCH FAIL: pattern not found' }"

echo [4/4] Antigravity 재시작...
start "" "C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"
echo.
echo ============================================
echo  완료! Antigravity가 재시작됩니다.
echo  문제 시 restore_binary.bat 실행
echo ============================================
pause
