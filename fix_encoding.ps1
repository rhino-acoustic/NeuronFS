# Step 1: Read current bat (CP949)
$batPath = "$env:USERPROFILE\NeuronFS\start_v4_swarm.bat"
$cp949 = [System.Text.Encoding]::GetEncoding(949)
$raw = [System.IO.File]::ReadAllBytes($batPath)

# Skip BOM if present
$offset = 0
if ($raw.Length -ge 3 -and $raw[0] -eq 0xEF -and $raw[1] -eq 0xBB -and $raw[2] -eq 0xBF) {
    $offset = 3
}
$content = $cp949.GetString($raw, $offset, $raw.Length - $offset)

# Step 2: Fix known broken Korean (from previous bad edits)
# These are mojibake patterns from CP949->UTF8->CP949 double encoding
$fixes = @{
    # From the damaged line 193 area
    "supervisor`u{AE30}" = "supervisor"
}

# Step 3: Fix NODE_OPTIONS line
$oldNode = 'set "NODE_OPTIONS=--require C:/Users/BASEMENT_ADMIN/NeuronFS/runtime/v4-hook.cjs"'
$newNode = ":: DISABLED: v4-hook patches http2 inside Antigravity, breaks chat persistence`r`n:: set ""NODE_OPTIONS=--require C:/Users/BASEMENT_ADMIN/NeuronFS/runtime/v4-hook.cjs"""
$content = $content.Replace($oldNode, $newNode)

# Step 4: Write as UTF-8 BOM
$utf8bom = New-Object System.Text.UTF8Encoding($true)
[System.IO.File]::WriteAllText($batPath, $content, $utf8bom)

# Verify
$verify = [System.IO.File]::ReadAllBytes($batPath)
$isBom = ($verify[0] -eq 0xEF -and $verify[1] -eq 0xBB -and $verify[2] -eq 0xBF)
Write-Host "UTF-8 BOM: $isBom"
Write-Host "Size: $($verify.Length) bytes"

# Show Korean lines to verify
$verifyContent = [System.Text.Encoding]::UTF8.GetString($verify, 3, $verify.Length - 3)
$lines = $verifyContent -split "`r?`n"
Write-Host "Lines: $($lines.Count)"
for ($i = 0; $i -lt $lines.Count; $i++) {
    $ln = $lines[$i]
    if ($ln -match "NODE_OPTIONS|supervisor|DISABLED") {
        Write-Host "L$($i+1): $ln"
    }
}
