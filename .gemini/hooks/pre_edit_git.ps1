#!/usr/bin/env pwsh
# NeuronFS Hook: 파일 수정 전 자동 git snapshot
# BeforeTool: replace_file_content, write_to_file, multi_replace_file_content
$ErrorActionPreference = "SilentlyContinue"

# stdin에서 JSON 읽기 (Gemini CLI가 도구 정보 전달)
$input_json = [Console]::In.ReadToEnd()

# git snapshot
Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>&1 | Out-Null
    $ts = (Get-Date).ToString("HH:mm:ss")
    git commit -m "[hook] pre-edit snapshot $ts" --allow-empty 2>&1 | Out-Null
    [Console]::Error.WriteLine("[HOOK] git snapshot created at $ts")
}

# 통과 (exit 0 = 허용)
Write-Output '{"decision":"allow"}'
exit 0
