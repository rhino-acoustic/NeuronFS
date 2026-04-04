# Simplified check - write to file
$out = @()

$dbFiles = Get-ChildItem "$env:APPDATA\Antigravity" -Recurse -Filter "*.vscdb" -ErrorAction SilentlyContinue
$out += "=== VSCDB Files ==="
foreach ($f in $dbFiles) {
    $kb = [math]::Round($f.Length / 1024, 1)
    $out += "$($f.Name) | $kb KB | $($f.LastWriteTime)"
}

$convDir = "$env:USERPROFILE\.gemini\antigravity\conversations"
if (Test-Path $convDir) {
    $pbFiles = Get-ChildItem $convDir -Filter "*.pb" -ErrorAction SilentlyContinue
    $out += ""
    $out += "=== Conversation .pb files ==="
    $out += "Total: $($pbFiles.Count) files"
    foreach ($pb in ($pbFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 5)) {
        $out += "$($pb.Name) | $([math]::Round($pb.Length/1024,1)) KB | $($pb.LastWriteTime)"
    }
}

$brainDir = "$env:USERPROFILE\.gemini\antigravity\brain"
if (Test-Path $brainDir) {
    $out += ""
    $out += "=== Brain conversation dirs ==="
    $dirs = Get-ChildItem $brainDir -Directory -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending | Select-Object -First 5
    foreach ($d in $dirs) {
        $out += "$($d.Name) | $($d.LastWriteTime)"
    }
}

$outPath = "$env:USERPROFILE\NeuronFS\storage_check.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
