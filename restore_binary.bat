@echo off
chcp 65001 >nul
echo ============================================
echo  RESTORE: language_server binary
echo ============================================
echo.

set "BIN=C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
set "BAK=%BIN%.bak_original"

if not exist "%BAK%" (
    echo ERROR: No backup found!
    echo   %BAK%
    pause
    exit /b 1
)

echo [1/3] Killing Antigravity...
taskkill /F /IM Antigravity.exe >nul 2>&1
taskkill /F /IM language_server_windows_x64.exe >nul 2>&1
echo   Waiting 5 seconds...
timeout /t 5 /nobreak >nul

echo [2/3] Restoring original binary...
copy /Y "%BAK%" "%BIN%" >nul
if %ERRORLEVEL% EQU 0 (
    echo   OK - Restored
) else (
    echo   FAIL - Could not copy
    pause
    exit /b 1
)

echo [3/3] Starting Antigravity...
start "" "C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"
echo.
echo ============================================
echo  DONE! Original binary restored.
echo ============================================
pause
