$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

& (Join-Path $PSScriptRoot 'check-baseline.ps1')
& (Join-Path $PSScriptRoot 'check-docs.ps1')
& (Join-Path $PSScriptRoot 'check-architecture.ps1')
& (Join-Path $PSScriptRoot 'check-format.ps1')
& (Join-Path $PSScriptRoot 'check-tui-acceptance.ps1')
& (Join-Path $PSScriptRoot 'check-coverage.ps1')

New-Item -ItemType Directory -Force -Path (Join-Path $Root 'dist') | Out-Null
$Exe = if ($env:OS -eq 'Windows_NT') { 'recomphamr.exe' } else { 'recomphamr' }
$Artifact = Join-Path $Root "dist/$Exe"
go build -o $Artifact ./cmd/recomphamr
if ($LASTEXITCODE -ne 0) {
    throw "go build failed with exit code $LASTEXITCODE"
}
Write-Host "go build: PASS ($Artifact)"

$Help = & $Artifact --help 2>&1 | Out-String
if ($LASTEXITCODE -ne 0) {
    throw "CLI smoke test failed with exit code $LASTEXITCODE"
}
if (-not $Help.Contains('recomphamr - barebones local-first coding-agent baseline')) {
    throw 'CLI smoke test failed: expected help text was not produced'
}
Write-Host 'CLI help smoke test: PASS'

Write-Host ''
Write-Host 'Automated verification complete.'
Write-Host 'Manual runtime and TUI screenshot acceptance are still required by docs/dev/workflows/baseline-gate.md.'
