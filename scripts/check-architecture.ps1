$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot

function Fail([string]$Message) {
    throw "architecture check failed: $Message"
}

# Stage C still permits inherited TUI -> runtime coupling while it is extracted,
# but backend packages must never depend on the TUI. internal/app is the sole
# composition root allowed to construct the concrete presentation.
$BackendRoots = @('internal/agent', 'internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/tools')
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

# Slice 2 owns model/tool orchestration in internal/agent. Presentation may
# schedule opaque work and render typed effects, but it must not recover direct
# tool execution, provider-error policy, or construct parallel agent roots.
$TuiProduction = Get-ChildItem -Path (Join-Path $Root 'internal/tui') -File -Filter '*.go' |
    Where-Object { $_.Name -notlike '*_test.go' }
foreach ($Pattern in @('internal/tools', 'tools.Execute', 'tools.InlineStatus', 'provider.ErrUnauthorized', 'provider.ErrUnreachable', 'agent.LocalToolExecutor', 'agent.NewTurnState', 'agent.NewStreamState', 'turn.Context', 'turn.CancelFunc')) {
    $Hit = $TuiProduction | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "presentation owns agent orchestration at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

$Entrypoint = Join-Path $Root 'cmd/recomphamr/main.go'
foreach ($Pattern in @('internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/tools', 'internal/tui')) {
    $Hit = Select-String -Path $Entrypoint -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "process entrypoint bypasses internal/app at $($Hit.Path):$($Hit.LineNumber): $Pattern"
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

Write-Host 'architecture (Stage C transition): PASS'
Write-Host 'note: internal/app composes the agent runtime; configuration/client persistence remains a later extraction.'
