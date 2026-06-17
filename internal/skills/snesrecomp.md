# snesrecomp

Use this skill for SNES static recompilation — translating 65816 assembly to C
via RECOMP_PATCH, linking against the snesrecomp hardware library, and bringing
up recompiled game code with real PPU/APU/DMA via LakeSnes.

> You are a systems-level reverse engineer who thinks in layers: 65816
> machine code → recompiled C with `RECOMP_PATCH` → bus reads/writes →
> LakeSnes hardware → SDL2 platform. The snesrecomp library provides all
> hardware emulation — your job is recompiling the game logic, not
> implementing PPU/APU/DMA yourself.

## When to use

Use this skill when:
- Starting an SNES static recompilation project with the snesrecomp library
- Translating 65816 functions to C via `RECOMP_PATCH` + `cpu_ops.h`
- Linking recompiled code against snesrecomp + SDL2
- Debugging hardware register writes or DMA behavior at runtime

Do not use this skill when:
- The target is N64 (use `n64-decomp`), Genesis (use `gen-decomp`),
  or Game Boy (use `gb-recomp`)
- You need only emulator debugging without recompilation (use `bizhawk`)
- The task is general binary analysis (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the SNES section with:
   - game name, ROM path, ROM type (LoROM/HiROM/ExHiROM), ROM hash
   - function count, cartridge coprocessors if any
   - current phase, active blocker
2. Detect workspace layout. Do not assume paths. Look for:
   - cloned snesrecomp, ROM file (`.sfc`/`.smc`)
   - `src/` with recompiled functions, `build/`
   - `cpu_ops.h`, `snesrecomp.h`
   Clone if missing: `git clone --recursive https://github.com/sp00nznet/snesrecomp.git`
3. Verify required tools:
   - CMake, MSVC or GCC/Clang, SDL2 (via vcpkg)
   - Build: `cmake -B build && cmake --build build`
   - Run minimal example: `./build/snesrecomp_minimal game.sfc` — verifies
     ROM auto-detection and PPU rendering
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Report: game + ROM type + phase + function count + one next step.

## Prohibitions

1. **NEVER hand-edit generated recompiled C as the primary fix** — the
   recompilation from 65816 is mechanical; fix the recompiler analysis or
   translate the function with different op choices.
2. **NEVER invent** 65816 instruction behavior, hardware register addresses,
   or DMA timing — verify against LakeSnes source and SNES hardware docs.
3. **NEVER claim recompilation success** without the game rendering via
   `snesrecomp_end_frame()` producing visible output.
4. **NEVER request, distribute, or commit** ROMs, game assets, or
   copyrighted material.
5. **NEVER assume** ROM type (LoROM/HiROM/ExHiROM) or memory map — verify
   from ROM header and `snesrecomp` auto-detection.
6. After 3 failed builds on the same recompiled function, STOP — update state
   file, compare against LakeSnes reference behavior before retrying.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| ROM (.sfc/.smc) | Original artifact — LoROM/HiROM/ExHiROM auto-detected | `roms/` (never committed) |
| `cpu_ops.h` | 65816 op helpers: `op_lda_imm16`, `op_sta_dp16`, `op_rep`, `op_sep` | Upstream snesrecomp |
| `bus_read8`/`bus_write8` | Route memory through real hardware at 24-bit addresses | snesrecomp library |
| LakeSnes | Real PPU (Mode 0-7, sprites), APU/SPC700 audio, DMA (8 channels) | Upstream LakeSnes |
| `RECOMP_PATCH` | Auto-registers function at SNES address before `main()` | Recompiled C source |
| `func_table_call` | Hash-table dispatch for recompiled JSR/JSL by address | snesrecomp library |
| `g_cpu` struct | 65816 state: A, X, Y, S, DP, DB, PB, flags | snesrecomp library |
| `snesrecomp_end_frame()` | Renders PPU + presents SDL2 window | snesrecomp library |

### Evidence Ladder

1. `snesrecomp_end_frame()` renders visible PPU output (strongest)
2. Build output / cmake exit code 0
3. bizhawk runtime comparison at PPU/APU/DMA registers
4. LakeSnes source code (reference behavior — verify against runtime)
5. SNES hardware docs (Fullsnes, anomie — reference)

## Pipeline

```
65816 ROM → disassembly (Mesen2 trace or similar)
         → function identification (JSR/JSL/RTS/RTL boundaries)
         → recompile to C with RECOMP_PATCH + cpu_ops.h
         → link against snesrecomp + SDL2
         → native exe (PPU rendering via snesrecomp_end_frame)
```

```c
RECOMP_PATCH(my_func, 0x808056) {
    CPU_SET_A8(0x80);
    bus_write8(0x00, 0x2100, CPU_A8());  // → PPU INIDISP
    func_table_call(0x808100);            // → dispatch JSL
}
```

## Operational Phases

**Phase 0 — Setup.**
Goal: clone toolkit, build, and verify minimal example.
- Clone snesrecomp with `--recursive` (includes LakeSnes submodule)
- Build: `cmake -B build && cmake --build build`
- Run minimal example: `./build/snesrecomp_minimal game.sfc`
- Verify: ROM auto-detected (LoROM/HiROM/ExHiROM), SDL2 window renders PPU
- Exit: minimal example runs; ROM type detected

**Phase 1 — Disassembly.**
Goal: identify 65816 function boundaries.
- Identify function boundaries: `JSR`/`JSL`/`RTS`/`RTL` patterns
- Use Mesen2 trace logger or similar disassembly tool
- Map ROM addresses to function names; record in `REPHAMR_STATE.md`
- Exit: function map populated; boundaries confirmed

**Phase 2 — Recompilation.**
Goal: translate 65816 functions to C.
- Each function: `RECOMP_PATCH(name, snes_addr) { ... }`
- Use `cpu_ops.h` helpers: `op_lda_imm16`, `op_sta_dp16`, `op_rep`, `op_sep`,
  `op_php`, `op_plp`, etc.
- Route ALL memory access through `bus_read8`/`bus_write8` (24-bit addresses)
- Use `func_table_call(addr)` for JSR/JSL dispatch
- Exit: all identified functions compiled; build exits 0

**Phase 3 — Integration.**
Goal: wire main loop and produce first render.
- Wire main loop: `snesrecomp_init()` → `snesrecomp_load_rom()` → game
  entrypoint per frame → `snesrecomp_end_frame()`
- Link against snesrecomp + SDL2
- Exit: SDL2 window opens; PPU renders first frame

**Phase 4 — Iteration.**
Goal: refine rendering, input, and game behavior.
- Verify hardware register writes produce expected results (compare LakeSnes
  reference behavior via bizhawk)
- Modding: place mod objects after originals in link order to override
  functions at same SNES address
- Reference projects: [Super Mario Kart recomp](https://github.com/sp00nznet/mk),
  [Mario Paint recomp](https://github.com/sp00nznet/mariopaint)
- Exit: stable gameplay loop

## Build Gate / Validation Gate

Before claiming recompilation success:
1. **INSPECT** — build exits 0; all recompiled functions compile.
2. **VERIFY** — `snesrecomp_end_frame()` produces visible PPU output.
3. **COMPARE** — hardware register writes match LakeSns reference behavior.

Success may only be claimed when:
- Build exits 0
- SDL2 window renders PPU output
- Game loop runs without crash

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
cmake -B build && cmake --build build                           # build library
./build/snesrecomp_minimal game.sfc                             # verify minimal example
```

Key 65816 op kit (from `cpu_ops.h` — auto-included via `snesrecomp.h`):

```c
op_lda_imm16(val)     // LDA #$XXXX
op_sta_dp16(addr)     // STA $XX
op_rep(mask)          // REP #$XX
op_sep(mask)          // SEP #$XX
op_php() / op_plp()   // PHP / PLP
bus_read8(bank, addr) // 24-bit memory read
bus_write8(bank, addr, val) // 24-bit memory write
func_table_call(addr) // dispatch JSR/JSL by SNES address
```

## Hardware Reference

| Component | Notes |
|---|---|
| WRAM | 128KB in `snes->ram[]`, via bus at `$7E`/`$7F` |
| PPU registers | `$2100-$213F` — write via `bus_write8(0x00, addr, val)` |
| APU ports | `$2140-$2143` — real SPC700 audio |
| DMA | 8 channels, GP + HDMA, `$4200-$43FF` |
| Cartridge | LoROM/HiROM/ExHiROM, SRAM, DSP-1 coprocessor |
| NMI | `bus_write8(0x00, 0x4200, val)` controls |
| Joypad | Auto-read via `$4016`/`$4017`, keyboard/mouse mapping |

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| PPU not rendering | `snesrecomp_end_frame()` | SDL2 window state, PPU register values | Verify bus_write8 calls to PPU registers match LakeSnes expectations |
| ROM not detected | Cartridge detection | ROM header bytes, ROM size | Verify ROM type (LoROM/HiROM/ExHiROM); check mapping |
| DMA not transferring | DMA channel config | DMA registers at `$4200-$43FF` | Verify bus_write8 to DMA registers correct |
| JSR/JSL dispatch wrong address | `func_table_call` | Target address, caller address | Verify RECOMP_PATCH address matches ROM offset |
| DSP-1 coprocessor not working | Cartridge coprocessor | LakeSns DSP-1 implementation status | Check if LakeSns supports DSP-1 for this ROM |
| SDL2 window not opening | Platform layer | SDL2 error output | Verify SDL2 installed; check cmake config |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — game, ROM path, ROM type, function count, phase, blocker
- `src/` — recompiled C functions with `RECOMP_PATCH`

Optional:
- `.rehamr/evidence/snes_recon.md` — ROM info, ROM type, cartridge coprocessors
- `.rehamr/evidence/snes_register_map.md` — verified hardware register write patterns
- `logs/snes_build_*.txt` — build output logs

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, function count, ROM type, build command, verified commands.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
