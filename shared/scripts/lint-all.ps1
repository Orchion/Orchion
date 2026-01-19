# Lint all Orchion components
# Usage: .\shared\scripts\lint-all.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Linting all Orchion components..." -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
if (-not (Test-GolangciLintInstalled)) {
    Write-Info "Or run: .\shared\scripts\setup-all.ps1"
    throw "golangci-lint is required for Go linting"
}

Write-Host ""

# Lint Go components (warnings are non-critical for Go linting)
foreach ($component in $script:Components.Go) {
    try {
        Lint-GoComponent -Component $component
        Write-Host "[PASS] $component linting passed" -ForegroundColor Green
    } catch {
        Write-Warning "$component linting failed (non-critical): $_"
        Write-Info "Run 'golangci-lint run ./...' in $component/ for details"
    }
    Write-Host ""
}

# Lint Node components (warnings are non-critical)
foreach ($component in $script:Components.Node) {
    try {
        Lint-NodeComponent -Component $component
        Write-Host "[PASS] $component linting passed" -ForegroundColor Green
    } catch {
        Write-Warning "$component linting failed (non-critical): $_"
        Write-Info "Run 'npm run lint' in $component/ for details"
    }
    Write-Host ""
}

Write-Host "Linting complete! Check individual component notes above for any issues." -ForegroundColor Green