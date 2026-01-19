# Clean all build artifacts
# Usage: .\shared\scripts\clean-all.ps1

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
Write-Host ""

# Clean Orchestrator
Write-Host "Cleaning Orchestrator..." -ForegroundColor Yellow
$orchestratorExe = "$projectRoot\orchestrator\orchestrator.exe"
if (Test-Path $orchestratorExe) {
    Remove-Item $orchestratorExe -Force
    Write-Host "  [OK] Removed orchestrator.exe" -ForegroundColor Gray
}

# Clean Node Agent
Write-Host "Cleaning Node Agent..." -ForegroundColor Yellow
$nodeAgentExe = "$projectRoot\node-agent\node-agent.exe"
if (Test-Path $nodeAgentExe) {
    Remove-Item $nodeAgentExe -Force
    Write-Host "  [OK] Removed node-agent.exe" -ForegroundColor Gray
}

# Clean Dashboard
Write-Host "Cleaning Dashboard..." -ForegroundColor Yellow
Push-Location "$projectRoot\dashboard"
try {
    if (Test-Path "build") {
        Remove-Item -Recurse -Force "build"
        Write-Host "  [OK] Removed build directory" -ForegroundColor Gray
    }
    if (Test-Path ".svelte-kit") {
        Remove-Item -Recurse -Force ".svelte-kit"
        Write-Host "  [OK] Removed .svelte-kit directory" -ForegroundColor Gray
    }
} catch {
    Write-Host "  [WARN] Dashboard clean skipped: $_" -ForegroundColor Yellow
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "[OK] Clean complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Note:" -ForegroundColor Yellow
Write-Host "   - Protobuf generated files are not removed" -ForegroundColor Yellow
Write-Host "   - Dashboard node_modules are not removed (use 'npm run clean' in dashboard/)" -ForegroundColor Yellow
Write-Host "   - Use 'git clean' if you want to remove all generated files" -ForegroundColor Yellow
