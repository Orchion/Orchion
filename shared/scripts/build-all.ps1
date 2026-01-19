# Build all Orchion components
# Usage: .\shared\scripts\build-all.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Building all Orchion components..." -ForegroundColor Cyan
Write-Host ""

# Build Go components
Build-GoComponent -Component 'orchestrator' -OutputName 'orchestrator'
Write-Host ""

Build-GoComponent -Component 'node-agent' -OutputName 'node-agent'
Write-Host ""

# Build Dashboard (with error handling for non-critical failures)
try {
    Build-Dashboard
} catch {
    Write-Warning "Dashboard build failed (non-critical): $_"
    Write-Info "Dashboard can still be run in development mode"
}

Write-Host ""
Write-Host "All components built successfully!" -ForegroundColor Green
