# n64-debug-mcp

Use this skill when N64 runtime debugging is needed via Mupen64Plus MCP —
reading registers and memory, setting breakpoints, tracing execution,
capturing frames, decoding display lists, and inspecting RSP state.

> Runtime state is evidence, but interpretation needs static confirmation
> against ROM data and known hardware behavior. Emulator output is not
> hardware truth — validate against static analysis.

## What it enables

- Read/write N64 RDRAM at runtime addresses via `n64-debug-mcp.n64_read_memory`
- Set execution breakpoints with `n64-debug-mcp.n64_add_breakpoint`
- Read all 32 GPRs + PC with `n64-debug-mcp.n64_get_registers`
- Decode GBI display lists with `n64-debug-mcp.n64_decode_display_list`
- Detect OS type, boot flow, and thread functions via `n64-debug-mcp.n64_detect_os`
- Capture PI DMA transfers with `n64-debug-mcp.n64_capture_pi_dma`
- Check RSP health and microcode state via `n64-debug-mcp.n64_check_rsp_health`
- Step frame-by-frame with `n64-debug-mcp.n64_step_frame`
- Mark game state labels with `n64-debug-mcp.n64_mark_game_state`

## When to use

Use this tool for:
- Guest runtime evidence when static analysis (Ghidra) is insufficient
- Breakpoint-driven crash analysis — capture registers at crash address
- Verifying overlay dispatch and `jalr` target resolution at runtime
- Frame-by-frame comparison of VI output via frame captures
- Display list decoding for GPU command verification
- Proving hit/miss at specific guest addresses (complements CDB on host side)

Do not use it for:
- Static analysis without runtime context (use `ghidra-mcp`)
- Host-side native EXE debugging (use `cdb-debug`)
- General RE methodology (use `core-re`)
- When the Mupen64MCP daemon is not running — provide setup steps, don't pretend

## Boot / Connection Check

1. Verify Mupen64MCP daemon is running:
   - `/mcp tools n64-debug-mcp` — lists all available tools
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Verify a ROM is loaded and emulation is active:
   - `n64-debug-mcp.n64_get_status` — returns `rom_loaded: true`, `running: true`
   - `n64-debug-mcp.n64_get_pc` — returns a valid PC (non-zero)
3. Verify the running core matches the target:
   - `n64-debug-mcp.n64_detect_os` — confirms OS type and ROM header
4. If unavailable:
   - Check `/mcp status n64-debug-mcp` for connection state
   - Verify `n64-debug-mcp` is on PATH or `RECOMPHAMR_MCP_N64_COMMAND` is set
   - Provide exact setup steps from `docs/mcp-n64-debug-mcp.md`
   - Ensure Mupen64Plus is launched with a ROM and the debugger plugin active

## Setup

1. Clone/build Mupen64MCP: `git clone https://github.com/DohmBoy64Bit/Mupen64MCP`
2. Follow the project's build instructions (MSYS2 MINGW64 on Windows)
3. Start Mupen64Plus with a ROM loaded and debugger plugin enabled
4. Start the n64-debug-daemon
5. Ensure `n64-debug-mcp` is on PATH (or set `RECOMPHAMR_MCP_N64_COMMAND`)
6. Start recomphamr — connect with `/mcp connect n64-debug-mcp`
7. Load `/skill n64-debug-mcp` — unlocks all `n64-debug-mcp.*` tools
8. Verify: `/mcp tools n64-debug-mcp`

## Evidence Protocol

Every runtime capture should be classified:
- Register dump at breakpoint → CONFIRMED (guest truth at that moment)
- Memory read at known address → CONFIRMED (verified bytes)
- Display list decode → CONFIRMED (GPU command stream)
- Trace events → CONFIRMED (execution path evidence)
- Interpretation of runtime state → HYPOTHESIS (needs static confirmation)

Save evidence to:
- `.rehamr/evidence/` — register dumps, memory snapshots, trace summaries
- `.rehamr/traces/` — raw trace events, frame captures
- `REPHAMR_STATE.md` — active breakpoints, last PC, runtime observations

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Read registers | `n64-debug-mcp.n64_get_registers` | All 32 GPRs + PC | Pause recommended before reading |
| Read memory | `n64-debug-mcp.n64_read_memory` | Hex dump | Specify address + size |
| Set breakpoint | `n64-debug-mcp.n64_add_breakpoint` | BP index | Execution breakpoint at virtual address |
| Wait for breakpoint | `n64-debug-mcp.n64_wait_for_breakpoint` | Status on hit | Blocks until BP fires or timeout |
| Step instruction | `n64-debug-mcp.n64_step_instruction` | New PC | Single MIPS instruction |
| Step frame | `n64-debug-mcp.n64_step_frame` | New frame count | Execute until next VBlank |
| Decode display list | `n64-debug-mcp.n64_decode_display_list` | GBI commands | Verify GPU command stream |
| Capture PI DMA | `n64-debug-mcp.n64_capture_pi_dma` | DMA registers | ROM→RDRAM transfer info |
| Detect OS | `n64-debug-mcp.n64_detect_os` | OS type + function addresses | Run once per session |
| Check RSP health | `n64-debug-mcp.n64_check_rsp_health` | RSP status, ucode hash | Diagnose RSP-HLE conflicts |
| Read framebuffer | `n64-debug-mcp.n64_read_framebuffer` | Pixel data | Compare VI output |
| Mark game state | `n64-debug-mcp.n64_mark_game_state` | Label | Tag trace events with context |

## Guardrails

1. Runtime state is evidence, but interpretation needs static confirmation
   against ROM data and known hardware behavior.
2. Emulator output is not hardware truth — always validate against Ghidra
   static analysis before claiming a fix.
3. Pause emulation before reading registers to avoid inconsistent state
   (`n64-debug-mcp.n64_pause` or breakpoint hit).
4. Capture before modifying — take register/memory snapshots before changing
   breakpoints or stepping, to preserve evidence.
5. Frame captures are large — use sparingly and only when visual evidence is
   required. Prefer targeted register/memory reads.
6. Not required for matching decomp or initial recomp triage — prefer guest
   evidence over static guesses when available, but static analysis alone is
   sufficient for many tasks.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `n64-debug-mcp` not on PATH | Daemon binary missing | Build/clone Mupen64MCP; set `RECOMPHAMR_MCP_N64_COMMAND` |
| Server Disconnected | Daemon not running or emulator not started | Start Mupen64Plus with ROM loaded; start daemon |
| `rom_loaded: false` | No ROM loaded in emulator | Load ROM via Mupen64Plus; verify `n64_get_status` |
| Breakpoint never hit | Address not reached at runtime | Verify address is in executed code path via `n64_detect_os` output |
| Register read returns zero | Emulator not in running state | `n64_pause` before reading; verify `running: true` |
| Display list decode fails | Invalid address or non-GBI data | Verify address points to display list in RDRAM |

## Session Close

1. Save all runtime captures to `.rehamr/evidence/` or `.rehamr/traces/`.
2. Update `REPHAMR_STATE.md` with active breakpoints, last PC, runtime observations.
3. Report: breakpoints hit, register snapshots captured, display lists decoded, frame captures saved, remaining addresses to investigate at runtime.
