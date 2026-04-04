# extension.js에서 base64 인코딩된 protobuf descriptor 추출
$content = [IO.File]::ReadAllText("C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\dist\extension.js")

# fileDesc 호출의 base64 문자열 추출 (패턴: fileDesc("...base64..."))
$regex = [regex]'fileDesc\("([A-Za-z0-9+/=]{20,})"\)'
$matches = $regex.Matches($content)

Write-Output "Found $($matches.Count) fileDesc calls"
$i = 0
foreach ($m in $matches) {
    $i++
    $b64 = $m.Groups[1].Value
    $bytes = [Convert]::FromBase64String($b64)
    $outPath = "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime\proto_desc_$i.bin"
    [IO.File]::WriteAllBytes($outPath, $bytes)
    
    # UTF-8로 읽어서 문자열 힌트를 확인
    $text = [Text.Encoding]::UTF8.GetString($bytes)
    $safePreview = $text.Substring(0, [Math]::Min($text.Length, 300)) -replace '[^\x20-\x7E]', '.'
    
    Write-Output ("=== Proto " + $i + " (" + $bytes.Length + " bytes) ===")
    Write-Output "Preview: $safePreview"
    Write-Output ""
}
