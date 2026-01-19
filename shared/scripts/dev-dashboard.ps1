# Start the dashboard development server
# Usage: .\shared\scripts\dev-dashboard.ps1

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Starting Orchion Dashboard..." -ForegroundColor Cyan
Write-Host ""

$dashboardDir = "$projectRoot\dashboard"

# Check if node_modules exists
if (-not (Test-Path "$dashboardDir\node_modules")) {
    Write-Host "Installing dependencies..." -ForegroundColor Yellow
    Push-Location $dashboardDir
    try {
        npm install
        if ($LASTEXITCODE -ne 0) {
            throw "npm install failed"
        }
        Write-Host "[OK] Dependencies installed" -ForegroundColor Green
    } finally {
        Pop-Location
    }
    Write-Host ""
}

# Start dev server
Write-Host "Starting development server..." -ForegroundColor Yellow
Push-Location $dashboardDir
try {
    npm run dev
} finally {
    Pop-Location
}
