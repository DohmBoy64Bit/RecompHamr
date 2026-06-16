# recomphamr

A terminal coding agent forked from [CodeHAMR](https://github.com/codehamr/codehamr),
specialized for reverse-engineering and native-code projects. Built for local
LLMs, also runs on OpenAI-compatible endpoints.

<img src="recomphamr.gif" width="640" alt="RecompHAMR demo">

## RE-first, local-first

recomphamr extends upstream CodeHAMR with RE-specific tooling: embedded skills
for reversing workflows, MCP servers for Ghidra and N64 debugging, project
handoff docs, and a system prompt tuned for unfamiliar codebases.

**Slash commands:** `/help`, `/clear`, `/models`, `/rehampass`, `/skills`,
`/skill`, `/init-re`, `/status-re`, `/doctor`, `/mcp`. Skills and MCP tools
wire into the system prompt dynamically.

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

On first run recomphamr seeds `.rehamr/config.yaml` with AMD-priority profiles:
`lmstudio-amd` (default), `lmstudio-fast`, `ollama-amd`, and `llama-vulkan`.
The system prompt and RE skills are embedded in the binary.

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

`/models` lists profiles, `/models <name>` switches. See
**[docs/profiles.md](docs/profiles.md)** for the full profile reference.

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
JSON-RPC 2.0, exposing their tools to the LLM alongside `bash`, `read_file`,
`write_file`, and `edit_file`.

Two servers ship with built-in configs: `ghidra` (20 tools by default) and
`n64-debug-mcp` (all tools). MCP tools are skill-gated to keep the token budget
lean — zero MCP tools are sent unless a matching skill is loaded.

```
/mcp                         show server status + tool counts
/mcp connect|disconnect <n>  launch or kill a server
/mcp tools <server>          list tools (* = enabled)
/mcp enable|disable <s> <t>  toggle individual or all tools
```

Full architecture, two-gate filtering, custom servers, and tool execution flow
are documented in **[docs/mcp.md](docs/mcp.md)**.

## Skills

Skills inject RE methodology and guardrails into the system prompt. Eight
are compiled into the binary; custom ones can be dropped in `.rehamr/skills/`.
MCP skills also gate which server tools the LLM sees.

```
/skills             list all skills (* = active, custom = from disk)
/skill <name>       load a skill (Tab to autocomplete)
```

Full details, built-in skill table, custom skill setup, and MCP pairing are
documented in **[docs/skills.md](docs/skills.md)**. For the distinction
between tools and skills, see **[docs/tools-vs-skills.md](docs/tools-vs-skills.md)**.

## License

[MIT](LICENSE). Fork of [CodeHAMR](https://github.com/codehamr/codehamr).
Star it if it earned one.
