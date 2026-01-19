# Start the dashboard development server
# Usage: .\shared\scripts\dev-dashboard.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Starting Orchion Dashboard..." -ForegroundColor Cyan
Write-Host ""

# Ensure dependencies are installed
Install-NodeDependencies -Component 'dashboard' -SkipIfExists
Write-Host ""

# Start dev server
Write-Step "Starting development server..."
Invoke-InDirectory "$script:ProjectRoot\dashboard" {
    npm run dev
}
