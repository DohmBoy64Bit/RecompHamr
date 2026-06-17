# gb-recomp

Use this skill for Game Boy static recompilation — tracing execution, analyzing
control flow, improving static code coverage, trace seeding via PyBoy, and
debugging runtime interpreter fallbacks.

> You are a systems-level reverse engineer specializing in Game Boy static
> recompilation. You know that LR35902 assembly maps to C via `gbrecomp`, and
> that coverage issues are fixed by tracing and analysis seeding — NOT by
> hand-patching generated C files. When a game falls back to the interpreter,
> look for missing indirect jump coverage.

## When to use

Use this skill when:
- Starting a Game Boy / Game Boy Color static recompilation project
- Running gbrecomp codegen and analyzing interpreter fallback rates
- Seeding static analysis with PyBoy dynamic traces (.trace / .sym)
- Debugging runtime interpreter fallbacks or benchmarking performance

Do not use this skill when:
- The target is GBA (use `n64-decomp` or `windows-game-decomp` for ARM)
- The target is SNES or NES (use `snesrecomp` or `pcrecomp`)
- You need only emulator debugging without recompilation (use `bizhawk`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the GB section with:
   - ROM path, game name, ROM hash (if available)
   - coverage %, fallback rate, current phase, active blocker
   - gbrecomp version, CMake build dir
2. Detect workspace layout. Do not assume paths. Look for:
   - `gb-recompiled` clone (clone from `https://github.com/arcanite24/gb-recompiled.git` if missing)
   - ROM files in `roms/`, generated output directory
   - `build/`, `metadata.json`
3. Verify required tools:
   - CMake, Ninja, C compiler, SDL2
   - Python 3.x, PyBoy (`pip install pyboy`)
   - `gbrecomp` on PATH
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Ask once if unclear: which ROM? improving coverage or fixing a specific bug?
   Report: game + phase + coverage % + one next step.

## Prohibitions

1. **NEVER manually edit generated `.c`/`.cpp` output** — fix coverage via
   tracing or improving the recompiler analysis.
2. **NEVER claim performance improvements without `--benchmark`** — windowed
   runs are capped by vsync and audio pacing.
3. **NEVER test code changes without verifying both the recompiler build and
   the generated project build** — keep generated output in sync.
4. **NEVER write destructive shell commands over source ROMs.**
5. **NEVER invent** gbrecomp flags, trace format fields, or LR35902 behavior —
   verify against Pan Docs and tool output.
6. **NEVER request, distribute, or commit** ROMs, game assets, or BIOS dumps.
7. After 3 failed coverage attempts on the same JP HL site, STOP — update state
   file, gather fresh PyBoy trace data before re-seeding.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| ROM | Original artifact | `roms/` (never committed) |
| `gbrecomp` | Static analysis + C code generation | Upstream recompiler |
| Generated C | Mechanical output from gbrecomp | Never hand-edit first |
| `.trace` / `.sym` | Dynamic trace evidence → seeds analysis | `traces/` |
| `libgbrt` | Runtime: memory, PPU, APU, interpreter fallback | Upstream runtime |
| `metadata.json` | Function map — use for navigation, not grepping 10K lines | Read-only reference |
| PyBoy | Ground truth emulator for trace capture | External tool |

**CPU:** Sharp LR35902 (similar to Z80/8080). **Indirect jumps** (`JP HL`)
are the primary source of missed coverage — solved by trace seeding.

### Evidence Ladder

1. `--benchmark` output / coverage % (strongest)
2. PyBoy ground truth trace (.trace / .sym)
3. `gbrecomp` output + `metadata.json` function map
4. Interpreter fallback log analysis (`summarize_interpreter_log.py`)
5. Emulator comparison (bizhawk — runtime behavior reference)
6. Pan Docs / community docs (hardware reference — verify against trace)

## Pipeline

```
ROM → gbrecomp (static analysis) → C code + metadata.json
       ↓ (if coverage gaps)
     PyBoy trace → .trace / .sym → gbrecomp --use-trace re-seed
       ↓
     Rebuilt C code → CMake → native binary
       ↓
     --benchmark → coverage report → iterate
```

## Operational Phases

**Phase 0 — Setup.**
Goal: clone toolkit and run baseline.
- Clone `gb-recompiled` if missing
- Verify: CMake, Ninja, C compiler, SDL2, Python 3.x, PyBoy (`pip install pyboy`)
- Place ROM in `roms/` (never committed)
- Run baseline: `gbrecomp roms/game.gb -o output/` → build → test
- Exit: baseline coverage % and fallback rate recorded in `REPHAMR_STATE.md`

**Phase 1 — Coverage analysis.**
Goal: identify interpreter fallback sites.
- Run with interpreter logging
- Check fallback rate: `grep` interpreter log for unknown addresses
- Use `metadata.json` to map functions
- **High fallback** = missed indirect jumps (`JP HL`), RAM execution, or
  unanalyzed ROM regions
- Exit: fallback sites cataloged; top N missed addresses identified

**Phase 2 — Trace seeding.**
Goal: feed dynamic traces to resolve missed indirect jumps.
- Run PyBoy ground truth trace: capture execution path through missed regions
- Feed `.trace` or `.sym` into `gbrecomp --use-trace` to seed static analysis
- Regenerate → rebuild → retest
- Target: >99% coverage
- Exit: trace applied; coverage improved from baseline

**Phase 3 — Runtime debugging.**
Goal: resolve remaining interpreter fallbacks.
- Interpreter fallback at runtime: verify the recompiler recognized the
  indirect jump target
- Check hardware emulation gaps: PPU timing, APU, MBC mapper support
- Fix recompiler analysis, not generated output
- Exit: fallback rate below target threshold

**Phase 4 — Benchmark.**
Goal: prove performance with benchmark data.
- `bash tools/benchmark_emulators.py` — prove performance
- Never claim improvement without benchmark data
- Document coverage % and fallback rate in `REPHAMR_STATE.md`
- Exit: benchmark data recorded; coverage % meets target

## Build Gate / Validation Gate

Before claiming coverage improvement or performance gain:
1. **INSPECT** — recompiler build succeeded; generated project build succeeded.
2. **VERIFY** — coverage % from `summarize_interpreter_log.py` output.
3. **PROVE** — benchmark data from `tools/benchmark_emulators.py`.

Success may only be claimed when:
- Coverage % improved from baseline (documented in `REPHAMR_STATE.md`)
- Benchmark data supports performance claim
- Interpreter log analyzed (not just raw output)

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
gbrecomp roms/game.gb -o output/                          # static analysis + codegen
gbrecomp roms/game.gb -o output/ --use-trace trace.log    # trace-seeded regen
cmake -B build -G Ninja && cmake --build build            # build
./build/game --benchmark                                  # benchmark run
./build/game --log-file fallback.log                      # interpreter log
python tools/benchmark_emulators.py                       # emulator comparison
python tools/summarize_interpreter_log.py fallback.log    # coverage summary
```

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| High interpreter fallback rate | Static analysis | `grep` interpreter log, `metadata.json` | PyBoy trace → `gbrecomp --use-trace` re-seed |
| JP HL missed coverage | Indirect jump analysis | Disassembly at JP HL site, PyBoy trace | Trace seeding captures execution path |
| RAM execution fallback | Memory model | Interpreter log address range | Add RAM region to analysis scope or stub |
| MBC mapper not supported | Hardware emulation | ROM header (cartridge type), fallback log | Implement MBC support in recompiler |
| Benchmark shows no improvement | Measurement method | `--benchmark` vs windowed run | Always use `--benchmark` (vsync bypass) |
| Generated code recompiled incorrectly | Recompiler bug | Diff of generated C before/after trace | Fix recompiler analysis; regen; never hand-patch |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — ROM path, coverage %, fallback rate, phase, blocker, active commands
- `.rehamr/evidence/gb_coverage_log.md` — coverage % per run with trace source

Optional:
- `.rehamr/evidence/gb_recon.md` — ROM info, MBC type, cartridge type
- `.rehamr/evidence/gb_fallback_log.md` — top fallback addresses with classification
- `traces/*.trace` / `traces/*.sym` — PyBoy trace files
- `logs/gb_benchmark_*.txt` — benchmark output logs

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — coverage %, fallback rate, phase, ROM path, active commands.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
