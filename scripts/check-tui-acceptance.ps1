$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot
$Harness = Join-Path $PSScriptRoot 'acceptance/Invoke-TuiAcceptance.ps1'
$Scenarios = Join-Path $PSScriptRoot 'acceptance/scenarios'

$tokens = $null
$errors = $null
[void][System.Management.Automation.Language.Parser]::ParseFile($Harness, [ref]$tokens, [ref]$errors)
if ($errors.Count -ne 0) {
    throw "TUI acceptance harness has $($errors.Count) PowerShell parse error(s)"
}

$scenarioFiles = @(Get-ChildItem -LiteralPath $Scenarios -Filter '*.json' -File)
if ($scenarioFiles.Count -eq 0) { throw 'no TUI acceptance scenarios found' }
foreach ($scenario in $scenarioFiles) {
    pwsh -NoProfile -File $Harness -ScenarioPath $scenario.FullName -ValidateOnly
    if ($LASTEXITCODE -ne 0) {
        throw "TUI acceptance scenario validation failed: $($scenario.Name)"
    }
}

$invalidScenario = Join-Path $PSScriptRoot 'acceptance/testdata/invalid-version.json'
pwsh -NoProfile -File $Harness -ScenarioPath $invalidScenario -ValidateOnly *> $null
if ($LASTEXITCODE -eq 0) { throw 'invalid TUI acceptance scenario unexpectedly passed validation' }

Write-Host "TUI acceptance harness: PASS ($($scenarioFiles.Count) scenarios)"
