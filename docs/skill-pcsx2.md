# pcsx2

PCSX2 emulator debug bridge via hkmodd/PCSX2-MCP. 30 tools for register
inspection, MIPS disassembly, breakpoints, and A/B comparison.

## What it enables

- Connect to PCSX2 DebugServer (port 21512)
- Read/write EE registers, PS2 RAM, disassemble MIPS at runtime
- Set breakpoints/watchpoints with conditions
- Call stack backtrace, thread/module listing
- Save/load states, memory diffing, pattern search

## Setup

1. Download [PCSX2-MCP](https://github.com/hkmodd/PCSX2-MCP) release
2. Run `setup-mcp.bat` (requires Node.js ≥ 18)
3. Launch `pcsx2-qt.exe`, load PS2 game
4. Ensure `pcsx2-mcp` on PATH (or set `RECOMPHAMR_MCP_PCSX2_COMMAND`)

## When to use

PS2 static recompilation A/B comparison — verifying recompiled output
against original PCSX2 behavior. Used by ps2recomp Phase 4.
