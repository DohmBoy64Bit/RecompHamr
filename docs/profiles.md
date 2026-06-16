# Profiles

recomphamr seeds four AMD-priority profiles on first run. Each targets a
different local inference backend and model. After first run the profile list
is yours — add, rename, or delete as needed.

## Purpose

Profiles decouple model selection from recomphamr's agent loop. Switching
profiles (`/models <name>`) instantly changes which backend and model the LLM
calls, without restarting. Active skills, conversation history, and MCP
connections survive the switch.

All built-in profiles set `context_size: 32768` (32k tokens) — enough for the
system prompt, twenty loaded skills, conversation history, and tool results
before packing kicks in. Raise it for larger models or longer sessions.

## Built-in profiles

### `lmstudio-amd` (default)

| Setting | Value |
|---|---|
| Model | `qwen/qwen3.6-35b-a3b` |
| URL | `http://localhost:1234` |
| Backend | LM Studio |

**Use for:** Primary RE sessions. Qwen 3.6 35B is a MoE (Mixture of Experts)
model — 35B total params, ~3B active per token — which fits comfortably in
16GB VRAM while delivering strong reasoning for decompilation, function
analysis, and evidence classification. LM Studio's ROCm support on AMD GPUs
makes this the default.

### `lmstudio-fast`

| Setting | Value |
|---|---|
| Model | `openai/gpt-oss-20b` |
| URL | `http://localhost:1234` |
| Backend | LM Studio |

**Use for:** Quick turns, simple edits, bash verification, or when the
primary model queue is busy. GPT-OSS 20B is smaller and faster — good for
build/fix loops where you need rapid iteration over correctness.

### `ollama-amd`

| Setting | Value |
|---|---|
| Model | `qwen3.6:27b` |
| URL | `http://localhost:11434` |
| Backend | Ollama |

**Use for:** Users who prefer Ollama over LM Studio. Qwen 3.6 27B is a
middle-weight model that balances speed and reasoning. Ollama's Vulkan backend
works across a wide range of AMD cards without ROCm setup.

### `llama-vulkan`

| Setting | Value |
|---|---|
| Model | `qwen3.6-35b-a3b` |
| URL | `http://localhost:8080` |
| Backend | llama.cpp (server mode) |

**Use for:** Users running llama.cpp directly with Vulkan acceleration.
Port 8080 is the default for `llama-server`. Uses the same Qwen 3.6 35B MoE
model as `lmstudio-amd` but through llama.cpp's native server — useful when
LM Studio isn't available or when you want direct control over sampling and
memory parameters.

## Switching profiles

```
/models                  list all profiles (▸ marks active)
/models lmstudio-fast    switch to lmstudio-fast
```

The switch rebuilds the LLM client in-place. No restart needed.

## Adding profiles

Edit `.rehamr/config.yaml` directly — any OpenAI-compatible endpoint works
(cloud, self-hosted, or local). The `key` field is optional for local
backends. See the **[Config section](../README.md#config)** for an example.
