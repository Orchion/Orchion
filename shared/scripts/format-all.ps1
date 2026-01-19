# Format all Orchion components
# Usage: .\shared\scripts\format-all.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Formatting all Orchion components..." -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
$goimportsInstalled = Get-Command goimports -ErrorAction SilentlyContinue
if (-not $goimportsInstalled) {
    Write-Error "goimports is not installed or not in PATH"
    Write-Info "Install with: go install golang.org/x/tools/cmd/goimports@latest"
    Write-Info "Or run: .\shared\scripts\setup-all.ps1"
    throw "goimports is required for Go formatting"
}

Write-Host ""

# Format Go components
foreach ($component in $script:Components.Go) {
    try {
        Format-GoComponent -Component $component
    } catch {
        Write-Error "$component formatting failed: $_"
        throw
    }
    Write-Host ""
}

# Format Node components (non-critical failures)
foreach ($component in $script:Components.Node) {
    try {
        Format-NodeComponent -Component $component
    } catch {
        Write-Warning "$component formatting failed (non-critical): $_"
    }
    Write-Host ""
}

Write-Host "All components formatted successfully!" -ForegroundColor Green