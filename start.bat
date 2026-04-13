@echo off
REM Kill previous NeuronFS cmd windows (이전 start.bat 창 닫기)
taskkill /FI "WINDOWTITLE eq NeuronFS_Supervisor*" /F >nul 2>&1
title NeuronFS_Supervisor

set "NFSDIR=%~dp0."
set "NEURONFS_AG_WORKSPACE=%~dp0."
if exist "%NFSDIR%\.secrets\groq_api_key" set /p GROQ_API_KEY=<"%NFSDIR%\.secrets\groq_api_key"
if exist "%NFSDIR%\.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%NFSDIR%\.secrets\anthropic_api_key"

REM Kill ALL previous neuronfs processes (clean slate)
taskkill /F /IM neuronfs.exe >nul 2>&1
timeout /t 2 /nobreak >nul
taskkill /F /IM neuronfs.exe >nul 2>&1

REM Build latest binary from source
echo [NeuronFS] Building from source...
pushd "%NFSDIR%\runtime"
go build -o "%NFSDIR%\neuronfs.exe" .
popd
echo [NeuronFS] Build complete.

REM Use main NeuronFS brain (SSOT)
set "BRAIN=%NFSDIR%\brain_v4"

:loop
echo [NeuronFS] Starting...
"%NFSDIR%\neuronfs.exe" "%BRAIN%" --supervisor

REM If _reboot_request exists, auto-restart
if exist "%NFSDIR%\_reboot_request" (
    del "%NFSDIR%\_reboot_request"
    echo [NeuronFS] Reboot requested. Restarting in 3 seconds...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 3 /nobreak >nul
    goto :loop
)
echo [NeuronFS] Process exited normally.
pause