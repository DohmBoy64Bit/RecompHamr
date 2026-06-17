# pcsx2

PCSX2 emulator debug bridge via hkmodd/PCSX2-MCP — A/B comparison, register
inspection, MIPS disassembly, breakpoint debugging, memory diffing, and
savestate management for PS2 static recompilation projects.

> Requires PCSX2 with DebugServer built-in. Install from
> [hkmodd/PCSX2-MCP](https://github.com/hkmodd/PCSX2-MCP) releases.

## What it enables

- Connect to PCSX2 DebugServer on port 21512 (auto-detects DebugServer or PINE fallback)
- Read/write EE registers — 7 categories: GPR, CP0, FPU, VU0, VU1, full 128-bit
- Set breakpoints with optional conditions and descriptions
- Set watchpoints for read/write/access/onChange memory monitoring
- Read/write PS2 RAM, find hex patterns with `??` wildcards, diff memory regions
- Native MIPS disassembly at runtime addresses + expression evaluation
- Call stack backtrace, thread list, IOP module list
- Save/load states, step/step-over/continue execution control
- Game info (title/ID/version via PINE), memory snapshots

## When to use

Use this tool for:
- PS2 static recompilation A/B comparison — verifying recompiled output
  matches original MIPS behavior at specific guest addresses
- Runtime register inspection to diagnose translation mismatches
- Breakpoint-driven crash analysis — capture EE registers at crash address
- Function-level comparison: set breakpoint → step → compare registers
- Used by `ps2recomp` during Phase 4 (A/B comparison)

Do not use it for:
- Static analysis without runtime context (use `ghidra-mcp`)
- Host-side native EXE debugging (use `cdb-debug`)
- Non-PS2 platforms (use `mcp-pine` for PS3, `n64-debug-mcp` for N64)
- When PCSX2 DebugServer is not running — check console for `[DebugServer] Listening`

## Boot / Connection Check

1. Verify PCSX2 MCP server is running:
   - `/mcp tools pcsx2` — lists all available tools (30 tools)
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Connect to PCSX2 DebugServer:
   - `pcsx2.pcsx2_connect` — auto-detects DebugServer or PINE fallback
   - Verify: `pcsx2.pcsx2_status` returns connection type + emulator status
3. Verify the correct game is loaded:
   - `pcsx2.pcsx2_game_info` — returns game title/ID/version (PINE)
   - Or check DebugServer console for loaded game
4. Confirm DebugServer port:
   - PCSX2 console should show `[DebugServer] Listening on 127.0.0.1:21512`
5. If unavailable:
   - Check `/mcp status pcsx2` for connection state
   - Verify `pcsx2-mcp` is on PATH or `RECOMPHAMR_MCP_PCSX2_COMMAND` is set
   - Ensure PCSX2 was launched from the PCSX2-MCP release (includes DebugServer)

## Setup

1. Download the latest release from [hkmodd/PCSX2-MCP](https://github.com/hkmodd/PCSX2-MCP/releases)
2. Extract the zip, run `setup-mcp.bat` (requires Node.js ≥ 18)
3. Launch `pcsx2-qt.exe`, load a PS2 game
4. Verify: PCSX2 console shows `[DebugServer] Listening on 127.0.0.1:21512`
5. Ensure `pcsx2-mcp` is on PATH (or set `RECOMPHAMR_MCP_PCSX2_COMMAND`)
6. Start recomphamr — connect with `/mcp connect pcsx2`
7. Load `/skill pcsx2` — unlocks `pcsx2.*` tools (30 tools)
8. Verify: `/mcp tools pcsx2`

## Evidence Protocol

Every runtime comparison should record:
- target game, TITLE ID, connection type (DebugServer or PINE)
- guest address compared, register values (recomp vs PCSX2)
- breakpoint address and hit count
- interpretation: CONFIRMED (match) / DIVERGENCE (mismatch) / HYPOTHESIS (unclear)

Save evidence to:
- `.rehamr/evidence/ps2_comparison_log.md` — A/B comparison results
- `.rehamr/evidence/` — register snapshots, memory dumps, disasm output
- `REPHAMR_STATE.md` — active breakpoints, verified addresses, divergence tracking

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Connect | `pcsx2.pcsx2_connect` | Connection type | Auto-detects DebugServer or PINE |
| Status | `pcsx2.pcsx2_status` | Connection + emulator state | Run after connect |
| Game info | `pcsx2.pcsx2_game_info` | Title/ID/version | PINE only |
| Read registers | `pcsx2.pcsx2_read_registers` | 128-bit EE registers | Primary diagnostic tool |
| Write register | `pcsx2.pcsx2_write_register` | Confirmation | Category + index |
| Disassemble | `pcsx2.pcsx2_disassemble` | Native MIPS at address | Runtime, not just ELF |
| Read memory | `pcsx2.pcsx2_read_memory` | Hex dump | Specify address + size |
| Write memory | `pcsx2.pcsx2_write_memory` | Confirmation | Hex data to write |
| Find pattern | `pcsx2.pcsx2_find_pattern` | Match addresses | `??` wildcards supported |
| Memory diff | `pcsx2.pcsx2_memory_diff` | Changed bytes | Snapshot → call again → diff |
| Set breakpoint | `pcsx2.pcsx2_set_breakpoint` | BP set | Optional condition + description |
| List breakpoints | `pcsx2.pcsx2_list_breakpoints` | BP list with status | Verify before running |
| Set watchpoint | `pcsx2.pcsx2_set_watchpoint` | Watchpoint set | Read/write/access/onChange |
| Step | `pcsx2.pcsx2_step` | One instruction | Single MIPS instruction |
| Step over | `pcsx2.pcsx2_step_over` | Next instruction | Skips JAL/JALR calls |
| Continue | `pcsx2.pcsx2_continue` | Execution resumed | Run until next breakpoint |
| Pause | `pcsx2.pcsx2_pause` | Emulation halted | REQUIRED before reading registers |
| Backtrace | `pcsx2.pcsx2_get_backtrace` | Call stack | Entry points, PCs, stack pointers |
| Threads | `pcsx2.pcsx2_get_threads` | EE/IOP BIOS threads | OS-level context |
| Modules | `pcsx2.pcsx2_get_modules` | IOP module list | Loaded runtime modules |
| Save state | `pcsx2.pcsx2_save_state` | State saved to slot | Checkpoint before risky ops |
| Load state | `pcsx2.pcsx2_load_state` | State restored | Restore checkpoint |

## A/B Comparison Protocol

The standard workflow for PS2 recomp validation:

```
1. pcsx2.pcsx2_connect → verify DebugServer
2. pcsx2.pcsx2_pause → REQUIRED before reading
3. pcsx2.pcsx2_set_breakpoint at target address
4. pcsx2.pcsx2_continue → wait for hit (or pcsx2_wait_for_breakpoint)
5. pcsx2.pcsx2_read_registers → capture EE state
6. pcsx2.pcsx2_step → single instruction
7. Compare: PCSX2 register values vs recompiled binary state
8. Document: match or divergence in comparison log
```

## Guardrails

1. **Pause before reading registers** — reading while running returns inconsistent
   state. Always `pcsx2.pcsx2_pause` or wait for breakpoint hit.
2. Runtime state is evidence, but interpretation needs static confirmation
   against Ghidra MIPS analysis and ps2tek hardware reference.
3. PCSX2 behavior is reference, not ground truth — the emulator may have
   inaccuracies. Cross-validate with static analysis.
4. Save state before writes or register modifications — irreversible changes
   invalidate comparison.
5. Breakpoints don't trigger in BIOS/menu — game must be running in-game.
6. PINE tools fail if PINE IPC is not enabled — use DebugServer path as primary.
7. Not required for initial recomp triage — prefer static analysis first.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `pcsx2-mcp` not on PATH | Server binary missing | Download PCSX2-MCP release; run setup-mcp.bat |
| No DebugServer connection | PCSX2 not launched from MCP release | Launch pcsx2-qt.exe from extracted folder |
| DebugServer not listening | Port 21512 not open | Check PCSX2 console for `[DebugServer] Listening` |
| PINE tools fail | PINE IPC not enabled | Enable: Settings → Advanced → Enable PINE IPC |
| Breakpoint never hits | Game not running or wrong address | Verify game is past BIOS/menu; check address |
| `read_registers` returns zero | Emulator not paused | `pcsx2.pcsx2_pause` before reading |
| Watchpoint not triggering | Access pattern mismatch | Verify watch type (read/write/access) matches |

## Session Close

1. Save all comparison results to `.rehamr/evidence/ps2_comparison_log.md`.
2. Save register snapshots and disasm output to `.rehamr/evidence/`.
3. Update `REPHAMR_STATE.md` with active breakpoints, verified addresses, divergences.
4. Report: addresses compared, registers matched, divergences found, remaining addresses to investigate.
