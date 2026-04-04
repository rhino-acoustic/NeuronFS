# state.vscdb is SQLite - check if it has conversation index
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"
$workDb = Get-ChildItem "$env:APPDATA\Antigravity\User\workspaceStorage" -Recurse -Filter "state.vscdb" -ErrorAction SilentlyContinue | Select-Object -First 1

$out = @()
$out += "=== Global state.vscdb ==="
$out += "Path: $globalDb"
$out += "Size: $([math]::Round((Get-Item $globalDb).Length/1024, 1)) KB"
$out += "Modified: $((Get-Item $globalDb).LastWriteTime)"

if ($workDb) {
    $out += ""
    $out += "=== Workspace state.vscdb ==="
    $out += "Path: $($workDb.FullName)"
    $out += "Size: $([math]::Round($workDb.Length/1024, 1)) KB"
    $out += "Modified: $($workDb.LastWriteTime)"
}

# Check if sqlite3 is available
$hasSqlite = $false
try { $null = Get-Command sqlite3 -ErrorAction Stop; $hasSqlite = $true } catch {}

if ($hasSqlite) {
    $out += ""
    $out += "=== Tables in globalStorage ==="
    $tables = sqlite3 $globalDb ".tables"
    $out += $tables
    $out += ""
    $out += "=== Keys containing 'chat' or 'conversation' or 'history' ==="
    $keys = sqlite3 $globalDb "SELECT key FROM ItemTable WHERE key LIKE '%chat%' OR key LIKE '%conversation%' OR key LIKE '%history%' OR key LIKE '%gemini%';"
    $out += $keys
} else {
    # Try reading raw bytes for conversation-related strings
    $out += ""
    $out += "=== Searching for conversation keys in raw bytes ==="
    $rawContent = [IO.File]::ReadAllText($globalDb, [Text.Encoding]::UTF8)
    $patterns = @('conversation', 'chatHistory', 'gemini.chat', 'chat_history', 'history')
    foreach ($p in $patterns) {
        $idx = $rawContent.IndexOf($p)
        if ($idx -ge 0) {
            $snippet = $rawContent.Substring([Math]::Max(0, $idx - 20), [Math]::Min(100, $rawContent.Length - $idx + 20))
            $clean = $snippet -replace '[^\x20-\x7E]', '.'
            $out += "FOUND '$p' at pos $idx : $clean"
        } else {
            $out += "NOT FOUND: '$p'"
        }
    }
}

$outPath = "$env:USERPROFILE\NeuronFS\vscdb_check.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
