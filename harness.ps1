# ═══════════════════════════════════════════════════════
# NeuronFS Harness — 같은 실수 반복 방지 자동 검증
# ═══════════════════════════════════════════════════════
# 
# 모든 禁 뉴런 + PD 교정 이력을 코드에서 자동 검증.
# 위반 시 경고. -Fix는 _proposals/에 제안만 기록 (직접 수정 금지).
# PD가 merge.ps1로 병합.
#
# 사용: .\harness.ps1 [-Fix]
# ═══════════════════════════════════════════════════════

param([switch]$Fix)

$ErrorActionPreference = "SilentlyContinue"
$brain = "C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4"
$runtime = "C:\Users\BASEMENT_ADMIN\NeuronFS\runtime"
$proposalDir = "C:\Users\BASEMENT_ADMIN\NeuronFS\_proposals"
$pass = 0; $fail = 0; $warn = 0
$violations = @()

function Check($id, $name, $test) {
    $result = & $test
    if ($result) {
        Write-Host "  ✅ [$id] $name" -ForegroundColor Green
        $script:pass++
        return $true
    } else {
        Write-Host "  ❌ [$id] $name" -ForegroundColor Red
        $script:fail++
        $script:violations += $name
        return $false
    }
}

function Warn($id, $name, $msg) {
    Write-Host "  ⚠️  [$id] $name — $msg" -ForegroundColor Yellow
    $script:warn++
}

function Fire($neuronPath) {
    $safePath = $neuronPath.Replace('\', '/')
    $obj = @{ path = $safePath }
    $jsonStr = ($obj | ConvertTo-Json -Compress)
    $body = [System.Text.Encoding]::UTF8.GetBytes($jsonStr)
    try {
        Invoke-RestMethod "http://localhost:9090/api/fire" -Method POST -ContentType "application/json; charset=utf-8" -Body $body -TimeoutSec 3 | Out-Null
    } catch {
        Write-Host "    fire skip: $neuronPath" -ForegroundColor DarkYellow
    }
}

function Propose($id, $title, $severity, $description, $fix) {
    if (-not $script:Fix) { return }
    if (-not (Test-Path $proposalDir)) { New-Item $proposalDir -ItemType Directory -Force | Out-Null }
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $filename = "${timestamp}_${id}.md"
    $content = @"
# $title
severity: $severity
harness_id: $id
created: $(Get-Date -Format "yyyy-MM-dd HH:mm")

## 문제
$description

## 수정 방법
``````powershell
$fix
``````
"@
    Set-Content (Join-Path $proposalDir $filename) $content -Encoding UTF8
    Write-Host "    → 제안 생성: _proposals/$filename" -ForegroundColor DarkCyan
}

Write-Host ""
Write-Host "╔═══════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   NeuronFS Harness — 실수 반복 방지   ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 0. 데이터 영속성 (OPERATIONS.md 체크리스트 A)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host "── 데이터 영속성 ──" -ForegroundColor Yellow

$r = Check "D01" ".gitignore에 brain_v4/ 없음" {
    $gi = Get-Content "C:\Users\BASEMENT_ADMIN\NeuronFS\.gitignore" -Raw -Encoding UTF8
    -not ($gi -match "^brain_v4/\s*$")
}
if (-not $r) {
    Propose "D01" ".gitignore에서 brain_v4/ 제거" "critical" `
        ".gitignore에 brain_v4/가 포함되어 Git이 뉴런 디렉토리를 추적하지 않음" `
        '(Get-Content ".gitignore") -replace "^brain_v4/\s*$","" | Set-Content ".gitignore"'
}

$r = Check "D02" "뉴런 카운트 250+" {
    try {
        $state = Invoke-RestMethod "http://localhost:9090/api/state" -TimeoutSec 5
        $state.totalNeurons -ge 250
    } catch { $false }
}

$r = Check "D03" "NAS 경로 접근 가능" {
    Test-Path "Z:\VOL1\VGVR\BRAIN\LW\system\neurons\brain_v4"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. 禁 뉴런 — 코드 위반 감지
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host ""
Write-Host "── 禁 규칙 검증 ──" -ForegroundColor Yellow

$r = Check "F01" "禁console.log" {
    $hits = Select-String -Path "$runtime\*.go" -Pattern "console\.log" 
    $hits.Count -eq 0
}
if ($r) { Fire "cortex/frontend\coding\禁console_log" }

$r = Check "F02" "禁inline style (Go 템플릿)" {
    $html = Get-Content "$runtime\dashboard_html.go" -Raw
    $inlines = [regex]::Matches($html, 'style="[^"]{30,}"')
    $inlines.Count -lt 5
}
if ($r) { Fire "cortex/frontend\coding\禁인라인스타일" }

$r = Check "F03" "禁평문 API 키" {
    $hits = Select-String -Path "$runtime\*.go" -Pattern "(sk-|gsk_|AKIA)[a-zA-Z0-9]{20,}" 
    ($hits -eq $null) -or ($hits.Count -eq 0)
}
if ($r) { Fire "cortex/security\禁평문_토큰" }

$r = Check "F04" "禁context stuffing (GEMINI.md < 15KB)" {
    $size = (Get-Item "C:\Users\BASEMENT_ADMIN\.gemini\GEMINI.md" -ErrorAction SilentlyContinue).Length
    $size -lt 15000
}
if ($r) { Fire "cortex/neuronfs\design\실재_온톨로지" }

Check "F05" "禁인위적 카운터 (max < 30)" {
    $high = Get-ChildItem $brain -Recurse -Filter "*.neuron" | Where-Object { $_.BaseName -match '^\d+$' -and [int]$_.BaseName -ge 30 }
    $high.Count -eq 0
} | Out-Null

$r = Check "F06" "禁brainstem 무단 변경" {
    $bs = "$brain\brainstem"
    $recent = Get-ChildItem $bs -Recurse -File | Where-Object { $_.LastWriteTime -gt (Get-Date).AddHours(-1) -and $_.Name -ne "_rules.md" }
    $recent.Count -eq 0
}
if ($r) { Fire "cortex/neuronfs/defense/brainstem_readonly" }

$r = Check "F07" "API fire UTF-8 encoding" {
    $testPath = "cortex/methodology/plan_then_execute"
    $obj = @{ path = $testPath }
    $jsonStr = ($obj | ConvertTo-Json -Compress)
    $body = [System.Text.Encoding]::UTF8.GetBytes($jsonStr)
    try {
        $res = Invoke-RestMethod "http://localhost:9090/api/fire" -Method POST -ContentType "application/json; charset=utf-8" -Body $body -TimeoutSec 3
        $res.status -eq "fired"
    } catch { $false }
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. PD 교정 이력 기반 검증
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host ""
Write-Host "── PD 교정 패턴 검증 ──" -ForegroundColor Yellow

$r = Check "P01" "禁SSOT 중복 (동일 뉴런)" {
    $inlineA = Test-Path "$brain\cortex\frontend\coding\禁inline_style"
    $inlineB = Test-Path "$brain\cortex\frontend\coding\禁인라인스타일"
    -not ($inlineA -and $inlineB)
}
if (-not $r) {
    Propose "P01" "SSOT 중복 제거" "warning" `
        "禁inline_style과 禁인라인스타일이 동시 존재" `
        'Remove-Item "$brain\cortex\frontend\coding\禁inline_style" -Recurse -Force'
}

$r = Check "P02" "강도 치환 작동 (절대/반드시 in _rules.md)" {
    $cortex = Get-Content "$brain\cortex\_rules.md" -Raw -Encoding UTF8
    $cortex -match "절대|반드시"
}

$r = Check "P03" "3-tier 분리 (7x _rules.md)" {
    $rules = Get-ChildItem $brain -Filter "_rules.md" -Recurse
    $rules.Count -ge 7
}

$r = Check "P04" "bomb 없음 (정상 운영)" {
    $bombs = Get-ChildItem $brain -Recurse -Filter "bomb.neuron"
    $bombs.Count -eq 0
}

Check "P05" "dormant 파일 존재 (가지치기 작동)" {
    $d = Get-ChildItem $brain -Recurse -Filter "*.dormant"
    $d.Count -gt 0
} | Out-Null

$r = Check "P06" "_rules.md 0바이트 없음" {
    $empty = Get-ChildItem $brain -Filter "_rules.md" -Recurse | Where-Object { $_.Length -eq 0 }
    $empty.Count -eq 0
}
if (-not $r) {
    Propose "P06" "빈 _rules.md 갱신" "warning" `
        "_rules.md가 0바이트인 리전 발견" `
        '& "C:\Users\BASEMENT_ADMIN\NeuronFS\neuronfs.exe" "C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4" --inject'
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. 멀티에이전트 검증
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host ""
Write-Host "── 멀티에이전트 검증 ──" -ForegroundColor Yellow

$r = Check "M01" "에이전트 디렉토리 구조 (bot1/entp/enfp/pm)" {
    (Test-Path "$brain\_agents\bot1\inbox") -and
    (Test-Path "$brain\_agents\entp\inbox") -and
    (Test-Path "$brain\_agents\enfp\inbox") -and
    (Test-Path "$brain\_agents\pm\inbox")
}

$r = Check "M02" "PM outbox 과적 아님 (< 150)" {
    $pm = @(Get-ChildItem "$brain\_agents\pm\outbox" -File -ErrorAction SilentlyContinue).Count
    $pm -lt 150
}
if (-not $r) {
    Propose "M02" "PM outbox pulse 정리" "warning" `
        "PM outbox에 파일 150개 이상 누적" `
        'Get-ChildItem "$brain\_agents\pm\outbox" -Filter "*pulse*" | Sort-Object LastWriteTime | Select-Object -SkipLast 10 | Remove-Item -Force'
}

Check "M03" "agent-bridge 프로세스" {
    $alive = $false
    Get-Process node -ErrorAction SilentlyContinue | ForEach-Object {
        try {
            $cmd = (Get-CimInstance Win32_Process -Filter "ProcessId=$($_.Id)").CommandLine
            if ($cmd -match "agent-bridge") { $alive = $true }
        } catch {}
    }
    $alive
} | Out-Null

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. 빌드 검증
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host ""
Write-Host "── 빌드 검증 ──" -ForegroundColor Yellow

$r = Check "B01" "Go build" {
    $tmpExe = Join-Path $env:TEMP "neuronfs_harness_test.exe"
    Remove-Item $tmpExe -Force -ErrorAction SilentlyContinue
    $saved = Get-Location
    Set-Location $script:runtime
    go build -o $tmpExe . 2>$null | Out-Null
    Set-Location $saved
    $exists = Test-Path $tmpExe
    Remove-Item $tmpExe -Force -ErrorAction SilentlyContinue
    $exists
}

$r = Check "B02" "API 응답 정상" {
    try {
        $state = Invoke-RestMethod "http://localhost:9090/api/state" -TimeoutSec 3
        $state.totalNeurons -gt 0
    } catch { $false }
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 결과
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write-Host ""
Write-Host "╔═══════════════════════════════════════╗" -ForegroundColor $(if ($fail -eq 0) { "Green" } else { "Red" })
Write-Host "║  PASS: $pass  |  FAIL: $fail  |  WARN: $warn" -ForegroundColor $(if ($fail -eq 0) { "Green" } else { "Red" })
Write-Host "╚═══════════════════════════════════════╝" -ForegroundColor $(if ($fail -eq 0) { "Green" } else { "Red" })

if ($violations.Count -gt 0) {
    Write-Host ""
    Write-Host "위반 사항:" -ForegroundColor Red
    $violations | ForEach-Object { Write-Host "  • $_" -ForegroundColor Red }
}

if ($Fix -and $fail -gt 0) {
    Write-Host ""
    Write-Host "제안이 _proposals/에 기록되었습니다." -ForegroundColor Cyan
    Write-Host "검토 후 .\merge.ps1 로 병합하세요." -ForegroundColor Cyan
}

# 세션 로그
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm"
$violationStr = if ($violations.Count -gt 0) { $violations -join ', ' } else { 'none' }
$logEntry = "$timestamp | PASS=$pass FAIL=$fail | violations: $violationStr"
$logDir = "$brain\hippocampus\session_log"
if (-not (Test-Path $logDir)) { New-Item $logDir -ItemType Directory -Force | Out-Null }
Add-Content "$logDir\harness_log.txt" $logEntry

if ($fail -eq 0) {
    Write-Host ""
    Write-Host "All clear. 같은 실수 없음." -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "$fail 개 위반 - _proposals/ 확인 후 merge.ps1 실행" -ForegroundColor Red
}
