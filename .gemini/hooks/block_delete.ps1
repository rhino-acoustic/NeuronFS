#!/usr/bin/env pwsh
# NeuronFS Hook: brain_v4 삭제 차단 → _quarantine 격리 강제
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "Remove-Item|del |rm |rmdir|Delete") {
    if ($input_json -match "brain_v4|NeuronFS") {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: brain_v4 직접 삭제 금지")
        Write-Output '{"decision":"block","reason":"brain_v4 파일 직접 삭제 금지. _quarantine으로 이동하세요."}'
        exit 2
    }
}
Write-Output '{"decision":"allow"}'
exit 0
