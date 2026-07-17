$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

& (Join-Path $PSScriptRoot 'evaluation/Invoke-SkillEvaluation.ps1') -ValidateOnly
if ($LASTEXITCODE -ne 0) { throw 'Stage G skill evaluation fixture validation failed' }

Write-Host 'Stage G skill evaluation fixtures: PASS'
