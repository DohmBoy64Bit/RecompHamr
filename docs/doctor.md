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
- `which` check for all 8 server binaries: `ghidra-mcp`, `n64-debug-mcp`,
  `pcrecomp-mcp`, `mcp-pine`, `objdiff-mcp`, `pcsx2-mcp`,
  `mcp-bizhawk`, `sega2asm-mcp`
- All MCP env vars: `RECOMPHAMR_MCP_<NAME>_*` pattern (COMMAND, URL, TOOLS)
  for all 8 servers: GHIDRA, N64, PCRECOMP, PINE, OBJDIFF, PCSX2, BIZHAWK,
  SEGA2ASM, plus `RECOMPHAMR_MCP_AUTOSTART`, `RECOMPHAMR_PCRECOMP_PATH`
- Full env var reference in **[mcp-common.md](mcp-common.md)**

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
