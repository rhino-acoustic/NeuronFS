@echo off
if exist "%~dp0.secrets\groq_api_key" set /p GROQ_API_KEY=<"%~dp0.secrets\groq_api_key"
if exist "%~dp0.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%~dp0.secrets\anthropic_api_key"

REM Kill previous neuronfs instances
taskkill /F /IM neuronfs.exe >nul 2>&1

REM Auto-upgrade binary
if exist "%~dp0neuronfs_new.exe" (
    echo [NeuronFS] Upgrading binary...
    move /Y "%~dp0neuronfs.exe" "%~dp0neuronfs.exe~" >nul 2>&1
    move /Y "%~dp0neuronfs_new.exe" "%~dp0neuronfs.exe" >nul 2>&1
)

:loop
echo [NeuronFS] Starting...
"%~dp0neuronfs.exe" "%~dp0brain_v4" --supervisor

if exist "%~dp0_reboot_request" (
    del "%~dp0_reboot_request"
    echo [NeuronFS] Reboot requested. Restarting in 3 seconds...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 3 /nobreak >nul
    goto :loop
)
echo [NeuronFS] Process exited normally.
pause