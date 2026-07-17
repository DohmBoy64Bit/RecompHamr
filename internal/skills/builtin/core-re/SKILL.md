---
name: core-re
description: Establish an evidence-first workflow for reverse engineering, decompilation, or static recompilation of an unfamiliar binary or codebase. Use at the start of RE work and when claims, names, edits, or next steps need grounding in artifacts; do not use for ordinary application development with no binary-analysis question or as a substitute for a platform-specific workflow.
compatibility: RecompHamr with filesystem and process tools; platform-specific analysis tools are optional.
---
# Core reverse-engineering workflow

Start by identifying the source of truth: target binaries, hashes, build files, scripts, configuration, logs, generated artifacts, and existing evidence. Read before editing.

For each material statement, classify it as:

- **CONFIRMED** — directly supported by a cited artifact, address, symbol, command, log, or reproducible observation.
- **HYPOTHESIS** — plausible but not yet proven; state what would falsify it.
- **TODO** — a concrete evidence-gathering or implementation step.
- **BLOCKED** — a named missing artifact, tool, permission, runtime, or user decision.

Use the smallest useful iteration:

1. Inspect the relevant artifact and record exact paths or identities.
2. Capture pre-change evidence sufficient to reproduce the observation.
3. Make one focused change only when the evidence supports it.
4. Run the narrowest check that would fail if the claim or change were wrong.
5. Record confirmed results and remaining uncertainty.

Never invent binary behavior, function boundaries, types, symbols, offsets, or format fields. Keep unknowns unknown until static or runtime evidence supports a name. If two attempts fail for the same reason, stop repeating them and select a new evidence-gathering strategy.

When the evidence workspace exists, update `.rehamr/EVIDENCE.md`, a focused file under `.rehamr/evidence/`, `.rehamr/CHANGELOG.md`, or `.rehamr/REPHAMR_STATE.md` as appropriate. Preserve user content and avoid recording secrets or unrelated data.

Finish with a concise user-facing record of changed artifacts, exact verification, confirmed findings, hypotheses, blockers, and the next evidence-producing steps. When asked to close a session, produce that final record directly rather than narrating private deliberation or merely requesting artifacts. Treat an observed divergence as a failed runtime check even when the build passes, and retain the exact last-known-good and first-divergent observations. Do not claim success from a build alone when runtime or binary equivalence is part of the task.
