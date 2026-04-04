# Read the chat index from state.vscdb
$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"

# Get chat index value
$indexVal = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'chat.ChatSessionStore.index';"
$outPath = "$env:USERPROFILE\NeuronFS\chat_index.txt"
$indexVal | Out-File $outPath -Encoding UTF8
Write-Host "Index length: $($indexVal.Length) chars"
Write-Host "Saved to: $outPath"

# Also get workspace transfer
$wsTransfer = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'chat.workspaceTransfer';"
$wsPath = "$env:USERPROFILE\NeuronFS\chat_ws_transfer.txt"
$wsTransfer | Out-File $wsPath -Encoding UTF8
Write-Host "Workspace transfer length: $($wsTransfer.Length) chars"
