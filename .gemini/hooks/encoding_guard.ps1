#!/usr/bin/env pwsh
# NeuronFS Hook: 한글 깨짐 방지 + WriteAllText 강제
# BeforeTool: run_command에서 Get-Content/Set-Content 감지 → 안전한 대안 안내
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()

# Get-Content 한글 깨짐 경고
if ($input_json -match "Get-Content" -and $input_json -notmatch "Encoding") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Get-Content without -Encoding UTF8 → 한글 깨짐 위험. [System.IO.File]::ReadAllText 사용 권장")
}

# Set-Content → WriteAllText 권장
if ($input_json -match "Set-Content" -and $input_json -notmatch "WriteAllText") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Set-Content → [System.IO.File]::WriteAllText 사용 권장 (BOM 방지)")
}

Write-Output '{"decision":"allow"}'
exit 0
