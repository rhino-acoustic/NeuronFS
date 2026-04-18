#!/usr/bin/env pwsh
# NeuronFS Hook: 파일 수정 전 자동 git snapshot
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>$null
    $ts = (Get-Date).ToString("HH:mm:ss")
    git commit -m "[hook] pre-edit snapshot $ts" --allow-empty 2>$null
    [Console]::Error.WriteLine("[HOOK] git snapshot at $ts")
}
Write-Output '{"decision":"allow"}'
exit 0
