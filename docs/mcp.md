# MCP Servers

recomphamr connects to MCP (Model Context Protocol) servers over stdio via
JSON-RPC 2.0, exposing their tools to the LLM alongside the built-in tools.

## Architecture

```
main.go
  └─ mcp.NewManager()
       ├─ Register(ServerConfig{...})  ← ghidra, n64, pcrecomp, mcp-pine, objdiff, pcsx2, bizhawk, sega2asm
       └─ ConnectAll()                 ← goroutine at startup

Model (TUI)
  └─ m.mcpManager
       ├─ buildTools() → ToolsForSkills(activeSkills) → []llm.Tool
       └─ tools.MCPExec hook            ← dispatch unknown tool calls
```

## Connection lifecycle

1. **Startup** — `BuiltinServers()` reads env vars for command paths and
   registers all servers. Set `RECOMPHAMR_MCP_AUTOSTART=1` to enable
   `ConnectAll()` on startup. By default, servers are registered but not
   auto-connected — use `/mcp connect <name>` to connect manually.

2. **Connect** — launches the server as a child process, speaks JSON-RPC 2.0
   over stdin/stdout. Handshake: `initialize` → `notifications/initialized` →
   `tools/list`. All operations have a 30s timeout.

3. **Disconnect** — kills the child process, resets state to `Disconnected`.

4. **Status** — shown on startup splash line and via `/mcp`.

## How tools reach the LLM (two gates)

MCP tools are NOT always sent to the LLM. Two per-server gates decide:

**Gate 1 — `RequireSkill`**
- `true` (built-in servers): tools only injected when a matching skill is
  loaded (e.g. `/skill ghidra-mcp` unlocks `ghidra.*` tools).
- `false` (custom servers): tools always injected, no skill needed.

**Gate 2 — `AllowedTools`**
- `nil`: all tools from the server are visible.
- `[list of N names]`: only those tools are visible.
- `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` → overrides to allow all.
- `RECOMPHAMR_MCP_GHIDRA_TOOLS=decompile_function,get_xrefs_to` → custom list.

**Result — token budget by skill state:**

| Active skills | Tools sent to LLM |
|---|---|
| (none) | Built-in tools only |
| `/skill ghidra-mcp` | Built-in + 20 ghidra tools |
| `/skill n64-debug-mcp` | Built-in + all n64 tools |
| `/skill pcrecomp` | Built-in + 8 pcrecomp tools |
| `/skill mcp-pine` | Built-in + all mcp-pine tools |
| `/skill objdiff` | Built-in + all objdiff tools |
| `/skill pcsx2` | Built-in + all pcsx2 tools |
| `/skill bizhawk` | Built-in + all bizhawk tools |
| `/skill sega2asm` | Built-in + all sega2asm tools |

No MCP skills loaded = zero MCP tools = same token cost as upstream CodeHAMR.

## Tool execution

When the LLM calls `ghidra.decompile_function`:

1. `tools.Execute()` → `runRaw()` → switch falls through to `default:`.
2. Checks the `tools.MCPExec` hook (wired during `New()`).
3. Hook calls `Manager.CallTool(ctx, "ghidra.decompile_function", args)`.
4. Manager splits on `.` → server `ghidra`, tool `decompile_function`.
5. Calls `Client.CallTool()` which sends JSON-RPC `tools/call`.
6. Result text returned as the tool message content in the conversation.

## Built-in servers

| Server | Default command | Env override |
|---|---|---|
| `ghidra` | `ghidra-mcp` | `RECOMPHAMR_MCP_GHIDRA_COMMAND` |
| `n64-debug-mcp` | `n64-debug-mcp` | `RECOMPHAMR_MCP_N64_COMMAND` |
| `pcrecomp` | `pcrecomp-mcp` | `RECOMPHAMR_MCP_PCRECOMP_COMMAND` |
| `mcp-pine` | `mcp-pine` | `RECOMPHAMR_MCP_PINE_COMMAND` |
| `objdiff` | `objdiff-mcp` | `RECOMPHAMR_MCP_OBJDIFF_COMMAND` |
| `pcsx2` | `pcsx2-mcp` | `RECOMPHAMR_MCP_PCSX2_COMMAND` |
| `bizhawk` | `mcp-bizhawk` | `RECOMPHAMR_MCP_BIZHAWK_COMMAND` |
| `sega2asm` | `sega2asm-mcp` | `RECOMPHAMR_MCP_SEGA2ASM_COMMAND` |

Ghidra ships with the 20 most-used RE tools enabled by default
(`RECOMPHAMR_MCP_GHIDRA_TOOLS=*` for all). All other servers allow all tools
by default. pcrecomp ships with 8 pipeline tools
(`RECOMPHAMR_MCP_PCRECOMP_TOOLS=*` for all).

## Runtime management

```
/mcp                         show all servers, connection state, tool counts
/mcp connect <name>          launch server + JSON-RPC handshake
/mcp disconnect <name>       kill child process
/mcp tools <server>          list every tool (* = enabled)
/mcp enable <server> <t|*>   allow one tool or all
/mcp disable <server> <t|*>   block one tool or all
```

## Custom servers

Any stdio MCP server can be registered at startup:

```go
mcp.Register(mcp.ServerConfig{
    Name:         "my-tools",
    Command:      "my-mcp-server",
    RequireSkill: false,  // always available, no skill needed
    AllowedTools: []string{"tool_a", "tool_b"},
})
```

For skill-gated servers, set `RequireSkill: true` — see
**[docs/skills.md](skills.md)** for how to pair a custom skill with a custom
MCP server.
