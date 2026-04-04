Get-ChildItem 'C:\Users\BASEMENT_ADMIN\NeuronFS' -Recurse -Include '*.jsonl','*.json' | ForEach-Object {
    $c = Get-Content $_.FullName -Raw -Encoding UTF8 -ErrorAction SilentlyContinue
    if ($c -and $c.Contains('bomb')) {
        Write-Host "FOUND: $($_.FullName)"
    }
}
Write-Host "Done searching"
