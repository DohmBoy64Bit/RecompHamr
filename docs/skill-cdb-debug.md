# cdb-debug

Windows CDB (Microsoft Console Debugger) skill. Teaches the LLM how to debug
recompiled native EXEs using CDB, PowerShell wrappers, MAP files, and
diagnostic logging.


## Kickoff

`/skill cdb-debug` — then "Build with /MAP, read the CDB wrapper, trace the entrypoint."

## What it teaches

- CDB workflow: build with /MAP → read wrapper script → run CDB → classify
  HIT/BYPASS/ABORT → archive evidence
- PowerShell wrapper discipline: never invent command lines, read the script first
- ICALL crash pattern: MAP file → caller function → breakpoint → classify VA
- Diagnostic logging pairing: stderr/fprintf + CDB trace = evidence
- Crash dump analysis: `!analyze -v` + stack trace

## What it references

- `REPHAMR_STATE.md` — persistent project memory
- `bash` — CDB is invoked through bash

## When to use

Debugging recompiled native Windows executables — proving hit/miss at
specific addresses, classifying ICALL crashes, comparing CDB evidence
against static analysis (Ghidra/TOML). Used by n64-decomp Track B and
xboxrecomp.
