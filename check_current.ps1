# 현재 state 확인
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"

$indexNow = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'chat.ChatSessionStore.index';"
$out = @()
$out += "=== Current index ==="
$out += "Length: $($indexNow.Length) chars"
$out += "Value: $($indexNow.Substring(0, [Math]::Min(500, $indexNow.Length)))"

# Check .pb files
$convDir = "$env:USERPROFILE\.gemini\antigravity\conversations"
$pbFiles = Get-ChildItem $convDir -Filter "*.pb" | Sort-Object LastWriteTime -Descending | Select-Object -First 5
$out += ""
$out += "=== Latest .pb files ==="
foreach ($pb in $pbFiles) {
    $out += "$($pb.Name) | $([math]::Round($pb.Length/1024,1)) KB | $($pb.LastWriteTime)"
}

# Check all keys related to chat
$out += ""
$out += "=== All chat keys ==="
$allKeys = sqlite3 $globalDb "SELECT key, length(value) FROM ItemTable WHERE key LIKE '%chat%' OR key LIKE '%Chat%';"
$out += $allKeys

$outPath = "$env:USERPROFILE\NeuronFS\current_state.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
