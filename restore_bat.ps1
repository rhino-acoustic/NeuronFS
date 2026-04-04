# Restore from backup first
$batPath = "$env:USERPROFILE\NeuronFS\start_v4_swarm.bat"
$bak = "$env:USERPROFILE\NeuronFS\start_v4_swarm.bat.bak"
Copy-Item $bak $batPath -Force
Write-Host "Restored from .bak"
Write-Host "Size: $((Get-Item $batPath).Length) bytes"
