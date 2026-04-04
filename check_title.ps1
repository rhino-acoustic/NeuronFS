Get-Process -Name "Antigravity" -ErrorAction SilentlyContinue | ForEach-Object {
    Write-Host "PID=$($_.Id) Title=[$($_.MainWindowTitle)]"
}
