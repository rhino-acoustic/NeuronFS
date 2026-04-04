@echo off
echo ==============================================
echo   NeuronFS Swarm Shutdown (Killing Daemons...)
echo ==============================================
echo.
echo [1/3] Killing Node.js Hijackers and Executors...
taskkill /F /FI "WINDOWTITLE eq NeuronFS*" /T >nul 2>&1
taskkill /F /IM node.exe /FI "COMM eq node.exe *hijack*" >nul 2>&1

echo [2/3] Killing NeuronFS Go Supervisor...
taskkill /F /IM neuronfs.exe /T >nul 2>&1

echo [3/3] Clearing Zombie processes...
taskkill /F /IM cmd.exe /FI "WINDOWTITLE eq NeuronFS*" >nul 2>&1

echo.
echo [OK] All Background Swarm Daemons have successfully terminated.
echo.
pause
