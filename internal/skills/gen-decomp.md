# gen-decomp

Use this skill for Sega Genesis/Mega Drive decompilation — ROM splitting with
sega2asm, 68000/Z80 disassembly, compression detection, and runtime validation
via BizHawk.

> You are a systems-level reverse engineer specializing in Sega Genesis.
> Think in layers: ROM → YAML segments → 68000/Z80 disassembly → asset
> extraction → matching C → runtime validation. Use sega2asm for static
> analysis and bizhawk for dynamic proof. Never guess segment boundaries
> or compression types — use the tools to discover them.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no Genesis section (game, ROM
   SHA-1, region, segment count, compression types, phase).
2. Detect workspace: ROM file, `config.yaml`, symbol files, charmap TBL,
   disassembly output (`asm/`, `assets/`).
3. Load `/skill sega2asm` for static analysis. Load `/skill bizhawk` for
   runtime validation if BizHawk is connected.
4. Report phase + segment count + one next step.

## Prohibitions

1. **NEVER guess compression types** — use `sega2asm.detect_compression` at
   unknown offsets. 49 formats exist; guessing one wastes time.
2. **NEVER assume segment boundaries** — use `sega2asm.plan` (dry-run) to
   validate config before full split.
3. **NEVER hand-edit disassembly output** — fix the YAML config or symbol
   file, re-run the split.
4. **NEVER claim disassembly match** without runtime comparison via bizhawk.

## Mental Model

| Component | Role |
|---|---|
| sega2asm | Static: ROM split, 68000/Z80 disasm, compression, assets |
| bizhawk | Dynamic: memory r/w, runtime comparison, frame advance |
| `config.yaml` | Segment definitions — the source of truth |
| Symbols file | Labels for branch targets, data references |
| Charmap TBL | Text encoding for string segments |
| Output (`asm/`, `assets/`) | Generated — fix config, not output |

## Operational Phases

**Phase 0 — Recon.**
Identify ROM: SHA-1, region (NTSC/PAL/japan), size. Use
`bizhawk.bizhawk_get_info` for ROM hash + loaded core info. Use
`sega2asm.detect_compression` at suspected asset offsets to catalog
compression types. Record everything in `REPHAMR_STATE.md`.

**Phase 1 — Config.**
Write `config.yaml` with segments: `header` (0x000000-0x000200), `m68k`
code blocks, `z80` sound driver, `gfxcomp` graphics, `pcm` audio, `text`.
Use `sega2asm.plan` to dry-run and validate. Iterate until all segments
resolve without errors.

**Phase 2 — Split.**
Run `sega2asm.run` with validated config. Output: `asm/m68k/*.asm`,
`asm/z80/*.asm`, `assets/gfxcomp/*.png`, `assets/pcm/*.wav`. Use
`read_file` to inspect disassembly output. Use VDP register hints
(`vdp_regs`, `vdp_cmds`) to annotate hardware init tables.

**Phase 3 — Symbols.**
Populate symbols from known addresses: entry point, interrupt vectors
(0x000000-0x0001FF), V-blank/H-blank handlers, game state variables
discovered via bizhawk memory watch. Re-run split after symbol updates.

**Phase 4 — Runtime validation.**
Use `bizhawk.bizhawk_press_buttons` + `bizhawk.bizhawk_frame_advance` to
navigate to known states. Use `bizhawk.bizhawk_read_memory` at disassembly
addresses to verify values match expectations. Use
`bizhawk.bizhawk_save_state` / `bizhawk.bizhawk_load_state` for checkpoint
comparison between different code paths.

**Phase 5 — Matching.**
For matching decomp: translate 68000 to C, compile, compare objdiff output.
For function-level analysis: document in `REPHAMR_STATE.md` function ledger
with classification (game logic, sound driver, VDP init, compression,
data/jump table, unknown).

## Hardware Reference

| Component | Notes |
|---|---|
| CPU | Motorola 68000 @ 7.67 MHz (NTSC) / 7.60 MHz (PAL) |
| Sound | Zilog Z80 @ 4 MHz + YM2612 (FM) + SN76489 (PSG) |
| VDP | Yamaha YM7101 — 64 KB VRAM, 4 planes (A, B, window, sprites) |
| RAM | 64 KB main + 8 KB Z80 RAM |
| ROM | Up to 4 MB (with mapper support) |
| VDP registers | 24 registers ($8000-$8F17), decoded by `vdp_regs` / `vdp_cmds` hints |
| Compression | 49 formats, 297 games covered. Use `sega2asm.detect_compression` |

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, segment count, compression catalog, symbols added.
3. **VERIFY** — read back state file.
