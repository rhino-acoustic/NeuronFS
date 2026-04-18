#!/usr/bin/env pwsh
# NeuronFS Hook: 삭제 도구 → _quarantine 격리 강제
# BeforeTool: run_command (rm/del/Remove-Item 감지)
$ErrorActionPreference = "SilentlyContinue"

$input_json = [Console]::In.ReadToEnd()

# 삭제 명령어 감지
if ($input_json -match "Remove-Item|del |rm |rmdir|Delete") {
    # brain_v4 내부 파일 삭제 시도 감지
    if ($input_json -match "brain_v4|NeuronFS") {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: brain_v4 파일 삭제 시도 차단. _quarantine으로 격리하세요.")
        Write-Output '{"decision":"block","reason":"brain_v4 파일 직접 삭제 금지. _quarantine 디렉토리로 이동하세요."}'
        exit 2  # exit 2 = 물리적 차단
    }
}

Write-Output '{"decision":"allow"}'
exit 0
