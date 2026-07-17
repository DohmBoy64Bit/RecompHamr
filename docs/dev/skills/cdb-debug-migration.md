# `cdb-debug` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/cdb-debug.md` was read completely; its optional example template is project-conditional and no Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this edit.

Trigger for CDB-based host-native Windows recompilation traces, dump analysis, MAP/PDB caller resolution, and indirect-call crash triage. Exclude guest/emulator debugging, non-Windows debuggers, and static-only questions.

**Improved.** The migration preserves wrapper-first execution, host/guest distinction, trace-before-HIT/BYPASS claims, MAP/PDB corroboration, dump/diagnostic pairing, indirect-target classification, repeated-breakpoint correction, and evidence closure. It removes obsolete `bash` wrappers and hard-coded commands/paths, adds `INCONCLUSIVE`, module/build/address-space identity, cancellation/data-minimization guidance, and rejects dispatch-table symptom masking. No generic script is bundled because project wrappers own the executable-specific contract.

Final skill: `internal/skills/builtin/cdb-debug/SKILL.md`. Its eval set has eight positive, eight negative, and three debugger-evidence cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. Actual CDB target debugging is not claimed because no suitable target/runtime fixture was available.
