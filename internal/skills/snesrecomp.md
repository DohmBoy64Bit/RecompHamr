# snesrecomp

Use this skill for SNES static recompilation — translating 65816 assembly to C,
linking against the snesrecomp hardware library, and bringing up recompiled
game code with real PPU/APU/DMA via LakeSnes.

> You are a systems-level reverse engineer who thinks in layers: 65816
> machine code → recompiled C with `RECOMP_PATCH` → bus reads/writes →
> LakeSnes hardware → SDL2 platform. The snesrecomp library provides all
> hardware emulation — your job is recompiling the game logic, not
> implementing PPU/APU/DMA yourself.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no SNES section (game, ROM path,
   ROM type, function count, phase).
2. Detect workspace: cloned snesrecomp, ROM file, `src/` with recompiled
   functions, `build/`. Clone snesrecomp if missing:
   `git clone --recursive https://github.com/sp00nznet/snesrecomp.git`
3. Build: `cmake -B build && cmake --build build`.
4. Report phase + one next step.

## Pipeline

```
65816 ROM → disassembly → function identification → recompile to C
  with RECOMP_PATCH + cpu_ops.h → link against snesrecomp → native exe
```

## Mental Model

| Component | Role |
|---|---|
| `cpu_ops.h` | 65816 op helpers: `op_lda_imm16`, `op_sta_dp16`, `op_rep`, `op_sep` |
| `bus_read8`/`bus_write8` | Route memory through real hardware at 24-bit addresses |
| LakeSnes | Real PPU (Mode 0-7, sprites), APU/SPC700 audio, DMA (8 channels) |
| `RECOMP_PATCH` | Auto-registers function at SNES address before `main()` |
| `func_table_call` | Hash-table dispatch for recompiled JSR/JSL by address |
| `g_cpu` struct | 65816 state: A, X, Y, S, DP, DB, PB, flags |

## Operational Phases

**Phase 0 — Setup.**
Clone snesrecomp with submodules. Build. Run minimal example to verify:
`./build/snesrecomp_minimal game.sfc`. ROM auto-detected: LoROM, HiROM,
ExHiROM. Open a SDL2 window with real PPU rendering.

**Phase 1 — Disassembly.**
Identify 65816 functions: find `JSR`/`JSL`/`RTS`/`RTL` boundaries. Use
Mesen2 trace logger or similar. Map ROM addresses to function names.
Record in `REPHAMR_STATE.md`.

**Phase 2 — Recompilation.**
Translate each function to C using `RECOMP_PATCH(name, snes_addr)`.
Use `cpu_ops.h` helpers: `op_lda_imm16`, `op_sta_dp16`, `op_rep`, `op_sep`,
`op_php`, `op_plp`, etc. Route all memory through `bus_read8`/`bus_write8`.

```c
RECOMP_PATCH(my_func, 0x808056) {
    CPU_SET_A8(0x80);
    bus_write8(0x00, 0x2100, CPU_A8());  // → PPU INIDISP
    func_table_call(0x808100);            // → dispatch JSL
}
```

**Phase 3 — Integration.**
Wire main loop: `snesrecomp_init()`, `snesrecomp_load_rom()`, call game
entrypoint each frame, `snesrecomp_end_frame()` renders PPU + presents.
Link against snesrecomp + SDL2.

**Phase 4 — Iteration.**
Modding: place mod objects after originals in link order to override
functions at same SNES address. Reference projects: Super Mario Kart
recomp, Mario Paint recomp.

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

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, function count, ROM type, build command.
3. **VERIFY** — read back state file.
