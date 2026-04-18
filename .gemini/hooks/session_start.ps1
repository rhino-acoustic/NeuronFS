#!/usr/bin/env pwsh
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "$env:USERPROFILE\NeuronFS"
$health = Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/ping" -TimeoutSec 3 -ErrorAction SilentlyContinue
if ($health.status -eq "ok") {
    Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/inject" -Method POST -TimeoutSec 10 -ErrorAction SilentlyContinue | Out-Null
    [Console]::Error.WriteLine("[HOOK] SessionStart: neuron inject OK")
}
$lastCommit = git log -1 --format="%h %s" 2>$null
[Console]::Error.WriteLine("[HOOK] SessionStart: $lastCommit")
Write-Output '{"decision":"allow"}'
exit 0
