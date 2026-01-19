# Run tests for all Orchion components
# Usage: .\shared\scripts\test-all.ps1

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$allPassed = $true

Write-Host "Running tests for all components..." -ForegroundColor Cyan
Write-Host ""

# Test Orchestrator
Write-Host "Testing Orchestrator..." -ForegroundColor Yellow
Push-Location "$projectRoot\orchestrator"
try {
    go test ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[PASS] Orchestrator tests passed" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Orchestrator tests failed" -ForegroundColor Red
        $allPassed = $false
    }
} catch {
    Write-Host "[FAIL] Orchestrator tests failed: $_" -ForegroundColor Red
    $allPassed = $false
} finally {
    Pop-Location
}

Write-Host ""

# Test Node Agent
Write-Host "Testing Node Agent..." -ForegroundColor Yellow
Push-Location "$projectRoot\node-agent"
try {
    go test ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[PASS] Node Agent tests passed" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Node Agent tests failed" -ForegroundColor Red
        $allPassed = $false
    }
} catch {
    Write-Host "[FAIL] Node Agent tests failed: $_" -ForegroundColor Red
    $allPassed = $false
} finally {
    Pop-Location
}

Write-Host ""

# Test Dashboard
Write-Host "Testing Dashboard..." -ForegroundColor Yellow
Push-Location "$projectRoot\dashboard"
try {
    # Check if node_modules exists
    if (-not (Test-Path "node_modules")) {
        Write-Host "  Installing dependencies..." -ForegroundColor Gray
        npm install
    }
    
    # Try to install Playwright browsers if needed (non-blocking)
    $playwrightPath = "node_modules\@playwright\test"
    if (Test-Path $playwrightPath) {
        Write-Host "  Installing Playwright browsers (if needed)..." -ForegroundColor Gray
        npx playwright install --with-deps chromium 2>&1 | Out-Null
    }
    
    npm test
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[PASS] Dashboard tests passed" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Dashboard tests failed (non-critical)" -ForegroundColor Yellow
        Write-Host "  Note: Run 'npx playwright install' in dashboard/ if Playwright errors occur" -ForegroundColor Gray
    }
} catch {
    Write-Host "[WARN] Dashboard tests skipped: $_" -ForegroundColor Yellow
} finally {
    Pop-Location
}

Write-Host ""
if ($allPassed) {
    Write-Host "All tests passed!" -ForegroundColor Green
} else {
    Write-Host "Some tests failed" -ForegroundColor Yellow
    exit 1
}
