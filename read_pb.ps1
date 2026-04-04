# Read first 500 bytes of the current conversation .pb file to understand structure
$convDir = "$env:USERPROFILE\.gemini\antigravity\conversations"
$pbFile = Join-Path $convDir "6aa264bc-e491-431d-a3f0-074a63c8fcaf.pb"

$raw = [IO.File]::ReadAllBytes($pbFile)
$out = @()
$out += "File size: $($raw.Length) bytes"
$out += ""

# Try reading as UTF-8 text (first 2000 chars)
$text = [Text.Encoding]::UTF8.GetString($raw, 0, [Math]::Min(2000, $raw.Length))
# Show printable content
$printable = $text -replace '[^\x20-\x7E\r\n]', '.'
$out += "=== First 2000 bytes (printable) ==="
$out += $printable

# Also check another smaller pb file for comparison
$smallPb = Join-Path $convDir "403325b5-83fd-43e4-9f8f-a03482de6fd2.pb"
if (Test-Path $smallPb) {
    $raw2 = [IO.File]::ReadAllBytes($smallPb)
    $text2 = [Text.Encoding]::UTF8.GetString($raw2, 0, [Math]::Min(1000, $raw2.Length))
    $printable2 = $text2 -replace '[^\x20-\x7E\r\n]', '.'
    $out += ""
    $out += "=== Small .pb file (first 1000 bytes) ==="
    $out += "Size: $($raw2.Length)"
    $out += $printable2
}

$outPath = "$env:USERPROFILE\NeuronFS\pb_structure.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
