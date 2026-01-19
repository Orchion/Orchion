# Setup all Orchion components - install dependencies and prerequisites
# Usage: .\shared\scripts\setup-all.ps1
#
# This script should be run once before building/running/testing the project.
# It installs all dependencies and sets up prerequisites for each component.

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host "Setting up Orchion development environment..." -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Host "Checking prerequisites..." -ForegroundColor Yellow

$prereqsOK = $true

# Check Go
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if (-not $goInstalled) {
    Write-Host "[ERROR] Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "  Install from: https://go.dev/dl/" -ForegroundColor Gray
    $prereqsOK = $false
} else {
    $goVersion = go version
    Write-Host "[OK] Go found: $goVersion" -ForegroundColor Green
}

# Check Node.js
$nodeInstalled = Get-Command node -ErrorAction SilentlyContinue
if (-not $nodeInstalled) {
    Write-Host "[ERROR] Node.js is not installed or not in PATH" -ForegroundColor Red
    Write-Host "  Install from: https://nodejs.org/" -ForegroundColor Gray
    $prereqsOK = $false
} else {
    $nodeVersion = node --version
    Write-Host "[OK] Node.js found: $nodeVersion" -ForegroundColor Green
}

# Check npm
$npmInstalled = Get-Command npm -ErrorAction SilentlyContinue
if (-not $npmInstalled) {
    Write-Host "[ERROR] npm is not installed or not in PATH" -ForegroundColor Red
    $prereqsOK = $false
} else {
    $npmVersion = npm --version
    Write-Host "[OK] npm found: v$npmVersion" -ForegroundColor Green
}

# Check protoc (optional but recommended)
$protocInstalled = Get-Command protoc -ErrorAction SilentlyContinue
if (-not $protocInstalled) {
    Write-Host "[WARN] protoc not found (optional, but needed for protobuf generation)" -ForegroundColor Yellow
    Write-Host "  Install from: https://github.com/protocolbuffers/protobuf/releases" -ForegroundColor Gray
} else {
    $protocVersion = protoc --version
    Write-Host "[OK] protoc found: $protocVersion" -ForegroundColor Green
}

# Check Go protobuf plugins (optional but recommended)
$protocGenGo = Get-Command protoc-gen-go -ErrorAction SilentlyContinue
$protocGenGoGrpc = Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue
if (-not $protocGenGo -or -not $protocGenGoGrpc) {
    Write-Host "[WARN] Go protobuf plugins not found (will install)" -ForegroundColor Yellow
} else {
    Write-Host "[OK] Go protobuf plugins found" -ForegroundColor Green
}

# Check Docker (optional, for container features)
$dockerInstalled = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerInstalled) {
    Write-Host "[INFO] Docker not found (optional, needed for vLLM/Ollama containers)" -ForegroundColor Gray
} else {
    $dockerVersion = docker --version
    Write-Host "[OK] Docker found: $dockerVersion" -ForegroundColor Green
}

Write-Host ""

if (-not $prereqsOK) {
    Write-Host "[ERROR] Missing required prerequisites. Please install them and run again." -ForegroundColor Red
    exit 1
}

# Install Go protobuf plugins if missing
if (-not $protocGenGo -or -not $protocGenGoGrpc) {
    Write-Host "Installing Go protobuf plugins..." -ForegroundColor Yellow
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    
    $goBinPath = "$env:USERPROFILE\go\bin"
    if ($env:Path -notlike "*$goBinPath*") {
        Write-Host "[WARN] Go bin directory not in PATH: $goBinPath" -ForegroundColor Yellow
        Write-Host "  Add to PATH or restart terminal after installation" -ForegroundColor Gray
    }
    Write-Host "[OK] Go protobuf plugins installed" -ForegroundColor Green
    Write-Host ""
}

# Setup Orchestrator
Write-Host "Setting up Orchestrator..." -ForegroundColor Yellow
Push-Location "$projectRoot\orchestrator"
try {
    Write-Host "  Installing Go dependencies..." -ForegroundColor Gray
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Orchestrator dependencies installed" -ForegroundColor Green
    } else {
        throw "go mod tidy failed"
    }
    
    # Generate protobuf if protoc is available
    if ($protocInstalled) {
        Write-Host "  Generating protobuf files..." -ForegroundColor Gray
        $makeAvailable = Get-Command make -ErrorAction SilentlyContinue
        if ($makeAvailable) {
            make proto 2>&1 | Out-Null
        } else {
            protoc -I ../shared/proto `
                --go_out=api/v1 --go_opt=paths=source_relative `
                --go-grpc_out=api/v1 --go-grpc_opt=paths=source_relative `
                ../shared/proto/v1/orchestrator.proto 2>&1 | Out-Null
        }
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Orchestrator protobuf generated" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "[ERROR] Orchestrator setup failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

Write-Host ""

# Setup Node Agent
Write-Host "Setting up Node Agent..." -ForegroundColor Yellow
Push-Location "$projectRoot\node-agent"
try {
    Write-Host "  Installing Go dependencies..." -ForegroundColor Gray
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Node Agent dependencies installed" -ForegroundColor Green
    } else {
        throw "go mod tidy failed"
    }
    
    # Generate protobuf if protoc is available
    if ($protocInstalled) {
        Write-Host "  Generating protobuf files..." -ForegroundColor Gray
        $makeAvailable = Get-Command make -ErrorAction SilentlyContinue
        if ($makeAvailable) {
            make proto 2>&1 | Out-Null
        } else {
            protoc -I ../shared/proto `
                --go_out=internal/proto --go_opt=paths=source_relative `
                --go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative `
                ../shared/proto/v1/orchestrator.proto 2>&1 | Out-Null
        }
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Node Agent protobuf generated" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "[ERROR] Node Agent setup failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

Write-Host ""

# Setup Dashboard
Write-Host "Setting up Dashboard..." -ForegroundColor Yellow
Push-Location "$projectRoot\dashboard"
try {
    Write-Host "  Installing npm dependencies..." -ForegroundColor Gray
    npm install
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Dashboard dependencies installed" -ForegroundColor Green
    } else {
        throw "npm install failed"
    }
    
    Write-Host "  Installing Playwright browsers..." -ForegroundColor Gray
    npx playwright install --with-deps chromium 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Playwright browsers installed" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Playwright browser installation failed (non-critical)" -ForegroundColor Yellow
        Write-Host "  Run 'npx playwright install' manually if needed" -ForegroundColor Gray
    }
} catch {
    Write-Host "[ERROR] Dashboard setup failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

Write-Host ""

# Summary
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Build: .\shared\scripts\build-all.ps1" -ForegroundColor Gray
Write-Host "  2. Run:   .\shared\scripts\run-all.ps1" -ForegroundColor Gray
Write-Host "  3. Test:  .\shared\scripts\test-all.ps1" -ForegroundColor Gray
Write-Host ""
