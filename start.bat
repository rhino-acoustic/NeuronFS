@echo off
REM Kill previous NeuronFS cmd windows
taskkill /FI "WINDOWTITLE eq NeuronFS_Supervisor*" /F >nul 2>&1
title NeuronFS_Supervisor

set "NFSROOT=%~dp0."
set "NEURONFS_AG_WORKSPACE=%NFSROOT%"
if exist "%NFSROOT%\.secrets\groq_api_key" set /p GROQ_API_KEY=<"%NFSROOT%\.secrets\groq_api_key"
if exist "%NFSROOT%\.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%NFSROOT%\.secrets\anthropic_api_key"

REM Check if NeuronFS is already running
tasklist /FI "IMAGENAME eq neuronfs.exe" 2>NUL | find /I /N "neuronfs.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo [NeuronFS] neuronfs.exe is already running. 
    echo [NeuronFS] Attempting graceful shutdown (SIGTERM)...
    taskkill /IM neuronfs.exe >nul 2>&1
    timeout /t 5 /nobreak
)

REM Last resort for stuck processes (Log instead of silent /F)
tasklist /FI "IMAGENAME eq neuronfs.exe" 2>NUL | find /I /N "neuronfs.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo [NeuronFS] WARNING: Process still alive. Manual intervention may be needed to follow 'No Force Kill' rule.
    echo [NeuronFS] If necessary, use: taskkill /F /IM neuronfs.exe
    pause
)

set "BRAIN=%NFSROOT%\brain_v4"

:loop
echo [NeuronFS] Starting...
"%NFSROOT%\neuronfs.exe" "%BRAIN%" --supervisor

REM If _reboot_request exists, auto-restart
if exist "%NFSROOT%\_reboot_request" (
    del "%NFSROOT%\_reboot_request"
    echo [NeuronFS] Reboot requested...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 3 /nobreak >nul
    goto :loop
)
echo [NeuronFS] Process exited normally.
pause