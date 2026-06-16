# N64 Debug MCP Setup

Connects to a running Mupen64Plus emulator for runtime debugging. The LLM can
read memory, set breakpoints, trace execution, capture frames, decode display
lists, and inspect RSP state.

## Dependencies

- **Mupen64Plus** — N64 emulator with debugger plugin
- **Python 3.10+** — the MCP daemon runtime
- **n64-debug-mcp** — the bridge from [DohmBoy64Bit/Mupen64MCP](https://github.com/DohmBoy64Bit/Mupen64MCP)
- **A ROM file** — loaded into the emulator

## Install

1. Clone: `git clone https://github.com/DohmBoy64Bit/Mupen64MCP`
2. Install Python deps: follow the project's `README.md` for dependencies
3. Build/install: follow the project's build instructions

## Enable

1. Start Mupen64Plus with a ROM loaded
2. Start the n64-debug-daemon (see Mupen64MCP docs for startup command)
3. Ensure `n64-debug-mcp` is on PATH (or set `RECOMPHAMR_MCP_N64_COMMAND`)
4. Start recomphamr — auto-connects at launch
5. Run `/skill n64-debug-mcp` — unlocks the skill text + `n64-debug-mcp.*` tools
6. Verify: `/mcp tools n64-debug-mcp` — all tools available by default

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
