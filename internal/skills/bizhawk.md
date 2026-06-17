# bizhawk

Multi-system emulator debug bridge via dmang-dev/mcp-bizhawk — memory r/w,
button input, frame advance, save states, and screenshots across NES, SNES,
GB/GBC/GBA, Genesis, N64, PS1, and 12+ more systems through one Lua bridge.

## What it enables

- Read/write memory by domain at u8/u16/u32 (WRAM, VRAM, RDRAM, etc.)
- Press/hold buttons by name per system (A, B, Start, Up, etc.)
- Frame advance, pause/unpause, reset, save/load state
- Screenshot capture, ROM info (hash, framecount, memory domains)
- Memory range read/write up to 4096 bytes per call
- All via one bridge — no per-system MCP setup

## When to use

Use this tool for:
- Cross-system emulator debugging — memory hunting, button scripting
- A/B comparison between recompiled output and reference emulator behavior
- Frame-precise input testing — verify game state at specific frame counts
- Runtime memory inspection at known addresses (WRAM, VRAM, RDRAM)
- Save-state checkpointing for reproducible test scenarios

Do not use it for:
- Static analysis without runtime context (use `ghidra-mcp`)
- Host-side native EXE debugging (use `cdb-debug`)
- Platform-specific debugging when a dedicated MCP bridge exists (use
  `n64-debug-mcp` for N64, `pcsx2` for PS2, `mcp-pine` for PS3)

## Boot / Connection Check

1. Verify BizHawk MCP server is running:
   - `/mcp tools bizhawk` — lists all available tools (16 tools)
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Verify bizhawk bridge connectivity:
   - `bizhawk.bizhawk_ping` — returns `pong` if bridge is healthy
3. Verify the correct ROM is loaded:
   - `bizhawk.bizhawk_get_info` — returns ROM name, hash, framecount, memory domains
   - Confirm framecount > 0 (emulator has initialized RAM)
4. List available memory domains for the loaded core:
   - `bizhawk.bizhawk_list_memory_domains` — names are case-sensitive
5. If unavailable:
   - Check `/mcp status bizhawk` for connection state
   - Verify `mcp-bizhawk` is on PATH or `RECOMPHAMR_MCP_BIZHAWK_COMMAND` is set
   - Ensure bridge.lua is loaded and polling (Lua console shows `frame loop active`)
   - Ensure BizHawk was launched with `--socket_ip=127.0.0.1 --socket_port=8766`

## Setup

1. Install: `npm install -g mcp-bizhawk`
2. Launch BizHawk: `EmuHawk.exe --socket_ip=127.0.0.1 --socket_port=8766 game.rom`
3. In BizHawk: Tools → Lua Console → Open Script → `lua/bridge.lua`
4. Verify: Lua console shows `frame loop active`
5. Ensure `mcp-bizhawk` is on PATH (or set `RECOMPHAMR_MCP_BIZHAWK_COMMAND`)
6. Start recomphamr — connect with `/mcp connect bizhawk`
7. Load `/skill bizhawk` — unlocks `bizhawk.*` tools (16 tools)
8. Verify: `/mcp tools bizhawk`

## Evidence Protocol

Every capture should record:
- target system, ROM name and hash (from `bizhawk_get_info`)
- memory domain used, address range, timestamp/framecount
- tool call used and output
- interpretation status: CONFIRMED / HYPOTHESIS / TODO / BLOCKED

Save evidence to:
- `.rehamr/evidence/` — memory dumps, register comparisons, trace summaries
- `.rehamr/traces/` — frame captures, button sequences
- `REPHAMR_STATE.md` — active comparisons, verified addresses, runtime observations

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Verify connectivity | `bizhawk.bizhawk_ping` | `pong` | Run first after connect |
| Get ROM info | `bizhawk.bizhawk_get_info` | ROM name, hash, framecount, domains | Run once per session |
| List memory domains | `bizhawk.bizhawk_list_memory_domains` | Domain name list | Names are case-sensitive |
| Read byte | `bizhawk.bizhawk_read8` | u8 value | Specify domain + address |
| Read word | `bizhawk.bizhawk_read16` | u16 LE value | For 16-bit values |
| Read dword | `bizhawk.bizhawk_read32` | u32 LE value | For 32-bit values |
| Read range | `bizhawk.bizhawk_read_range` | Up to 4096 bytes | Specify domain + start + size |
| Write byte | `bizhawk.bizhawk_write8` | Confirmation | Specify domain + address + value |
| Press buttons | `bizhawk.bizhawk_press_buttons` | Joypad state set | Key=button name, value=bool |
| Frame advance | `bizhawk.bizhawk_frame_advance` | N frames stepped | Step by N frames |
| Pause | `bizhawk.bizhawk_pause` | Emulation paused | Required before reading |
| Screenshot | `bizhawk.bizhawk_screenshot` | PNG saved to path | Visual evidence |
| Save state | `bizhawk.bizhawk_save_state` | State saved to file | Checkpoint before risky ops |
| Load state | `bizhawk.bizhawk_load_state` | State restored | Restore checkpoint |

## Memory Domains by System

Names come from BizHawk's core implementation. Use `bizhawk_list_memory_domains`
to see the exact set for the loaded ROM.

| System | Main RAM domain | Other common domains |
|---|---|---|
| NES | `RAM` | `PPU`, `OAM`, `PRG ROM`, `CHR` |
| SNES | `WRAM` | `VRAM`, `CARTROM`, `CARTRAM` |
| GB/GBC | `WRAM` | `VRAM`, `HRAM`, `OAM`, `ROM` |
| GBA | `EWRAM`, `IWRAM` | `VRAM`, `PALRAM`, `OAM`, `ROM` |
| Genesis | `68K RAM` | `VRAM`, `Z80 RAM`, `CARTRAM` |
| N64 | `RDRAM` | `SP DMEM`, `SP IMEM`, `PI Reg` |
| PSX | `MainRAM` | `VRAM`, `Scratchpad`, `BIOS` |

## Guardrails

1. Runtime state is evidence, but interpretation needs static confirmation
   against ROM data and known hardware behavior.
2. Screenshots/logs/traces do not prove root cause alone — pair with
   register/memory evidence and static analysis.
3. Do not mutate emulator/game state unless the task calls for it.
4. Prefer save states before writes or risky button sequences.
5. Memory reads may return zeros for the first few frames after boot —
   verify framecount > 0.
6. Domain names are case-sensitive — mismatch returns `unknown memory domain`.
7. Some cores expose slightly different surface — check `bizhawk_get_info`
   capabilities map.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `mcp-bizhawk` not on PATH | Server binary missing | `npm install -g mcp-bizhawk` |
| Server Disconnected | Bridge not polling or BizHawk not launched | Load bridge.lua; verify `--socket_ip/port` flags |
| `unreachable` | BizHawk not running | Launch BizHawk with socket flags; verify process |
| `unknown memory domain` | Domain name typo or case mismatch | Run `bizhawk_list_memory_domains` for exact names |
| Memory returns zeros | RAM not initialized yet | Advance frames until framecount > 0 |
| Screenshot/savestate not available | Core doesn't expose feature | Check `bizhawk_get_info` capabilities map |

## Session Close

1. Save all runtime captures to `.rehamr/evidence/` or `.rehamr/traces/`.
2. Update `REPHAMR_STATE.md` with active system, ROM info, verified addresses.
3. Report: memory domains explored, addresses verified, button sequences tested, screenshots saved, remaining states to investigate.
