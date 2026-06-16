# pcsx2 Setup

PCSX2 emulator debug bridge via hkmodd/PCSX2-MCP. Connects to a running
PCSX2 DebugServer for A/B comparison, register inspection, MIPS disassembly,
and breakpoint debugging.

## Dependencies

- **PCSX2** with DebugServer — from PCSX2-MCP release
- **Node.js ≥ 18** — for the MCP server bridge
- **PS2 game** — loaded in the emulator

## Install

1. Download the latest release from [hkmodd/PCSX2-MCP](https://github.com/hkmodd/PCSX2-MCP/releases)
2. Extract the zip, run `setup-mcp.bat`
3. Launch `pcsx2-qt.exe`, load a PS2 game
4. Verify: PCSX2 console shows `[DebugServer] Listening on 127.0.0.1:21512`
5. Ensure `pcsx2-mcp` is on PATH (or set `RECOMPHAMR_MCP_PCSX2_COMMAND`)

## Enable

1. Start recomphamr — auto-connects at launch
2. Run `/skill pcsx2` — unlocks `pcsx2.*` tools (30 debugging tools)
3. Verify: `/mcp tools pcsx2`

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
