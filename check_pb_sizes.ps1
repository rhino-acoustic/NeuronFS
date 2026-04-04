$dir = "$env:USERPROFILE\.gemini\antigravity\conversations"
$files = Get-ChildItem $dir -Filter "*.pb" | Sort-Object Length
Write-Host "=== Smallest 5 .pb files ==="
$files | Select-Object -First 5 | ForEach-Object {
    Write-Host ("{0} | {1:N1} KB" -f $_.Name, ($_.Length/1024))
}
Write-Host ""
Write-Host "=== Largest 5 .pb files ==="
$files | Select-Object -Last 5 | ForEach-Object {
    Write-Host ("{0} | {1:N1} KB" -f $_.Name, ($_.Length/1024))
}
Write-Host ""
Write-Host ("Total: {0} files | Min: {1:N1} KB | Max: {2:N1} KB" -f $files.Count, ($files[0].Length/1024), ($files[-1].Length/1024))
