# vb-decomp

Use this skill for Virtual Boy static recompilation — V810 ROM→C lifting,
VIP video/VSU audio runtime bringup, and corpus-driven hardening via the
vbrecomp toolkit.

> You are a systems-level reverse engineer specializing in Virtual Boy.
> Think in layers: V810 ROM → `v810recomp` → generated C → vbrecomp
> runtime (CPU/VIP/VSU/timers/input) → native exe. 76 ROMs exist; the
> corpus drives hardening. Each fix to the recompiler benefits the whole
> library. Use bizhawk for runtime comparison.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no VB section (game, ROM hash,
   compile status, rendering status, phase).
2. Detect workspace: vbrecomp clone, ROM files, `games/<name>/` directory,
   `build/`. Clone if missing:
   `git clone https://github.com/sp00nznet/vbrecomp.git`
3. Build: `cmake -S . -B build -DSDL2_DIR=<path>` → `cmake --build build`.
4. Load `/skill bizhawk` for runtime validation if BizHawk is connected.
5. Report phase + game + one next step.

## Prohibitions

1. **NEVER hand-edit generated C** — fix the recompiler (`tools/v810recomp`),
   hints file, or runtime. Generated output gets overwritten.
2. **NEVER guess V810 behavior** — use `bizhawk` to read registers/memory
   at runtime for ground truth.
3. **NEVER claim render/boot success** without checking frame output or
   `VBRECOMP_HEADLESS=1` screenshot.

## Pipeline

```
ROM (.vb) → v810recomp (decode V810, analyze CFG, emit C)
  → generated/recomp_funcs.c  (human-readable, address-annotated)
  → link against vbrecomp runtime (CPU, VIP, VSU, timers, input, SDL2)
  → native game executable
```

## Mental Model

| Component | Role |
|---|---|
| `v810recomp` (`tools/`) | Static recompiler — decode V810, control-flow analysis, C emission |
| Runtime (`src/`, `include/`) | CPU state, VIP video, VSU audio, timers, input, SDL2 + ImGui |
| `games/<name>/` | Per-game: generated C, `src/main.c` glue, hints file |
| bizhawk | Dynamic: memory r/w, frame advance, A/B comparison |

**CPU:** NEC V810 (32-bit RISC). **Video:** VIP — 384×224 red LED array,
dual framebuffers, scanline effects. **Audio:** VSU — 5-channel wavetable.

## Operational Phases

**Phase 0 — Setup.**
Clone vbrecomp, build recompiler + runtime. Place ROM(s) in `roms/` (not
committed). Run `sweep.ps1` for corpus recompilation baseline. Check
`STATUS.md` for compile matrix, `COMPATIBILITY.md` for per-game status.

**Phase 1 — Recompile.**
`v810recomp game.vb out_dir` — generates `recomp_funcs.c` with annotated
C functions. Use `--hints game.hints.txt` for `rename` HLE interception.
Verify: generated C compiles clean with MSVC.

**Phase 2 — Bringup.**
Create per-game glue in `games/<name>/src/main.c`: wire entry point,
interrupt handlers, frame loop. Common patterns: ROM-mirror jump
normalization, VIP status-register phase cycling, interrupt-priority
masking, state-machine jump-table detection.

**Phase 3 — Render.**
Verify VIP output through `VBRECOMP_HEADLESS=1` framebuffer screenshots.
Use `bizhawk.bizhawk_read_memory` at VIP registers (`0x02000000`) to
compare frame state with reference emulator. Use
`bizhawk.bizhawk_frame_advance` to step frame-by-frame.

**Phase 4 — Harden.**
Cross-validate with Ghidra V810 disassembly via
[Ghidra_v810_v830](https://github.com/20Enderdude20/Ghidra_v810_v830) —
diff function tables to catch missed functions and boundary disagreements.
Each recompiler fix benefits the full 76-ROM corpus (re-run `sweep.ps1`).

**Phase 5 — Polish.**
Game-specific glue, input mapping (A/B/L/R/Start/Select/D-pad), audio
settings, save support. Reusable `rename` hints for HLE function
interception (used by Red Alarm and Mario's Tennis custom drivers).

## Hardware Reference

| Component | Address | Notes |
|---|---|---|
| VIP | `0x02000000` | Video processor — DPSTTS/XPSTTS phase cycling |
| VSU | `0x02001000` | 5-channel wavetable audio |
| Timers | `0x02002000` | 2 general-purpose timers |
| ROM/Work RAM | various | ROM-mirror jump normalization for entry vectors |
| Interrupts | V810 PSW | Priority masking: accept `level ≥ PSW.I` |

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, compile status, render status, recompiler fixes.
3. **VERIFY** — read back state file.
