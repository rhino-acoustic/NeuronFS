[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$brain = "C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4"

$map = @{}
$map["cortex\frontend\coding\no_console_log"] = "절대 금지: console log"
$map["cortex\tools\avoid_general_commands"] = "절대 금지: general commands"
$map["cortex\tools\avoid_ls"] = "절대 금지: ls usage"
$map["cortex\tools\avoid_g"] = "절대 금지: g usage"
$map["cortex\tools\avoid_ls_and_g"] = "절대 금지: ls and g"
$map["cortex\tools\avoid_ls_for_directories"] = "절대 금지: ls for directories"
$map["cortex\tools\adopt_list_dir"] = "추천: list dir"
$map["cortex\tools\adopt_precise_tools"] = "추천: precise tools"
$map["cortex\tools\adopt_specific_tools"] = "추천: specific tool usage"
$map["cortex\tools\list_dir_instead_of_ls"] = "추천: list dir"
$map["cortex\tools\precise_tool_usage"] = "추천: precise tool usage"
$map["cortex\tools\specific_tool_usage"] = "추천: specific tool usage"
$map["cortex\tools\use_list_dir"] = "추천: list dir"
$map["cortex\tools\use_precise_tools"] = "추천: precise tools"
$map["cortex\tools\use_specific_tools"] = "추천: specific tools"
$map["cortex\tool_usage\precise_tools_only"] = "추천: precise tools only"
$map["cortex\communication"] = "소통"
$map["cortex\thought"] = "사고"
$map["cortex\strategy\strategic_depth"] = "전략적 깊이"
$map["cortex\security\least_agency_principle"] = "최소권한 원칙"
$map["cortex\neuronfs\naming"] = "명명규칙"
$map["cortex\neuronfs\emit_kanji_dedup"] = "한자 중복제거"
$map["cortex\neuronfs\dual_gemini_sync"] = "듀얼 gemini 동기화"
$map["cortex\neuronfs\runtime\idle_auto_decay"] = "유휴자동감쇠"
$map["cortex\neuronfs\runtime\modtime_sync_fixed"] = "수정시간 동기화"
$map["cortex\skills\crawler\instagram_cdp_pipeline"] = "인스타 CDP 파이프라인"
$map["cortex\frontend\typography"] = "타이포그래피"
$map["cortex\neuronfs\design"] = "설계"
$map["cortex\neuronfs\ops"] = "운영"
$map["cortex\backend\devops"] = "데브옵스"
$map["ego\proverbs"] = "격언"
$map["ego\communication\concise_execution"] = "간결 실행중심"
$map["ego\communication\structured_and_systematic"] = "구조화 체계적"
$map["ego\language\korean_thought"] = "한국어 사고"
$map["hippocampus\methodology"] = "방법론"
$map["hippocampus\quality"] = "품질"
$map["hippocampus\session_log"] = "세션로그"
$map["sensors\brand"] = "브랜드"
$map["sensors\environment"] = "환경"
$map["prefrontal\project\github_public_preparation"] = "깃허브 공개 준비"
$map["prefrontal\todo\groq_auto_neuronize"] = "Groq 자동뉴런화"

$done = 0; $merged = 0; $skipped = 0

foreach ($key in $map.Keys) {
    $oldPath = Join-Path $brain $key
    $newName = $map[$key]
    if (-not (Test-Path $oldPath)) { $skipped++; continue }
    
    $parentDir = Split-Path $oldPath -Parent
    $newPath = Join-Path $parentDir $newName
    
    if (Test-Path $newPath) {
        $oldN = Get-ChildItem $oldPath -Filter '*.neuron' -ErrorAction SilentlyContinue
        $newN = Get-ChildItem $newPath -Filter '*.neuron' -ErrorAction SilentlyContinue
        $o = 0; $n = 0
        if ($oldN) { $o = ($oldN | ForEach-Object { [int]($_.BaseName -replace '\D') } | Sort-Object -Descending)[0] }
        if ($newN) { $n = ($newN | ForEach-Object { [int]($_.BaseName -replace '\D') } | Sort-Object -Descending)[0] }
        $total = $o + $n
        Get-ChildItem $newPath -Filter '*.neuron' | Remove-Item -Force
        New-Item (Join-Path $newPath "$total.neuron") -ItemType File -Force | Out-Null
        Remove-Item $oldPath -Recurse -Force
        Write-Host "MERGE: $key -> $newName ($o+$n=$total)"
        $merged++
    } else {
        Rename-Item $oldPath $newName
        Write-Host "RENAME: $key -> $newName"
        $done++
    }
}
Write-Host "`nResult: $done renamed, $merged merged, $skipped skipped"
