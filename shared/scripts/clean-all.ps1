# Clean all build artifacts
# Usage: .\shared\scripts\clean-all.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
Write-Host ""

# Clean Go components
Clean-GoComponent -Component 'orchestrator' -OutputName 'orchestrator'
Clean-GoComponent -Component 'node-agent' -OutputName 'node-agent'

# Clean Dashboard
try {
    Clean-Dashboard
} catch {
    Write-Warning "Dashboard clean failed: $_"
}

Write-Host ""
Write-Host "[OK] Clean complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Note:" -ForegroundColor Yellow
Write-Host "   - Protobuf generated files are not removed" -ForegroundColor Yellow
Write-Host "   - Dashboard node_modules are not removed (use 'npm run clean' in dashboard/)" -ForegroundColor Yellow
Write-Host "   - Use 'git clean' if you want to remove all generated files" -ForegroundColor Yellow
