# cdb-debug

Use this skill for debugging recompiled native Windows executables with
Microsoft Console Debugger (CDB). CDB is part of the Windows SDK Debugging
Tools and runs from the command line — no GUI required.

> CDB operates on the **host** native binary, not the guest console. It proves
> whether recompiled code is hit, bypassed, or crashing at specific addresses.
> Always pair CDB evidence with static analysis (Ghidra/TOML/symbols) —
> dynamic hit/miss alone doesn't explain why.

## Prerequisites

- **Windows SDK** (Debugging Tools) — `cdb.exe` on PATH
- **Build with `/MAP`** — produces `.map` file mapping addresses to function names
- **PowerShell wrappers** — `tools/*cdb*.ps1` scripts in the project (never
  invent `cdb.exe` command lines — read the wrapper script first)

## Boot

1. Read `REPHAMR_STATE.md` — populate CDB section if needed (last trace path,
   HIT/BYPASS/ABORT status).
2. Verify: `which cdb` or `Get-Command cdb` returns a path.
3. Locate project wrappers: `ls tools/*cdb*.ps1`. Read the wrapper script
   before suggesting any CDB command.
4. Report trace status + one next step.

## Workflow

```
Build with /MAP → read wrapper script → run CDB → capture .cdb.txt →
classify HIT/BYPASS/ABORT → archive evidence in REPHAMR_STATE.md →
if BYPASS: fix TOML/stubs/recompiler → rebuild → retrace
```

## Common Patterns

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

## ICALL Crash Pattern (xboxrecomp, n64-decomp Track B)

When the native EXE shows `ICALL FAIL: VA=0x........`:

1. Build with `/MAP` to produce the map file
2. In the map file, search the caller address to find the function name
3. Set a CDB breakpoint at the caller function's entry
4. Step through to the indirect call site, inspect the target VA
5. Classify: garbage (corrupted vtable) vs valid code (missing dispatch) vs
   kernel range (0xFE000000+)

## Diagnostic Logging (stderr/fprintf)

When the recompiled binary has diagnostic logging:

1. Run with stderr redirected: `./game.exe 2> diag.txt`
2. Read the log: `read_file diag.txt` (use offset/limit for large logs)
3. Match timestamps/addressing patterns to CDB trace addresses
4. Archive relevant log excerpts in `REPHAMR_STATE.md` alongside CDB evidence

Diagnostic logs alone are **not proof** — they show what the program reports,
not what actually executed. Always pair with CDB hit/miss evidence.

## Archiving Evidence

After each CDB session, update `REPHAMR_STATE.md`:
- **Last CDB trace:** path + HIT/BYPASS/ABORT
- **Crash table:** guest PC, structural cause, fix layer, status
- **Active commands:** verbatim wrapper invocation that worked

Template: `examples/cdb-trace-evidence-template.txt` if the project has one.

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — trace path, HIT/BYPASS/ABORT status, fix layer.
3. **VERIFY** — read back state file.
