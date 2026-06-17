# bizhawk

Multi-system emulator debug bridge via dmang-dev/mcp-bizhawk. 16 tools for
memory r/w, button input, frame advance, and save states across 12+ systems.


## Kickoff

`/skill bizhawk` — then "Connect to BizHawk and show me the memory domains for the loaded ROM."

## What it enables

- Read/write memory by domain (WRAM, VRAM, RDRAM, etc.) at u8/u16/u32
- Press/hold buttons by name per system (A, B, Start, Up, etc.)
- Frame advance, pause, reset, save/load state, screenshot
- One bridge for NES, SNES, GB/GBC, GBA, Genesis, N64, PS1, and more

## Setup

1. `npm install -g mcp-bizhawk`
2. BizHawk: `EmuHawk.exe --socket_ip=127.0.0.1 --socket_port=8766 game.rom`
3. Tools → Lua Console → Open Script → `bridge.lua`
4. Ensure `mcp-bizhawk` on PATH (or set `RECOMPHAMR_MCP_BIZHAWK_COMMAND`)

## When to use

Cross-system emulator debugging — memory hunting, button scripting,
frame-precise input testing, A/B comparison.
