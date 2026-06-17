# Ghidra MCP Setup

Bridges Ghidra's analysis engine to the LLM, exposing ~100+ tools for
decompilation, cross-references, symbol management, and data analysis.

## Dependencies

- **Ghidra 12.1.2** — installed and a project open with a binary loaded
- **Java 21** — required by Ghidra
- **Python 3.10+** — the MCP bridge runtime
- **ghidra-mcp** — the bridge itself

## Install

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

## Enable

1. Ensure `ghidra-mcp` is on PATH (or set `RECOMPHAMR_MCP_GHIDRA_COMMAND`)
2. Start recomphamr — connect with `/mcp connect ghidra`
3. Run `/skill ghidra-mcp` — unlocks the skill text + `ghidra.*` tools
4. Verify: `/mcp tools ghidra` — shows 20 enabled tools by default
5. For all tools: set `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` or `/mcp enable ghidra *`

## Tool filtering

20 most-used RE tools enabled by default. Set `RECOMPHAMR_MCP_GHIDRA_TOOLS=*`
for all ~100+. See `--help` for the env var or use `/mcp enable ghidra <tool>`
at runtime. Refer to [common.md](mcp-common.md) for shared env vars and
management commands.
