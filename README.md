# recomphamr

A terminal coding agent forked from [CodeHAMR](https://github.com/codehamr/codehamr),
specialized for reverse-engineering and native-code projects. Built for local
LLMs, also runs on OpenAI-compatible endpoints.

<img src="recomphamr.gif" width="640" alt="RecompHAMR demo">

## RE-first, local-first

recomphamr extends upstream CodeHAMR with RE-specific tooling: embedded skills
for reversing workflows, MCP servers for Ghidra and N64 debugging, project
handoff docs, and a system prompt tuned for unfamiliar codebases.

**Slash commands:** `/help`, `/models`, `/skills`, `/skill`, `/init-re`,
`/status-re`, `/doctor`, `/mcp`. Skills and MCP tools wire into the system
prompt dynamically.

## Install

Linux, macOS:

```bash
curl -fsSL https://recomphamr.com/install.sh | bash
```

Windows:

```cmd
curl -fsSL https://recomphamr.com/install.cmd -o install.cmd && install.cmd
```

Then run `recomphamr` in your project.

> **Warning:** AI agents run model-generated shell commands with full filesystem
> access. Best run inside safe sandboxes like devcontainers or isolated VMs.

## Config

On first run recomphamr seeds `.rehamr/config.yaml` with AMD-priority profiles
(`lmstudio-amd` default, `ollama-amd`). The system prompt and RE
skills are embedded in the binary.

Any OpenAI-compatible endpoint works. Example profiles:

```yaml
active: lmstudio-amd
models:
    lmstudio-amd:
        llm: deepseek-coder-v2
        url: http://localhost:1234
        key: ""
        context_size: 131072
    openai:
        llm: gpt-4.1
        url: https://api.openai.com
        key: sk-...
        context_size: 200000
```

`/models` lists profiles, `/models <name>` switches.

## Hardware

For RE workloads we target **AMD GPUs** with ROCm or Vulkan backends via
llama.cpp. A **~30B-class** model on **32 GB+ VRAM** recommended. Works with
LM Studio, Ollama, or any OpenAI-compatible endpoint.

## Give the agent a runtime

recomphamr verifies by running things, so give its sandbox the toolchains your
project needs; it cannot install them itself. If a check can't run, it reports
`unverified:` instead of pretending.

## Compare

| Tool | Pick if |
|---|---|
| **Frontier** | you want commercial heavyweight polish from Claude Code or Codex and accept the subscription cost |
| **[opencode](https://github.com/anomalyco/opencode)** | you want a great, loaded Swiss army knife and embrace plugin complexity |
| **[pi-agent](https://github.com/badlogic/pi-mono)** | you want something lighter than opencode and accept configuring your own extensions |
| **[CodeHAMR](https://github.com/codehamr/codehamr)** | you want the lightest coding agent with no skills, no plugins, just three slash commands |
| **recomphamr** | you do RE work (binaries, decompilation, unknown codebases), want MCP tool integration for Ghidra/Mupen64, and prefer lightweight skill-backed tooling |

## MCP Servers

recomphamr connects to MCP (Model Context Protocol) servers over stdio via
JSON-RPC 2.0, exposing their tools to the LLM alongside the four built-in
tools (`bash`, `read_file`, `write_file`, `edit_file`).

### Architecture

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

### Connection lifecycle

1. **Startup** — `BuiltinServers()` reads env vars for command paths and
   registers both servers. `ConnectAll()` runs in a background goroutine so
   the TUI isn't blocked. Set `RECOMPHAMR_MCP_AUTOSTART=0` to skip.

2. **Connect** — launches the server as a child process, speaks JSON-RPC 2.0
   over stdin/stdout. Handshake: `initialize` → `notifications/initialized` →
   `tools/list`. All operations have a 30s timeout.

3. **Disconnect** — kills the child process, resets state to `Disconnected`.

4. **Status** — shown on startup splash line and via `/mcp`.

### How tools reach the LLM (two gates)

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
| `/skill n64-debug-mcp` | 4 + 47 n64 tools |
| both skills | 4 + 20 + 47 = 71 |

No MCP skills loaded = zero MCP tools = same token cost as upstream CodeHAMR.

### Tool execution

When the LLM calls `ghidra.decompile_function`:

1. `tools.Execute()` → `runRaw()` → switch falls through to `default:`.
2. Checks the `tools.MCPExec` hook (wired during `New()`).
3. Hook calls `Manager.CallTool(ctx, "ghidra.decompile_function", args)`.
4. Manager splits on `.` → server `ghidra`, tool `decompile_function`.
5. Calls `Client.CallTool()` which sends JSON-RPC `tools/call`.
6. Result text returned as the tool message content in the conversation.

### Built-in servers

| Server | Default command | Env override |
|---|---|---|
| `ghidra` | `ghidra-mcp` | `RECOMPHAMR_MCP_GHIDRA_COMMAND` |
| `n64-debug-mcp` | `n64-debug-mcp` | `RECOMPHAMR_MCP_N64_COMMAND` |

Ghidra ships with the 20 most-used RE tools enabled by default
(`RECOMPHAMR_MCP_GHIDRA_TOOLS=*` for all). n64-debug-mcp allows all 47 tools
by default.

### Runtime management

```
/mcp                         show all servers, connection state, tool counts
/mcp connect <name>          launch server + JSON-RPC handshake
/mcp disconnect <name>       kill child process
/mcp tools <server>          list every tool (* = enabled)
/mcp enable <server> <t|*>   allow one tool or all
/mcp disable <server> <t|*>   block one tool or all
```

### Custom servers

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

## Skills

Skills inject focused context into the system prompt, giving the LLM
methodology and guardrails for specific tasks. MCP skills also gate which
server tools the LLM sees.

### How skills work

When loaded via `/skill <name>`, the skill's full markdown body is appended
to the system prompt under `## Active RE Skills`. This text travels with every
turn until recomphamr is restarted — skills survive `/clear` and `/models`
switches.

Skills also unlock MCP tools by convention: if a registered MCP server has
`RequireSkill: true`, loading a skill whose name matches the server (or maps
to it via the built-in `SkillServers` table) exposes that server's tools to
the LLM. For example, `/skill ghidra-mcp` maps to the `ghidra` server,
injecting `ghidra.*` tools.

### Built-in skills

Eight skills are compiled into the binary:

| `/skill <name>` | Purpose |
|---|---|
| `core-re` | General RE workflow |
| `evidence-mode` | Evidence-first methodology |
| `build-fix-loop` | Iterate on build failures |
| `file-format-reversing` | Binary format analysis |
| `function-discovery` | Find and classify functions |
| `ghidra-mcp` | Ghidra MCP integration (gates `ghidra.*` tools) |
| `n64-debug-mcp` | N64 runtime debugging via Mupen64Plus MCP (gates `n64-debug-mcp.*` tools) |
| `project-handoff` | Generate project docs for handoff |

List them with `/skills`; active skills are marked `*`.

### Custom skills

Drop `.md` files into `.rehamr/skills/` and they appear in `/skills` with a
`(custom)` label. Custom skills take precedence over built-in skills with the
same name.

```
.rehamr/
├── config.yaml
├── skills/
│   ├── my-workflow.md       # /skill my-workflow
│   └── my-mcp.md            # gates my-mcp.* tools if server registered
└── history
```

To pair a custom skill with a custom MCP server, name the skill after the
server and register it with `RequireSkill: true`:

## License

[MIT](LICENSE). Fork of [CodeHAMR](https://github.com/codehamr/codehamr).
Star it if it earned one.
