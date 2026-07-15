$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

# First enforce the durable documentation contract: required files must exist,
# be non-empty, and retain the required project facts/terms.
go run ./cmd/docscheck
if ($LASTEXITCODE -ne 0) {
    throw "documentation contract check failed with exit code $LASTEXITCODE"
}
Write-Host 'documentation contract: PASS'

# Then independently verify that local Markdown links resolve.
$MarkdownFiles = Get-ChildItem -Path $Root -Recurse -File -Filter '*.md' |
    Where-Object { $_.FullName -notmatch '[\\/]\.git[\\/]' }

$Broken = New-Object System.Collections.Generic.List[string]
$LinkPattern = [regex]'\[[^\]]+\]\(([^)]+)\)'

foreach ($File in $MarkdownFiles) {
    $Text = Get-Content -Raw $File.FullName
    foreach ($Match in $LinkPattern.Matches($Text)) {
        $Target = $Match.Groups[1].Value.Trim()
        if ($Target -eq '' -or $Target.StartsWith('#')) { continue }
        if ($Target -match '^[a-zA-Z][a-zA-Z0-9+.-]*://') { continue }
        if ($Target.StartsWith('mailto:')) { continue }

        $PathOnly = ($Target -split '#', 2)[0]
        if ($PathOnly -eq '') { continue }
        $Decoded = [System.Uri]::UnescapeDataString($PathOnly)
        $Resolved = Join-Path $File.DirectoryName $Decoded
        if (-not (Test-Path $Resolved)) {
            $Broken.Add("$($File.FullName): $Target")
        }
    }
}

if ($Broken.Count -gt 0) {
    throw "documentation check failed: broken relative links:`n$($Broken -join "`n")"
}

Write-Host "documentation links: PASS ($($MarkdownFiles.Count) Markdown files)"
