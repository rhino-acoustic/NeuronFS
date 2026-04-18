#!/usr/bin/env pwsh
# NeuronFS Hook: go vet 실패 시 커밋 차단
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "git commit" -and $input_json -match "runtime") {
    Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"
    $vet = go vet ./runtime/... 2>&1
    if ($LASTEXITCODE -ne 0) {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: go vet failed")
        [Console]::Error.WriteLine($vet)
        Write-Output '{"decision":"block","reason":"go vet failed"}'
        exit 2
    }
    [Console]::Error.WriteLine("[HOOK] go vet PASS")
}
Write-Output '{"decision":"allow"}'
exit 0
