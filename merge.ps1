# ═══════════════════════════════════════════════════════
# NeuronFS merge.ps1 — Strangler Fig 병합 도구
# ═══════════════════════════════════════════════════════
# 
# _proposals/ 에 쌓인 변경 제안을 PD가 검토 후 병합.
# AI/harness는 시스템 파일을 직접 수정할 수 없음.
# 이 스크립트는 PD(인간)만 실행.
#
# 사용:
#   .\merge.ps1               # 대기 중 제안 목록 표시
#   .\merge.ps1 -Apply <ID>   # 특정 제안 병합
#   .\merge.ps1 -Reject <ID>  # 특정 제안 거절
#   .\merge.ps1 -ApplyAll     # 모든 제안 병합
# ═══════════════════════════════════════════════════════

param(
    [string]$Apply,
    [string]$Reject,
    [switch]$ApplyAll
)

$proposalDir = Join-Path $PSScriptRoot "_proposals"
$archiveDir = Join-Path $proposalDir "_archive"

if (-not (Test-Path $proposalDir)) {
    Write-Host "❌ _proposals/ 디렉토리 없음" -ForegroundColor Red
    exit 1
}

# 대기 중 제안 목록
$pending = Get-ChildItem $proposalDir -Filter "*.md" -File | Where-Object { $_.Name -notmatch "^_" }

if (-not $Apply -and -not $Reject -and -not $ApplyAll) {
    Write-Host ""
    Write-Host "═══ NeuronFS 제안 목록 ═══" -ForegroundColor Cyan
    Write-Host ""
    if ($pending.Count -eq 0) {
        Write-Host "  (대기 중 제안 없음)" -ForegroundColor DarkGray
    } else {
        foreach ($p in $pending) {
            $content = Get-Content $p.FullName -Raw -Encoding UTF8
            $title = ($content -split "`n" | Select-Object -First 1).TrimStart("# ").Trim()
            $severity = if ($content -match "severity:\s*(\w+)") { $Matches[1] } else { "info" }
            $color = switch ($severity) { "critical" { "Red" } "warning" { "Yellow" } default { "White" } }
            Write-Host "  [$($p.BaseName)] $title" -ForegroundColor $color
        }
        Write-Host ""
        Write-Host "  사용법:" -ForegroundColor DarkGray
        Write-Host "    .\merge.ps1 -Apply <ID>    제안 병합" -ForegroundColor DarkGray
        Write-Host "    .\merge.ps1 -Reject <ID>   제안 거절" -ForegroundColor DarkGray
        Write-Host "    .\merge.ps1 -ApplyAll      전체 병합" -ForegroundColor DarkGray
    }
    Write-Host ""
    exit 0
}

# 아카이브 디렉토리
if (-not (Test-Path $archiveDir)) { New-Item $archiveDir -ItemType Directory -Force | Out-Null }

function Process-Proposal {
    param([System.IO.FileInfo]$File, [string]$Action)
    
    $content = Get-Content $File.FullName -Raw -Encoding UTF8
    $title = ($content -split "`n" | Select-Object -First 1).TrimStart("# ").Trim()
    
    if ($Action -eq "apply") {
        # 제안 내 patch 블록 실행
        $patchBlocks = [regex]::Matches($content, '```patch\s*\n([\s\S]*?)```')
        $scriptBlocks = [regex]::Matches($content, '```powershell\s*\n([\s\S]*?)```')
        
        if ($scriptBlocks.Count -gt 0) {
            Write-Host "  ⚡ 실행 중: $title" -ForegroundColor Yellow
            foreach ($block in $scriptBlocks) {
                $script = $block.Groups[1].Value.Trim()
                Write-Host "    $($script.Substring(0, [Math]::Min(80, $script.Length)))..." -ForegroundColor DarkGray
                try {
                    Invoke-Expression $script
                    Write-Host "    ✅ 완료" -ForegroundColor Green
                } catch {
                    Write-Host "    ❌ 실패: $_" -ForegroundColor Red
                }
            }
        } elseif ($patchBlocks.Count -gt 0) {
            Write-Host "  📋 패치 블록 $($patchBlocks.Count)개 — 수동 적용 필요" -ForegroundColor Yellow
            foreach ($block in $patchBlocks) {
                Write-Host $block.Groups[1].Value -ForegroundColor DarkGray
            }
        } else {
            Write-Host "  📋 수동 적용 제안: $title" -ForegroundColor Yellow
        }
        
        # 아카이브로 이동 (applied_ 접두어)
        $archiveName = "applied_$($File.Name)"
        Move-Item $File.FullName (Join-Path $archiveDir $archiveName) -Force
        Write-Host "  ✅ 병합 완료 → _archive/$archiveName" -ForegroundColor Green
        
    } elseif ($Action -eq "reject") {
        $archiveName = "rejected_$($File.Name)"
        Move-Item $File.FullName (Join-Path $archiveDir $archiveName) -Force
        Write-Host "  🗑 거절됨 → _archive/$archiveName" -ForegroundColor DarkYellow
    }
}

if ($Apply) {
    $file = $pending | Where-Object { $_.BaseName -eq $Apply -or $_.Name -eq $Apply }
    if (-not $file) { Write-Host "❌ 제안 '$Apply' 없음" -ForegroundColor Red; exit 1 }
    Process-Proposal -File $file -Action "apply"
}

if ($Reject) {
    $file = $pending | Where-Object { $_.BaseName -eq $Reject -or $_.Name -eq $Reject }
    if (-not $file) { Write-Host "❌ 제안 '$Reject' 없음" -ForegroundColor Red; exit 1 }
    Process-Proposal -File $file -Action "reject"
}

if ($ApplyAll) {
    Write-Host ""
    Write-Host "═══ 전체 병합 ($($pending.Count)개) ═══" -ForegroundColor Cyan
    foreach ($p in $pending) {
        Process-Proposal -File $p -Action "apply"
    }
    Write-Host ""
    Write-Host "✅ 전체 병합 완료" -ForegroundColor Green
}
