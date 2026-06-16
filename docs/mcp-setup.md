# MCP Server Setup

Setup instructions for the two built-in MCP servers. Both are external Python
applications — recomphamr launches them as child processes over stdio; no HTTP
ports or config files needed from recomphamr's side.

## Ghidra MCP

Bridges Ghidra's analysis engine to the LLM, exposing ~100+ tools for
decompilation, cross-references, symbol management, and data analysis.

### Dependencies

- **Ghidra 12.1.2** — installed and a project open with a binary loaded
- **Java 21** — required by Ghidra
- **Python 3.10+** — the MCP bridge runtime
- **ghidra-mcp** — the bridge itself

### Install

The recommended way is via [REPlugins](https://github.com/DohmBoy64Bit/REPlugins),
a pre-built extension pack for Ghidra 12.1.2:

1. Clone: `git clone https://github.com/DohmBoy64Bit/REPlugins`
2. In Ghidra Project Window: **File → Install Extensions → Add** (+)
3. Install `GhidraMCP/GhidraMCP-5.13.1.zip`
4. Install Python deps: `pip install -r GhidraMCP/requirements.txt`
5. Restart Ghidra

Alternatively, install from source:
```
pip install ghidra-mcp
```
(from [bethington/ghidra-mcp](https://github.com/bethington/ghidra-mcp))

### Enable

1. Ensure `ghidra-mcp` is on PATH (or set `RECOMPHAMR_MCP_GHIDRA_COMMAND`)
2. Start recomphamr — auto-connects at launch
3. Run `/skill ghidra-mcp` — unlocks the skill text + `ghidra.*` tools
4. Verify: `/mcp tools ghidra` — shows 20 enabled tools by default
5. For all tools: set `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` or `/mcp enable ghidra *`

---

## N64 Debug MCP

Connects to a running Mupen64Plus emulator for runtime debugging. The LLM can
read memory, set breakpoints, trace execution, capture frames, decode display
lists, and inspect RSP state.

### Dependencies

- **Mupen64Plus** — N64 emulator with debugger plugin
- **Python 3.10+** — the MCP daemon runtime
- **n64-debug-mcp** — the bridge from [DohmBoy64Bit/Mupen64MCP](https://github.com/DohmBoy64Bit/Mupen64MCP)
- **A ROM file** — loaded into the emulator

### Install

1. Clone: `git clone https://github.com/DohmBoy64Bit/Mupen64MCP`
2. Install Python deps: follow the project's `README.md` for dependencies
3. Build/install: follow the project's build instructions

### Enable

1. Start Mupen64Plus with a ROM loaded
2. Start the n64-debug-daemon (see Mupen64MCP docs for startup command)
3. Ensure `n64-debug-mcp` is on PATH (or set `RECOMPHAMR_MCP_N64_COMMAND`)
4. Start recomphamr — auto-connects at launch
5. Run `/skill n64-debug-mcp` — unlocks the skill text + `n64-debug-mcp.*` tools
6. Verify: `/mcp tools n64-debug-mcp` — all tools available by default

---

## PCRECOMP MCP

Wraps PCRECOMP-Next PC static recompilation tools. 8 of 30+ tools enabled
by default (PE analysis, disassembly, lifting, classification, Ghidra batch).

### Dependencies

- [PCRECOMP-Next](https://github.com/DohmBoy64Bit/PCRECOMP-Next) — cloned locally
- **Python 3.10+** — with `pip install capstone pefile`
- **Ghidra 11.0+** — optional, for headless decompilation tools
- **CMake + C compiler** — for building lifted code

### Install

1. Clone: `git clone https://github.com/DohmBoy64Bit/PCRECOMP-Next`
2. Install Python deps: `pip install capstone pefile`
3. Set `RECOMPHAMR_PCRECOMP_PATH` to the clone directory
4. Ensure `pcrecomp-mcp` is on PATH (or set `RECOMPHAMR_MCP_PCRECOMP_COMMAND`)

### Enable

1. Start recomphamr — auto-connects at launch
2. Run `/skill pcrecomp` — unlocks skill text + 8 default `pcrecomp.*` tools
3. Verify: `/mcp tools pcrecomp` — shows 8 enabled tools
4. Enable all: `RECOMPHAMR_MCP_PCRECOMP_TOOLS=*` or `/mcp enable pcrecomp *`

### Default tools (8 of 30+)

| Tool | Phase |
|---|---|
| `pcrecomp.pe.analyze` | Binary identity, sections, compiler hints |
| `pcrecomp.pe.extract_imports` | DLL imports |
| `pcrecomp.disasm32.run` | Recursive descent disassembly |
| `pcrecomp.disasm32.callgraph` | Who-calls-who graph |
| `pcrecomp.lift32.run` | x86-32 to readable C |
| `pcrecomp.classify.run` | SDK vs custom classification |
| `pcrecomp.ghidra.decompile_all` | Batch Ghidra headless decompile |
| `pcrecomp.ghidra.function_stats` | Function statistics and counts |

Disabled by default: 16-bit DOS, 16-bit NE/Win16, DRM, asset extractors,
Ghidra extras. Enable groups with `/mcp enable pcrecomp <tool>` or `*` for all.

---

## Common

| Env var | Purpose |
|---|---|
| `RECOMPHAMR_MCP_AUTOSTART` | Set to `0` to skip auto-connect on startup |
| `RECOMPHAMR_MCP_GHIDRA_COMMAND` | Override ghidra MCP server command/path |
| `RECOMPHAMR_MCP_N64_COMMAND` | Override n64-debug-mcp server command/path |
| `RECOMPHAMR_MCP_PCRECOMP_COMMAND` | Override pcrecomp MCP server command/path |
| `RECOMPHAMR_MCP_GHIDRA_TOOLS` | Comma-separated tool list or `*` for all |
| `RECOMPHAMR_MCP_PCRECOMP_TOOLS` | Comma-separated tool list or `*` for all |

Servers appear on the startup splash — `* Connected` with tool counts,
or `  Disconnected` if not running.

Connect/disconnect manually with `/mcp connect <name>` /
`/mcp disconnect <name>`. Detailed architecture in [docs/mcp.md](mcp.md).
