[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$ScenarioPath,

    [string]$AppPath,
    [string]$Workspace,
    [string]$ArtifactDirectory,
    [switch]$ValidateOnly
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Resolve-FullPath([string]$Path, [string]$Base) {
    if ([System.IO.Path]::IsPathRooted($Path)) {
        return [System.IO.Path]::GetFullPath($Path)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $Base $Path))
}

function Resolve-ContainedPath([string]$Path, [string]$Root, [string]$Purpose) {
    $full = Resolve-FullPath $Path $Root
    $rootFull = [System.IO.Path]::GetFullPath($Root).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $prefix = $rootFull + [System.IO.Path]::DirectorySeparatorChar
    if ($full -ne $rootFull -and -not $full.StartsWith($prefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw "$Purpose path escapes its configured root"
    }
    return $full
}

function Get-RequiredProperty($Object, [string]$Name, [string]$Context) {
    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property -or $null -eq $property.Value -or [string]::IsNullOrWhiteSpace([string]$property.Value)) {
        throw "$Context requires '$Name'"
    }
    return $property.Value
}

function Test-Scenario($Scenario) {
    if ($Scenario.version -ne 1) { throw 'scenario version must be 1' }
    [void](Get-RequiredProperty $Scenario 'name' 'scenario')
    if ($null -eq $Scenario.terminal -or [int]$Scenario.terminal.columns -lt 20 -or [int]$Scenario.terminal.rows -lt 8) {
        throw 'scenario terminal requires columns >= 20 and rows >= 8'
    }
    if ($null -eq $Scenario.steps -or $Scenario.steps.Count -eq 0) {
        throw 'scenario requires at least one step'
    }
    $allowed = @('launch', 'wait_event', 'assert_event_count', 'assert_event_sequence', 'type_text', 'key', 'resize', 'screenshot', 'write_file', 'assert_file', 'remove_file', 'sleep', 'close_window')
    $labels = @{}
    foreach ($step in $Scenario.steps) {
        $type = [string](Get-RequiredProperty $step 'type' 'step')
        $label = [string](Get-RequiredProperty $step 'label' "step '$type'")
        if ($type -notin $allowed) { throw "step '$label' has unsupported type '$type'" }
        if ($labels.ContainsKey($label)) { throw "duplicate step label '$label'" }
        $labels[$label] = $true
        switch ($type) {
            'wait_event' { [void](Get-RequiredProperty $step 'category' "step '$label'") }
            'assert_event_count' {
                [void](Get-RequiredProperty $step 'category' "step '$label'")
                [void](Get-RequiredProperty $step 'count' "step '$label'")
            }
            'assert_event_sequence' {
                if ($null -eq $step.categories -or $step.categories.Count -eq 0) { throw "step '$label' requires categories" }
            }
            'type_text' { [void](Get-RequiredProperty $step 'text' "step '$label'") }
            'key' { [void](Get-RequiredProperty $step 'key' "step '$label'") }
            'screenshot' { [void](Get-RequiredProperty $step 'file' "step '$label'") }
            'write_file' {
                [void](Get-RequiredProperty $step 'path' "step '$label'")
                [void](Get-RequiredProperty $step 'text' "step '$label'")
            }
            'assert_file' { [void](Get-RequiredProperty $step 'path' "step '$label'") }
            'remove_file' { [void](Get-RequiredProperty $step 'path' "step '$label'") }
        }
    }
}

$scenarioFile = (Resolve-Path -LiteralPath $ScenarioPath).Path
$scenarioDirectory = Split-Path -Parent $scenarioFile
$scenario = Get-Content -LiteralPath $scenarioFile -Raw | ConvertFrom-Json
Test-Scenario $scenario
if ($ValidateOnly) {
    Write-Host "TUI acceptance scenario: PASS ($($scenario.name))"
    exit 0
}

if ($env:OS -ne 'Windows_NT') { throw 'TUI acceptance execution requires Windows' }

$resolvedApp = if ($AppPath) { Resolve-FullPath $AppPath (Get-Location).Path } else { Resolve-FullPath ([string]$scenario.app_path) $scenarioDirectory }
$resolvedWorkspace = if ($Workspace) { Resolve-FullPath $Workspace (Get-Location).Path } else { Resolve-FullPath ([string]$scenario.workspace) $scenarioDirectory }
$resolvedArtifacts = if ($ArtifactDirectory) { Resolve-FullPath $ArtifactDirectory (Get-Location).Path } else { Resolve-FullPath ([string]$scenario.artifact_directory) $scenarioDirectory }
$debugLog = Resolve-ContainedPath ([string]$scenario.debug_log) $resolvedWorkspace 'debug log'

if (-not (Test-Path -LiteralPath $resolvedApp -PathType Leaf)) { throw "application not found: $resolvedApp" }
if (-not (Test-Path -LiteralPath $resolvedWorkspace -PathType Container)) { throw "workspace not found: $resolvedWorkspace" }
[void](New-Item -ItemType Directory -Force -Path $resolvedArtifacts)

Add-Type -AssemblyName System.Drawing
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
using System.Text;
public static class TuiHarnessWin32 {
    public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
    [StructLayout(LayoutKind.Sequential)]
    public struct RECT { public int Left, Top, Right, Bottom; }
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc callback, IntPtr lParam);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
    [DllImport("user32.dll")] public static extern bool MoveWindow(IntPtr hWnd, int x, int y, int width, int height, bool repaint);
    [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
    [DllImport("user32.dll")] public static extern void keybd_event(byte virtualKey, byte scanCode, uint flags, UIntPtr extraInfo);
    [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int command);
    [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint processId);
}
'@

function Get-TerminalWindows {
    $script:terminalWindows = New-Object System.Collections.Generic.List[long]
    $callback = [TuiHarnessWin32+EnumWindowsProc]{
        param([IntPtr]$handle, [IntPtr]$unused)
        if (-not [TuiHarnessWin32]::IsWindowVisible($handle)) { return $true }
        [uint32]$processId = 0
        [void][TuiHarnessWin32]::GetWindowThreadProcessId($handle, [ref]$processId)
        $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
        if ($null -ne $process -and $process.ProcessName -eq 'WindowsTerminal') {
            $script:terminalWindows.Add($handle.ToInt64())
        }
        return $true
    }
    [void][TuiHarnessWin32]::EnumWindows($callback, [IntPtr]::Zero)
    return @($script:terminalWindows)
}

function Find-NewTerminalWindow([long[]]$ExistingHandles, [int]$TimeoutSeconds = 15) {
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        $newHandle = @(Get-TerminalWindows | Where-Object { $_ -notin $ExistingHandles } | Select-Object -First 1)
        if ($newHandle.Count -eq 1) { return [IntPtr]$newHandle[0] }
        Start-Sleep -Milliseconds 100
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "a new Windows Terminal window was not found within $TimeoutSeconds seconds"
}

function Wait-WindowClosed([IntPtr]$Handle, [int]$TimeoutSeconds = 10) {
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        if ($Handle.ToInt64() -notin @(Get-TerminalWindows)) { return }
        Start-Sleep -Milliseconds 100
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "Windows Terminal window did not close within $TimeoutSeconds seconds"
}

function Focus-Window([IntPtr]$Handle) {
    [void][TuiHarnessWin32]::ShowWindow($Handle, 9)
    [void][TuiHarnessWin32]::SetForegroundWindow($Handle)
    if ([TuiHarnessWin32]::GetForegroundWindow() -ne $Handle) {
        [TuiHarnessWin32]::keybd_event(0x12, 0, 0, [UIntPtr]::Zero)
        [TuiHarnessWin32]::keybd_event(0x12, 0, 2, [UIntPtr]::Zero)
        [void][TuiHarnessWin32]::SetForegroundWindow($Handle)
    }
    if ([TuiHarnessWin32]::GetForegroundWindow() -ne $Handle) {
        throw 'could not focus the acceptance window'
    }
    Start-Sleep -Milliseconds 150
}

function Get-EventCategories([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return @() }
    $content = Get-Content -LiteralPath $Path -Raw
    $matches = [regex]::Matches($content, '(?m)^\[[^\]]+\]\s+([A-Za-z0-9_]+)\r?$')
    return @($matches | ForEach-Object { $_.Groups[1].Value })
}

function Wait-Event([string]$Path, [string]$Category, [int]$Minimum, [int]$TimeoutSeconds) {
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        $categories = @(Get-EventCategories $Path)
        if (@($categories | Where-Object { $_ -eq $Category }).Count -ge $Minimum) { return $categories }
        Start-Sleep -Milliseconds 100
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "event category '$Category' did not reach count $Minimum within $TimeoutSeconds seconds"
}

function Assert-EventSequence([string[]]$Actual, [string[]]$Expected) {
    $position = 0
    foreach ($category in $Actual) {
        if ($position -lt $Expected.Count -and $category -eq $Expected[$position]) { $position++ }
    }
    if ($position -ne $Expected.Count) {
        throw "event category sequence was not observed; matched $position of $($Expected.Count)"
    }
}

function Save-WindowScreenshot([IntPtr]$Handle, [string]$Path) {
    Focus-Window $Handle
    $rect = New-Object TuiHarnessWin32+RECT
    if (-not [TuiHarnessWin32]::GetWindowRect($Handle, [ref]$rect)) { throw 'GetWindowRect failed' }
    $width = $rect.Right - $rect.Left
    $height = $rect.Bottom - $rect.Top
    $bitmap = New-Object Drawing.Bitmap $width, $height
    $graphics = [Drawing.Graphics]::FromImage($bitmap)
    try {
        $graphics.CopyFromScreen($rect.Left, $rect.Top, 0, 0, $bitmap.Size)
        $parent = Split-Path -Parent $Path
        [void](New-Item -ItemType Directory -Force -Path $parent)
        $bitmap.Save($Path, [Drawing.Imaging.ImageFormat]::Png)
    } finally {
        $graphics.Dispose()
        $bitmap.Dispose()
    }
}

function Send-NamedKey([IntPtr]$Handle, [string]$Key) {
    $map = @{
        'enter' = '{ENTER}'; 'tab' = '{TAB}'; 'escape' = '{ESC}'; 'up' = '{UP}'; 'down' = '{DOWN}'
        'page_up' = '{PGUP}'; 'page_down' = '{PGDN}'; 'ctrl_c' = '^c'; 'ctrl_d' = '^d'; 'ctrl_l' = '^l'; 'alt_f4' = '%{F4}'
    }
    if (-not $map.ContainsKey($Key)) { throw "unsupported key '$Key'" }
    Focus-Window $Handle
    $shell = New-Object -ComObject WScript.Shell
    $shell.SendKeys($map[$Key])
}

function Send-Paste([IntPtr]$Handle, [string]$Text) {
    $hadClipboard = $false
    $oldClipboard = $null
    try {
        try { $oldClipboard = Get-Clipboard -Raw; $hadClipboard = $true } catch {}
        Set-Clipboard -Value $Text
        Focus-Window $Handle
        $shell = New-Object -ComObject WScript.Shell
        $shell.SendKeys('^v')
        Start-Sleep -Milliseconds 150
    } finally {
        if ($hadClipboard) { Set-Clipboard -Value $oldClipboard } else { Set-Clipboard -Value $null }
    }
}

function Resize-Window([IntPtr]$Handle, [int]$FromColumns, [int]$FromRows, [int]$ToColumns, [int]$ToRows) {
    $rect = New-Object TuiHarnessWin32+RECT
    if (-not [TuiHarnessWin32]::GetWindowRect($Handle, [ref]$rect)) { throw 'GetWindowRect failed' }
    $width = [Math]::Max(320, [int](($rect.Right - $rect.Left) * $ToColumns / $FromColumns))
    $height = [Math]::Max(200, [int](($rect.Bottom - $rect.Top) * $ToRows / $FromRows))
    if (-not [TuiHarnessWin32]::MoveWindow($Handle, 20, 20, $width, $height, $true)) { throw 'MoveWindow failed' }
    Start-Sleep -Milliseconds 350
}

$steps = New-Object System.Collections.Generic.List[object]
$windowHandle = [IntPtr]::Zero
$columns, $rows = [int]$scenario.terminal.columns, [int]$scenario.terminal.rows
$started = [DateTime]::UtcNow
$status = 'passed'
$failure = $null
$failureDetail = $null

try {
    foreach ($step in $scenario.steps) {
        $stepStarted = [DateTime]::UtcNow
        $label, $type = [string]$step.label, [string]$step.type
        switch ($type) {
            'launch' {
                $existingHandles = @(Get-TerminalWindows)
                $encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes("& '$($resolvedApp.Replace("'", "''"))'"))
                $arguments = @('-w', 'new', '--size', "$columns,$rows", '-d', $resolvedWorkspace, 'powershell.exe', '-NoExit', '-EncodedCommand', $encoded)
                Start-Process -FilePath 'wt.exe' -ArgumentList $arguments
                $windowHandle = Find-NewTerminalWindow $existingHandles ([int]($(if ($step.timeout_seconds) { $step.timeout_seconds } else { 20 })))
            }
            'wait_event' {
                $minimum = if ($step.minimum) { [int]$step.minimum } else { 1 }
                $timeout = if ($step.timeout_seconds) { [int]$step.timeout_seconds } else { 120 }
                [void](Wait-Event $debugLog ([string]$step.category) $minimum $timeout)
            }
            'assert_event_count' {
                $actual = @((Get-EventCategories $debugLog) | Where-Object { $_ -eq [string]$step.category }).Count
                $expected = [int]$step.count
                if ($actual -ne $expected) { throw "event '$($step.category)' count $actual, expected $expected" }
            }
            'assert_event_sequence' { Assert-EventSequence @(Get-EventCategories $debugLog) @($step.categories) }
            'type_text' { Send-Paste $windowHandle ([string]$step.text) }
            'key' { Send-NamedKey $windowHandle ([string]$step.key) }
            'resize' {
                $newColumns, $newRows = [int]$step.columns, [int]$step.rows
                Resize-Window $windowHandle $columns $rows $newColumns $newRows
                $columns, $rows = $newColumns, $newRows
            }
            'screenshot' {
                $target = Resolve-ContainedPath ([string]$step.file) $resolvedArtifacts 'screenshot'
                Save-WindowScreenshot $windowHandle $target
            }
            'write_file' {
                $target = Resolve-ContainedPath ([string]$step.path) $resolvedWorkspace 'written file'
                $parent = Split-Path -Parent $target
                [void](New-Item -ItemType Directory -Force -Path $parent)
                [IO.File]::WriteAllText($target, [string]$step.text, [Text.UTF8Encoding]::new($false))
            }
            'assert_file' {
                $target = Resolve-ContainedPath ([string]$step.path) $resolvedWorkspace 'asserted file'
                $timeout = if ($step.PSObject.Properties['timeout_seconds']) { [int]$step.timeout_seconds } else { 10 }
                $shouldExist = if ($step.PSObject.Properties['exists']) { [bool]$step.exists } else { $true }
                if (-not $shouldExist) {
                    if (Test-Path -LiteralPath $target -PathType Leaf) { throw "asserted file unexpectedly exists: $target" }
                } else {
                    $deadline = [DateTime]::UtcNow.AddSeconds($timeout)
                    while (-not (Test-Path -LiteralPath $target -PathType Leaf) -and [DateTime]::UtcNow -lt $deadline) {
                        Start-Sleep -Milliseconds 100
                    }
                    if (-not (Test-Path -LiteralPath $target -PathType Leaf)) { throw "asserted file does not exist after $timeout seconds: $target" }
                    if ($null -ne $step.PSObject.Properties['sha256']) {
                        $actual = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash
                        if ($actual -ne [string]$step.sha256) { throw "asserted file hash mismatch: $target" }
                    }
                    if ($null -ne $step.PSObject.Properties['text']) {
                        $actual = Get-Content -LiteralPath $target -Raw
                        if ($actual -cne [string]$step.text) { throw "asserted file content mismatch: $target" }
                    }
                }
            }
            'remove_file' {
                $target = Resolve-ContainedPath ([string]$step.path) $resolvedWorkspace 'removed file'
                if (Test-Path -LiteralPath $target -PathType Leaf) { Remove-Item -LiteralPath $target -Force }
            }
            'sleep' { Start-Sleep -Milliseconds ([int]$step.milliseconds) }
            'close_window' {
                Send-NamedKey $windowHandle 'alt_f4'
                Wait-WindowClosed $windowHandle ([int]($(if ($step.PSObject.Properties['timeout_seconds']) { $step.timeout_seconds } else { 10 })))
                $windowHandle = [IntPtr]::Zero
            }
        }
        $steps.Add([pscustomobject]@{ label = $label; type = $type; status = 'passed'; duration_ms = [int]([DateTime]::UtcNow - $stepStarted).TotalMilliseconds })
    }
} catch {
    $status = 'failed'
    $failure = "step '$($step.label)' ($($step.type)) failed"
    $failureDetail = $_.Exception.Message
    $steps.Add([pscustomobject]@{ label = [string]$step.label; type = [string]$step.type; status = 'failed'; duration_ms = [int]([DateTime]::UtcNow - $stepStarted).TotalMilliseconds })
    if ($windowHandle -ne [IntPtr]::Zero) {
        try { Save-WindowScreenshot $windowHandle (Join-Path $resolvedArtifacts 'failure.png') } catch {}
        try {
            Send-NamedKey $windowHandle 'alt_f4'
            Wait-WindowClosed $windowHandle 10
            $windowHandle = [IntPtr]::Zero
        } catch {}
    }
} finally {
    $categories = @(Get-EventCategories $debugLog)
    $safeTail = @($categories | Select-Object -Last 20)
    $report = [ordered]@{
        schema_version = 1
        scenario = [string]$scenario.name
        status = $status
        started_utc = $started.ToString('o')
        duration_ms = [int]([DateTime]::UtcNow - $started).TotalMilliseconds
        terminal = [ordered]@{ columns = $columns; rows = $rows }
        steps = $steps
        event_categories_tail = $safeTail
        failure = $failure
    }
    $report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $resolvedArtifacts 'report.json') -Encoding utf8
}

if ($status -ne 'passed') { throw "TUI acceptance scenario failed: $failureDetail" }
Write-Host "TUI acceptance: PASS ($($scenario.name))"
