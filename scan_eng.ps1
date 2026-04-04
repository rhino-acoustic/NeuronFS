$brain = "C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4"
$dirs = Get-ChildItem -Recurse -Directory $brain | Where-Object {
    $_.Name -match '^[a-zA-Z_]+$' -and
    (Get-ChildItem $_.FullName -Filter '*.neuron' -ErrorAction SilentlyContinue).Count -gt 0
}
foreach ($d in $dirs) {
    $neurons = Get-ChildItem $d.FullName -Filter '*.neuron'
    $max = ($neurons | ForEach-Object { [int]($_.BaseName -replace '\D','') } | Sort-Object -Descending | Select-Object -First 1)
    $rel = $d.FullName.Replace($brain + "\", "")
    Write-Host "$max`t$rel"
}
