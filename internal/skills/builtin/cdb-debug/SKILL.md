---
name: cdb-debug
description: Debug a recompiled native Windows executable with Microsoft CDB using project wrappers, MAP/PDB symbols, breakpoint traces, crash dumps, and paired diagnostics. Use to prove host-native HIT/BYPASS/ABORT behavior or triage indirect-call crashes; do not use for guest-console debugging, non-Windows targets, or static analysis with no runtime question.
compatibility: Windows with cdb.exe from Debugging Tools for Windows and a debuggable native build; project wrapper and MAP/PDB artifacts are strongly recommended.
---
# CDB debugging for recompiled Windows hosts

CDB observes the native host executable, not guest virtual addresses directly. Establish the exact executable/PDB/MAP identities and distinguish guest addresses from relocated native addresses.

Before running anything:

1. Locate and read the project's CDB PowerShell wrapper and build configuration. Do not invent quoting, symbol, source-path, environment, or breakpoint syntax.
2. Confirm `cdb.exe`, the exact executable, symbols/MAP, wrapper inputs, expected trace path, and whether dump capture is configured.
3. Preserve the current failure and target address with its address space.

Run the wrapper with existing cancellation/time bounds and read the resulting trace before classifying:

- **HIT** — the intended native breakpoint was observed.
- **BYPASS** — the process completed the relevant path without the breakpoint; this proves only that the breakpoint was not observed under that run.
- **ABORT** — an earlier crash or termination prevented the target observation.
- **INCONCLUSIVE** — symbols, breakpoint binding, timeout, process selection, or trace integrity is uncertain.

Resolve native caller addresses from the actual MAP/PDB rather than assumed formats. Pair diagnostic logs with debugger evidence; application logging alone cannot prove control flow. For a dump, use the project-approved command or documented CDB dump workflow, retain exception code/address, loaded-module identity, stack, and decisive analysis—not secret-bearing full dumps in project text.

For indirect-call failures, inspect the call site and target provenance before changing dispatch. A plausible guest code address may indicate missing registration; a nonsensical target may indicate corruption or object initialization; a kernel/import range needs its owning bridge. Never add a dispatch entry solely to silence a bad pointer.

After repeated binding failures, verify address translation, module load base, symbol status, wrapper commands, and build identity before retrying. Record trace path, command, executable hash/build, HIT/BYPASS/ABORT/INCONCLUSIVE status, static corroboration, fix layer, and remaining uncertainty in the evidence workspace.
