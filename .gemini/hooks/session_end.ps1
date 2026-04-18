#!/usr/bin/env pwsh
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "$env:USERPROFILE\NeuronFS"
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>$null
    git commit -m "[session-end] auto-save" --allow-empty 2>$null
    [Console]::Error.WriteLine("[HOOK] SessionEnd: auto-commit")
}
Write-Output '{"decision":"allow"}'
exit 0
