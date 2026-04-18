@echo off
title NeuronFS_Supervisor
color 0A

REM ═══════════════════════════════════════════════
REM  NeuronFS Supervisor Launcher
REM  - 이전 프로세스 정리 후 단일 인스턴스 재구동
REM ═══════════════════════════════════════════════

set "NFSROOT=%~dp0."
set "NEURONFS_AG_WORKSPACE=%NFSROOT%"

REM 시크릿 로드
if exist "%NFSROOT%\.secrets\groq_api_key" set /p GROQ_API_KEY=<"%NFSROOT%\.secrets\groq_api_key"
if exist "%NFSROOT%\.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%NFSROOT%\.secrets\anthropic_api_key"

REM ── 1. 이전 프로세스 강제 정리 ──
echo [NeuronFS] Cleaning up previous processes...
taskkill /F /IM neuronfs.exe >nul 2>&1
taskkill /FI "WINDOWTITLE eq NeuronFS_Supervisor*" /FI "PID ne %PID%" /F >nul 2>&1
timeout /t 2 /nobreak >nul

REM ── 2. 잔여 확인 ──
tasklist /FI "IMAGENAME eq neuronfs.exe" 2>NUL | find /I "neuronfs.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo [NeuronFS] WARNING: Process still alive, force killing...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
)

set "BRAIN=%NFSROOT%\brain_v4"

REM ── 3. 메인 루프 ──
:loop
echo.
echo ╔══════════════════════════════════════╗
echo ║  NeuronFS Starting...               ║
echo ╚══════════════════════════════════════╝
echo.

"%NFSROOT%\neuronfs.exe" "%BRAIN%" --supervisor

REM 종료 후 reboot 요청 체크
if exist "%NFSROOT%\_reboot_request" (
    del "%NFSROOT%\_reboot_request" >nul 2>&1
    echo [NeuronFS] Reboot requested, restarting in 3s...
    timeout /t 3 /nobreak >nul
    goto :loop
)

echo.
echo [NeuronFS] Process exited. Press any key to restart, or Ctrl+C to quit.
pause >nul
goto :loop