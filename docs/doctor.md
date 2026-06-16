# Doctor

`/doctor` runs environment diagnostics to validate that your RE setup is
ready. It checks system info, active model, profiles, GPU drivers, toolchain,
MCP server availability, endpoint reachability, and workspace state.

## Sections

### System
OS, architecture, Go runtime, project and config paths.

### Active model
Profile name, model ID, endpoint URL, context size.

### Profiles
All configured profiles with model and URL. The active profile is marked `*`.

### Memory / GPU hints
- **Linux:** `/proc/meminfo` (MemTotal / MemAvailable / SwapTotal)
- **Windows:** total physical memory via PowerShell
- Queries `rocm-smi`, `vulkaninfo`, `nvidia-smi`, `lspci` if available

### Toolchain
`which` check for: `git`, `go`, `python`, `python3`, `cmake`, `ninja`,
`make`, `ghidraRun`, `java`. Reports path if found, `missing` if not.

### MCP servers
- `which ghidra-mcp`, `which n64-debug-mcp`, `which pcrecomp-mcp`, `which mcp-pine`, `which objdiff-mcp`, `which pcsx2-mcp`, `which bizhawk-mcp`
- All MCP env vars: `RECOMPHAMR_MCP_GHIDRA_COMMAND`,
  `RECOMPHAMR_MCP_N64_COMMAND`, `RECOMPHAMR_MCP_PCRECOMP_COMMAND`,
  `RECOMPHAMR_MCP_GHIDRA_TOOLS`, `RECOMPHAMR_MCP_PCRECOMP_TOOLS`,
  `RECOMPHAMR_MCP_AUTOSTART`

### Endpoint check
HTTP GET to the active model's URL (`/v1/models`), reports status code
and reachability. 4-second timeout.

### Workspace (`.rehamr/`)
Each workspace file shows `present (size)` or `missing`:
`PROJECT.md`, `REPHAMR_STATE.md`, `EVIDENCE.md`, `BLOCKERS.md`,
`CHANGELOG.md`, `repomix-instruction.md`.

Also reports custom skill count (`.rehamr/skills/`) and cached repo count
(`.rehamr/repos/`).

### PCRECOMP-Next
- `RECOMPHAMR_PCRECOMP_PATH` — path to PCRECOMP-Next clone (or unset)
- `which python` — Python availability for PCRECOMP scripts
