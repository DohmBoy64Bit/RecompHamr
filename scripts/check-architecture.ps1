$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot

function Fail([string]$Message) {
    throw "architecture check failed: $Message"
}

# Stage A permits inherited TUI -> runtime coupling for baseline comparison, but
# backend packages must never depend on the TUI.
$BackendRoots = @('internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/tools')
foreach ($Relative in $BackendRoots) {
    $Dir = Join-Path $Root $Relative
    if (-not (Test-Path $Dir)) { continue }
    $Hit = Get-ChildItem -Path $Dir -Recurse -File -Filter '*.go' |
        Select-String -SimpleMatch 'internal/tui' |
        Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "backend imports presentation at $($Hit.Path):$($Hit.LineNumber)"
    }
}

# Removed feature packages must not be imported under a different file layout.
$AllGo = Get-ChildItem -Path $Root -Recurse -File -Filter '*.go'
foreach ($Pattern in @('/mcp', '/skills', '/update', '/classifier', '/doctor', '/project')) {
    $Hit = $AllGo | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "removed feature dependency remains at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

Write-Host 'architecture (Stage A): PASS'
Write-Host 'note: TUI-to-runtime coupling remains allowed only until the baseline gate closes.'
