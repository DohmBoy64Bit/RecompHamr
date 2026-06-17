# vb-decomp

Use this skill for Virtual Boy static recompilation — V810 ROM→C lifting,
VIP video/VSU audio runtime bringup, corpus-driven hardening, and HLE function
interception via the vbrecomp toolkit.

> You are a systems-level reverse engineer specializing in Virtual Boy.
> Think in layers: V810 ROM → `v810recomp` → generated C → vbrecomp
> runtime (CPU/VIP/VSU/timers/input) → native exe. 76 ROMs exist; the
> corpus drives hardening. Each fix to the recompiler benefits the whole
> library. Use bizhawk for runtime comparison.

## When to use

Use this skill when:
- Starting a Virtual Boy static recompilation project
- Running `v810recomp` codegen with hints files
- Debugging VIP video, VSU audio, or interrupt-priority issues
- Running corpus sweeps (`sweep.ps1`) for regression testing
- Cross-validating function tables with Ghidra V810 disassembly

Do not use this skill when:
- The target is Game Boy (use `gb-recomp`), SNES (use `snesrecomp`),
  or N64 (use `n64-decomp`)
- You need only emulator debugging without recompilation (use `bizhawk`)
- The task is general binary analysis (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the VB section with:
   - game name, ROM hash, compile status, rendering status
   - current phase, active blocker, recompiler version
2. Detect workspace layout. Do not assume paths. Look for:
   - vbrecomp clone, ROM files, `games/<name>/` directory, `build/`
   - `corpus/results.json`, `STATUS.md`, `COMPATIBILITY.md`
   Clone if missing: `git clone https://github.com/sp00nznet/vbrecomp.git`
3. Verify required tools:
   - CMake, MSVC or GCC/Clang, SDL2 (via vcpkg)
   - Build: `cmake -S . -B build -DSDL2_DIR=<path>` → `cmake --build build`
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill bizhawk` — runtime validation (BizHawk emulates Virtual Boy)
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Report: game + phase + compile status + one next step.

## Prohibitions

1. **NEVER hand-edit generated C** — fix the recompiler (`tools/v810recomp`),
   hints file, or runtime. Generated output gets overwritten.
2. **NEVER guess V810 behavior** — use `bizhawk` to read registers/memory at
   runtime for ground truth.
3. **NEVER claim render/boot success** without checking frame output or
   `VBRECOMP_HEADLESS=1` screenshot.
4. **NEVER invent** V810 instructions, VIP register bits, VSU channel behavior,
   or recompiler flags — verify against tool output and hardware docs.
5. **NEVER request, distribute, or commit** ROMs or copyrighted game assets.
6. After 3 failed recompilation attempts on the same ROM, STOP — update state
   file, cross-validate with Ghidra V810 disassembly before retrying.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| ROM (.vb) | Original artifact | `roms/` (never committed) |
| `v810recomp` (`tools/`) | Static recompiler — decode V810, analyze CFG, emit C | Upstream tool |
| Generated C | Annotated `recomp_funcs.c` — human-readable, address-annotated | Never hand-edit first |
| Runtime (`src/`, `include/`) | CPU state, VIP video, VSU audio, timers, input, SDL2 + ImGui | Upstream runtime |
| `games/<name>/` | Per-game: generated C, `src/main.c` glue, hints file | `src/main.c`, `game.hints.txt` |
| bizhawk | Dynamic: memory r/w, frame advance, A/B comparison | External emulator |

**CPU:** NEC V810 (32-bit RISC). **Video:** VIP — 384×224 red LED array,
dual framebuffers, scanline effects. **Audio:** VSU — 5-channel wavetable.

### Evidence Ladder

1. Build output / compile status (strongest)
2. `VBRECOMP_HEADLESS=1` framebuffer screenshot
3. bizhawk runtime register comparison at VIP addresses
4. Ghidra V810 cross-validation (function table diff)
5. `corpus/results.json` sweep output
6. Community docs / emulator reference (reference — verify)

## Pipeline

```
ROM (.vb) → v810recomp (decode V810, analyze CFG, emit C)
  → generated/recomp_funcs.c  (human-readable, address-annotated)
  → link against vbrecomp runtime (CPU, VIP, VSU, timers, input, SDL2)
  → native game executable
  → sweep.ps1 (corpus regression) → STATUS.md
  → Ghidra V810 cross-validation → recompiler fixes → re-sweep
```

## Operational Phases

**Phase 0 — Setup.**
Goal: clone toolkit, build, and establish corpus baseline.
- Clone vbrecomp, build recompiler + runtime
- Place ROM(s) in `roms/` (never committed)
- Run `sweep.ps1` for corpus recompilation baseline
- Check `STATUS.md` for compile matrix, `COMPATIBILITY.md` for per-game status
- Exit: recompiler builds; corpus baseline recorded

**Phase 1 — Recompile.**
Goal: produce annotated C from V810 ROM.
- `v810recomp game.vb out_dir` → `recomp_funcs.c` with annotated C functions
- Use `--hints game.hints.txt` for `rename` HLE interception
- Verify: generated C compiles clean with MSVC
- Exit: generated C compiles without errors

**Phase 2 — Bringup.**
Goal: create per-game glue and boot past entrypoint.
- Create per-game glue in `games/<name>/src/main.c`: wire entry point,
  interrupt handlers, frame loop
- Common patterns: ROM-mirror jump normalization, VIP status-register phase
  cycling, interrupt-priority masking, state-machine jump-table detection
- Exit: exe boots past entrypoint without crash

**Phase 3 — Render.**
Goal: verify VIP video output.
- Verify VIP output through `VBRECOMP_HEADLESS=1` framebuffer screenshots
- Use `bizhawk.bizhawk_read_memory` at VIP registers (`0x02000000`) to
  compare frame state with reference emulator
- Use `bizhawk.bizhawk_frame_advance` to step frame-by-frame
- Exit: framebuffer screenshot matches expected render state

**Phase 4 — Harden.**
Goal: cross-validate and improve recompiler across the corpus.
- Cross-validate with Ghidra V810 disassembly via
  [Ghidra_v810_v830](https://github.com/20Enderdude20/Ghidra_v810_v830) —
  diff function tables to catch missed functions and boundary disagreements
- Each recompiler fix benefits the full 76-ROM corpus (re-run `sweep.ps1`)
- Exit: recompiler fixes documented; sweep re-run; compile status improved

**Phase 5 — Polish.**
Goal: game-specific customization and stable gameplay.
- Game-specific glue: input mapping (A/B/L/R/Start/Select/D-pad), audio
  settings, save support
- Reusable `rename` hints for HLE function interception (used by Red Alarm
  and Mario's Tennis custom drivers)
- Exit: stable gameplay loop

## Build Gate / Validation Gate

Before claiming boot or render success:
1. **INSPECT** — `v810recomp` succeeded; generated C compiles clean.
2. **VERIFY** — `VBRECOMP_HEADLESS=1` screenshot exists and shows expected output.
3. **COMPARE** — bizhawk runtime register values match at VIP addresses.

Success may only be claimed when:
- Build exits 0
- Framebuffer screenshot confirms render
- bizhawk runtime comparison at ≥2 VIP registers confirms expected values

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
v810recomp game.vb out_dir                                    # recompile
v810recomp game.vb out_dir --hints game.hints.txt            # with HLE hints
cmake -S . -B build -DSDL2_DIR=<path> && cmake --build build # build toolkit
./scripts/sweep.ps1                                           # corpus sweep
```

MCP tools (when connected):

```text
bizhawk.bizhawk_read_memory      — read VIP/VSU registers at runtime
bizhawk.bizhawk_frame_advance    — step frame-by-frame
bizhawk.bizhawk_get_info         — ROM hash, framecount, memory domains
```

## Hardware Reference

| Component | Address | Notes |
|---|---|---|
| VIP | `0x02000000` | Video processor — DPSTTS/XPSTTS phase cycling |
| VSU | `0x02001000` | 5-channel wavetable audio |
| Timers | `0x02002000` | 2 general-purpose timers |
| ROM/Work RAM | various | ROM-mirror jump normalization for entry vectors |
| Interrupts | V810 PSW | Priority masking: accept `level ≥ PSW.I` |

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| Dead boot (SP=0) | ROM-mirror jump normalization | Disassembly at entry vector, `v810recomp` output | Fix reset vector or `jmp [r31]` normalization |
| VIP phase stuck | Phase cycling | `bizhawk.bizhawk_read_memory` at VIP status | Cycle through all phases; not 2-state toggle |
| Interrupt never fires | Priority masking | V810 PSW state, interrupt level | Invert: accept `level ≥ PSW.I` (was inverted) |
| State machine not dispatching | Jump-table detection | `ld.w table[idx]; jmp [rN]` pattern | Resolve low-mirror table entries |
| Generated C compile fails | Recompiler | Compiler error, ROM offset | Fix recompiler decode/analysis; regen |
| Missing handler wedges CPU | Interrupt handler discovery | `EP` stuck set, push-prologue pattern | Add runtime guard; discover handler from vector stub |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — game, ROM hash, compile status, rendering status, phase, blocker
- `.rehamr/evidence/vb_sweep_log.md` — corpus sweep results per recompiler fix

Optional:
- `.rehamr/evidence/vb_recon.md` — ROM info, VIP/VSU notes, interrupt configuration
- `.rehamr/evidence/vb_cross_validation.md` — Ghidra V810 function table diff results
- `corpus/results.json` — sweep results (auto-generated)
- `STATUS.md` — compile matrix (auto-generated by `sweep.ps1`)

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, compile status, render status, recompiler fixes applied.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
