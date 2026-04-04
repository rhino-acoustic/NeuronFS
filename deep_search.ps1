# Deep search for ALL relevant keys in state.vscdb
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"
$workDbs = Get-ChildItem "$env:APPDATA\Antigravity\User\workspaceStorage" -Recurse -Filter "state.vscdb" -ErrorAction SilentlyContinue

$out = @()

# Global DB - ALL keys
$out += "=== ALL keys in globalStorage ==="
$allKeys = sqlite3 $globalDb "SELECT key, length(value) as len FROM ItemTable ORDER BY len DESC;"
foreach ($k in $allKeys) { $out += $k }

# Search for trajectory
$out += ""
$out += "=== Keys containing 'trajectory' ==="
$trajKeys = sqlite3 $globalDb "SELECT key, length(value) FROM ItemTable WHERE key LIKE '%trajectory%' OR key LIKE '%Trajectory%';"
if ($trajKeys) { foreach ($k in $trajKeys) { $out += $k } } else { $out += "NOT FOUND in global" }

# Check workspace DBs
foreach ($wdb in $workDbs) {
    $out += ""
    $out += "=== Workspace: $($wdb.Directory.Name) ==="
    $wKeys = sqlite3 $wdb.FullName "SELECT key, length(value) as len FROM ItemTable WHERE key LIKE '%chat%' OR key LIKE '%trajectory%' OR key LIKE '%Chat%' OR key LIKE '%gemini%' OR key LIKE '%conversation%' ORDER BY len DESC;"
    if ($wKeys) { foreach ($k in $wKeys) { $out += $k } } else { $out += "No chat/trajectory keys" }
}

$outPath = "$env:USERPROFILE\NeuronFS\deep_keys.txt"
$out | Out-File $outPath -Encoding UTF8
Write-Host "Done: $outPath"
