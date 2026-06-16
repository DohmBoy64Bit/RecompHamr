# recomphamr

A terminal coding agent forked from [CodeHAMR](https://github.com/codehamr/codehamr),
specialized for reverse-engineering and native-code projects. Built for local
LLMs, also runs on OpenAI-compatible endpoints.

![recomphamr demo](recomphamr.gif)

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
(`lmstudio-amd` default, `ollama-amd`, `hamrpass`). The system prompt and RE
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
    hamrpass:
        llm: hamrpass
        url: https://recomphamr.com
        key: hp_...
```

`/models` lists profiles, `/models <name>` switches.

## Hardware

Local LLMs finally caught up. For RE workloads we recommend a **~30B-class**
model on **32 GB+ unified RAM / VRAM**. AMD RX 7000-series GPUs with ROCm or
Vulkan backends (llama.cpp) are the default target.

Ollama users: raise the **Context length** slider to **64k+** (RAM/VRAM
permitting) and set `context_size` in `.rehamr/config.yaml` to match.

For coding, a ~30B-class model typically wants `temperature 0.6`, `top_p 0.95`,
`top_k 20`, and never greedy decoding (temp 0), which loops. If it still loops,
add `presence_penalty` and check your server applies it.

If the model prints tool calls as text instead of acting, enable your server's
tool-call parser; recomphamr warns you when that happens.

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

## HamrPass

HamrPass is optional. It's there if you want to support the project, or skip
benchmarking the latest open-weight model and tuning every parameter. We do that
work and ship it as one endpoint with sensible defaults.

There's a waitlist at [recomphamr.com](https://recomphamr.com). HamrPass only
gets built if real demand shows up there.

## License

[MIT](LICENSE). Fork of [CodeHAMR](https://github.com/codehamr/codehamr).
Star it if it earned one.
