# MCP Common Reference

Shared concepts, env vars, and management commands that apply to all MCP
servers. For server-specific setup, see the per-server docs.

## How MCP servers work

recomphamr launches MCP servers as child processes over stdio using JSON-RPC
2.0. Each server registers on startup and auto-connects in a background
goroutine. Tools are skill-gated — they only appear in the LLM's tool list
when a matching skill is loaded.

Full architecture details in **[mcp.md](mcp.md)**.

## Env vars

| Var | Purpose |
|---|---|
| `RECOMPHAMR_MCP_AUTOSTART` | Set to `0` to skip auto-connect on startup |
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
| `n64-debug-mcp` | [mcp-n64.md](mcp-n64.md) |
| `pcrecomp` | [mcp-pcrecomp.md](mcp-pcrecomp.md) |
| `mcp-pine` | [mcp-pine.md](mcp-pine.md) |
| `objdiff` | [mcp-objdiff.md](mcp-objdiff.md) |
| `pcsx2` | [mcp-pcsx2.md](mcp-pcsx2.md) |
| `bizhawk` | [mcp-bizhawk.md](mcp-bizhawk.md) |
| `sega2asm` | [mcp-sega2asm.md](mcp-sega2asm.md) |
