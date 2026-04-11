@echo off
set "NFSDIR=%~dp0dist\neuronfs"
set "NEURONFS_AG_WORKSPACE=%~dp0."
if exist "%NFSDIR%\.secrets\groq_api_key" set /p GROQ_API_KEY=<"%NFSDIR%\.secrets\groq_api_key"
if exist "%NFSDIR%\.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%NFSDIR%\.secrets\anthropic_api_key"

REM Kill previous neuronfs instances
taskkill /F /IM neuronfs.exe >nul 2>&1

REM Auto-upgrade binary if new version exists
if exist "%NFSDIR%\neuronfs_new.exe" (
    echo [NeuronFS] Upgrading binary...
    move /Y "%NFSDIR%\neuronfs.exe" "%NFSDIR%\neuronfs.exe~" >nul 2>&1
    move /Y "%NFSDIR%\neuronfs_new.exe" "%NFSDIR%\neuronfs.exe" >nul 2>&1
)

REM Use main NeuronFS brain, not dist copy
set "BRAIN=%~dp0brain_v4"

:loop
echo [NeuronFS] Starting...
"%NFSDIR%\neuronfs.exe" "%BRAIN%" --supervisor

REM If _reboot_request exists, auto-restart (telegram 159487 code)
if exist "%NFSDIR%\_reboot_request" (
    del "%NFSDIR%\_reboot_request"
    echo [NeuronFS] Reboot requested. Restarting in 3 seconds...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 3 /nobreak >nul
    goto :loop
)
echo [NeuronFS] Process exited normally.
pause