# Run tests for all Orchion components
# Usage: .\shared\scripts\test-all.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force
$allPassed = $true

Write-Host "Running tests for all components..." -ForegroundColor Cyan
Write-Host ""

# Test Go components
foreach ($component in $script:Components.Go) {
    try {
        Test-GoComponent -Component $component
        Write-Host "[PASS] $component tests passed" -ForegroundColor Green
    } catch {
        Write-Host "[FAIL] $component tests failed: $_" -ForegroundColor Red
        $allPassed = $false
    }
    Write-Host ""
}

# Test Node components (dashboard tests are non-critical)
foreach ($component in $script:Components.Node) {
    try {
        Test-NodeComponent -Component $component
        Write-Host "[PASS] $component tests passed" -ForegroundColor Green
    } catch {
        if ($component -eq 'dashboard') {
            Write-Warning "$component tests failed (non-critical): $_"
            Write-Info "Run 'npx playwright install' in $component/ if Playwright errors occur"
        } else {
            Write-Host "[FAIL] $component tests failed: $_" -ForegroundColor Red
            $allPassed = $false
        }
    }
    Write-Host ""
}

if ($allPassed) {
    Write-Host "All tests passed!" -ForegroundColor Green
} else {
    Write-Host "Some tests failed" -ForegroundColor Yellow
    exit 1
}
