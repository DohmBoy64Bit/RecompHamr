# bizhawk Setup

Multi-system emulator debug bridge via dmang-dev/mcp-bizhawk. Connects to
BizHawk for memory r/w, button input, frame advance, and save states across
NES, SNES, GB/GBC/GBA, Genesis, N64, PS1, and 12+ more systems.

## Dependencies

- **BizHawk 2.6.2+** — installed, ROM loaded
- **Node.js 22+** — for the MCP server
- **mcp-bizhawk** — install via npm

## Install

1. `npm install -g mcp-bizhawk`
2. Launch BizHawk: `EmuHawk.exe --socket_ip=127.0.0.1 --socket_port=8766 game.rom`
3. Load bridge: Tools → Lua Console → Open Script → `lua/bridge.lua`
4. Verify: Lua console shows `frame loop active`
5. Ensure `mcp-bizhawk` is on PATH (or set `RECOMPHAMR_MCP_BIZHAWK_COMMAND`)

## Enable

1. Start recomphamr — auto-connects at launch
2. Load `/skill bizhawk` — unlocks `bizhawk.*` tools (16 tools)
3. Verify: `/mcp tools bizhawk`

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
