# xbox360-decomp

Use this skill for Xbox 360 static recompilation — XEX/STFS extraction, PPC to
C++ lifting via ReXGlue or XenonRecomp, runtime stubs, crash triage, and VdSwap
QPC/switch table/ROV/MSAA polish.

> You are a systems-level reverse engineer who thinks in layers: Xenon PPC →
> generated C++ → ReXGlue/Xenon runtime → host OS/GPU. Diagnose which layer
> is broken before patching. `generated/` is machine output — durable fixes
> live in TOML, stubs, and SDK patches. Ask: *"Is the PPC translation wrong,
> or is the runtime/environment incomplete?"*

## When to use

Use this skill when:
- Starting an Xbox 360 static recompilation project from XEX/STFS/ISO
- Running `rexglue init`/`codegen` or XenonRecomp pipeline steps
- Debugging runtime crashes, unregistered VA, or VdSwap half-speed issues
- Applying ReXGlue SDK patches 0001–0005 to fix known patterns

Do not use this skill when:
- The target is OG Xbox (use `xboxrecomp`) or PC/Win32 (use `pcrecomp`)
- You need XEX extraction only without pipeline (use `bash` with 360tools)
- The task is general RE methodology without platform-specific pipeline (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the 360 section with:
   - game title, XEX path, XEX SHA-256, IMAGE_BASE (typically `0x82000000`)
   - current track (A/B/C/D), current phase, active blocker
   - build dir path, verified CMake command
2. Detect workspace layout. Do not assume paths. Look for:
   - `extracted/default.xex`, `*_config.toml`, `generated/`
   - CMake `build/`, `code/` reference repo
   - `run_game_agent.bat` (adapt from template if missing)
3. Verify required tools:
   - 360toolsUpdated clone (`tools/extract_stfs.py`, `tools/extract-xiso`)
   - ReXGlue SDK or XenonRecomp (per track)
   - MSVC + clang-cl on Windows; CMake + Ninja
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill ghidra-mcp` — PPC guest-VA static analysis
   - `/skill core-re` — RE workflow discipline
5. Detect track from workspace evidence. Report: game + track + phase + one next
   step. Wait for go-ahead on destructive refactors.

## Prohibitions

1. **NEVER patch `generated/` as the primary fix** — TOML, stubs, regen first.
2. **NEVER clean the build** without user approval — no `--clean-first`,
   `--target clean`, or deleting `build/`.
3. **NEVER invent** ReXGlue/Xenon hook APIs, config fields, or paths — cite
   SDK `file:line` or verified tool output.
4. **NEVER cast guest VA to host pointers** without documented translation layer.
5. **NEVER default to XenonRecomp** on 360toolsUpdated — ReXGlue-native unless
   user confirmed sp00nznet legacy tree.
6. **NEVER run destructive git** (`checkout`, `clean`, `reset`) without explicit
   request.
7. **NEVER claim build success** without reading full output + exit code 0.
8. **NEVER assume** image base, paths, or guest semantics — verify XEX hash,
   PPC, Ghidra.
9. **NEVER commit, request, or redistribute** retail binaries, XEX files, SDK
   leaks, or proprietary keys.
10. After same crash twice with same patch — STOP, update state file, gather
    fresh PPC evidence via Ghidra before next fix.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| 360toolsUpdated Python | Extract + triage + switch-table fallback | Upstream tools |
| ReXGlue SDK | PPC lift + Xbox 360 OS/runtime on PC | `rexglue-sdk` |
| `generated/` | Mechanical codegen — never hand-edit for durable fix | Re-run `rexglue codegen` |
| Game project | Stubs, TOML, CMake, durable config | `stubs.cpp`, `*_config.toml` |
| Ghidra MCP | Evidence at guest VA → feeds TOML/stubs | Static analysis |

**CPU:** Xenon PowerPC, big-endian guest view. **GPU:** Xenos (not NV2A — that's OG Xbox).
**Not emulation** — PPC lifted to C++ with runtime stubs.

### Evidence Ladder

1. Build output / cmake exit code 0 (strongest)
2. Ghidra PPC disassembly at guest VA
3. `ghidra.get_xrefs_to` / `ghidra.read_memory` (static PPC truth)
4. `ghidra.decompile_function` (PPC hint only — never final proof)
5. `code/` reference repo comparison (if available — verify, not copy)

## Tracks

| Track | Use When | Pipeline | Success Criteria |
|---|---|---|---|
| A — ReXGlue | User has ReXGlue SDK installed | XEX → config → `rexglue codegen` → hooks → native exe | Boot past entrypoint; VI/audio stable |
| B — XenonRecomp | User confirmed sp00nznet legacy tree | XEX → XenonAnalyse → codegen → runtime → native exe | Same as A, XenonRecomp path |
| C — Matching decomp | Goal is PPC function-level match | XEX → Ghidra/PPC → handwritten C → verify vs disasm | Function-level PPC match |
| D — 360toolsUpdated | XBLA/ISO/LIVE/PIRS/CON packages | XBLA/ISO → `extract_stfs`/`extract-xiso` → `rexglue init` → `rexglue codegen` → cmake | Full pipeline from package to exe |

## Pipeline

```text
XEX/STFS/ISO → extract (360toolsUpdated) → rexglue init → *_{}_config.toml
              → rexglue codegen → generated/ → cmake --build → native EXE
              → crash triage (Ghidra PPC evidence) → TOML/stub fixes → regen
```

## Four Fix Tools

1. **Config TOML** — `[[switch_tables]]`, `[[functions]]`
2. **Game stubs** — `stubs.cpp` / `templates/advanced/`
3. **ReXGlue SDK patch** — `patches/0001`–`0005` on rexglue-sdk
4. **Regen** — `rexglue codegen` refreshes `generated/`

## Operational Phases

**Phase 0 — Extract.**
Goal: extract XEX from package and record identity.
- Track D: `python tools/extract_stfs.py <package>` or `extract-xiso <iso>`
- Record: title, XEX SHA-256, IMAGE_BASE (typically `0x82000000`), XEX sections
- Output to `extracted/`
- Exit: XEX hash recorded; IMAGE_BASE confirmed via `xex_info.py`

**Phase 1 — Init.**
Goal: generate project skeleton and first codegen.
- `rexglue init default.xex` → `*_config.toml`, `generated/`, CMake
- Verify paths in `REPHAMR_STATE.md`
- Generate `run_game_agent.bat` from template if missing
- Exit: project skeleton created; config.toml present; generated/ populated

**Phase 2 — First build.**
Goal: produce first build (expect linker errors).
- `cmake --build build/` — read full output
- Expect linker errors (missing stubs). Do NOT clean.
- Exit: build completes (may have linker errors)

**Phase 3 — Runtime bringup.**
Goal: resolve linker errors and boot past entrypoint.
- Stub missing imports in `stubs.cpp`
- Fix VdSwap timing if half-speed (`docs/speed-fix.md`)
- Handle unregistered VA: add to TOML `[[functions]]`
- Switch tables: `extract_switch_tables.py` or Ghidra MCP → TOML `[[switch_tables]]`
- Exit: build exits 0; exe boots past entrypoint

**Phase 4 — Crash triage.**
Goal: diagnose and fix runtime crashes with PPC evidence.
- Guest VA crash → check PPC translation in `generated/` → verify runtime
  registration → Ghidra MCP for PPC truth at crash VA → TOML override or stub fix
- Use `ghidra.decompile_function`, `ghidra.get_xrefs_to`, `ghidra.read_memory`
- Exit: crash table documented; fixes applied via TOML/stubs, not generated/

**Phase 5 — Polish.**
Goal: full graphics, audio, and stable gameplay.
- Graphics: VdSwap QPC, ROV, MSAA
- Audio, save/load
- Optional ReXGlue SDK patches `0001`–`0005` applied per symptom + SDK source check
- Exit: stable boot + full gameplay loop

## Build Gate / Validation Gate

Before every `cmake --build`:
1. **INSPECT** — no `--clean-first`, `--target clean`, or delete in command.
2. **VERIFY ENV** — MSVC + clang-cl on Windows (primary path). CMake + Ninja.
3. **VERIFY DIR** — build dir from `REPHAMR_STATE.md`; confirm exists.
4. **EXECUTE** — `cmake --build <build_dir>`; read full output; verify exit code 0.

After TOML/stub changes that affect `generated/`: `rexglue codegen` to regen.

Success may only be claimed when:
- Build exits 0
- Exe boots past entrypoint (Phase 3+)
- Graphics and audio stable (Phase 5)

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
python tools/extract_stfs.py <package>     # Track D: extract from STFS
extract-xiso <iso>                          # Track D: extract from ISO
rexglue init default.xex                    # generate project skeleton
rexglue codegen                             # regen generated/ after config change
cmake --build build/                        # incremental build (never clean)
```

MCP tools (when connected):

```text
ghidra.decompile_function        — PPC hint only, never final proof
ghidra.get_xrefs_to              — trace guest addresses at crash VA
ghidra.read_memory               — verify PPC bytes at crash VA
ghidra.rename_function_by_address — name after evidence
ghidra.analyze_function_complete — full dump: xrefs, callees, callers, vars
```

**IMAGE_BASE** from `xex_info.py` (typically `0x82000000`). Never ask user to
click Ghidra when MCP is connected.

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| Build fails: linker error | Stubs | Linker output, missing symbol name | Add stub in `stubs.cpp` |
| Unregistered VA crash | TOML | Crash VA, `ghidra.read_memory` | Add to `[[functions]]` in config TOML |
| Half-speed VdSwap | Recomp config | Visual frame rate, `docs/speed-fix.md` | Apply VdSwap QPC fix |
| Switch table not resolved | Recompiler / TOML | `extract_switch_tables.py` or Ghidra | Add to TOML `[[switch_tables]]` |
| Same crash twice, same patch | Evidence gap | `ghidra.get_xrefs_to` at crash VA | STOP — update state, gather fresh PPC evidence |
| `generated/` edited directly | Discipline | Git diff in generated/ | Revert; fix in TOML/stubs; regen |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — game, track, phase, XEX SHA-256, IMAGE_BASE, blocker, crash table
- `.rehamr/evidence/x360_crash_log.md` — crash table with guest VA, structural cause, fix

Optional:
- `.rehamr/evidence/x360_recon.md` — XEX sections, imports, switch tables
- `.rehamr/evidence/x360_patches_applied.md` — SDK patches 0001–0005 applied with dates
- `logs/x360_build_*.txt` — build output logs
- `logs/x360_trace.txt` — CDB trace if using native EXE debugging

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — track, phase, blocker, build command, crash table, verified paths.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
