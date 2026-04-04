# Read trajectorySummaries
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"
$val = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'antigravityUnifiedStateSync.trajectorySummaries';"
$outPath = "$env:USERPROFILE\NeuronFS\trajectory_summaries.json"
$val | Out-File $outPath -Encoding UTF8 -NoNewline
Write-Host "Length: $($val.Length)"
Write-Host "Saved to: $outPath"

# Also read sidebarWorkspaces
$wsVal = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'antigravityUnifiedStateSync.sidebarWorkspaces';"
Write-Host ""
Write-Host "sidebarWorkspaces length: $($wsVal.Length)"
$wsVal | Out-File "$env:USERPROFILE\NeuronFS\sidebar_workspaces.json" -Encoding UTF8 -NoNewline
