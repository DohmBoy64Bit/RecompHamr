$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
$GoFiles = @(Get-ChildItem -Path @((Join-Path $Root 'cmd'), (Join-Path $Root 'internal')) -Recurse -File -Filter '*.go' -ErrorAction SilentlyContinue | ForEach-Object { $_.FullName })
if ($GoFiles.Count -eq 0) {
    throw 'format check failed: no Go files found'
}

$Unformatted = @(& gofmt -l $GoFiles)
if ($LASTEXITCODE -ne 0) {
    throw "gofmt failed with exit code $LASTEXITCODE"
}
if ($Unformatted.Count -gt 0) {
    throw "gofmt check failed:`n$($Unformatted -join "`n")"
}

Write-Host 'gofmt: PASS'
