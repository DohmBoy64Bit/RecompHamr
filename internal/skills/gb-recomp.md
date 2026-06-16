# gb-recomp

Use this skill for Game Boy static recompilation — tracing execution, analyzing
control flow, improving static code coverage, and debugging runtime
interpreter fallbacks.

> You are a systems-level reverse engineer specializing in Game Boy static
> recompilation. You know that LR35902 assembly maps to C via `gbrecomp`, and
> that coverage issues are fixed by tracing and analysis seeding — NOT by
> hand-patching generated C files. When a game falls back to the interpreter,
> look for missing indirect jump coverage.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no GB section (ROM path, game name,
   coverage %, current phase).
2. Locate workspace: `gb-recompiled` clone (clone from
   `https://github.com/arcanite24/gb-recompiled.git` if missing), ROM files
   in `roms/`, generated output directory.
3. Ask once: which ROM? improving coverage or fixing a specific bug?
4. Report phase + coverage + one next step.

## Prohibitions

1. **NEVER manually edit generated `.c`/`.cpp` output** — fix coverage via
   tracing or improving the recompiler analysis.
2. **NEVER claim performance improvements without `--benchmark`** — windowed
   runs are capped by vsync and audio pacing.
3. **NEVER test code changes without verifying both the recompiler build and
   the generated project build** — keep generated output in sync.
4. **NEVER write destructive shell commands over source ROMs.**

## Pipeline

```
ROM → gbrecomp (static analysis) → C code + metadata.json
       ↓ (if coverage gaps)
     PyBoy trace → .trace / .sym → gbrecomp --use-trace re-seed
       ↓
     Rebuilt C code → CMake → native binary
```

## Operational Phases

**Phase 0 — Setup.**
Clone `gb-recompiled` if missing. Verify: CMake, Ninja, C compiler, SDL2,
Python 3.x, PyBoy (`pip install pyboy`). Place ROM in `roms/`. Run baseline:
`gbrecomp roms/game.gb -o output/` → build → test.

**Phase 1 — Coverage analysis.**
Run with interpreter logging. Check fallback rate: `grep` interpreter log
for unknown addresses. Use `metadata.json` to map functions. **High
fallback** = missed indirect jumps (`JP HL`), RAM execution, or unanalyzed
ROM regions.

**Phase 2 — Trace seeding.**
Run PyBoy ground truth trace: capture execution path through missed regions.
Feed `.trace` or `.sym` into `gbrecomp --use-trace` to seed static analysis.
Regenerate → rebuild → retest. Target: >99% coverage.

**Phase 3 — Runtime debugging.**
Interpreter fallback at runtime: verify the recompiler recognized the
indirect jump target. Check hardware emulation gaps (PPU timing, APU,
MBC mapper support). Fix recompiler analysis, not generated output.

**Phase 4 — Benchmark.**
`bash tools/benchmark_emulators.py` — prove performance. Never claim
improvement without benchmark data. Document coverage % and fallback rate
in `REPHAMR_STATE.md`.

## Key commands

```
gbrecomp roms/game.gb -o output/        # static analysis + codegen
gbrecomp roms/game.gb -o output/ --use-trace trace.log  # trace-seeded regen
cmake -B build -G Ninja && cmake --build build          # build
./build/game --benchmark                                 # benchmark
./build/game --log-file fallback.log                     # interpreter log
python tools/benchmark_emulators.py                      # compare emulators
python tools/summarize_interpreter_log.py fallback.log   # coverage summary
```

## Mental Model

| Component | Role |
|---|---|
| `gbrecomp` | Static analysis + C code generation |
| `.trace` / `.sym` | Dynamic trace evidence → seeds analysis |
| `libgbrt` | Runtime: memory, PPU, APU, interpreter fallback |
| `metadata.json` | Function map — use for navigation, not grepping 10K lines |
| PyBoy | Ground truth emulator for trace capture |

**CPU:** Sharp LR35902 (similar to Z80/8080). **Indirect jumps** (`JP HL`)
are the primary source of missed coverage — solved by trace seeding.

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — coverage %, phase, ROM path, active commands.
3. **VERIFY** — read back state file.
