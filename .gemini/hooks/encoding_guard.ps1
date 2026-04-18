#!/usr/bin/env pwsh
# NeuronFS Hook: 한글 인코딩 안전장치
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "Get-Content" -and $input_json -notmatch "Encoding") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Get-Content without -Encoding → ReadAllText 권장")
}
if ($input_json -match "Set-Content" -and $input_json -notmatch "WriteAllText") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Set-Content → WriteAllText 권장 (BOM 방지)")
}
Write-Output '{"decision":"allow"}'
exit 0
