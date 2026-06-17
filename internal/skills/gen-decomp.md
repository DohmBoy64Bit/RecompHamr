# gen-decomp

Use this skill for Sega Genesis/Mega Drive decompilation — ROM splitting with
sega2asm, 68000/Z80 disassembly, compression detection, VDP register decoding,
and runtime validation via BizHawk.

> You are a systems-level reverse engineer specializing in Sega Genesis.
> Think in layers: ROM → YAML segments → 68000/Z80 disassembly → asset
> extraction → matching C → runtime validation. Use sega2asm for static
> analysis and bizhawk for dynamic proof. Never guess segment boundaries
> or compression types — use the tools to discover them.

## When to use

Use this skill when:
- Starting a Sega Genesis / Mega Drive decompilation or split project
- Building a YAML segment config and running sega2asm splits
- Cataloging Genesis compression types (49 formats, 297 games)
- Annotating 68000/Z80 disassembly with symbols and VDP register hints
- Validating disassembly accuracy with BizHawk runtime comparison

Do not use this skill when:
- The target is SNES (use `snesrecomp`), N64 (use `n64-decomp`), or GBA
- You need only emulator debugging without ROM splitting (use `bizhawk`)
- The task is general binary analysis without Genesis context (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the Genesis section with:
   - game name, ROM SHA-1, region (NTSC/PAL/Japan), ROM size
   - segment count, compression types cataloged, current phase, active blocker
   - symbol file path, charmap TBL path
2. Detect workspace layout. Do not assume paths. Look for:
   - ROM file, `config.yaml`, symbol files (`.txt`, `.sym`), charmap TBL
   - disassembly output (`asm/`, `assets/`), `sega2asm` clone
3. Verify required tools:
   - `sega2asm` binary on PATH (`go install github.com/hansbonini/sega2asm@latest`)
   - BizHawk 2.6.2+ with `bridge.lua` loaded
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill sega2asm` — ROM splitting, compression detection, disassembly
   - `/skill bizhawk` — runtime validation via BizHawk
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Report: game + region + phase + segment count + one next step.

## Prohibitions

1. **NEVER guess compression types** — use `sega2asm.detect_compression` at
   unknown offsets. 49 formats exist across 297 games; guessing wastes time.
2. **NEVER assume segment boundaries** — use `sega2asm.plan` (dry-run) to
   validate config before full split.
3. **NEVER hand-edit disassembly output** — fix the YAML config or symbol
   file, re-run the split. The generated `.asm` is output, not source.
4. **NEVER claim disassembly accuracy** without runtime comparison via bizhawk.
5. **NEVER invent** VDP register values, compression format names, or M68K/Z80
   instruction behavior — verify against sega2asm output and hardware docs.
6. **NEVER request, distribute, or commit** ROMs or copyrighted game assets.
7. After 3 failed split attempts with the same segment, STOP — update state
   file, gather fresh compression evidence via `sega2asm.detect_compression`
   before retrying.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| ROM | Original artifact — SHA-1 verified | `roms/` (never committed) |
| `config.yaml` | Segment definitions — the source of truth | `config.yaml` |
| sega2asm | Static: ROM split, 68000/Z80 disasm, compression, assets | External tool |
| Symbols file | Labels for branch targets, data references, text decode | `.txt` / `.sym` |
| Charmap TBL | Text encoding for string segments | `.tbl` |
| Generated output | `asm/`, `assets/` — fix config, not output | Re-run split after config/symbol fix |
| bizhawk | Dynamic: memory r/w, runtime comparison, frame advance | External emulator |

### Evidence Ladder

1. bizhawk runtime comparison at disassembly addresses (strongest)
2. `sega2asm.plan` dry-run validation
3. `sega2asm.run` split output + VDP register decode
4. `sega2asm.detect_compression` result at specific offset
5. Community decomp projects / Genesis hardware docs (reference — verify)

## Pipeline

```text
ROM → sega2asm.detect_compression (catalog types)
    → config.yaml (segment definitions) → sega2asm.plan (dry-run validate)
    → sega2asm.run (full split) → asm/ (M68K+Z80) + assets/ (PNG+WAV+text)
    → symbols populated (entry point, vectors, bizhawk-verified)
    → re-run split → read_file disasm → bizhawk runtime validation
    → matching C (68000→C, objdiff)
```

## Operational Phases

**Phase 0 — Recon.**
Goal: identify ROM and catalog compression types.
- Identify ROM: SHA-1, region (NTSC/PAL/Japan), size
- Use `bizhawk.bizhawk_get_info` for ROM hash + loaded core info
- Use `sega2asm.detect_compression` at suspected asset offsets
- Record everything in `REPHAMR_STATE.md`
- Exit: ROM hash recorded; compression catalog started

**Phase 1 — Config.**
Goal: produce validated YAML segment config.
- Write `config.yaml`: `header` (0x000000-0x000200), `m68k` code blocks,
  `z80` sound driver, `gfxcomp` graphics, `pcm` audio, `text` segments
- Use `sega2asm.plan` to dry-run and validate
- Iterate until all segments resolve without errors
- Exit: all segments pass dry-run validation

**Phase 2 — Split.**
Goal: produce full disassembly and asset extraction.
- Run `sega2asm.run` with validated config
- Output: `asm/m68k/*.asm`, `asm/z80/*.asm`, `assets/gfxcomp/*.png`,
  `assets/pcm/*.wav`
- Use `read_file` to inspect disassembly output
- Use VDP register hints (`vdp_regs`, `vdp_cmds`) to annotate hardware init tables
- Exit: split completes with all segments; assets extracted

**Phase 3 — Symbols.**
Goal: populate labels and validate with runtime evidence.
- Populate symbols from known addresses: entry point, interrupt vectors
  (0x000000-0x0001FF), V-blank/H-blank handlers
- Discover game state variables via bizhawk memory watch
- Re-run split after symbol updates
- Exit: symbol file populated; runtime addresses verified

**Phase 4 — Runtime validation.**
Goal: prove disassembly accuracy with BizHawk comparison.
- Use `bizhawk.bizhawk_press_buttons` + `bizhawk.bizhawk_frame_advance` to
  navigate to known states
- Use `bizhawk.bizhawk_read_memory` at disassembly addresses to verify values
- Use `bizhawk.bizhawk_save_state` / `bizhawk.bizhawk_load_state` for
  checkpoint comparison between code paths
- Exit: key addresses verified; discrepancies documented

**Phase 5 — Matching.**
Goal: translate to C and validate match.
- For matching decomp: translate 68000 to C, compile, compare objdiff output
- For function-level analysis: document in `REPHAMR_STATE.md` function ledger
  (game logic, sound driver, VDP init, compression, data/jump table, unknown)
- Exit: function ledger populated; objdiff clean for matched functions

## Build Gate / Validation Gate

Before claiming disassembly accuracy or match:
1. **INSPECT** — `sega2asm.run` completed without errors.
2. **VERIFY** — key runtime addresses checked via `bizhawk.bizhawk_read_memory`.
3. **DOCUMENT** — verification results in `REPHAMR_STATE.md`.

Success may only be claimed when:
- All segments split without errors
- Runtime comparison at ≥3 key addresses confirms expected values
- Symbols populated from known addresses + bizhawk verification

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
sega2asm -c config.yaml -s symbols.txt -t charmap.tbl -v   # full split
sega2asm -c config.yaml --dry-run                           # validate config
```

MCP tools (when connected):

```text
sega2asm.detect_compression   — identify compression at ROM offset
sega2asm.plan                 — dry-run validate config.yaml
sega2asm.run                  — execute full ROM split
bizhawk.bizhawk_get_info      — ROM hash, framecount, memory domains
bizhawk.bizhawk_read_memory   — read runtime values at disasm addresses
bizhawk.bizhawk_frame_advance — step frame-by-frame
bizhawk.bizhawk_save_state    — checkpoint before risky operations
bizhawk.bizhawk_load_state    — restore checkpoint
bizhawk.bizhawk_press_buttons — set joypad state for state navigation
```

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

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| Segment fails to split | `config.yaml` | `sega2asm.plan` dry-run output | Fix segment boundaries in config |
| Unknown compression type | Asset detection | Raw bytes at offset, visual inspection | Use `sega2asm.detect_compression` |
| Disassembly missing labels | Symbols file | Branch target addresses in .asm | Add symbols; re-run split |
| Runtime values don't match disasm | Symbol/label mapping | `bizhawk.bizhawk_read_memory` at address | Verify symbol address; update symbol file |
| VDP register table not decoded | Hint config | Raw `dc.w`/`dc.l` values in .asm | Add `vdp_regs`/`vdp_cmds` hints to config |
| Text segment garbage | Charmap TBL | Raw byte values, expected text | Apply correct charmap; re-run split |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — game, ROM SHA-1, region, segment count, compression catalog, phase
- `config.yaml` — validated segment definitions (source of truth)

Optional:
- `.rehamr/evidence/genesis_recon.md` — ROM info, compression catalog, hardware notes
- `.rehamr/evidence/genesis_symbols.md` — known symbols with bizhawk verification
- `.rehamr/functions/inventory.csv` — function ledger (Phase 5)
- `symbols.txt` / `charmap.tbl` — project-specific symbol and charmap files

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, segment count, compression catalog, symbols added, verified addresses.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
