$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Root = Split-Path -Parent $PSScriptRoot

function Fail([string]$Message) {
    throw "architecture check failed: $Message"
}

# Backend packages must never depend on the concrete presentation.
$BackendRoots = @('internal/agent', 'internal/config', 'internal/ctx', 'internal/llm', 'internal/logging', 'internal/provider', 'internal/session', 'internal/skills', 'internal/tools', 'internal/workspace')
foreach ($Relative in $BackendRoots) {
    $Dir = Join-Path $Root $Relative
    if (-not (Test-Path $Dir)) { continue }
    $Hit = Get-ChildItem -Path $Dir -Recurse -File -Filter '*.go' -ErrorAction SilentlyContinue |
        Select-String -SimpleMatch 'internal/tui' |
        Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "backend imports presentation at $($Hit.Path):$($Hit.LineNumber)"
    }
}

# The core application package is independently buildable without Bubble Tea
# or the concrete TUI. Only internal/app/terminal may own that wiring edge.
$CoreAppProduction = Get-ChildItem -Path (Join-Path $Root 'internal/app') -File -Filter '*.go' |
    Where-Object { $_.Name -notlike '*_test.go' }
foreach ($Pattern in @('internal/tui', 'charmbracelet/bubbletea')) {
    $Hit = $CoreAppProduction | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "core app imports terminal presentation at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

# Presentation may import only the neutral frontend contract among project
# runtime packages and may only schedule opaque Work values.
$TuiProduction = Get-ChildItem -Path (Join-Path $Root 'internal/tui') -File -Filter '*.go' |
    Where-Object { $_.Name -notlike '*_test.go' }
foreach ($Pattern in @('internal/agent', 'internal/session', 'internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/skills', 'internal/tools', 'internal/logging', 'internal/workspace', '.ApplyDelivery(', '.ApplyToolResult(', '.StartRound(', '.NextTool(', '.DecideClose(', '.CancelTurn(', '.ResetConversation(', '.Reachability(', 'ProbeWork', 'ToolDelivery', 'StreamDelivery', 'TurnState', 'StreamState', 'LoopState', 'cancelFunc', 'processHandle')) {
    $Hit = $TuiProduction | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "presentation imports backend lifecycle at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

# The neutral contract must import neither backend packages nor Bubble Tea.
$FrontendProduction = Get-ChildItem -Path (Join-Path $Root 'internal/frontend') -File -Filter '*.go' |
    Where-Object { $_.Name -notlike '*_test.go' }
foreach ($Pattern in @('internal/agent', 'internal/app', 'internal/session', 'internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/tools', 'internal/logging', 'internal/workspace', 'charmbracelet/bubbletea')) {
    $Hit = $FrontendProduction | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "frontend contract imports concrete runtime at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

$Entrypoint = Join-Path $Root 'cmd/recomphamr/main.go'
foreach ($Pattern in @('internal/config', 'internal/ctx', 'internal/llm', 'internal/provider', 'internal/tools', 'internal/tui', 'internal/agent', 'internal/session', 'internal/workspace')) {
    $Hit = Select-String -Path $Entrypoint -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "process entrypoint bypasses internal/app at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}
if (-not (Select-String -Path $Entrypoint -SimpleMatch 'internal/app/terminal' -Quiet)) {
    Fail 'process entrypoint does not delegate through internal/app/terminal'
}
$DirectAppImport = Select-String -Path $Entrypoint -Pattern 'internal/app"' | Select-Object -First 1
if ($null -ne $DirectAppImport) {
    Fail "process entrypoint bypasses terminal adapter at $($DirectAppImport.Path):$($DirectAppImport.LineNumber)"
}

# go list is the positive deletion-boundary proof: core application and all
# backend owners resolve without either concrete TUI or Bubble Tea dependencies.
$CorePackages = @('./internal/app', './internal/app/controller', './internal/frontend', './internal/agent', './internal/session', './internal/config', './internal/ctx', './internal/llm', './internal/provider', './internal/skills', './internal/tools', './internal/logging', './internal/workspace')
$Deps = & go list -deps @CorePackages
if ($LASTEXITCODE -ne 0) { Fail 'core/backend package graph does not build' }
foreach ($Pattern in @('github.com/DohmBoy64Bit/RecompHamr/internal/tui', 'github.com/charmbracelet/bubbletea')) {
    if ($Deps -contains $Pattern) { Fail "core/backend deletion graph contains $Pattern" }
}

# Removed feature packages must not be imported under a different file layout.
$AllGo = Get-ChildItem -Path @((Join-Path $Root 'cmd'), (Join-Path $Root 'internal')) -Recurse -File -Filter '*.go' -ErrorAction SilentlyContinue
if ($AllGo.Count -eq 0) { Fail 'active Go source could not be enumerated' }
foreach ($Pattern in @('/mcp', '/update', '/classifier', '/doctor', '/project')) {
    $Hit = $AllGo | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) {
        Fail "removed feature dependency remains at $($Hit.Path):$($Hit.LineNumber): $Pattern"
    }
}

# Workspace filesystem/state capability belongs only to core application
# composition. It must not enter agent, session, frontend, or presentation.
$WorkspaceImports = $AllGo |
    Where-Object { $_.Name -notlike '*_test.go' -and $_.FullName -ne (Join-Path $Root 'internal/app/app.go') } |
    Select-String -SimpleMatch 'internal/workspace'
if ($null -ne $WorkspaceImports) {
    $Hit = $WorkspaceImports | Select-Object -First 1
    Fail "workspace capability bypasses core app at $($Hit.Path):$($Hit.LineNumber)"
}

# Concrete Stage F cache configuration belongs only to core application
# composition. Agent policy sees an injected executor and presentation sees
# only display-safe statuses; neither may construct a privileged tool set.
$ToolSetConstruction = $AllGo |
    Where-Object { $_.Name -notlike '*_test.go' -and $_.FullName -ne (Join-Path $Root 'internal/app/app.go') } |
    Select-String -SimpleMatch 'tools.NewSet('
if ($null -ne $ToolSetConstruction) {
    $Hit = $ToolSetConstruction | Select-Object -First 1
    Fail "tool-set cache authority bypasses core app at $($Hit.Path):$($Hit.LineNumber)"
}
foreach ($Pattern in @('RepomixrDir', 'RecompRefDir', 'MCPExec')) {
    $Hit = $AllGo | Where-Object { $_.Name -notlike '*_test.go' } | Select-String -SimpleMatch $Pattern | Select-Object -First 1
    if ($null -ne $Hit) { Fail "unsafe/deferred tool extension state remains at $($Hit.Path):$($Hit.LineNumber): $Pattern" }
}

Write-Host 'architecture (Stage G skills + Stage F tools + Stage D workspace + Stage C frontend boundary): PASS'
Write-Host 'skills/workspace/tool authority remains below presentation; core/backend deletion graph excludes internal/tui and Bubble Tea.'
