# Test job submission to Orchion
# Usage: .\shared\scripts\test-job.ps1

$ErrorActionPreference = "Stop"

Write-Host "Testing Orchion Job Execution..." -ForegroundColor Cyan
Write-Host ""

$apiUrl = "http://localhost:8080/v1/chat/completions"

# Test job payload
$jobPayload = @{
    model = "llama2"
    messages = @(
        @{
            role = "user"
            content = "Hello, how are you?"
        }
    )
    max_tokens = 50
    stream = $false
} | ConvertTo-Json

Write-Host "Submitting job to $apiUrl..." -ForegroundColor Yellow
Write-Host "Payload: $jobPayload" -ForegroundColor Gray
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri $apiUrl -Method Post -Body $jobPayload -ContentType "application/json"

    if ($response) {
        Write-Host "[OK] Job submitted and completed!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Response:" -ForegroundColor Cyan
        Write-Host ($response | ConvertTo-Json -Depth 10) -ForegroundColor Gray
    } else {
        Write-Host "[WARN] Empty response received" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Job submission failed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""

    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "HTTP Status Code: $statusCode" -ForegroundColor Red

        $streamReader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorResponse = $streamReader.ReadToEnd()
        $streamReader.Close()

        if ($errorResponse) {
            Write-Host "Error Response:" -ForegroundColor Red
            Write-Host $errorResponse -ForegroundColor Gray
        }
    }

    Write-Host ""
    Write-Host "Make sure both orchestrator and node-agent are running:" -ForegroundColor Yellow
    Write-Host "   .\shared\scripts\run-all.ps1" -ForegroundColor Gray
    exit 1
}