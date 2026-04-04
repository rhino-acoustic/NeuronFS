$batPath = Join-Path $env:USERPROFILE "NeuronFS\start_v4_swarm.bat"
$raw = [IO.File]::ReadAllBytes($batPath)
$crCount = 0
$lfCount = 0
foreach ($b in $raw) {
    if ($b -eq 13) { $crCount++ }
    if ($b -eq 10) { $lfCount++ }
}
Write-Host "CR: $crCount, LF: $lfCount"
Write-Host "First 3 bytes: $($raw[0]) $($raw[1]) $($raw[2])"
if ($crCount -eq 0 -and $lfCount -gt 0) {
    Write-Host "PROBLEM: LF only (Unix line endings). cmd.exe needs CRLF!"
    # Fix: convert LF to CRLF
    $content = [IO.File]::ReadAllText($batPath)
    $content = $content.Replace("`r`n", "`n").Replace("`n", "`r`n")
    [IO.File]::WriteAllText($batPath, $content, [Text.UTF8Encoding]::new($false))
    Write-Host "FIXED: Converted to CRLF"
    # Verify
    $raw2 = [IO.File]::ReadAllBytes($batPath)
    $cr2 = 0
    foreach ($b in $raw2) { if ($b -eq 13) { $cr2++ } }
    Write-Host "After fix - CR: $cr2, Size: $($raw2.Length)"
}
