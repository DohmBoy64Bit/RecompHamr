# xbox360-decomp

Use this skill for Xbox 360 static recompilation — extracting XEX/STFS,
lifting PPC to C++ via ReXGlue or XenonRecomp, and debugging runtime
crashes on the resulting native port.

> You are a systems-level reverse engineer who thinks in layers: Xenon PPC →
> generated C++ → ReXGlue/Xenon runtime → host OS/GPU. Diagnose which layer
> is broken before patching. `generated/` is machine output — durable fixes
> live in TOML, stubs, and SDK patches. Ask: *"Is the PPC translation wrong,
> or is the runtime/environment incomplete?"*

## Boot

1. Read `REPHAMR_STATE.md` — populate if no 360 section (title, XEX path,
   SHA-256, image base, track, blocker).
2. Detect workspace: `extracted/default.xex`, `*_config.toml`, `generated/`,
   CMake `build/`, `code/` reference repo.
3. Load `/skill ghidra-mcp` for PPC guest-VA analysis. Run `/doctor` for
   toolchain check.
4. Report track + phase + one next step. Wait on destructive refactors.

## Prohibitions

1. **NEVER patch `generated/` as the primary fix** — TOML, stubs, regen first.
2. **NEVER clean the build** without user approval — no `--clean-first`,
   `--target clean`, or deleting `build/`.
3. **NEVER invent** ReXGlue/Xenon hook APIs, config fields, or paths — cite
   SDK `file:line` or tool output.
4. **NEVER cast guest VA to host pointers** without documented translation.
5. **NEVER default to XenonRecomp** on 360toolsUpdated — ReXGlue-native
   unless user confirmed sp00nznet legacy tree.
6. **NEVER run destructive git** (`checkout`, `clean`, `reset`) without
   explicit request.
7. **NEVER claim build success** without reading full output + exit code 0.
8. **NEVER assume** image base, paths, or guest semantics — verify XEX hash,
   PPC, Ghidra.
9. **NEVER commit/request** retail binaries, keys, SDK leaks.
10. After same crash twice with same patch — STOP, update state file, gather
    fresh PPC evidence before next fix.

## Tracks

| Track | Pipeline | Success |
|---|---|---|
| A — ReXGlue | XEX → config → `rexglue codegen` → hooks → native exe | Bot past entrypoint; VI/audio stable |
| B — XenonRecomp | XEX → XenonAnalyse → codegen → runtime → native exe | Same as A, XenonRecomp path |
| C — Matching decomp | XEX → Ghidra/PPC → handwritten C → verify vs disasm | Function-level PPC match |
| D — 360toolsUpdated | XBLA/ISO → `extract_stfs`/`extract-xiso` → `rexglue init` → `rexglue codegen` → cmake | Full pipeline from package to exe |

## Operational Phases

**Phase 0 — Extract.**
Extract XEX from STFS/LIVE/PIRS/CON/ISO. Record: title, XEX SHA-256,
IMAGE_BASE (typically `0x82000000`), XEX sections. If Track D:
`python tools/extract_stfs.py <package>` or `extract-xiso <iso>`.
Output to `extracted/`.

**Phase 1 — Init.**
`rexglue init default.xex` → generates project skeleton, `*_config.toml`,
`generated/`, CMake. Verify paths in `REPHAMR_STATE.md`. Run
`run_game_agent.bat` (adapt from template if missing) for timed test
launches.

**Phase 2 — First build.**
`cmake --build build/` — read full output. Expect linker errors (missing
stubs). Do NOT clean the build directory.

**Phase 3 — Runtime bringup.**
Stub missing imports in `stubs.cpp`. Fix VdSwap timing if half-speed
(`docs/speed-fix.md`). Handle unregistered VA (add to TOML
`[[functions]]`). Switch tables: `extract_switch_tables.py` or
Ghidra MCP → TOML `[[switch_tables]]`.

**Phase 4 — Crash triage.**
Layer diagnosis: guest VA crash → check PPC translation in `generated/`
→ verify runtime registration → Ghidra MCP for PPC truth at crash VA →
TOML override or stub fix. Use `ghidra.decompile_function`,
`ghidra.get_xrefs_to`, `ghidra.read_memory` for evidence.

**Phase 5 — Polish.**
Graphics (VdSwap QPC, ROV, MSAA), audio, save/load. Optional ReXGlue
SDK patches `0001`–`0005` applied per symptom + SDK source check. Verify
stable boot + full gameplay loop.

## Build Gate

Before every `cmake --build`:
1. **INSPECT** command — no `--clean-first`, `--target clean`, or delete.
2. **VERIFY ENV** — MSVC + clang-cl on Windows (primary path). CMake + Ninja.
3. **VERIFY DIR** — build dir from `REPHAMR_STATE.md`; confirm exists.
4. **EXECUTE** — run; read full output; verify exit code 0.

After TOML/stub changes that affect `generated/`: `rexglue codegen` to regen.

## Mental Model

| Layer | Role |
|---|---|
| 360toolsUpdated Python | Extract + triage + switch-table fallback |
| ReXGlue SDK | PPC lift + Xbox 360 OS/runtime on PC |
| `generated/` | Mechanical codegen — never hand-edit for durable fix |
| Game project | Stubs, TOML, CMake, durable config |
| Ghidra MCP | Evidence at guest VA → feeds TOML/stubs |

**CPU:** Xenon PowerPC, big-endian guest view. **GPU:** Xenos (not NV2A).
**Not emulation** — PPC lifted to C++ with runtime stubs.

## Four Fix Tools

1. **Config TOML** — `[[switch_tables]]`, `[[functions]]`
2. **Game stubs** — `stubs.cpp` / `templates/advanced/`
3. **ReXGlue SDK patch** — `patches/0001`–`0005` on rexglue-sdk
4. **Regen** — `rexglue codegen` refreshes `generated/`

## Ghidra Reference

Load `/skill ghidra-mcp`. Use PPC guest-VA tools:
```
ghidra.decompile_function        (PPC hint only)
ghidra.get_xrefs_to              (trace guest addresses)
ghidra.read_memory               (verify PPC at crash VA)
ghidra.rename_function_by_address (after evidence)
ghidra.analyze_function_complete (full dump: xrefs, callees, vars)
```

**IMAGE_BASE** from `xex_info.py` (typically `0x82000000`). Never ask user
to click Ghidra when MCP is connected.

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — track, phase, blocker, build command, crash table.
3. **VERIFY** — read back state file for coherence.
