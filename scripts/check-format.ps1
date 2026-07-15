$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
$GoFiles = @(Get-ChildItem -Path $Root -Recurse -File -Filter '*.go' | ForEach-Object { $_.FullName })
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
