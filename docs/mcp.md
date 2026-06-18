# MCP Servers

recomphamr connects to MCP (Model Context Protocol) servers over stdio or
streamable HTTP via JSON-RPC 2.0, exposing their tools to the LLM alongside
the built-in tools.

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

1. **Startup** — `BuiltinServers()` returns defaults; `.rehamr/mcp.json`
   merges overrides and adds custom servers; `RECOMPHAMR_MCP_*` env vars
   apply as final overrides. Set `RECOMPHAMR_MCP_AUTOSTART=1` to enable
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

| Server | Default command | Env override prefix |
|---|---|---|
| `ghidra` | `ghidra-mcp` | `RECOMPHAMR_MCP_GHIDRA_*` |
| `n64-debug-mcp` | `n64-debug-mcp` | `RECOMPHAMR_MCP_N64_*` |
| `pcrecomp` | `pcrecomp-mcp` | `RECOMPHAMR_MCP_PCRECOMP_*` |
| `mcp-pine` | `mcp-pine` | `RECOMPHAMR_MCP_PINE_*` |
| `objdiff` | `objdiff-mcp` | `RECOMPHAMR_MCP_OBJDIFF_*` |
| `pcsx2` | `pcsx2-mcp` | `RECOMPHAMR_MCP_PCSX2_*` |
| `bizhawk` | `mcp-bizhawk` | `RECOMPHAMR_MCP_BIZHAWK_*` |
| `sega2asm` | `sega2asm-mcp` | `RECOMPHAMR_MCP_SEGA2ASM_*` |

All servers support `_COMMAND` (stdio path), `_URL` (HTTP endpoint), and
`_TOOLS` (comma-separated list or `*`). Ghidra ships with 20 default tools;
pcrecomp ships with 8; all others allow all tools by default.

## Runtime management

See **[mcp-common.md](mcp-common.md)** for the full `/mcp` command reference
and env var configuration.

## Custom servers

Add servers in `.rehamr/mcp.json` — the preferred non-code approach.
Servers defined there merge with built-in ones (omitted fields keep defaults)
or register as new entries:

```json
{
  "my-tools": {
    "command": "my-mcp-server",
    "tools": ["tool_a", "tool_b"]
  }
}
```

To register servers programmatically at startup, use the Go API:

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
