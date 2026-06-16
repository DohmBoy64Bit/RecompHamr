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

recomphamr connects to MCP (Model Context Protocol) servers over stdio,
exposing their tools to the LLM alongside built-in tools. Two servers ship
with built-in configs:

| Server | Default command | Env override |
|---|---|---|
| `ghidra` | `ghidra-mcp` | `RECOMPHAMR_MCP_GHIDRA_COMMAND` |
| `n64-debug-mcp` | `n64-debug-mcp` | `RECOMPHAMR_MCP_N64_COMMAND` |

Servers auto-connect on startup (set `RECOMPHAMR_MCP_AUTOSTART=0` to disable).
MCP tools are **scoped to active skills** — `ghidra.*` tools only inject when
`/skill ghidra-mcp` is loaded, `n64-debug-mcp.*` only with
`/skill n64-debug-mcp`. Without active MCP skills, zero extra tools are sent
to the LLM, keeping context lean.

**Tool filtering:** ghidra ships with only the top 20 most-used tools enabled
by default. Set `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` to enable all, or specify a
comma-separated list. n64-debug-mcp tools are all enabled by default.

```
/mcp                         show server status
/mcp connect <name>          connect to a server
/mcp disconnect <name>       disconnect from a server
/mcp tools <server>          list tools (* = enabled)
/mcp enable <server> <t|*>   enable one tool or all
/mcp disable <server> <t|*>  disable one tool or all
```

Server status appears on the startup splash — `* Connected (20 tools)` or
`  Disconnected`.

## Skills

Eight embedded RE skills are loaded on demand via `/skill <name>`:

| Command | Effect |
|---|---|
| `/skill core-re` | General RE workflow |
| `/skill evidence-mode` | Evidence-first methodology |
| `/skill build-fix-loop` | Iterate on build failures |
| `/skill file-format-reversing` | Binary format analysis |
| `/skill function-discovery` | Find and classify functions |
| `/skill ghidra-mcp` | Ghidra MCP integration |
| `/skill n64-debug-mcp` | N64 runtime debugging via Mupen64Plus MCP |
| `/skill project-handoff` | Generate project docs for handoff |

Skills inject targeted context into the system prompt when active. MCP skills
also gate which server tools the LLM sees.

## License

[MIT](LICENSE). Fork of [CodeHAMR](https://github.com/codehamr/codehamr).
Star it if it earned one.
