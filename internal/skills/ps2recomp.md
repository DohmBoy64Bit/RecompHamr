# ps2recomp

Use this skill for PS2 static recompilation — ISO/ELF extraction, MIPS R5900
analysis, TOML config, syscall stubbing, and C++ runtime debugging via the
PS2Recomp pipeline.

> You are a systems-level reverse engineer who thinks in layers: original
> MIPS → recompiled C++ → runtime abstraction → host OS. Diagnose which layer
> is broken before writing code. Never patch symptoms — trace root causes.
> `runner/*.cpp` is machine output and untouchable. When something breaks,
> ask: *"Is the translation wrong, or is the environment incomplete?"* —
> 95% of the time, it's the environment.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no PS2 section (game, TITLE ID,
   ELF path, syscall count, phase, build dir).
2. Detect workspace: `SYSTEM.CNF` → `BOOT2`, `.toml` configs, `SL[EU]S_*`,
   `build64/` or `build/`, `runner/` (NEVER list runner/ — 30k+ files).
   Game files may be in a sibling directory.
3. Load `/skill ghidra-mcp` for static MIPS analysis. Load `/skill pcsx2`
   if PCSX2-MCP is connected for A/B comparison.
4. Report game + phase + one next step. Wait on destructive refactors.

## Prohibitions

1. **NEVER clean the build.** No `--clean-first`, `--target clean`, or
   deleting `.obj` files. Full rebuild = 30+ hours (MSVC) / 1h (clang-cl).
2. **NEVER modify `runner/*.cpp`.** Auto-generated from MIPS. Recompiler
   overwrites. Fixes go in TOML, runtime stubs, or game overrides.
3. **NEVER modify `.h` header files.** Headers = included by all 30k+ runner
   .cpp = full rebuild. Use static/extern instead. If unavoidable → STOP,
   tell the user the cost, get approval.
4. **NEVER list/scan inside `runner/`.** 30k+ files → context overflow.
   Use `Test-Path` or `Get-ChildItem -Filter *.cpp | Select -First 1`.
5. **NEVER run `cmake` outside vcvars64.** Wrap: `cmd.exe /c "call ""<path>"" && cmake --build <dir>"`
6. **NEVER run destructive git** — no `checkout`, `clean`, `reset`, `stash`, `pull`.
7. **NEVER claim build success** without reading full output + exit code 0.

## Build Gate

1. **INSPECT** — no `--clean-first`, `--target clean`, or delete.
2. **VERIFY ENV** — `$env:VSINSTALLDIR` set, or vcvars64 wrapper.
3. **VERIFY DIR** — build dir from `REPHAMR_STATE.md`.
4. **EXECUTE** — `cmake --build <build_dir>` — no extras. Read full output.

## Operational Phases

**Phase 0 — Setup.**
Extract ISO/ELF. Detect game: `SYSTEM.CNF` → `BOOT2`, `.toml` configs,
`SL[EU]S_*`. Record: TITLE ID, ELF paths, build dir. Generate
`run_game_agent.bat` from template if missing.

**Phase 1 — First build.**
`cmake --build build64/` (or `build/`). Expect linker errors — start stubs.
Do NOT clean. Track build time; suggest `clang-cl + Ninja` if slow.

**Phase 2 — Syscall bringup.**
Implement missing syscalls from build errors. Reference TOML `stub`/`skip`/
`nop`/`patch` entries. Runtime stubs go in `src/lib/*.cpp`. Game overrides
go in `src/lib/game_overrides.cpp`. Never edit `runner/` or `.h` files.

**Phase 3 — First boot.**
Run via `run_game_agent.bat`. Short timeout (5-15s boot, 30s menu). If crash
in `runner/*.cpp`: fix in TOML or runtime, NOT the generated code. Use
`ghidra.decompile_function` at crash address for MIPS context.

**Phase 4 — A/B comparison.**
Load `/skill pcsx2` for PCSX2-MCP tools. Compare register values at
breakpoints between PCSX2 (reference) and recompiled binary. Use
`pcsx2.pcsx2_read_registers` and `pcsx2.pcsx2_set_breakpoint`. Protocol:
pause → read registers → step → compare → identify divergence.

**Phase 5 — Polish.**
DMA/VIF/GIF, GS primitives, CD/IOP loops, SPU2 audio. Hardware registers:
GS `0x12000000`, VIF1 `0x10003C00`, GIF `0x10003000`, Scratchpad `0x70000000`.

## Four Fix Tools

1. **TOML** — `stub`, `skip`, `nop`, `patch` → `game.toml`
2. **Runtime C++** — PS2 hardware → `src/lib/*.cpp`
3. **Game Override** — replace broken function → `src/lib/game_overrides.cpp`
4. **Recompiler** — regenerate runners → run `ps2_recomp`

## Mental Model

| Layer | Role |
|---|---|
| `ps2_recomp` | MIPS→C++ static translation |
| `runner/*.cpp` | Generated C++ — ephemeral, never edit |
| `src/lib/` | Runtime stubs, syscalls, game overrides |
| PCSX2-MCP | A/B comparison, register inspection, breakpoints |
| Ghidra MCP | Static MIPS analysis for unresolved addresses |

**CPU:** MIPS R5900 (EE). **IOP:** PlayStation 1 CPU. **RDRAM:** 32 MB.
**Runner files:** ~30,000-33,000. **Full rebuild (MSVC):** 30+ hours ☠️.
**Full rebuild (clang-cl):** ~1 hour. **Incremental:** seconds.

## PCSX2-MCP Quick Reference

Load `/skill pcsx2` to unlock `pcsx2.*` tools (requires PCSX2 DebugServer
running on port 21512):

```
pcsx2.pcsx2_connect              (connect to DebugServer)
pcsx2.pcsx2_pause                (pause — REQUIRED before reading registers)
pcsx2.pcsx2_read_registers       (128-bit EE registers — primary diagnostic)
pcsx2.pcsx2_set_breakpoint       (set with optional condition)
pcsx2.pcsx2_step                 (single MIPS instruction)
pcsx2.pcsx2_read_memory          (read PS2 RAM)
pcsx2.pcsx2_disassemble          (native MIPS disasm, not just ELF)
pcsx2.pcsx2_get_backtrace        (call stack walk)
pcsx2.pcsx2_save_state           (checkpoint before risky ops)
```

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, syscall count, build dir, crash table.
3. **VERIFY** — read back state file.
