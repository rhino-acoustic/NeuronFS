$dir = "$env:USERPROFILE\.gemini\antigravity\conversations"
$files = Get-ChildItem $dir -Filter "*.pb" | Sort-Object LastWriteTime -Descending | Select-Object -First 10
foreach ($f in $files) {
    $kb = [math]::Round($f.Length/1024, 1)
    Write-Host "$($f.Name) | $kb KB | $($f.LastWriteTime)"
}
Write-Host ""
Write-Host "Total .pb files: $((Get-ChildItem $dir -Filter '*.pb').Count)"
