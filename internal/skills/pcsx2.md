# pcsx2

PCSX2 emulator debug bridge. Connects to a running PCSX2 DebugServer for
A/B comparison, register inspection, MIPS disassembly, and breakpoint
debugging. Load this skill to unlock `pcsx2.*` tools.

## What it enables

- Connect to PCSX2 DebugServer (port 21512)
- Read/write EE registers (128-bit GPR, CP0, FPU, VU)
- Set breakpoints/watchpoints with conditions
- Read/write PS2 RAM, find hex patterns, diff memory
- Native MIPS disassembly at runtime addresses
- Call stack backtrace, thread list, module list
- Save/load states for checkpoint-based debugging

Requires PCSX2 with DebugServer built-in. Install from
[hkmodd/PCSX2-MCP](https://github.com/hkmodd/PCSX2-MCP) releases.

## Setup

1. Download PCSX2-MCP release from GitHub
2. Run `setup-mcp.bat` (requires Node.js ≥ 18)
3. Launch `pcsx2-qt.exe`, load a PS2 game
4. Ensure `pcsx2-mcp` is on PATH (or set `RECOMPHAMR_MCP_PCSX2_COMMAND`)
5. Start recomphamr — auto-connects at launch
6. Load `/skill pcsx2` — unlocks `pcsx2.*` tools

## When to use

PS2 static recompilation A/B comparison — verifying recompiled C++ output
matches original MIPS behavior at specific guest addresses. Used by the
`ps2recomp` skill during Phase 4 (A/B comparison).
