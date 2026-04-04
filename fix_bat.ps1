$batPath = "C:\Users\BASEMENT_ADMIN\NeuronFS\start_v4_swarm.bat"
$bytes = [System.IO.File]::ReadAllBytes($batPath)
$content = [System.Text.Encoding]::Default.GetString($bytes)

# Line 29: comment out NODE_OPTIONS
$oldLine = 'set "NODE_OPTIONS=--require C:/Users/BASEMENT_ADMIN/NeuronFS/runtime/v4-hook.cjs"'
$newLine = ":: DISABLED: v4-hook patches http2 inside Antigravity, breaks chat persistence`r`n:: set ""NODE_OPTIONS=--require C:/Users/BASEMENT_ADMIN/NeuronFS/runtime/v4-hook.cjs"""

if ($content.Contains($oldLine)) {
    $content = $content.Replace($oldLine, $newLine)
    [System.IO.File]::WriteAllBytes($batPath, [System.Text.Encoding]::Default.GetBytes($content))
    Write-Host "DONE: NODE_OPTIONS line commented out"
} else {
    Write-Host "SKIP: line already modified or not found"
}
