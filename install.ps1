<#
.SYNOPSIS
NeuronFS Installer for Windows
Usage: iwr https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.ps1 -useb | iex
#>

Write-Host "🧠 Installing NeuronFS..." -ForegroundColor Cyan

# 1. Check requirements
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Error: 'go' is not installed. Please install Go (1.22+) first." -ForegroundColor Red
    exit 1
}
if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Error: 'git' is not installed." -ForegroundColor Red
    exit 1
}

$InstallDir = Join-Path $env:USERPROFILE ".neuronfs"
$BinDir = Join-Path $env:USERPROFILE "AppData\Local\Microsoft\WindowsApps" # Common PATH dir, or custom

# 2. Setup directories
if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir | Out-Null }

$RepoDir = Join-Path $InstallDir "repo"

# 3. Clone or Update
if (Test-Path $RepoDir) {
    Write-Host "🔄 Updating existing installation..." -ForegroundColor Yellow
    Set-Location $RepoDir
    git pull origin main
} else {
    Write-Host "📦 Cloning NeuronFS repository..." -ForegroundColor Yellow
    git clone https://github.com/rhino-acoustic/NeuronFS.git $RepoDir
    Set-Location $RepoDir
}

# 4. Build
Write-Host "🔨 Building core engine..." -ForegroundColor Yellow
Set-Location (Join-Path $RepoDir "runtime")
go build -o "neuronfs.exe" .

# 5. Move to PATH
$TargetBin = Join-Path $BinDir "neuronfs.exe"
Copy-Item "neuronfs.exe" -Destination $TargetBin -Force

Write-Host "✅ Installed successfully to $TargetBin!" -ForegroundColor Green
Write-Host "🚀 Run 'neuronfs --init ./my_brain' to create your first autonomous brain." -ForegroundColor Cyan
