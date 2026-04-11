param (
    [Parameter(Mandatory=$true)]
    [string]$Path,

    [Parameter(Mandatory=$false)]
    [string]$Text = ""
)

$exePath = "c:\Users\BASEMENT_ADMIN\NeuronFS\neuronfs.exe"
$brainRoot = "c:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4"

$payload = @{ path = $Path }
if ($Text -ne "") { $payload.text = $Text }

$jsonStr = $payload | ConvertTo-Json -Compress
$jsonStr = $jsonStr.Replace('"', '\"')

cmd.exe /c """$exePath"" ""$brainRoot"" --tool grow ""$jsonStr"""
