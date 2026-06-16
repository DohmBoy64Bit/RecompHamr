# MCP Servers

recomphamr connects to MCP (Model Context Protocol) servers over stdio via
JSON-RPC 2.0, exposing their tools to the LLM alongside the four built-in
tools (`bash`, `read_file`, `write_file`, `edit_file`).

## Architecture

```
main.go
  └─ mcp.NewManager()
       ├─ Register(ServerConfig{...})  ← ghidra + n64-debug-mcp
       └─ ConnectAll()                 ← goroutine at startup

Model (TUI)
  └─ m.mcpManager
       ├─ buildTools() → ToolsForSkills(activeSkills) → []llm.Tool
       └─ tools.MCPExec hook            ← dispatch unknown tool calls
```

## Connection lifecycle

1. **Startup** — `BuiltinServers()` reads env vars for command paths and
   registers both servers. `ConnectAll()` runs in a background goroutine so
   the TUI isn't blocked. Set `RECOMPHAMR_MCP_AUTOSTART=0` to skip.

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
| (none) | 4 (bash, read, write, edit) |
| `/skill ghidra-mcp` | 4 + 20 ghidra tools |
| `/skill n64-debug-mcp` | 4 + all n64 tools |
| both skills | 4 + 20 + all n64 tools |

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

Ghidra ships with the 20 most-used RE tools enabled by default
(`RECOMPHAMR_MCP_GHIDRA_TOOLS=*` for all). n64-debug-mcp allows all tools
by default.

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

For skill-gated servers, set `RequireSkill: true` and name the skill `.md`
file after the server:

```go
mcp.Register(mcp.ServerConfig{
    Name:         "my-mcp",
    Command:      "my-mcp-server",
    RequireSkill: true,  // only activated via /skill my-mcp
})
```

Then drop `.rehamr/skills/my-mcp.md` with your methodology. Loading
`/skill my-mcp` injects both the skill text and the `my-mcp.*` tools.
