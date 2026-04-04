# NeuronFS 디렉토리에서 CP949 (non-UTF8) 파일 스캔
$scanDir = "$env:USERPROFILE\NeuronFS"
$exts = @("*.bat", "*.cmd", "*.txt", "*.cfg", "*.ini", "*.conf")
$results = @()

foreach ($ext in $exts) {
    $files = Get-ChildItem $scanDir -Filter $ext -Recurse -ErrorAction SilentlyContinue |
        Where-Object { -not $_.PSIsContainer -and $_.Length -gt 0 -and $_.FullName -notmatch 'node_modules|\.git' }
    
    foreach ($f in $files) {
        $raw = [System.IO.File]::ReadAllBytes($f.FullName)
        if ($raw.Length -lt 3) { continue }
        
        $isUtf8Bom = ($raw[0] -eq 0xEF -and $raw[1] -eq 0xBB -and $raw[2] -eq 0xBF)
        $isUtf16 = ($raw[0] -eq 0xFF -and $raw[1] -eq 0xFE)
        
        # Check if content has high bytes (non-ASCII = likely Korean)
        $hasHighBytes = $false
        foreach ($b in $raw) {
            if ($b -gt 127) { $hasHighBytes = $true; break }
        }
        
        if ($hasHighBytes -and -not $isUtf8Bom -and -not $isUtf16) {
            $rel = $f.FullName.Replace($scanDir, ".")
            $results += "$rel ($($f.Length) bytes) - NOT UTF-8 BOM, has non-ASCII"
        }
    }
}

Write-Host "=== CP949/Non-UTF8 files with Korean ==="
if ($results.Count -eq 0) {
    Write-Host "None found. All clear."
} else {
    foreach ($r in $results) {
        Write-Host $r
    }
    Write-Host "`nTotal: $($results.Count) files at risk"
}
