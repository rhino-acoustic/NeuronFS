# Decode trajectorySummaries (base64 -> bytes -> printable text)
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"
$val = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'antigravityUnifiedStateSync.trajectorySummaries';"

# Decode base64
$bytes = [Convert]::FromBase64String($val)
$text = [Text.Encoding]::UTF8.GetString($bytes)
$printable = $text -replace '[^\x20-\x7E\r\n]', '.'

$out = @()
$out += "Raw length: $($val.Length)"
$out += "Decoded bytes: $($bytes.Length)"
$out += ""
$out += "=== Decoded (printable) ==="
$out += $printable
$out += ""
$out += "=== Hex dump (first 100 bytes) ==="
$hexLine = ""
for ($i = 0; $i -lt [Math]::Min(100, $bytes.Length); $i++) {
    $hexLine += "{0:X2} " -f $bytes[$i]
}
$out += $hexLine

$outPath = "$env:USERPROFILE\NeuronFS\trajectory_decoded.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
