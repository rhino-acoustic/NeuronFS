$settingsPath = Join-Path $env:APPDATA "Antigravity\User\settings.json"
$json = @'
{
    "files.autoGuessEncoding": true,
    "files.encoding": "utf8"
}
'@
[IO.File]::WriteAllText($settingsPath, $json, [Text.UTF8Encoding]::new($false))
Write-Host "DONE: $(Get-Content $settingsPath -Raw)"
