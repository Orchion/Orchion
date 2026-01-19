# Run all Orchion components
# Usage: .\shared\scripts\run-all.ps1
# The dashboard runs in this window. Closing it (Ctrl+C) will also close the other components.

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

# Store process IDs for cleanup
$orchestratorProcess = $null
$nodeAgentProcess = $null

# Cleanup function to kill child processes and close terminal windows
function Cleanup-Processes {
    Write-Host ""
    Write-Host "Cleaning up processes..." -ForegroundColor Yellow
    
    if ($orchestratorProcess -and -not $orchestratorProcess.HasExited) {
        Write-Host "Stopping Orchestrator..." -ForegroundColor Gray
        # Use taskkill with /T to kill the entire process tree (window + child exe)
        taskkill /T /F /PID $orchestratorProcess.Id 2>$null | Out-Null
    }
    
    if ($nodeAgentProcess -and -not $nodeAgentProcess.HasExited) {
        Write-Host "Stopping Node Agent..." -ForegroundColor Gray
        # Use taskkill with /T to kill the entire process tree (window + child exe)
        taskkill /T /F /PID $nodeAgentProcess.Id 2>$null | Out-Null
    }
    
    Write-Host "[OK] All processes stopped" -ForegroundColor Green
}

# Register cleanup on script exit
Register-EngineEvent PowerShell.Exiting -Action { Cleanup-Processes } | Out-Null
trap { Cleanup-Processes; break }

Write-Host "Starting all Orchion components..." -ForegroundColor Cyan
Write-Host ""

# Check if executables exist
$orchestratorExe = "$projectRoot\orchestrator\orchestrator.exe"
$nodeAgentExe = "$projectRoot\node-agent\node-agent.exe"

if (-not (Test-Path $orchestratorExe)) {
    Write-Host "[ERROR] Orchestrator not found. Run build-all.ps1 first." -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $nodeAgentExe)) {
    Write-Host "[ERROR] Node Agent not found. Run build-all.ps1 first." -ForegroundColor Red
    exit 1
}

# Start Orchestrator in new window and track the process
Write-Host "Starting Orchestrator..." -ForegroundColor Yellow
$orchestratorProcess = Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$projectRoot\orchestrator'; .\orchestrator.exe" -WindowStyle Normal -PassThru

Start-Sleep -Seconds 2

# Start Node Agent in new window and track the process
Write-Host "Starting Node Agent..." -ForegroundColor Yellow
$nodeAgentProcess = Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$projectRoot\node-agent'; .\node-agent.exe" -WindowStyle Normal -PassThru

Start-Sleep -Seconds 2

Write-Host ""
Write-Host "[OK] Backend components started" -ForegroundColor Green
Write-Host ""
Write-Host "Started:" -ForegroundColor Cyan
Write-Host "  - Orchestrator (gRPC: 50051, HTTP: 8080)" -ForegroundColor Gray
Write-Host "  - Node Agent (connected to 127.0.0.1:50051)" -ForegroundColor Gray
Write-Host ""

# Start Dashboard in current window (blocking)
Write-Host "Starting Dashboard..." -ForegroundColor Yellow
Write-Host "  Dashboard will run in this window." -ForegroundColor Gray
Write-Host "  Press Ctrl+C to stop all components." -ForegroundColor Gray
Write-Host ""

$dashboardDir = "$projectRoot\dashboard"

# Check if node_modules exists
if (-not (Test-Path "$dashboardDir\node_modules")) {
    Write-Host "Installing dashboard dependencies..." -ForegroundColor Yellow
    Push-Location $dashboardDir
    try {
        npm install
        if ($LASTEXITCODE -ne 0) {
            throw "npm install failed"
        }
        Write-Host "[OK] Dependencies installed" -ForegroundColor Green
    } finally {
        Pop-Location
    }
    Write-Host ""
}

# Start dev server (blocking - script will wait here)
Push-Location $dashboardDir
try {
    npm run dev
} finally {
    Pop-Location
    # Cleanup will be called automatically via trap/Register-EngineEvent
    Cleanup-Processes
}