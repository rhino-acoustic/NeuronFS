# 결과를 파일로 출력
$out = "C:\Users\BASEMENT_ADMIN\NeuronFS\diag_result.txt"

$procs = Get-CimInstance Win32_Process | Where-Object { $_.Name -eq 'Antigravity.exe' }
$sb = [System.Text.StringBuilder]::new()

$i = 0
foreach ($p in $procs) {
    $i++
    [void]$sb.AppendLine("=== PROCESS $i ===")
    [void]$sb.AppendLine("PID: $($p.ProcessId)")
    [void]$sb.AppendLine("ParentPID: $($p.ParentProcessId)")
    [void]$sb.AppendLine("CMD: $($p.CommandLine)")
    [void]$sb.AppendLine("")
}
[void]$sb.AppendLine("Total: $i processes")
[void]$sb.AppendLine("")

[void]$sb.AppendLine("=== ENV ===")
[void]$sb.AppendLine("NODE_OPTIONS: [$($env:NODE_OPTIONS)]")
[void]$sb.AppendLine("HTTPS_PROXY: [$($env:HTTPS_PROXY)]")
[void]$sb.AppendLine("HTTP_PROXY: [$($env:HTTP_PROXY)]")
[void]$sb.AppendLine("NODE_TLS: [$($env:NODE_TLS_REJECT_UNAUTHORIZED)]")

[void]$sb.AppendLine("")
[void]$sb.AppendLine("=== v4-hook.cjs ===")
$hookPath = "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime\v4-hook.cjs"
if (Test-Path $hookPath) {
    $hookContent = Get-Content $hookPath -Raw -ErrorAction SilentlyContinue
    [void]$sb.AppendLine($hookContent)
} else {
    [void]$sb.AppendLine("NOT FOUND")
}

[System.IO.File]::WriteAllText($out, $sb.ToString(), [System.Text.Encoding]::UTF8)
Write-Host "Done. Output: $out"
