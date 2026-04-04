# Rebuild chat index from .pb files
# Since Antigravity is currently running and maintains the index in memory,
# we need to query the LIVE state.vscdb

$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"

# First check if the current running Antigravity has updated the index
$indexNow = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'chat.ChatSessionStore.index';"
Write-Host "Current index: $indexNow"

if ($indexNow -match '"entries":\s*\{\}') {
    Write-Host "[PROBLEM] Index is empty even while Antigravity is running!"
    Write-Host ""
    Write-Host "This means Antigravity stores the index in memory and the DB is only"
    Write-Host "written on graceful shutdown. When bat kills the process, index is lost."
    Write-Host ""
    
    # Build a minimal index from .pb file list
    $convDir = "$env:USERPROFILE\.gemini\antigravity\conversations"
    $pbFiles = Get-ChildItem $convDir -Filter "*.pb" -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending
    
    Write-Host "Found $($pbFiles.Count) .pb files"
    Write-Host "Building index entries..."
    
    $entries = @{}
    foreach ($pb in $pbFiles) {
        $id = $pb.BaseName
        $ts = $pb.LastWriteTime.ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ss.fffZ")
        $entries[$id] = @{
            "conversationId" = $id
            "lastModified" = $ts
            "title" = ""
        }
    }
    
    $index = @{
        "version" = 1
        "entries" = $entries
    }
    
    $json = $index | ConvertTo-Json -Depth 3 -Compress
    Write-Host "Built index with $($entries.Count) entries"
    Write-Host "Sample: $($json.Substring(0, [Math]::Min(200, $json.Length)))..."
    
    # Save backup
    $json | Out-File "$env:USERPROFILE\NeuronFS\chat_index_rebuilt.json" -Encoding UTF8 -NoNewline
    Write-Host "Saved rebuilt index to chat_index_rebuilt.json"
    
    # Write to DB (Antigravity must be stopped first for this to stick)
    $escaped = $json.Replace("'", "''")
    sqlite3 $globalDb "UPDATE ItemTable SET value = '$escaped' WHERE key = 'chat.ChatSessionStore.index';"
    Write-Host "[OK] Index written to state.vscdb"
} else {
    Write-Host "[OK] Index is not empty - backing up"
    $indexNow | Out-File "$env:USERPROFILE\NeuronFS\chat_index_backup.json" -Encoding UTF8 -NoNewline
}
