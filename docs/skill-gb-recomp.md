# gb-recomp

Game Boy static recompilation skill using arcanite24/gb-recompiled. Trace-guided
coverage improvement with PyBoy ground truth.

## What it teaches

- Pipeline: ROM → gbrecomp → C code + metadata.json → CMake → native binary
- Trace seeding: PyBoy dynamic trace → `.trace`/`.sym` → gbrecomp `--use-trace`
- Interpreter fallback analysis: missed indirect jumps (`JP HL`), RAM execution
- 5 operational phases: Setup → Coverage → Trace → Debug → Benchmark
- Prohibitions: never hand-patch generated C, benchmark before claiming perf
- Key commands: gbrecomp, cmake, benchmark_emulators.py, summarize_interpreter_log.py

## What it references

- `REPHAMR_STATE.md` — persistent project memory

## When to use

Game Boy / Game Boy Color static recompilation — improving static code
coverage, resolving interpreter fallbacks, trace-guided analysis seeding,
or validating recompiler improvements with benchmark data.
