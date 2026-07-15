$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$CoverageProfile = Join-Path ([System.IO.Path]::GetTempPath()) ("recomphamr-coverage-{0}.out" -f [guid]::NewGuid().ToString('N'))

try {
    go test ./... -covermode=atomic "-coverprofile=$CoverageProfile"
    if ($LASTEXITCODE -ne 0) {
        throw "go test with coverage failed with exit code $LASTEXITCODE"
    }
    Write-Host 'go test with coverage: PASS'

    go run ./cmd/coveragecheck $CoverageProfile
    if ($LASTEXITCODE -ne 0) {
        throw "100% statement coverage gate failed with exit code $LASTEXITCODE"
    }
    Write-Host 'statement coverage: PASS (100.0%)'
}
finally {
    if (Test-Path $CoverageProfile) {
        Remove-Item -Force $CoverageProfile
    }
}
