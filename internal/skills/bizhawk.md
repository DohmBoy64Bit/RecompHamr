# bizhawk

Multi-system emulator debug bridge via dmang-dev/mcp-bizhawk. Connects to
BizHawk for memory r/w, button input, frame advance, save states, and
screenshots across NES, SNES, GB/GBC/GBA, Genesis, N64, PS1, and more.
Load this skill to unlock `bizhawk.*` tools.

## What it enables

- Read/write memory at u8/u16/u32 by domain (WRAM, VRAM, RDRAM, etc.)
- Press/ hold buttons by name per system (A, B, Start, Up, etc.)
- Frame advance, pause/unpause, reset, save/load state
- Screenshot capture, ROM info (hash, framecount, domains)
- All via one bridge — no per-system setup

Requires BizHawk 2.6.2+ running with the bridge.lua script loaded.

## Setup

1. Install: `npm install -g mcp-bizhawk`
2. Launch BizHawk: `EmuHawk.exe --socket_ip=127.0.0.1 --socket_port=8766 game.rom`
3. In BizHawk: Tools → Lua Console → Open Script → `lua/bridge.lua`
4. Verify: Lua console shows `frame loop active`
5. Ensure `mcp-bizhawk` is on PATH (or set `RECOMPHAMR_MCP_BIZHAWK_COMMAND`)
6. Start recomphamr — auto-connects at launch
7. Load `/skill bizhawk` — unlocks `bizhawk.*` tools

## When to use

Cross-system emulator debugging — memory hunting, button scripting, A/B
comparison, frame-precise input testing across 12+ systems.
