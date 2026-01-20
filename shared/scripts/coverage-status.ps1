# Display current test coverage status
# Usage: .\shared\scripts\coverage-status.ps1

$ErrorActionPreference = "Stop"
Import-Module "$PSScriptRoot\Orchion.Common.psm1" -Force

Write-Host "Orchion Test Coverage Status" -ForegroundColor Cyan
Write-Host "=" * 40 -ForegroundColor Cyan
Write-Host ""

$components = $script:Components.Go
$results = @()

foreach ($component in $components) {
    try {
        $coveragePercent = 0.0
        $testCount = 0

        Invoke-InDirectory "$script:ProjectRoot\$component" {
            # Run tests and capture output
            $testOutput = go test -v ./... 2>&1
            $testLines = $testOutput | Where-Object { $_ -match '^=== RUN' }
            $script:testCount = $testLines.Count

            # Run coverage test
            go test -coverprofile=coverage.out -covermode=atomic ./... 2>$null | Out-Null

            # Extract coverage percentage
            $coverageLine = go tool cover -func=coverage.out 2>$null | Select-String -Pattern "total:" | Select-Object -Last 1
            if ($coverageLine) {
                $match = [regex]::Match($coverageLine, '(\d+\.\d+)%')
                if ($match.Success) {
                    $script:coveragePercent = [double]$match.Groups[1].Value
                }
            }
        }

        $results += @{
            Component = $component
            Coverage = $coveragePercent
            Tests = $testCount
        }

    } catch {
        $results += @{
            Component = $component
            Coverage = 0.0
            Tests = 0
        }
    }
}

# Display results in a table
Write-Host ("{0,-20} {1,-10} {2,-8}" -f "Component", "Coverage", "Tests") -ForegroundColor Yellow
Write-Host ("{0,-20} {1,-10} {2,-8}" -f "---------", "--------", "-----") -ForegroundColor Gray

foreach ($result in $results) {
    $status = if ($result.Coverage -ge 95) { "✅" } elseif ($result.Coverage -ge 80) { "⚠️ " } else { "❌" }
    $color = if ($result.Coverage -ge 95) { "Green" } elseif ($result.Coverage -ge 80) { "Yellow" } else { "Red" }

    Write-Host ("{0,-20} {1,-10:F1}% {2,-8}" -f $result.Component, $result.Coverage, $result.Tests) -NoNewline
    Write-Host " $status" -ForegroundColor $color
}

Write-Host ""
$avgCoverage = ($results | Measure-Object -Property Coverage -Average).Average
$totalTests = ($results | Measure-Object -Property Tests -Sum).Sum

Write-Host "Summary:" -ForegroundColor Cyan
Write-Host ("  Total Tests: {0}" -f $totalTests) -ForegroundColor White
Write-Host ("  Average Coverage: {0:F1}%" -f $avgCoverage) -ForegroundColor $(if ($avgCoverage -ge 95) { "Green" } elseif ($avgCoverage -ge 80) { "Yellow" } else { "Red" })

$passingComponents = ($results | Where-Object { $_.Coverage -ge 95 }).Count
$totalComponents = $results.Count

if ($passingComponents -eq $totalComponents) {
    Write-Host ""
    Write-Host "All components meet the 95% coverage requirement!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Components meeting 95% threshold: $passingComponents/$totalComponents" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To check detailed coverage: .\shared\scripts\check-coverage.ps1" -ForegroundColor Gray
    Write-Host "To generate reports: .\shared\scripts\test-all.ps1 -Coverage" -ForegroundColor Gray
}