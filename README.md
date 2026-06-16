# recomphamr

A terminal coding agent forked from [CodeHAMR](https://github.com/codehamr/codehamr),
specialized for reverse-engineering and native-code projects. Built for local
LLMs, also runs on OpenAI-compatible endpoints.

<img src="recomphamr.gif" width="640" alt="RecompHAMR demo">

## RE-first, local-first

recomphamr extends upstream CodeHAMR with RE-specific tooling: embedded skills
for reversing workflows, project handoff docs, and a system prompt tuned for
unfamiliar codebases. It keeps upstream's simplicity — one loop, minimal tool
set, context-efficient packing — and adds only what RE work needs.

**Slash commands:** `/help`, `/models`, `/skills`, `/skill`, `/init-re`,
`/status-re`, `/doctor`. Skills wire into the system prompt dynamically.

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
| **recomphamr** | you do RE work (binaries, decompilation, unknown codebases) and want lightweight skill-backed tooling |

## Skills

Seven embedded RE skills are loaded on demand via `/skill <name>`:

| Command | Effect |
|---|---|
| `/skill core-re` | General RE workflow |
| `/skill evidence-mode` | Evidence-first methodology |
| `/skill build-fix-loop` | Iterate on build failures |
| `/skill file-format-reversing` | Binary format analysis |
| `/skill function-discovery` | Find and classify functions |
| `/skill ghidra-mcp` | Ghidra MCP integration |
| `/skill project-handoff` | Generate project docs for handoff |

Skills inject targeted context into the system prompt when active, keeping the
default prompt lean.

## License

[MIT](LICENSE). Fork of [CodeHAMR](https://github.com/codehamr/codehamr).
Star it if it earned one.
