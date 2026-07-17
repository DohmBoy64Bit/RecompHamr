$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot

function Fail([string]$Message) {
    throw "baseline check failed: $Message"
}

$RemovedDirectories = @(
    'internal/cloud',
    'internal/mcp',
    'internal/update',
    'internal/classifier',
    'internal/doctor',
    'internal/project'
)

foreach ($Relative in $RemovedDirectories) {
    if (Test-Path (Join-Path $Root $Relative)) {
        Fail "removed subsystem still exists: $Relative"
    }
}

$GoMod = Get-Content -Raw (Join-Path $Root 'go.mod')
$Pinned = @(
    'github.com/charmbracelet/bubbles v0.20.0',
    'github.com/charmbracelet/bubbletea v1.2.4',
    'github.com/charmbracelet/glamour v0.8.0',
    'github.com/charmbracelet/lipgloss v1.0.0'
)
foreach ($Needle in $Pinned) {
    if (-not $GoMod.Contains($Needle)) {
        Fail "frozen TUI dependency drifted: $Needle"
    }
}

# Active Go source is confined to cmd/ and internal/. Do not recurse through
# project runtime state such as the deliberately protected .rehamr directory.
$GoFiles = Get-ChildItem -Path @((Join-Path $Root 'cmd'), (Join-Path $Root 'internal')) -Recurse -File -Filter '*.go' -ErrorAction SilentlyContinue
if ($GoFiles.Count -eq 0) {
    Fail 'active Go source could not be enumerated'
}
$ForbiddenCodePatterns = @(
    'internal/cloud',
    'internal/mcp',
    'internal/update',
    'internal/classifier',
    'internal/doctor',
    'internal/project',
    'codehamr.com',
    'CODEHAMR_',
    '.codehamr'
)
foreach ($Pattern in $ForbiddenCodePatterns) {
    $Hit = $GoFiles | Select-String -SimpleMatch -Pattern $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "forbidden active code reference '$Pattern' at $($Hit.Path):$($Hit.LineNumber)"
    }
}

$ToolSource = (Get-ChildItem -Path (Join-Path $Root 'internal/tools') -File -Filter '*.go' |
    Where-Object { $_.Name -notlike '*_test.go' } |
    ForEach-Object { Get-Content -Raw $_.FullName }) -join "`n"
foreach ($Name in @('powershell', 'read_file', 'write_file', 'edit_file', 'repomixr', 'recomp_reference')) {
    if (-not $ToolSource.Contains($Name)) {
        Fail "accepted tool surface is missing $Name"
    }
}
foreach ($RemovedTool in @('MCPExec')) {
    if ($ToolSource.Contains($RemovedTool)) {
        Fail "removed tool/extension hook is still active: $RemovedTool"
    }
}

$TodoHit = $GoFiles | Select-String -Pattern '\bTODO\b|\bFIXME\b' | Select-Object -First 1
if ($null -ne $TodoHit) {
    Fail "unfinished marker at $($TodoHit.Path):$($TodoHit.LineNumber)"
}

Write-Host 'baseline policy: PASS'
