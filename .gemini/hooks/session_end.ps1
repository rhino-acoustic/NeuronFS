#!/usr/bin/env pwsh
# NeuronFS Hook: SessionEnd — 세션 종료 시 진행상황 저장
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()

Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"

# 1. 변경사항 자동 커밋
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>$null
    $ts = (Get-Date).ToString("yyyy-MM-dd_HH:mm")
    git commit -m "[session-end] auto-save $ts" --allow-empty 2>$null
    [Console]::Error.WriteLine("[HOOK] SessionEnd: auto-commit at $ts")
}

# 2. 전사 파일 보존 알림
$txDir = "C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4\_transcripts"
$today = Get-ChildItem $txDir -Filter "*$(Get-Date -Format 'yyyy-MM-dd')*" -ErrorAction SilentlyContinue
[Console]::Error.WriteLine("[HOOK] SessionEnd: $($today.Count) transcripts today")

Write-Output '{"decision":"allow"}'
exit 0
