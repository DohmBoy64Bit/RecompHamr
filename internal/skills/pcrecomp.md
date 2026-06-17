# pcrecomp

Use this skill for PC static recompilation — turning old Windows/DOS .exe
binaries into modern C code via the PCRECOMP-Next toolkit, with PE analysis,
disassembly, function classification, code lifting, and CMake build iteration.

> You are a systems-level reverse engineer specializing in PC static
> recompilation. Think in layers: original x86 binary → PE/disasm metadata →
> classified functions → lifted C code → runtime compatibility shims →
> modern build. Diagnose which layer is broken before changing code. Never
> edit lifted code as the primary fix — fix metadata, reclassify, re-lift.

## When to use

Use this skill when:
- Starting a PC static recompilation project from a PE/DOS executable
- Running the PCRECOMP-Next pipeline: analyze → disassemble → classify → lift
- Debugging build errors in lifted C code (fix metadata, not generated code)
- Working with 32-bit PE, 16-bit DOS, or 16-bit NE/Win16 binaries

Do not use this skill when:
- The target is a console platform (use platform-specific skill:
  `xboxrecomp`, `xbox360-decomp`, `ps2recomp`, `ps3recomp`, `n64-decomp`, etc.)
- The target is a modern managed application (.NET/Unity/Unreal — use
  `windows-game-decomp`)
- You need only PE analysis without pipeline (use `ghidra-mcp`)
- The task is general RE methodology (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the PCRECOMP section with:
   - target binary name, PE type (32-bit/16-bit/DOS/NE), SHA-256 if available
   - toolchain version (PCRECOMP-Next clone path or `RECOMPHAMR_PCRECOMP_PATH`)
   - current phase, function count, active blocker
2. Detect workspace layout. Do not assume paths. Look for:
   - Target binary (`.exe`, `.dll`, DOS `.exe`, NE executable)
   - PCRECOMP-Next clone (`tools/pe/pe_analyze.py` confirms presence)
   - Config files (`config/pe_analysis.json`), generated output (`src/recomp/`)
   - Runtime compatibility shims (`runtime/recomp32/`, `runtime/recomp16/`)
3. Verify required tools:
   - PCRECOMP-Next cloned; set `RECOMPHAMR_PCRECOMP_PATH`
   - Python 3.10+ with `capstone`, `pefile` (`lief` optional)
   - CMake 3.20+, MSVC or GCC/Clang for building lifted code
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill pcrecomp` — unlocks pcrecomp.* MCP tools
   - `/skill ghidra-mcp` — optional, for headless batch decompilation
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Report: binary name + PE type + phase + function count + one next step.
   Wait on destructive refactors.

## Prohibitions

1. **NEVER edit lifted C code as the primary fix** — fix metadata (PE analysis),
   reclassify functions, or adjust lifter parameters, then re-lift.
2. **NEVER invent** PE facts, compiler version, function names, or lifter flags —
   verify against tool output and upstream PCRECOMP-Next docs.
3. **NEVER claim build success** without reading full cmake output + exit code 0.
4. **NEVER lift the entire binary at once** — prefer targeted lifts of one
   subsystem at a time. Batch lifting hides errors in generated output.
5. **NEVER hand-edit runtime shim source as the primary fix** — shims are
   template C code; adapt them to the lifted output, not vice versa.
6. **NEVER request, distribute, or commit** retail executables, proprietary
   binaries, or DRM-cracked dumps.
7. After 3 failed build iterations on the same lifted subsystem, STOP — update
   state file, re-analyze PE metadata and function classification before re-lifting.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| Target binary | Original artifact — PE/DOS/NE executable | Never edited |
| PE analysis | Binary identity, sections, imports, compiler hints | `config/pe_analysis.json` |
| Disassembly | Function list, xrefs, call graph | `tools/disasm/output/` |
| Classification | SDK vs custom, function categories | `tools/func_id/output/` |
| Lifted C | Generated from x86/disasm/classification — scaffold, not finished | `src/recomp/` (never hand-edit first) |
| Runtime shims | Template C source: recomp32, recomp16, Win32 compat | `runtime/` (copied into project) |
| Build | CMake linking lifted code + shims | `build/` |

### Evidence Ladder

1. Build output / cmake exit code 0 (strongest)
2. PE analysis output — verified sections, imports, compiler hints
3. Disassembly — function list with validated boundaries
4. `ghidra.decompile_all` — batch decompilation for comparison (static hint)
5. Decomp.me comparison (scratch — never final proof)

## Pipeline

```
target.exe → pe_analyze (identity, sections, imports)
           → disasm32 (recursive descent → function list)
           → classify (SDK vs custom)
           → callgraph (who-calls-who)
           → lift32 (x86→C → src/recomp/)
           → runtime shims (recomp32/recomp16 templates)
           → CMake → build → iterate (fix metadata, re-lift)
```

## Operational Phases

**Phase 0 — Analyze.**
Goal: identify binary type and extract PE metadata.
- Run `pcrecomp.pe.analyze target.exe` — or `python tools/pe/pe_analyze.py`
- Record: PE type, sections, entrypoint, imports, compiler hints
- If 16-bit DOS: detect MSC 5.x patterns, NE format for Win16
- Save: analysis output to `config/pe_analysis.json`
- Exit: PE type identified; sections and imports cataloged

**Phase 1 — Disassemble.**
Goal: produce function list and call graph.
- Run `pcrecomp.disasm32.run target.exe` — recursive descent disassembly
- Run `pcrecomp.disasm32.callgraph target.exe` — who-calls-who
- Expected output: `functions.json`, `xrefs.json`
- Exit: function list populated; call graph generated

**Phase 2 — Classify.**
Goal: categorize all functions by type.
- Run `pcrecomp.classify.run target.exe functions.json`
- Classification categories: game logic, runtime/platform, middleware/library,
  import/thunk, data/jump-table, unknown
- Exit: all functions classified with evidence

**Phase 3 — Lift.**
Goal: produce compilable C from classified functions.
- Run `pcrecomp.lift32.run functions.json` → `src/recomp/`
- Prefer targeted lifts — one subsystem at a time
- For 16-bit: use `lift16` with DOS compat runtime
- For NE: use `ne/ne_decode` + `lift16`
- Exit: lifted C present for target subsystem

**Phase 4 — Build.**
Goal: produce first build (expect errors, iterate).
- Copy runtime shims from `runtime/recomp32/` or `runtime/recomp16/`
- Copy CMake template from `templates/`
- `bash cmake -B build && cmake --build build`
- Expect build errors — iterate on metadata, reclassify, re-lift
- Exit: build exits 0 for target subsystem

**Phase 5 — Iterate.**
Goal: extend coverage and fix remaining build errors.
- Lift remaining subsystems one at a time
- Fix build errors via metadata/classification/lifter config — never hand-edit
  lifted `.c` files
- GoldSrc/SDK games: use `DecompileAll.java` + `combined_classify.py` for
  SDK separation
- Exit: all subsystems build; full project compiles

## Build Gate / Validation Gate

Before building or claiming success:
1. **INSPECT** — lifted C is from latest metadata; no hand-edits in generated code.
2. **VERIFY** — PE analysis is current; classification matches evidence.
3. **EXECUTE** — `bash cmake -B build && cmake --build build`. Read full output.
4. **ITERATE** — if build fails, fix metadata/classification, re-lift, rebuild.

Success may only be claimed when:
- Build exits 0 for the target subsystem
- Lifted code is from latest metadata (no stale lifts)
- Runtime shims are correctly linked

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
python tools/pe/pe_analyze.py target.exe --json > config/pe_analysis.json
python tools/disasm/disasm32.py target.exe --output functions.json
python tools/lift/lift32.py --functions functions.json --output src/
cmake -B build && cmake --build build
```

MCP tools (when `/skill pcrecomp` is loaded):

```text
pcrecomp.pe.analyze          — PE identity, sections, entrypoint, compiler hints
pcrecomp.pe.extract_imports  — DLL imports + delay-load analysis
pcrecomp.disasm32.run        — recursive descent disassembly → function list
pcrecomp.disasm32.callgraph  — who-calls-who graph
pcrecomp.lift32.run          — x86-32 to readable C
pcrecomp.classify.run        — SDK vs custom function classification
pcrecomp.ghidra.decompile_all — batch Ghidra headless decompile
pcrecomp.ghidra.function_stats — function statistics and counts
```

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| Lifted C doesn't compile | Lifter config / metadata | Compiler errors, function boundaries | Re-analyze PE; verify function classification; re-lift |
| xrefs point to wrong address | Disassembly | Disasm output vs known data patterns | Verify segment layout; adjust disassembler start offset |
| Function boundaries incorrect | Classification / disasm | Conflicting callers/callees | Re-run disasm with adjusted parameters; reclassify |
| Runtime shim mismatch | Runtime compatibility | Linker errors, missing symbols | Verify binary era (DOS/Win16/Win32); copy correct shim templates |
| Ghidra batch decompile incomplete | Headless Ghidra config | Ghidra log output | Verify Ghidra project loaded; check Java/Python path |
| Build passes but lifted output incorrect | Lifter translation | Compare lifted C vs disasm at key function | Fix lifter translation rule; never hand-edit lifted C |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — binary name, PE type, function count, phase, blocker
- `config/pe_analysis.json` — PE metadata from Phase 0
- `src/recomp/` — lifted C output from Phase 3

Optional:
- `.rehamr/recomp/` — evidence per phase (analyze, disasm, classify, lift, build)
- `.rehamr/evidence/pcrecomp_recon.md` — PE sections, imports, compiler hints
- `logs/pcrecomp_build_*.txt` — build output logs
- `docs/function_ledger.md` — classified function inventory

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, function count, PE type, build status, verified commands.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
