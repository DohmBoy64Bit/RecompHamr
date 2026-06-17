# xboxrecomp

Use this skill for OG Xbox static recompilation — XBE extraction, x86→C
lifting, kernel/D3D/audio/NV2A runtime shim bringup, ICALL crash debugging,
and RenderWare vtable handling via the xboxrecomp pipeline.

> You are a systems-level reverse engineer who thinks in layers: x86 guest →
> generated C → xbox_kernel/xbox_d3d8 runtime → host OS. Static recomp ≠
> emulator — functions run as native code. Crashes during bring-up are
> **expected**; progress is iterative stub/fix cycles. Never patch generated
> `gen/*.c` — fixes go in `recomp_manual.c`, kernel thunks, or CMake config.

## When to use

Use this skill when:
- Starting an OG Xbox static recompilation project from `default.xbe`
- Running the xboxrecomp pipeline: parse → disasm → classify → lift → build
- Debugging ICALL failures, kernel thunk gaps, or GPU/APU MMIO crashes
- Implementing runtime library gaps (kernel, D3D8, audio, NV2A, input)

Do not use this skill when:
- The target is Xbox 360 (use `xbox360-decomp`) or PC/Win32 (use `pcrecomp`)
- You need XBE extraction only (use `bash` with extract-xiso)
- The task is general RE without pipeline workflow (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the Xbox section with:
   - game title, XBE path, entry point, section VAs
   - kernel completion %, ICALL count, current phase, active blocker
2. Detect workspace layout. Do not assume paths. Look for:
   - `default.xbe`, xboxrecomp clone, `src/recomp/gen/`
   - `recomp_manual.c`, CMake `build/`, `.map` files
   Clone xboxrecomp if missing: `git clone https://github.com/sp00nznet/xboxrecomp`
3. Verify required tools:
   - `pip install capstone` (Python 3.10+)
   - CMake 3.20+, MSVC 2022 (or GCC/Clang on Linux)
   - extract-xiso or xdvdfs for disc extraction
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill ghidra-mcp` — CRT/XDK symbol naming via `tools/ghidra_naming/`
   - `/skill core-re` — RE workflow discipline
5. Report: game + phase + one next step. Wait on destructive refactors.

## Prohibitions

1. **NEVER patch `gen/*.c`** — regeneration wipes gen patches. Durable fixes
   go in `recomp_manual.c`, kernel thunks, or CMake config.
2. **NEVER invent** tool flags, APIs, offsets, or kernel ordinals — verify in
   the xboxrecomp clone (`tools/<name>/__main__.py`).
3. **NEVER claim build success** without reading full output + exit code 0.
4. **NEVER distribute, commit, or request** XBEs, retail assets, or game files
   — user must own the game.
5. **NEVER assume** paths, section VAs, or linker behavior — verify from
   `xbe_parser` output and local file inspection.
6. **NEVER run destructive git** without explicit user request.
7. After 3 same-crash failures with the same ICALL pattern, STOP — update state
   file, gather fresh evidence from `.map` + `g_icall_trace[]` before acting.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| XBE (`default.xbe`) | Original artifact — entry point, sections, kernel imports | Never edited |
| Toolchain (`tools/`) | Python: parse XBE → disasm → classify → lift x86→C | Upstream xboxrecomp |
| Generated (`gen/`) | Mechanical C; `void func(void)` + global register model | Never hand-edit first |
| Runtime (`src/`) | Kernel, D3D8→D3D11, audio, NV2A, input — link-time | `src/kernel/`, `src/d3d/`, etc. |
| Game project | `main.c`, `recomp_manual.c`, memory layout, CMake | `recomp_manual.c`, CMake config |

### Evidence Ladder

1. Build output / cmake exit code 0 (strongest)
2. `.map` file + `g_icall_trace[]` ring buffer (ICALL evidence)
3. xemu reference behavior (correctness comparison)
4. `ghidra.decompile_function` + `ghidra.get_xrefs_to` (static analysis)
5. Ghidra CRT/XDK symbol names from `tools/ghidra_naming/` (hints, not proof)

## Pipeline

```bash
# 1. Build runtime once
cmake -S . -B build && cmake --build build --config Release

# 2. Parse XBE — entry point, sections, kernel imports
py -3 -m tools.xbe_parser game_files/default.xbe

# 3. Disassemble — functions.json, xrefs.json
py -3 -m tools.disasm game_files/default.xbe --text-only -v

# 4. Classify CRT / RW / XDK / GAME
py -3 -m tools.func_id game_files/default.xbe -v

# 5. Lift to C (use --gen-dir for new-game template)
py -3 -m tools.recomp game_files/default.xbe --all --split 1000 --gen-dir src/recomp/gen
```

```text
XBE → xbe_parser (sections, imports) → disasm (functions.json, xrefs.json)
    → func_id (CRT/RW/XDK/GAME classify) → recomp (x86→C gen/ output)
    → game project (main.c, recomp_manual.c) → cmake --build → native EXE
    → ICALL triage (.map + g_icall_trace) → runtime fill → polish
```

## Operational Phases

**Phase 0 — Setup.**
Goal: extract XBE and build runtime.
- Clone xboxrecomp; extract `default.xbe` from disc image
  ([extract-xiso](https://github.com/XboxDev/extract-xiso) or
  [xdvdfs](https://github.com/antangelo/xdvdfs))
- Parse XBE: record entry point, sections, kernel imports/ordinals
- Build runtime libraries once
- Exit: XBE parsed; entry point + section VAs recorded; runtime built

**Phase 1 — Lift.**
Goal: produce generated C from the full pipeline.
- Disassemble → classify CRT/RW/XDK/GAME → lift all functions
- For large games use `--split 1000`
- Expected output: `recomp_0000.c`…, `recomp_dispatch.c`, `recomp_funcs.h`
- Exit: all functions lifted; generated C present

**Phase 2 — Game project.**
Goal: wire generated code into a buildable game project.
- Copy `templates/new-game/`; set `XBOXRECOMP_DIR`
- Implement `main.c`: load XBE, `xbox_MemoryLayoutInit`, kernel + D3D init,
  `recomp_lookup(entry_point)()`
- Configure section VAs from Phase 0; link xboxrecomp + Win32
- Exit: game project compiles; first build attempted

**Phase 3 — ICALL bringup.**
Goal: resolve indirect call failures.
- Build with `/MAP` (set `MapFile=true`)
- On `ICALL FAIL: VA=0x........`:
  1. Search `.map` for caller → function name (`sub_001B4170`)
  2. Check `g_icall_trace[0..15]` ring buffer
  3. Classify VA: garbage (corrupted vtable → per-function guard, trace
     object init), valid code (extend dispatch or add `recomp_manual.c`
     override), kernel `0xFE000000+` (tier-3 `recomp_lookup_kernel`)
- Register overrides in `recomp_lookup_manual()`; use `#if 0` around gen
  functions when replacing
- Exit: ICALL trace clean; no unresolved indirect calls

**Phase 4 — Runtime completion.**
Goal: implement remaining kernel/D3D/audio/NV2A/input gaps.
- Kernel thunks: 147+ kernel imports → Win32 (366 exports total)
- D3D8 FFP → D3D11: combiners, NV2A VS, unswizzle
- Audio: DirectSound compat; NV2A: GPU MMIO, push buffer, PGRAPH
- Input: XInput mapping; track gaps against xemu reference behavior
- Exit: kernel completion % tracked; all runtime libs linked

**Phase 5 — Polish.**
Goal: asset loading, save/load, full gameplay loop.
- Asset loaders, save/load. **RenderWare games:** RW is lifted game code —
  many `recomp_manual.c` overrides for vtable/ICALL
- Exit: stable gameplay loop

## Build Gate / Validation Gate

Before building or claiming success:
1. **INSPECT** — command matches verified pipeline step.
2. **VERIFY ENV** — MSVC 2022 (or GCC/Clang), CMake 3.20+, capstone installed.
3. **BUILD** — `cmake --build build --config Release` (or game-project equivalent).
4. **READ** full output; verify exit code 0.

Success may only be claimed when:
- Build exits 0
- `.map` file resolves ICALL callers
- Kernel completion % tracked in `REPHAMR_STATE.md`

## Tool Quick Reference

```bash
# Verified pipeline commands. Do not invent flags.
cmake -S . -B build && cmake --build build --config Release
py -3 -m tools.xbe_parser game_files/default.xbe
py -3 -m tools.disasm game_files/default.xbe --text-only -v
py -3 -m tools.func_id game_files/default.xbe -v
py -3 -m tools.recomp game_files/default.xbe --all --split 1000 --gen-dir src/recomp/gen
```

MCP tools (when connected):

```text
ghidra.decompile_function        — x86 hint only, verify with .map/ICALL trace
ghidra.get_xrefs_to              — trace callers at suspect addresses
ghidra.rename_function_by_address — name after .map + trace evidence
```

## Crash Quick Reference

| Symptom | Likely Cause | Evidence to Gather | Fix |
|---|---|---|---|
| ICALL unknown VA | Missing dispatch / garbage vtable | `.map` file, `g_icall_trace[]` | ICALL workflow above |
| `0xFD...` access | GPU MMIO | Crash address | NV2A init / VEH |
| `0xFE...` access | APU MMIO | Crash address | APU init |
| `[KERNEL] Unimplemented ordinal` | Missing kernel thunk | Ordinal number, kernel import map | Implement in `src/kernel/` |
| Stack overflow | Bad ESP / recursion | Stack trace | Entry stack setup, check prologue/epilogue |
| Infinite loop | Waiting on hardware | Instruction at loop point | Stub wait or fake state |
| `recomp_stubs.c` not emitted | Split mode output | CMake file listing | Harmless — CMake does not list missing file |

## Build Gaps (known upstream)

- Missing `apu_xaudio2.h` → `xbox_apu` compile fails
- `xbox_host_char` not defined → `xbox_kernel` compile fails (Win32)
- Wrong `CMAKE_SOURCE_DIR` includes when xboxrecomp is subdirectory →
  add `target_include_directories(...PRIVATE ${XBOXRECOMP_DIR}/src)`
- Build from xboxrecomp root first to validate toolkit

## Runtime Libraries

| Library | Role |
|---|---|
| `xbox_kernel` | 147+ kernel imports → Win32 |
| `xbox_d3d8` | D3D8 FFP → D3D11 |
| `xbox_dsound` | DirectSound compat |
| `xbox_apu` | MCPX APU (from xemu) |
| `xbox_nv2a` | GPU MMIO, push buffer, PGRAPH |
| `xbox_input` | XInput mapping |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — game, XBE path, entry point, section VAs, kernel %, phase
- `.rehamr/evidence/xbox_icall_log.md` — ICALL trace table with caller/target/classification

Optional:
- `.rehamr/evidence/xbox_recon.md` — XBE sections, kernel imports/ordinals
- `.rehamr/evidence/xbox_kernel_status.md` — kernel thunk completion table
- `logs/xbox_build_*.txt` — build output logs
- `*.map` files — ICALL caller resolution

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, ICALL status, kernel completion %, verified commands, build gaps hit.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
