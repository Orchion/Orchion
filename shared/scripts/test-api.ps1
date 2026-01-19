# Test the Orchion API
# Usage: .\shared\scripts\test-api.ps1

$ErrorActionPreference = "Stop"

Write-Host "Testing Orchion API..." -ForegroundColor Cyan
Write-Host ""

$apiUrl = "http://localhost:8080/api/nodes"

try {
    Write-Host "Fetching nodes from $apiUrl..." -ForegroundColor Yellow
    
    $response = Invoke-RestMethod -Uri $apiUrl -Method Get
    
    if ($response) {
        Write-Host "[OK] API is responding!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Registered Nodes:" -ForegroundColor Cyan
        Write-Host ($response | ConvertTo-Json -Depth 10) -ForegroundColor Gray
        
        $nodeCount = if ($response -is [Array]) { $response.Count } else { 1 }
        Write-Host ""
        Write-Host "Total nodes: $nodeCount" -ForegroundColor Green
    } else {
        Write-Host "[WARN] API returned empty response" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to connect to API" -ForegroundColor Red
    Write-Host ""
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Make sure the orchestrator is running:" -ForegroundColor Yellow
    Write-Host "   .\shared\scripts\run-all.ps1" -ForegroundColor Gray
    exit 1
}
