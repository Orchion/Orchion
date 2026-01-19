# Orchion.Common.psm1
# Shared PowerShell module for Orchion project scripts

# Configuration
$script:ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$script:Components = @{
    Go = @('orchestrator', 'node-agent', 'shared/logging')
    Node = @('dashboard', 'vscode-extension/orchion-tools')
}

# Common utility functions
function Get-ProjectRoot {
    return $script:ProjectRoot
}

function Write-Step {
    param([string]$Message, [string]$Color = "Yellow")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Gray
}

function Invoke-InDirectory {
    param(
        [string]$Path,
        [scriptblock]$ScriptBlock,
        [string]$ErrorMessage = "Command failed"
    )

    Push-Location $Path
    try {
        & $ScriptBlock
        if ($LASTEXITCODE -ne 0) {
            throw $ErrorMessage
        }
    } finally {
        Pop-Location
    }
}

# Tool checking functions
function Test-GoInstalled {
    $goCmd = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goCmd) {
        Write-Error "Go is not installed or not in PATH"
        Write-Info "Install from: https://go.dev/dl/"
        return $false
    }
    Write-Success "Go found: $(go version)"
    return $true
}

function Test-NodeInstalled {
    $nodeCmd = Get-Command node -ErrorAction SilentlyContinue
    if (-not $nodeCmd) {
        Write-Error "Node.js is not installed or not in PATH"
        Write-Info "Install from: https://nodejs.org/"
        return $false
    }
    Write-Success "Node.js found: $(node --version)"
    return $true
}

function Test-NpmInstalled {
    $npmCmd = Get-Command npm -ErrorAction SilentlyContinue
    if (-not $npmCmd) {
        Write-Error "npm is not installed or not in PATH"
        return $false
    }
    Write-Success "npm found: v$(npm --version)"
    return $true
}

function Test-ProtocInstalled {
    $protocCmd = Get-Command protoc -ErrorAction SilentlyContinue
    if (-not $protocCmd) {
        Write-Warning "protoc not found (optional, needed for protobuf generation)"
        Write-Info "Install from: https://github.com/protocolbuffers/protobuf/releases"
        return $false
    }
    Write-Success "protoc found: $(protoc --version)"
    return $true
}

function Test-GolangciLintInstalled {
    $golangciCmd = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if (-not $golangciCmd) {
        Write-Error "golangci-lint is not installed or not in PATH"
        Write-Info "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        return $false
    }
    Write-Success "golangci-lint found"
    return $true
}

# Component management functions
function Install-GoDependencies {
    param([string]$Component)

    Write-Step "Installing Go dependencies for $Component..."
    Invoke-InDirectory "$script:ProjectRoot\$Component" {
        go mod tidy
    }
    Write-Success "$Component dependencies installed"
}

function Install-NodeDependencies {
    param([string]$Component, [switch]$SkipIfExists)

    $componentPath = "$script:ProjectRoot\$Component"

    if ($SkipIfExists -and (Test-Path "$componentPath\node_modules")) {
        Write-Info "$Component dependencies already installed"
        return
    }

    Write-Step "Installing npm dependencies for $Component..."
    Invoke-InDirectory $componentPath {
        npm install
    }
    Write-Success "$Component dependencies installed"
}

function Build-GoComponent {
    param([string]$Component, [string]$OutputName)

    Write-Step "Building $Component..."
    Invoke-InDirectory "$script:ProjectRoot\$Component" {
        go build -o "$OutputName.exe" "./cmd/$Component"
    }
    Write-Success "$Component built successfully"
}

function Test-GoComponent {
    param([string]$Component)

    Write-Step "Testing $Component..."
    Invoke-InDirectory "$script:ProjectRoot\$Component" {
        go test ./...
    }
    Write-Success "$Component tests passed"
}

function Lint-GoComponent {
    param([string]$Component)

    Write-Step "Linting $Component..."
    Invoke-InDirectory "$script:ProjectRoot\$Component" {
        golangci-lint run ./...
    }
    Write-Success "$Component linting passed"
}

function Format-GoComponent {
    param([string]$Component)

    Write-Step "Formatting $Component..."
    Invoke-InDirectory "$script:ProjectRoot\$Component" {
        gofmt -w .
        goimports -w .
    }
    Write-Success "$Component formatted"
}

function Test-NodeComponent {
    param([string]$Component)

    Write-Step "Testing $Component..."
    $componentPath = "$script:ProjectRoot\$Component"

    # Ensure dependencies are installed
    Install-NodeDependencies -Component $Component -SkipIfExists

    # Install Playwright browsers if needed for dashboard
    if ($Component -eq 'dashboard') {
        Invoke-InDirectory $componentPath {
            $playwrightPath = "node_modules\@playwright\test"
            if (Test-Path $playwrightPath) {
                Write-Info "Installing Playwright browsers..."
                npx playwright install --with-deps chromium 2>$null | Out-Null
            }
        }
    }

    Invoke-InDirectory $componentPath {
        npm test
    }
    Write-Success "$Component tests passed"
}

function Lint-NodeComponent {
    param([string]$Component)

    Write-Step "Linting $Component..."
    $componentPath = "$script:ProjectRoot\$Component"

    # Ensure dependencies are installed
    Install-NodeDependencies -Component $Component -SkipIfExists

    Invoke-InDirectory $componentPath {
        npm run lint
    }
    Write-Success "$Component linting passed"
}

function Format-NodeComponent {
    param([string]$Component)

    Write-Step "Formatting $Component..."
    $componentPath = "$script:ProjectRoot\$Component"

    # Ensure dependencies are installed
    Install-NodeDependencies -Component $Component -SkipIfExists

    Invoke-InDirectory $componentPath {
        npm run format
    }
    Write-Success "$Component formatted"
}

function Build-Dashboard {
    Write-Step "Building Dashboard..."
    $dashboardPath = "$script:ProjectRoot\dashboard"

    # Ensure dependencies are installed
    Install-NodeDependencies -Component 'dashboard' -SkipIfExists

    Invoke-InDirectory $dashboardPath {
        npm run build
    }
    Write-Success "Dashboard built successfully"
}

function Clean-GoComponent {
    param([string]$Component, [string]$OutputName)

    $exePath = "$script:ProjectRoot\$Component\$OutputName.exe"
    if (Test-Path $exePath) {
        Remove-Item $exePath -Force
        Write-Success "$Component executable removed"
    }
}

function Clean-Dashboard {
    Write-Step "Cleaning Dashboard..."
    Invoke-InDirectory "$script:ProjectRoot\dashboard" {
        if (Test-Path "build") { Remove-Item "build" -Recurse -Force }
        if (Test-Path ".svelte-kit") { Remove-Item ".svelte-kit" -Recurse -Force }
    }
    Write-Success "Dashboard build artifacts removed"
}

# Protobuf generation
function Generate-Protobuf {
    param([string]$Component)

    $protocInstalled = Test-ProtocInstalled
    if (-not $protocInstalled) { return }

    Write-Step "Generating protobuf files for $Component..."

    $componentPath = "$script:ProjectRoot\$Component"
    $protoPath = "$script:ProjectRoot\shared\proto"

    Invoke-InDirectory $componentPath {
        $makeCmd = Get-Command make -ErrorAction SilentlyContinue
        if ($makeCmd) {
            make proto 2>$null | Out-Null
        } else {
            # Fallback to direct protoc commands
            $outDir = if ($Component -eq 'orchestrator') { 'api/v1' } else { 'internal/proto' }
            protoc -I "$protoPath" `
                --go_out="$outDir" --go_opt=paths=source_relative `
                --go-grpc_out="$outDir" --go-grpc_opt=paths=source_relative `
                "$protoPath/v1/orchestrator.proto" 2>$null | Out-Null
        }
    }
    Write-Success "$Component protobuf generated"
}

# Install Go tools
function Install-GoTools {
    Write-Step "Installing Go protobuf plugins..."

    $protocGenGo = Get-Command protoc-gen-go -ErrorAction SilentlyContinue
    $protocGenGoGrpc = Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue

    if (-not $protocGenGo -or -not $protocGenGoGrpc) {
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
        Write-Success "Go protobuf plugins installed"
    } else {
        Write-Info "Go protobuf plugins already installed"
    }

    Write-Step "Installing Go linting/formatting tools..."

    $golangciLint = Get-Command golangci-lint -ErrorAction SilentlyContinue
    $goimports = Get-Command goimports -ErrorAction SilentlyContinue

    if (-not $golangciLint) {
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    }
    if (-not $goimports) {
        go install golang.org/x/tools/cmd/goimports@latest
    }

    Write-Success "Go linting/formatting tools installed"
}

# Export functions
Export-ModuleMember -Function * -Variable ProjectRoot, Components