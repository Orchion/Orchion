# Check test coverage across all Go components
# Usage: .\shared\scripts\check-coverage.ps1 [-Threshold <percentage>]

param(
    [double]$Threshold = 95.0,
    [switch]$GenerateReports,
    [switch]$FailOnThreshold
)

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Checking test coverage across all Go components..." -ForegroundColor Cyan
Write-Host "Coverage threshold: $Threshold%" -ForegroundColor Yellow
Write-Host ""

$components = $script:Components.Go
$allPassed = $true
$results = @()

foreach ($component in $components) {
    try {
        Write-Step "Checking coverage for $component..."

        $coveragePercent = 0.0
        $coverageValid = $false

        Invoke-InDirectory "$script:ProjectRoot\$component" {
            # Run tests with coverage
            go test -race -coverprofile=coverage.out -covermode=atomic ./... 2>$null

            # Generate HTML report if requested
            if ($GenerateReports) {
                go tool cover -html=coverage.out -o coverage.html
            }

            # Extract coverage percentage
            $coverageLine = go tool cover -func=coverage.out 2>$null | Select-String -Pattern "total:" | Select-Object -Last 1
            if ($coverageLine) {
                $match = [regex]::Match($coverageLine, '(\d+\.\d+)%')
                if ($match.Success) {
                    $script:coveragePercent = [double]$match.Groups[1].Value
                    $script:coverageValid = $true
                }
            }
        }

        if ($coverageValid) {
            $status = if ($coveragePercent -ge $Threshold) { "✅ PASS" } else { "❌ FAIL" }
            Write-Host ("{0,-15} {1,6:F1}% {2}" -f "$component`:", $coveragePercent, $status) -ForegroundColor $(if ($coveragePercent -ge $Threshold) { "Green" } else { "Red" })

            $results += @{
                Component = $component
                Coverage = $coveragePercent
                Passed = ($coveragePercent -ge $Threshold)
            }

            if ($coveragePercent -lt $Threshold) {
                $allPassed = $false
            }
        } else {
            Write-Host ("{0,-15} {1,6} ❌ FAIL" -f "$component`:", "ERROR") -ForegroundColor Red
            $results += @{
                Component = $component
                Coverage = 0.0
                Passed = $false
            }
            $allPassed = $false
        }

        if ($GenerateReports) {
            Write-Info "Coverage report: $component/coverage.html"
        }

    } catch {
        Write-Host ("{0,-15} {1,6} ❌ FAIL" -f "$component`:", "ERROR") -ForegroundColor Red
        Write-Error "Error testing $component`: $_"
        $results += @{
            Component = $component
            Coverage = 0.0
            Passed = $false
        }
        $allPassed = $false
    }
}

Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan

$passedCount = ($results | Where-Object { $_.Passed }).Count
$totalCount = $results.Count
$avgCoverage = ($results | Measure-Object -Property Coverage -Average).Average

Write-Host "Components: $passedCount/$totalCount passed" -ForegroundColor $(if ($passedCount -eq $totalCount) { "Green" } else { "Red" })
Write-Host ("Average coverage: {0:F1}%" -f $avgCoverage) -ForegroundColor $(if ($avgCoverage -ge $Threshold) { "Green" } else { "Red" })

if ($FailOnThreshold -and -not $allPassed) {
    Write-Host ""
    Write-Host "❌ Coverage requirements not met!" -ForegroundColor Red
    exit 1
} elseif ($allPassed) {
    Write-Host ""
    Write-Host "✅ All coverage requirements met!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "⚠️  Some components below coverage threshold" -ForegroundColor Yellow
}