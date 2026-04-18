@echo off
title NeuronFS_Supervisor
color 0A

set "NFSROOT=%~dp0."
set "NEURONFS_AG_WORKSPACE=%NFSROOT%"

if exist "%NFSROOT%\.secrets\groq_api_key" set /p GROQ_API_KEY=<"%NFSROOT%\.secrets\groq_api_key"
if exist "%NFSROOT%\.secrets\anthropic_api_key" set /p ANTHROPIC_API_KEY=<"%NFSROOT%\.secrets\anthropic_api_key"

echo [NeuronFS] Cleaning up previous processes...
taskkill /F /IM neuronfs.exe >nul 2>&1
timeout /t 2 /nobreak >nul

tasklist /FI "IMAGENAME eq neuronfs.exe" 2>NUL | find /I "neuronfs.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo [NeuronFS] Force killing remaining...
    taskkill /F /IM neuronfs.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
)

set "BRAIN=%NFSROOT%\brain_v4"

:loop
echo.
echo [NeuronFS] Starting...
echo.

"%NFSROOT%\neuronfs.exe" "%BRAIN%" --supervisor

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