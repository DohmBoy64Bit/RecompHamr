# PCRECOMP MCP Setup

Wraps PCRECOMP-Next PC static recompilation tools — PE analysis, disassembly,
code lifting, and function classification. 8 of 30+ tools enabled by default.

## Dependencies

- [PCRECOMP-Next](https://github.com/DohmBoy64Bit/PCRECOMP-Next) — cloned locally
- **Python 3.10+** — with `pip install capstone pefile`
- **Ghidra 11.0+** — optional, for headless decompilation tools
- **CMake + C compiler** — for building lifted code

## Install

1. Clone: `git clone https://github.com/DohmBoy64Bit/PCRECOMP-Next`
2. Install Python deps: `pip install capstone pefile`
3. Set `RECOMPHAMR_PCRECOMP_PATH` to the clone directory
4. Ensure `pcrecomp-mcp` is on PATH (or set `RECOMPHAMR_MCP_PCRECOMP_COMMAND`)

## Enable

1. Start recomphamr — auto-connects at launch
2. Run `/skill pcrecomp` — unlocks skill text + 8 default `pcrecomp.*` tools
3. Verify: `/mcp tools pcrecomp` — shows 8 enabled tools
4. Enable all: `RECOMPHAMR_MCP_PCRECOMP_TOOLS=*` or `/mcp enable pcrecomp *`

## Default tools (8 of 30+)

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

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
