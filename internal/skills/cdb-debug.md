# cdb-debug

Use this skill for debugging recompiled native Windows executables with
Microsoft Console Debugger (CDB) — breakpoint verification, MAP file caller
resolution, ICALL crash triage, crash dump analysis, and diagnostic log pairing.

> CDB operates on the **host** native binary, not the guest console. It proves
> whether recompiled code is hit, bypassed, or crashing at specific addresses.
> Always pair CDB evidence with static analysis (Ghidra/TOML/symbols) —
> dynamic hit/miss alone doesn't explain why.

## When to use

Use this skill when:
- Verifying whether a recompiled function is reached at runtime (HIT/BYPASS)
- Diagnosing ICALL crashes with MAP file + breakpoint workflow
- Analyzing crash dumps (`!analyze -v`) for recompiled native EXEs
- Pairing diagnostic logging output with CDB trace evidence

Do not use this skill when:
- You need only static analysis without runtime proof (use `ghidra-mcp`)
- The target is not a Windows native EXE (use `bash` with platform debugger)
- The task is general RE methodology without CDB-specific workflow (use `core-re`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the CDB section with:
   - last CDB trace path, HIT/BYPASS/ABORT status
   - active wrapper script path, build MAP file path
   - crash table entries with guest PC and structural cause
2. Verify required tools:
   - Windows SDK Debugging Tools — `cdb.exe` on PATH (`where cdb` or `Get-Command cdb`)
   - Build with `/MAP` — produces `.map` file for address-to-function resolution
   - PowerShell wrappers — `tools/*cdb*.ps1` in the project root
3. Locate and read the wrapper script before suggesting any CDB command:
   `bash Get-Content tools/run_cdb.ps1`
   The wrapper handles symbol paths, source paths, environment setup, and
   breakpoint syntax. Never construct CDB command lines from memory.
4. Report: trace status + wrapper found + one next step.

## Prohibitions

1. **NEVER invent CDB command lines** — read the project's PowerShell wrapper
   script first. It handles symbol paths, env setup, and breakpoint syntax.
2. **NEVER claim hit/miss without reading the `.cdb.txt` output** — verify the
   trace log before recording HIT/BYPASS/ABORT.
3. **NEVER trust diagnostic logs alone** — they show what the program reports,
   not what actually executed. Always pair with CDB hit/miss evidence.
4. **NEVER assume MAP file format or function naming** — search the actual
   `.map` file for the caller address.
5. After 3 failed breakpoint attempts at the same address, STOP — update state
   file, verify the address from Ghidra/TOML before retrying.

## Mental Model

| Layer | Role | Evidence Type |
|---|---|---|
| MAP file | Maps native addresses to function names | Build artifact |
| Wrapper script | PowerShell script handling CDB setup | Project source |
| CDB trace | `.cdb.txt` — shows HIT/BYPASS/ABORT per breakpoint | Runtime proof |
| Diagnostic logs | stderr/fprintf output from recompiled binary | Corroborating evidence (not standalone) |
| Crash dump | `.dmp` file from VEH/WER | Crash analysis |
| Static analysis | Ghidra/TOML/symbols | Explains why a hit/miss occurred |

### Evidence Ladder

1. CDB trace `.cdb.txt` — HIT/BYPASS/ABORT at specific address (strongest)
2. MAP file — native address → function name resolution
3. Crash dump `!analyze -v` + stack trace
4. Diagnostic logs (stderr/fprintf) — paired with CDB evidence
5. `ghidra.decompile_function` at caller address (static context)

## Workflow

```
Build with /MAP → read wrapper script → run CDB → capture .cdb.txt →
classify HIT/BYPASS/ABORT → archive evidence in REPHAMR_STATE.md →
if BYPASS: fix TOML/stubs/recompiler → rebuild → retrace
if CRASH: analyze dump → classify cause → fix layer → rebuild → retrace
```

## Common Operations

### Read the wrapper script first
```
bash Get-Content tools/run_cdb.ps1
```
Never construct CDB command lines from memory. The wrapper handles symbol
paths, source paths, environment setup, and breakpoint syntax.

### Run a trace
```
bash .\tools\run_cdb.ps1
```
Output goes to `logs/cdb_trace.txt` (or path defined in the wrapper).

### Classify the result

| Status | Meaning | Action |
|---|---|---|
| HIT | Breakpoint at target address was reached | Record address + function in state file |
| BYPASS | Target address was never reached, no crash | Fix TOML/runtime registration — target isn't in dispatch |
| ABORT | Process crashed before reaching target | Read crash address, stack trace, classify cause |

### Capture crash evidence
```
bash cdb -z crash.dmp -c "!analyze -v; k; q"
```
If the project has crash dump capture set up (VEH or WER), analyze the dump
before restarting the trace.

## ICALL Crash Pattern

When the native EXE shows `ICALL FAIL: VA=0x........` (xboxrecomp, n64-decomp Track B):

1. Build with `/MAP` to produce the map file
2. In the map file, search the caller address to find the function name
3. Set a CDB breakpoint at the caller function's entry
4. Step through to the indirect call site, inspect the target VA
5. Classify:
   - **Garbage** (e.g. `0x1C45BA68`) — corrupted vtable → per-function guard,
     trace object init. Do not add dispatch entries.
   - **Valid code VA** — not in table → extend dispatch or add `recomp_manual.c`
     override.
   - **Kernel range** `0xFE000000+` — tier-3 `recomp_lookup_kernel`.

## Diagnostic Logging

When the recompiled binary has diagnostic logging (stderr/fprintf):

1. Run with stderr redirected: `./game.exe 2> diag.txt`
2. Read the log: `read_file diag.txt` (use offset/limit for large logs)
3. Match timestamps/addressing patterns to CDB trace addresses
4. Archive relevant log excerpts in `REPHAMR_STATE.md` alongside CDB evidence

Diagnostic logs alone are **not proof** — they show what the program reports,
not what actually executed. Always pair with CDB hit/miss evidence.

## Tool Quick Reference

```bash
# Verified commands only. Do not invent CDB command lines.
bash Get-Content tools/run_cdb.ps1                  # read wrapper script first
bash .\tools\run_cdb.ps1                            # run project CDB trace
bash cdb -z crash.dmp -c "!analyze -v; k; q"        # analyze crash dump
./game.exe 2> diag.txt                              # capture diagnostic log
```

## Failure Patterns

| Symptom | Likely Cause | Evidence to Gather | Fix |
|---|---|---|---|
| Breakpoint never hit (BYPASS) | Missing dispatch | `.cdb.txt` trace, MAP file | Fix TOML/runtime — target not in dispatch table |
| ABORT before breakpoint | Upstream crash | Crash dump `!analyze -v`, stack trace | Fix crash at earlier address; retrace |
| MAP file doesn't resolve caller | MAP generation | MAP file path, build command | Verify `/MAP` flag in build; rebuild |
| ICALL garbage VA (e.g. `0x1C45BA68`) | Corrupted vtable | MAP file, `g_icall_trace[]` ring buffer | Per-function vtable guard; do not add dispatch entry |
| Wrapper script not found | Project layout | Directory listing for `tools/*cdb*.ps1` | Adapt from template; record in state file |
| stderr log contradicts CDB trace | Logging vs reality | `.cdb.txt` trace (authoritative) | Trust CDB over diagnostic log |

## Archiving Evidence

After each CDB session, update `REPHAMR_STATE.md`:
- **Last CDB trace:** path + HIT/BYPASS/ABORT
- **Crash table:** guest PC, structural cause, fix layer, status
- **Active commands:** verbatim wrapper invocation that worked

Template: `examples/cdb-trace-evidence-template.txt` if the project has one.

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — last trace path, HIT/BYPASS/ABORT status, crash table
- `logs/cdb_trace.txt` — CDB trace output (or wrapper-defined path)

Optional:
- `.rehamr/evidence/cdb_trace_log.md` — trace summaries with MAP file resolution
- `.rehamr/evidence/cdb_crash_log.md` — crash table with dump analysis

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — trace path, HIT/BYPASS/ABORT status, fix layer, crash table.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
