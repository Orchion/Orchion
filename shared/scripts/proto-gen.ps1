# Generate protobuf files for all components
# Usage: .\shared\scripts\proto-gen.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Generating protobuf files..." -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
if (-not (Test-ProtocInstalled)) {
    Write-Error "protoc is required for protobuf generation"
    exit 1
}

# Generate protobuf for Go components
foreach ($component in $script:Components.Go) {
    if ($component -ne 'shared/logging') {  # shared/logging doesn't have protobuf
        Generate-Protobuf -Component $component
        Write-Host ""
    }
}

Write-Host "All protobuf files generated!" -ForegroundColor Green
