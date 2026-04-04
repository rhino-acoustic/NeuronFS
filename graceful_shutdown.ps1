# Antigravity graceful shutdown - FULL LOGGING version
$logFile = Join-Path $PSScriptRoot "shutdown_log.txt"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

function Log($msg) {
    $line = "[$((Get-Date).ToString('HH:mm:ss.fff'))] $msg"
    Write-Host $line
    Add-Content -Path $logFile -Value $line -Encoding UTF8
}

Log "=== GRACEFUL SHUTDOWN START ($timestamp) ==="

# Check if Antigravity is running
$agProcs = Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue
if (-not $agProcs) {
    Log "Antigravity is NOT running. Nothing to shut down."
    exit 0
}
Log "Found $($agProcs.Count) Antigravity process(es)"
foreach ($p in $agProcs) {
    Log "  PID=$($p.Id) WindowHandle=$($p.MainWindowHandle) StartTime=$($p.StartTime)"
}

# Count .pb files BEFORE shutdown
$pbDir = "$env:USERPROFILE\.gemini\antigravity\conversations"
$pbBefore = (Get-ChildItem $pbDir -Filter "*.pb" -ErrorAction SilentlyContinue).Count
Log "PB files before shutdown: $pbBefore"

$cdpPort = 9000

# Step 1: Try CDP Browser.close via WebSocket
Log "Step 1: CDP Browser.close via WebSocket..."
try {
    $response = Invoke-WebRequest -Uri "http://127.0.0.1:$cdpPort/json/version" -UseBasicParsing -TimeoutSec 3
    $json = $response.Content | ConvertFrom-Json
    $wsUrl = $json.webSocketDebuggerUrl
    Log "CDP version endpoint OK. wsUrl=$wsUrl"

    if ($wsUrl) {
        Add-Type -AssemblyName System.Net.WebSockets
        $ws = [System.Net.WebSockets.ClientWebSocket]::new()
        $cts = [System.Threading.CancellationTokenSource]::new(5000)
        
        try {
            $ws.ConnectAsync([Uri]$wsUrl, $cts.Token).Wait()
            Log "WebSocket connected"
            
            $closeCmd = '{"id":1,"method":"Browser.close"}'
            $bytes = [System.Text.Encoding]::UTF8.GetBytes($closeCmd)
            $segment = [System.ArraySegment[byte]]::new($bytes)
            $ws.SendAsync($segment, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, $cts.Token).Wait()
            Log "Browser.close command SENT"
            
            # Wait for response
            $buf = [byte[]]::new(4096)
            $seg = [System.ArraySegment[byte]]::new($buf)
            try {
                $result = $ws.ReceiveAsync($seg, $cts.Token).Result
                $resp = [System.Text.Encoding]::UTF8.GetString($buf, 0, $result.Count)
                Log "CDP response: $resp"
            } catch {
                Log "CDP response timeout (expected if browser closed immediately)"
            }
            $ws.Dispose()
        } catch {
            Log "WebSocket FAILED: $_"
            try { $null = Invoke-WebRequest -Uri "http://127.0.0.1:$cdpPort/json/close" -UseBasicParsing -TimeoutSec 3 -ErrorAction SilentlyContinue } catch {}
            Log "Fallback: /json/close sent"
        }
    }
} catch {
    Log "CDP not available: $_"
}

# Step 2: CloseMainWindow
Log "Step 2: Sending CloseMainWindow..."
$procs = Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue
$windowCount = 0
foreach ($p in $procs) {
    if ($p.MainWindowHandle -ne 0) {
        $p.CloseMainWindow() | Out-Null
        $windowCount++
    }
}
Log "CloseMainWindow sent to $windowCount window(s)"

# Step 3: Wait for graceful exit (max 30 seconds)
Log "Step 3: Waiting for process exit (max 30s)..."
$waited = 0
while ($waited -lt 30) {
    Start-Sleep -Seconds 2
    $waited += 2
    $still = Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue
    if (-not $still) {
        Log "Antigravity EXITED gracefully after $waited seconds"
        $pbAfter = (Get-ChildItem $pbDir -Filter "*.pb" -ErrorAction SilentlyContinue).Count
        Log "PB files after shutdown: $pbAfter (delta: $($pbAfter - $pbBefore))"
        Log "=== SHUTDOWN COMPLETE ==="
        exit 0
    }
    Log "  Still alive after $waited sec ($($still.Count) procs)"
}

# Step 4: Force kill
Log "FORCE KILL after 30s timeout"
Stop-Process -Name "Antigravity" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2
$pbAfter = (Get-ChildItem $pbDir -Filter "*.pb" -ErrorAction SilentlyContinue).Count
Log "PB files after FORCE kill: $pbAfter (delta: $($pbAfter - $pbBefore))"
Log "=== SHUTDOWN COMPLETE (FORCED) ==="
