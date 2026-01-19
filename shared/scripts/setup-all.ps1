# Setup all Orchion components - install dependencies and prerequisites
# Usage: .\shared\scripts\setup-all.ps1
#
# This script should be run once before building/running/testing the project.
# It installs all dependencies and sets up prerequisites for each component.

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Setting up Orchion development environment..." -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Step "Checking prerequisites..."

$prereqsOK = $true
$prereqsOK = $prereqsOK -and (Test-GoInstalled)
$prereqsOK = $prereqsOK -and (Test-NodeInstalled)
$prereqsOK = $prereqsOK -and (Test-NpmInstalled)

# Check optional tools
Test-ProtocInstalled | Out-Null
Test-GolangciLintInstalled | Out-Null

# Check Docker (optional)
$dockerInstalled = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerInstalled) {
    Write-Info "Docker not found (optional, needed for vLLM/Ollama containers)"
} else {
    Write-Success "Docker found: $(docker --version)"
}

Write-Host ""

if (-not $prereqsOK) {
    Write-Error "Missing required prerequisites. Please install them and run again."
    exit 1
}

# Install Go tools
Install-GoTools
Write-Host ""

# Setup Go components
foreach ($component in $script:Components.Go) {
    Install-GoDependencies -Component $component
    Generate-Protobuf -Component $component
    Write-Host ""
}

# Setup Node components
foreach ($component in $script:Components.Node) {
    Install-NodeDependencies -Component $component

    # Special handling for dashboard Playwright browsers
    if ($component -eq 'dashboard') {
        Write-Step "Installing Playwright browsers..."
        Invoke-InDirectory "$script:ProjectRoot\$component" {
            try {
                npx playwright install --with-deps chromium 2>$null | Out-Null
                Write-Success "Playwright browsers installed"
            } catch {
                Write-Warning "Playwright browser installation failed (non-critical)"
                Write-Info "Run 'npx playwright install' manually if needed"
            }
        }
    }

    Write-Host ""
}

# Summary
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Build: .\shared\scripts\build-all.ps1" -ForegroundColor Gray
Write-Host "  2. Run:   .\shared\scripts\run-all.ps1" -ForegroundColor Gray
Write-Host "  3. Test:  .\shared\scripts\test-all.ps1" -ForegroundColor Gray
Write-Host ""
