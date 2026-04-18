@echo off
chcp 65001 >nul
echo ============================================
echo  NeuronFS System Prompt RESTORE (원복)
echo  language_server 바이너리 원본 복원
echo ============================================
echo.

set "BIN=C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
set "BAK=%BIN%.bak_original"

if not exist "%BAK%" (
    echo ERROR: 원본 백업이 없습니다!
    echo   %BAK%
    pause
    exit /b 1
)

echo [1/3] Antigravity 종료...
taskkill /F /IM Antigravity.exe >nul 2>&1
taskkill /F /IM language_server_windows_x64.exe >nul 2>&1
timeout /t 5 /nobreak >nul

echo [2/3] 원본 복원...
copy /Y "%BAK%" "%BIN%" >nul
echo   복원 완료

echo [3/3] Antigravity 재시작...
start "" "C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"
echo.
echo ============================================
echo  원복 완료! 원본 바이너리로 복원됨.
echo ============================================
pause
