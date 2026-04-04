@echo off
echo ============================================
echo   ALPN MITM Reality Test
echo ============================================
echo.

cd /d "C:\Users\BASEMENT_ADMIN\NeuronFS"

echo [1/4] Stopping Antigravity...
taskkill /IM "Antigravity.exe" >nul 2>&1
timeout /t 5 /nobreak >nul
taskkill /F /IM "Antigravity.exe" >nul 2>&1

echo [2/4] Starting MITM Proxy (context-hijacker.mjs)...
start /B "" node "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime\hijackers\context-hijacker.mjs"
timeout /t 3 /nobreak >nul

echo [3/4] Starting Antigravity (MITM + CDP mode)...
set "HTTPS_PROXY=http://127.0.0.1:8080"
set "HTTP_PROXY=http://127.0.0.1:8080"
set "NODE_TLS_REJECT_UNAUTHORIZED=0"
start "" "%LOCALAPPDATA%\Programs\Antigravity\Antigravity.exe" --remote-debugging-port=9000 --ignore-certificate-errors

set RETRIES=0
:WAIT
timeout /t 2 /nobreak >nul
set /a RETRIES+=1
powershell -Command "try{$null=Invoke-WebRequest -Uri 'http://127.0.0.1:9000/json/list' -UseBasicParsing -TimeoutSec 2;exit 0}catch{exit 1}" >nul 2>&1
if %errorlevel% equ 0 goto READY
if %RETRIES% gtr 15 goto TIMEOUT
echo   Waiting... (%RETRIES%/15)
goto WAIT

:TIMEOUT
echo [ERROR] CDP timeout. Antigravity may not support CDP.
pause
exit /b 1

:READY
echo [4/4] CDP connected!

start /B "" node "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime\hijackers\hijack-launcher.mjs"

echo.
echo ============================================
echo   Ready! Now:
echo   1. Say something to AI in Antigravity
echo   2. Come back here and press any key
echo.
echo   Capture locations:
echo     MITM: brain_v4\_agents\global_inbox\grpc_dumps\
echo     CDP:  brain_v4\_transcripts\
echo ============================================
echo.
pause

echo.
echo === MITM Capture Results ===
echo.
if exist "brain_v4\_agents\global_inbox\latest_hijacked_context.md" (
    echo --- latest_hijacked_context.md ---
    type "brain_v4\_agents\global_inbox\latest_hijacked_context.md"
) else (
    echo   No MITM capture found
)

echo.
echo === MITM grpc_dumps ===
if exist "brain_v4\_agents\global_inbox\grpc_dumps" (
    dir /b "brain_v4\_agents\global_inbox\grpc_dumps\raw_dump_*" 2>nul
    for /f "delims=" %%F in ('dir /b /o-d "brain_v4\_agents\global_inbox\grpc_dumps\raw_dump_*" 2^>nul') do (
        echo --- %%F ---
        type "brain_v4\_agents\global_inbox\grpc_dumps\%%F" 2>nul
        echo.
        goto :DONE_DUMP
    )
    :DONE_DUMP
) else (
    echo   No grpc_dumps directory
)

echo.
echo === CDP Transcript ===
for /f "delims=" %%F in ('dir /b /o-d "brain_v4\_transcripts\*.txt" 2^>nul') do (
    echo --- %%F ---
    powershell -Command "Get-Content 'brain_v4\_transcripts\%%F' -Tail 10 -Encoding utf8" 2>nul
    goto :DONE_TR
)
:DONE_TR

echo.
echo ============================================
echo   If you see plaintext chat above = SUCCESS
echo   If empty = MITM blocked (cert pinning)
echo ============================================
echo.
echo Press any key to cleanup (stop MITM Antigravity)
pause

taskkill /F /IM "Antigravity.exe" >nul 2>&1
powershell -Command "Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match 'context-hijacker' } | Invoke-CimMethod -MethodName Terminate" >nul 2>&1
echo Done. Run start_v4_swarm.bat to restart clean.
pause
