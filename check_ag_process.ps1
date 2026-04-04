$procs = Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue | Where-Object { $_.MainWindowHandle -ne 0 } | Select-Object -First 1
if (-not $procs) {
    $procs = Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue | Select-Object -First 1
}
if ($procs) {
    $wmi = Get-CimInstance Win32_Process -Filter "ProcessId = $($procs.Id)"
    Write-Host "PID: $($procs.Id)"
    Write-Host "StartTime: $($procs.StartTime)"
    Write-Host "CommandLine: $($wmi.CommandLine)"
    Write-Host ""
    Write-Host "=== Key Env Vars ==="
    # Check env vars via WMI
    $envBlock = [System.Diagnostics.Process]::GetProcessById($procs.Id)
    Write-Host "WorkingSet: $([math]::Round($procs.WorkingSet64/1MB, 1)) MB"
} else {
    Write-Host "Antigravity not running"
}
