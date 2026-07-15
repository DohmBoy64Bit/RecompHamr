# Baseline Tools

The baseline exposes exactly four tools to the model.

## `powershell`

Runs a fresh non-interactive PowerShell process in the current project directory.

Arguments:

- `script` — required PowerShell code.
- `timeout_seconds` — optional; defaults to 120 seconds and is capped at 3600 seconds.

PowerShell 7 (`pwsh`) is preferred. On Windows, Windows PowerShell is also accepted as a fallback.

## `read_file`

Reads one file exactly and returns bounded output.

## `write_file`

Creates or replaces a file with the supplied content.

## `edit_file`

Performs one exact, unique string replacement. Ambiguous, missing, or no-op replacements fail explicitly.

## Deliberate omissions

There is no unknown-tool extension fallback in the baseline. MCP, skills, `repomixr`, `recomp_reference`, and other RecompHamr-specific tools remain excluded until their integration stages.
