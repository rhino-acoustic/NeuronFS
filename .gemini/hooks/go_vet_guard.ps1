#!/usr/bin/env pwsh
# NeuronFS Hook: go vet 검증 없이 커밋/재시작 차단
# BeforeTool: run_command에서 git commit 또는 restart 감지 시 go vet 선행
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()

# go 코드 수정 후 commit 시도 감지
if ($input_json -match "git commit" -and $input_json -match "runtime") {
    Set-Location "C:\Users\BASEMENT_ADMIN\NeuronFS"
    $vet = go vet ./runtime/... 2>&1
    if ($LASTEXITCODE -ne 0) {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: go vet 실패 — 커밋 불가")
        [Console]::Error.WriteLine($vet)
        Write-Output '{"decision":"block","reason":"go vet failed. Fix errors before commit."}'
        exit 2
    }
    [Console]::Error.WriteLine("[HOOK] go vet PASS")
}

# restart/start.bat 시도 시 빌드 확인
if ($input_json -match "start\.bat|neuronfs\.exe" -and $input_json -match "Start-Process|cmd.exe") {
    $exe = "C:\Users\BASEMENT_ADMIN\NeuronFS\neuronfs.exe"
    $src = "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime"
    if (Test-Path $exe) {
        $exeTime = (Get-Item $exe).LastWriteTime
        $srcTime = Get-ChildItem $src -Filter "*.go" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
        if ($srcTime -and $srcTime.LastWriteTime -gt $exeTime) {
            [Console]::Error.WriteLine("[HOOK] WARNING: 소스가 바이너리보다 새로움 — 빌드 먼저 권장")
        }
    }
}

Write-Output '{"decision":"allow"}'
exit 0
