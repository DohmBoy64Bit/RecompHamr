# mcp-pine

RPCS3 emulator debug bridge via PINE IPC — A/B comparison, dynamic memory
probing, savestate management, and register inspection for PS3 static
recompilation projects.

> The PINE IPC interface must be enabled in RPCS3 settings. The mcp-pine
> server connects over localhost — no remote exposure.

## What it enables

- Read/write PS3 guest memory at runtime addresses
- Save and load savestates for reproducible A/B comparison
- Read PS3 guest registers (PPU/SPU state)
- Dynamic probing of unresolved addresses and NIDs
- Compare recompiled native behavior against original RPCS3 execution
- Step execution, set breakpoints for targeted comparison

## When to use

Use this tool for:
- PS3 static recompilation A/B comparison — verifying recompiled output
  matches original PPU/SPU behavior at specific guest addresses
- Runtime register inspection to diagnose translation mismatches
- Checkpoint-based debug workflows — save state before risky operations
- Used by `ps3recomp` during Phase 4 (runtime debugging)

Do not use it for:
- Static analysis without runtime context (use `ghidra-mcp`)
- Host-side native EXE debugging (use `cdb-debug`)
- Non-PS3 platforms (use platform-specific bridge: `pcsx2` for PS2,
  `n64-debug-mcp` for N64)
- When RPCS3 is not running or PINE IPC is disabled

## Boot / Connection Check

1. Verify mcp-pine server is running:
   - `/mcp tools mcp-pine` — lists all available tools
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Verify RPCS3 is running with PINE IPC enabled:
   - Confirm PINE IPC is enabled: Settings → Advanced → Enable PINE IPC
   - RPCS3 must be running with the target game loaded
3. Verify connection is local-only:
   - mcp-pine connects over localhost — never expose remotely
4. If unavailable:
   - Check `/mcp status mcp-pine` for connection state
   - Verify `mcp-pine` is on PATH or `RECOMPHAMR_MCP_PINE_COMMAND` is set
   - Enable PINE IPC in RPCS3 settings and restart

## Setup

1. Enable PINE IPC in RPCS3: Settings → Advanced → Enable PINE IPC
2. Install mcp-pine: `pip install mcp-pine` or from source
3. Start RPCS3 with the target game loaded
4. Ensure `mcp-pine` is on PATH (or set `RECOMPHAMR_MCP_PINE_COMMAND`)
5. Start recomphamr — connect with `/mcp connect mcp-pine`
6. Load `/skill mcp-pine` — unlocks `mcp-pine.*` tools
7. Verify: `/mcp tools mcp-pine`

## Evidence Protocol

Every runtime comparison should record:
- target game, TITLE ID, running state
- guest address being compared, register values (recomp vs RPCS3)
- savestate slot used, timestamp
- interpretation: CONFIRMED (match) / DIVERGENCE (mismatch) / HYPOTHESIS (unclear)

Save evidence to:
- `.rehamr/evidence/ps3_comparison_log.md` — A/B comparison results
- `.rehamr/evidence/` — register snapshots, memory dumps
- `REPHAMR_STATE.md` — active comparisons, verified addresses, divergence tracking

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Read memory | `mcp-pine.mcp_pine_read_memory` | Hex dump at address | Specify address + size |
| Write memory | `mcp-pine.mcp_pine_write_memory` | Confirmation | Hex data to write |
| Read registers | `mcp-pine.mcp_pine_read_registers` | Register state | PPU/SPU register values |
| Save state | `mcp-pine.mcp_pine_savestate_save` | State saved to slot | Checkpoint before risky ops |
| Load state | `mcp-pine.mcp_pine_load_state` | State restored from slot | Restore checkpoint for re-test |
| Step execution | `mcp-pine.mcp_pine_step` | Advanced one instruction | Single-step for comparison |
| Set breakpoint | `mcp-pine.mcp_pine_set_breakpoint` | BP set at address | Target specific guest addresses |
| Continue | `mcp-pine.mcp_pine_continue` | Execution resumed | Run until next breakpoint |
| Pause | `mcp-pine.mcp_pine_pause` | Emulation paused | Required before reading state |

## Guardrails

1. Runtime state is evidence, but interpretation needs static confirmation
   against Ghidra PPU/SPU analysis.
2. Always save state before writes or register modifications — irreversible
   changes can make comparison invalid.
3. Pause emulation before reading registers to avoid inconsistent state.
4. RPCS3 behavior is reference, not ground truth — the emulator itself may
   have inaccuracies. Cross-validate with static analysis.
5. Do not expose mcp-pine beyond localhost — PINE IPC is designed for local
   debugging only.
6. Not required for initial recomp triage — prefer static analysis first,
   use A/B comparison when static evidence is insufficient.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `mcp-pine` not on PATH | Server binary missing | `pip install mcp-pine`; set `RECOMPHAMR_MCP_PINE_COMMAND` |
| Server Disconnected | RPCS3 not running or PINE IPC disabled | Start RPCS3; enable PINE IPC in settings; restart |
| `read_registers` returns zero | Emulator not in running state | Pause before reading; verify game is loaded |
| Savestate fails | Slot invalid or RPCS3 state issue | Try different slot; verify RPCS3 supports savestates for this game |
| Connection refused | PINE IPC not listening | Verify PINE IPC enabled; check RPCS3 console for errors |

## Session Close

1. Save all comparison results to `.rehamr/evidence/ps3_comparison_log.md`.
2. Update `REPHAMR_STATE.md` with verified addresses, divergence points, active comparisons.
3. Report: addresses compared, matches confirmed, divergences found, remaining addresses to investigate.
