# MCP Common Reference

Shared concepts, env vars, and management commands that apply to all MCP
servers. For server-specific setup, see the per-server docs.

## How MCP servers work

recomphamr launches MCP servers as child processes over stdio using JSON-RPC
2.0. Each server registers on startup and auto-connects in a background
goroutine. Tools are skill-gated — they only appear in the LLM's tool list
when a matching skill is loaded.

Full architecture details in **[mcp.md](mcp.md)**.

## How to configure a server

MCP servers are launched as child processes. Two ways to tell recomphamr where
the binary is:

1. **Put it on PATH** — the binary name (e.g. `ghidra-mcp`) must be
   findable by your OS when recomphamr starts.
2. **Set the env var** — point to the full path when the binary isn't on
   PATH or you want to override the default name.

**Verify with `/doctor`** — it shows every MCP env var and whether it's set.
Run it after configuring to confirm your settings.

### Setting env vars

**Windows (PowerShell — per-session):**
```
$env:RECOMPHAMR_MCP_GHIDRA_COMMAND = "C:\Tools\ghidra-mcp.exe"
```

**Windows (persistent — survives terminal restarts):**
```
[System.Environment]::SetEnvironmentVariable("RECOMPHAMR_MCP_GHIDRA_COMMAND", "C:\Tools\ghidra-mcp.exe", "User")
```
Then restart your terminal for the change to take effect.

**Linux / macOS (per-session):**
```bash
export RECOMPHAMR_MCP_GHIDRA_COMMAND=/opt/ghidra-mcp/bridge.py
```

**Linux / macOS (persistent — add to `~/.bashrc` or `~/.zshrc`):**
```bash
echo 'export RECOMPHAMR_MCP_GHIDRA_COMMAND=/opt/ghidra-mcp/bridge.py' >> ~/.bashrc
```

### Why PATH first?

If a server binary is on PATH (e.g. installed via `pip install ghidra-mcp`,
`npm install -g mcp-bizhawk`, or `go install`), recomphamr finds it
automatically — no env var needed. Env vars exist as overrides for custom
install locations or renamed binaries.

Note: MCP server paths are **never** stored in `.rehamr/config.yaml`. They
are strictly environment variables — this keeps credentials and local paths
out of version-controlled config files.

## Env vars

| Var | Purpose |
|---|---|
| `RECOMPHAMR_MCP_AUTOSTART` | Set to `1` to enable auto-connect on startup |
| `RECOMPHAMR_MCP_GHIDRA_COMMAND` | Override ghidra MCP server command/path |
| `RECOMPHAMR_MCP_GHIDRA_TOOLS` | Ghidra tool list or `*` for all |
| `RECOMPHAMR_MCP_N64_COMMAND` | Override n64-debug-mcp server command/path |
| `RECOMPHAMR_MCP_PCRECOMP_COMMAND` | Override pcrecomp MCP server command/path |
| `RECOMPHAMR_MCP_PCRECOMP_TOOLS` | PCRECOMP tool list or `*` for all |
| `RECOMPHAMR_MCP_PINE_COMMAND` | Override mcp-pine server command/path |
| `RECOMPHAMR_MCP_OBJDIFF_COMMAND` | Override objdiff MCP server command/path |
| `RECOMPHAMR_MCP_PCSX2_COMMAND` | Override pcsx2 MCP server command/path |
| `RECOMPHAMR_MCP_BIZHAWK_COMMAND` | Override bizhawk MCP server command/path |
| `RECOMPHAMR_MCP_SEGA2ASM_COMMAND` | Override sega2asm MCP server command/path |
| `RECOMPHAMR_PCRECOMP_PATH` | Path to PCRECOMP-Next clone directory |

## Runtime management

```
/mcp                         show all servers, connection state, tool counts
/mcp path <server> [path]    set or show server binary path
/mcp connect <name>          launch server + JSON-RPC handshake
/mcp disconnect <name>       kill child process
/mcp tools <server>          list every tool (* = enabled)
/mcp enable <server> <t|*>   allow one tool or all
/mcp disable <server> <t|*>   block one tool or all
```

Servers appear on the startup splash — `* Connected` with tool counts,
or `  Disconnected` if not running.

## Per-server docs

| Server | Setup guide |
|---|---|
| `ghidra` | [mcp-ghidra.md](mcp-ghidra.md) |
| `n64-debug-mcp` | [mcp-n64-debug-mcp.md](mcp-n64-debug-mcp.md) |
| `pcrecomp` | [mcp-pcrecomp.md](mcp-pcrecomp.md) |
| `mcp-pine` | [mcp-pine.md](mcp-pine.md) |
| `objdiff` | [mcp-objdiff.md](mcp-objdiff.md) |
| `pcsx2` | [mcp-pcsx2.md](mcp-pcsx2.md) |
| `bizhawk` | [mcp-bizhawk.md](mcp-bizhawk.md) |
| `sega2asm` | [mcp-sega2asm.md](mcp-sega2asm.md) |
