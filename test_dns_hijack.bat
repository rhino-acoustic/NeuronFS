@echo off
echo ============================================
echo   DNS Hijack Reality Test (Layer 5)
echo ============================================
echo.
echo This test will:
echo   1. Resolve real IP of cloudcode-pa.googleapis.com
echo   2. Add hosts file entry (127.0.0.1)
echo   3. Start DNS Hijack Proxy on port 443
echo   4. Restart Antigravity
echo   5. Capture LLM traffic
echo.
echo REQUIRES: Run as Administrator!
echo.

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Please run as Administrator!
    echo Right-click this file and select "Run as administrator"
    pause
    exit /b 1
)

cd /d "C:\Users\BASEMENT_ADMIN\NeuronFS"

echo [1/5] Resolving real IP first (before hosts change)...
for /f "tokens=*" %%i in ('powershell -NoProfile -Command "(Resolve-DnsName cloudcode-pa.googleapis.com -Type A | Select-Object -First 1).IPAddress"') do set REAL_IP=%%i
echo   Real IP: %REAL_IP%
if "%REAL_IP%"=="" (
    echo [ERROR] Cannot resolve DNS. Check internet connection.
    pause
    exit /b 1
)

echo.
echo [2/5] Modifying hosts file...
set HOSTS=%SystemRoot%\System32\drivers\etc\hosts
findstr /C:"cloudcode-pa.googleapis.com" "%HOSTS%" >nul 2>&1
if %errorlevel% equ 0 (
    echo   Entry already exists in hosts file
) else (
    echo 127.0.0.1  cloudcode-pa.googleapis.com >> "%HOSTS%"
    echo   Added: 127.0.0.1  cloudcode-pa.googleapis.com
)
echo   Flushing DNS cache...
ipconfig /flushdns >nul 2>&1

echo.
echo [3/5] Stopping existing Antigravity...
taskkill /IM "Antigravity.exe" >nul 2>&1
timeout /t 3 /nobreak >nul
taskkill /F /IM "Antigravity.exe" >nul 2>&1

echo.
echo [4/5] Starting DNS Hijack Proxy (needs port 443)...
set "NODE_TLS_REJECT_UNAUTHORIZED=0"
start "DNS-Hijack-Proxy" cmd /c "node C:\Users\BASEMENT_ADMIN\NeuronFS\runtime\hijackers\dns-hijack-proxy.mjs 2>&1 | tee C:\Users\BASEMENT_ADMIN\NeuronFS\dns_hijack.log"
timeout /t 3 /nobreak >nul

echo.
echo [5/5] Starting Antigravity...
set "NODE_TLS_REJECT_UNAUTHORIZED=0"
start "" "%LOCALAPPDATA%\Programs\Antigravity\Antigravity.exe" --remote-debugging-port=9000 --ignore-certificate-errors

echo.
echo ============================================
echo   DNS Hijack Proxy running!
echo.
echo   Say something to AI in Antigravity, then
echo   come back here and press any key to check.
echo.
echo   Captures: brain_v4\_transcripts\_dns_hijack_dumps\
echo ============================================
echo.
pause

echo.
echo === DNS Hijack Captures ===
if exist "brain_v4\_transcripts\_dns_hijack_dumps" (
    for /f "delims=" %%F in ('dir /b /o-d "brain_v4\_transcripts\_dns_hijack_dumps\capture_*.json" 2^>nul') do (
        echo --- %%F ---
        type "brain_v4\_transcripts\_dns_hijack_dumps\%%F" 2>nul
        echo.
        goto :DONE_CAP
    )
    :DONE_CAP
    echo.
    echo Total captures:
    dir /b "brain_v4\_transcripts\_dns_hijack_dumps\capture_*.json" 2>nul | find /c /v ""
) else (
    echo   No captures found
)

echo.
echo Press any key to cleanup (restore hosts, stop proxy)...
pause

echo.
echo [Cleanup] Removing hosts entry...
powershell -NoProfile -Command "(Get-Content '%HOSTS%') | Where-Object { $_ -notmatch 'cloudcode-pa\.googleapis\.com' } | Set-Content '%HOSTS%'"
ipconfig /flushdns >nul 2>&1

echo [Cleanup] Stopping proxy...
powershell -NoProfile -Command "Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match 'dns-hijack-proxy' } | Invoke-CimMethod -MethodName Terminate" >nul 2>&1

echo [Cleanup] Stopping Antigravity...
taskkill /F /IM "Antigravity.exe" >nul 2>&1

echo.
echo Done. Run start_v4_swarm.bat to restart clean.
pause
