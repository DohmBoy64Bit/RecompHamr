# Tools vs Skills

In recomphamr, tools and skills serve different roles. Understanding the
distinction is key to using the agent effectively.

## Tools: what the LLM can *do*

Tools execute actions. They live in the `tools[]` array sent with every API
request — the LLM sees each tool's name, description, and parameter schema,
then chooses which to invoke by name. The result returns to the conversation
as a `tool` message.

**Built-in tools** (always available):
| Tool | What it does |
|---|---|
| `bash` | Run shell commands with a configurable timeout |
| `read_file` | Read a file from disk |
| `write_file` | Write a file to disk |
| `edit_file` | Perform exact string replacements in a file |

**MCP tools** (available when connected + skill loaded):
| Server | Example tool |
|---|---|
| `ghidra` | `ghidra.decompile_function`, `ghidra.get_xrefs_to` |
| `n64-debug-mcp` | `n64-debug-mcp.n64_read_memory`, `n64-debug-mcp.n64_set_breakpoint` |

## Skills: what the LLM *knows*

Skills inject methodology and guardrails into the system prompt. They're
Markdown text appended under `## Active RE Skills` in the system message —
the LLM reads them as instructions, not as callable functions. Skills don't
appear in the tools array.

**Built-in skills:**
| Skill | What it teaches |
|---|---|
| `core-re` | General RE workflow — navigate, trace, rename, repeat |
| `evidence-mode` | Evidence-first methodology — classify, cite, save |
| `build-fix-loop` | Iterate on build failures — compile, read error, fix |
| `file-format-reversing` | Binary format analysis — magic bytes, structs, offsets |
| `function-discovery` | Find and classify functions — entry points, tool patterns |
| `ghidra-mcp` | Ghidra MCP integration — exports, decompiler output, evidence |
| `n64-debug-mcp` | N64 runtime debugging — breakpoints, traces, frame capture |
| `project-handoff` | Generate project docs — state summary, ledger, next steps |

## The overlap: MCP skills

MCP skills are the bridge. Loading `/skill ghidra-mcp` does two things:

1. **Skill injection** — the skill's methodology text ("use Ghidra outputs as
   evidence, save to `.rehamr/evidence/`") enters the system prompt.
2. **Tool gating** — the `ghidra.*` tools (e.g. `ghidra.decompile_function`)
   are added to `buildTools()`, so the LLM can now call them.

Without the skill loaded, the Ghidra server could be connected and running —
the LLM simply never sees its tools. This keeps the token budget lean: zero
MCP tools unless explicitly activated.

## Key differences

| Aspect | Tools | Skills |
|---|---|---|
| Location | `tools[]` array in API request | System prompt |
| Purpose | Execute actions | Provide knowledge + guardrails |
| Invocation | LLM calls by name, returns result | LLM reads as instructions |
| Token cost | Schema per tool (~200 tokens each) | Body per skill (~200-400 tokens each) |
| Lifetime | Always available (built-in), skill-gated (MCP) | Loaded until restart |
| Addition | Add MCP servers via `Register()` | Drop `.md` files in `.rehamr/skills/` |
