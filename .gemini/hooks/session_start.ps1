#!/usr/bin/env pwsh
# NeuronFS Hook: SessionStart — 세션 시작 시 뉴런 컨텍스트 완전 로드
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()

Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"

# 1. GEMINI.md 최신화 (emit 트리거)
$health = Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/ping" -TimeoutSec 3 -ErrorAction SilentlyContinue
if ($health.status -eq "ok") {
    Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/inject" -Method POST -TimeoutSec 10 -ErrorAction SilentlyContinue | Out-Null
    [Console]::Error.WriteLine("[HOOK] SessionStart: neuron inject OK (uptime=$($health.uptime))")
} else {
    [Console]::Error.WriteLine("[HOOK] SessionStart: API offline — skipping inject")
}

# 2. TODO 큐 로드
$todoPath = "C:\Users\BASEMENT_ADMIN\.gemini\antigravity\knowledge\neuronfs_session_progress\artifacts\session_progress.md"
if (Test-Path $todoPath) {
    $todos = Select-String "^- \[ \]" $todoPath | Measure-Object
    [Console]::Error.WriteLine("[HOOK] SessionStart: $($todos.Count) pending TODOs")
}

# 3. 마지막 커밋 알림
$lastCommit = git log -1 --format="%h %s" 2>$null
[Console]::Error.WriteLine("[HOOK] SessionStart: last commit = $lastCommit")

Write-Output '{"decision":"allow"}'
exit 0
