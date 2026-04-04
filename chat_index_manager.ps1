# Antigravity Chat Index Backup/Restore
# bat에서 시작 전/후에 호출

param(
    [string]$Action  # "backup" or "restore"
)

$globalDb = "$env:APPDATA\Antigravity\User\globalStorage\state.vscdb"
$backupFile = "$env:USERPROFILE\NeuronFS\chat_index_backup.json"

if ($Action -eq "backup") {
    # Antigravity 실행 중에 인덱스를 백업
    try {
        $indexVal = sqlite3 $globalDb "SELECT value FROM ItemTable WHERE key = 'chat.ChatSessionStore.index';"
        if ($indexVal -and $indexVal.Length -gt 30) {
            $indexVal | Out-File $backupFile -Encoding UTF8 -NoNewline
            Write-Host "[OK] Chat index backed up ($($indexVal.Length) chars)"
        } else {
            Write-Host "[SKIP] Index is empty, nothing to backup"
        }
    } catch {
        Write-Host "[ERROR] Backup failed: $_"
    }
}
elseif ($Action -eq "restore") {
    # Antigravity 종료 상태에서 인덱스를 복원
    if (Test-Path $backupFile) {
        try {
            $backup = Get-Content $backupFile -Raw -Encoding UTF8
            if ($backup -and $backup.Length -gt 30) {
                # Escape single quotes for SQLite
                $escaped = $backup.Replace("'", "''")
                sqlite3 $globalDb "UPDATE ItemTable SET value = '$escaped' WHERE key = 'chat.ChatSessionStore.index';"
                Write-Host "[OK] Chat index restored ($($backup.Length) chars)"
            } else {
                Write-Host "[SKIP] Backup is empty"
            }
        } catch {
            Write-Host "[ERROR] Restore failed: $_"
        }
    } else {
        Write-Host "[SKIP] No backup file found"
    }
}
else {
    Write-Host "Usage: script.ps1 -Action backup|restore"
}
