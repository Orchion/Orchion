# Build all Orchion components
# Usage: .\shared\scripts\build-all.ps1

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Building all Orchion components..." -ForegroundColor Cyan
Write-Host ""

# Build Orchestrator
Write-Host "Building Orchestrator..." -ForegroundColor Yellow
Push-Location "$projectRoot\orchestrator"
try {
    go build -o orchestrator.exe ./cmd/orchestrator
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Orchestrator built successfully" -ForegroundColor Green
    } else {
        throw "Orchestrator build failed"
    }
} finally {
    Pop-Location
}

Write-Host ""

# Build Node Agent
Write-Host "Building Node Agent..." -ForegroundColor Yellow
Push-Location "$projectRoot\node-agent"
try {
    go build -o node-agent.exe ./cmd/node-agent
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Node Agent built successfully" -ForegroundColor Green
    } else {
        throw "Node Agent build failed"
    }
} finally {
    Pop-Location
}

Write-Host ""

# Build Dashboard (production build)
Write-Host "Building Dashboard..." -ForegroundColor Yellow
Push-Location "$projectRoot\dashboard"
try {
    # Check if node_modules exists
    if (-not (Test-Path "node_modules")) {
        Write-Host "  Installing dependencies..." -ForegroundColor Gray
        npm install
        if ($LASTEXITCODE -ne 0) {
            throw "Dashboard npm install failed"
        }
    }
    
    npm run build
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Dashboard built successfully" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Dashboard build failed (non-critical)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[WARN] Dashboard build skipped: $_" -ForegroundColor Yellow
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "All components built successfully!" -ForegroundColor Green
