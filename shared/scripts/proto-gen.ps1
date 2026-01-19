# Generate protobuf files for all components
# Usage: .\shared\scripts\proto-gen.ps1

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Generating protobuf files..." -ForegroundColor Cyan
Write-Host ""

# Check if make is available
$makeAvailable = Get-Command make -ErrorAction SilentlyContinue
if (-not $makeAvailable) {
    Write-Host "[WARN] Make not found. Trying direct protoc commands..." -ForegroundColor Yellow
}

# Generate for Orchestrator
Write-Host "Generating protobuf for Orchestrator..." -ForegroundColor Yellow
Push-Location "$projectRoot\orchestrator"
try {
    if ($makeAvailable) {
        make proto
    } else {
        protoc -I ../shared/proto `
            --go_out=api/v1 --go_opt=paths=source_relative `
            --go-grpc_out=api/v1 --go-grpc_opt=paths=source_relative `
            ../shared/proto/v1/orchestrator.proto
    }
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Orchestrator protobuf generated" -ForegroundColor Green
    } else {
        throw "Orchestrator protobuf generation failed"
    }
} finally {
    Pop-Location
}

Write-Host ""

# Generate for Node Agent
Write-Host "Generating protobuf for Node Agent..." -ForegroundColor Yellow
Push-Location "$projectRoot\node-agent"
try {
    if ($makeAvailable) {
        make proto
    } else {
        protoc -I ../shared/proto `
            --go_out=internal/proto --go_opt=paths=source_relative `
            --go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative `
            ../shared/proto/v1/orchestrator.proto
    }
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Node Agent protobuf generated" -ForegroundColor Green
    } else {
        throw "Node Agent protobuf generation failed"
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "All protobuf files generated!" -ForegroundColor Green
