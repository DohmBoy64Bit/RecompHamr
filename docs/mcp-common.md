# MCP Common Reference

Shared concepts, env vars, and management commands that apply to all MCP
servers. For server-specific setup, see the per-server docs.

## How MCP servers work

recomphamr connects to MCP servers over stdio (launched as child processes) or
streamable HTTP (already-running endpoints), using JSON-RPC 2.0. Each server
registers on startup. Tools are skill-gated — they only appear in the LLM's
tool list when a matching skill is loaded. Servers do not auto-connect by
default; use `/mcp connect <name>` or set `RECOMPHAMR_MCP_AUTOSTART=1`.

Full architecture details in **[mcp.md](mcp.md)**.

## How to configure a server

Three ways to tell recomphamr how to reach each MCP server:

1. **`.rehamr/mcp.json`** — the preferred persistent config. Define servers as
   JSON with `command`/`args` (stdio) or `url` (HTTP), plus optional `tools`
   whitelist and `requireSkill`. For example:
   ```json
   {
     "ghidra": {
       "command": "python",
       "args": ["C:/Tools/ghidra-mcp/bridge_mcp_ghidra.py"],
       "tools": ["*"]
     }
   }
   ```
   Omitted fields keep built-in defaults. Create this file in your project's
   `.rehamr/` directory.

2. **Put it on PATH** — the binary name (e.g. `ghidra-mcp`) must be
   findable by your OS when recomphamr starts.

3. **Set the env var** — point to the full path when the binary isn't on
   PATH or you want to override mcp.json at runtime.

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
belong in `.rehamr/mcp.json` or environment variables — this keeps
version-controlled project config (mcp.json, checked in) separate from
secrets (env vars, never checked in).

## Env vars

All MCP servers support the `RECOMPHAMR_MCP_<NAME>_*` pattern. The `<NAME>`
suffix is the server key: `GHIDRA`, `N64`, `PCRECOMP`, `PINE`, `OBJDIFF`,
`PCSX2`, `BIZHAWK`, `SEGA2ASM`. Three per-server vars are supported:

| Var suffix | Purpose |
|---|---|
| `RECOMPHAMR_MCP_<NAME>_COMMAND` | Override MCP server stdio command/path |
| `RECOMPHAMR_MCP_<NAME>_URL` | Override MCP server HTTP endpoint (streamable-http transport) |
| `RECOMPHAMR_MCP_<NAME>_TOOLS` | Tool list or `*` for all |

Global vars:

| Var | Purpose |
|---|---|
| `RECOMPHAMR_MCP_AUTOSTART` | Set to `1` to enable auto-connect on startup |
| `RECOMPHAMR_PCRECOMP_PATH` | Path to PCRECOMP-Next clone directory |

See the per-server docs for specific setup guides; see [mcp.md](mcp.md) for
architecture details.

## Runtime management

```
/mcp                         show all servers, connection state, tool counts
/mcp connect <name>          launch server (stdio) or connect (HTTP) + handshake
/mcp disconnect <name>       kill child process or close HTTP connection
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
