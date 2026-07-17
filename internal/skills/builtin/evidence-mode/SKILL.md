---
name: evidence-mode
description: Classify and cite reverse-engineering findings before naming symbols, documenting binary behavior, or promoting claims to fact. Use when reviewing or writing RE evidence and when separating confirmed facts from hypotheses, TODOs, and blockers; do not use merely to format citations for ordinary prose or when no artifact-backed technical claim is being made.
compatibility: RecompHamr; works with any static or runtime evidence source.
---
# Evidence classification

Apply this taxonomy to every material reverse-engineering claim:

- **CONFIRMED** — reproduced or directly supported by a specific artifact. Cite the file and location, binary identity and address, command and relevant output, symbol record, trace, or log.
- **HYPOTHESIS** — plausible but unproven. State the observation behind it and the narrow result that would confirm or falsify it.
- **TODO** — a concrete next action and the evidence it should produce.
- **BLOCKED** — the exact missing artifact, tool, permission, runtime, dependency, or decision and what resolves it.

Do not infer semantic names from instruction shapes, proximity, intuition, or decompiler guesses alone. Naming a function, field, structure, asset, or section requires evidence such as symbols, callers/callees, xrefs, data flow, strings used in context, format comparisons, or runtime observation. Preserve earlier notes unless stronger evidence explicitly supersedes them; record the correction rather than silently rewriting history.

Use short excerpts or precise references, not indiscriminate dumps. Never copy credentials, private prompts, or unrelated user data into evidence. Distinguish the command that was run from the interpretation of its output.

Before closing:

1. Check every CONFIRMED item has a reproducible source.
2. Demote unsupported statements to HYPOTHESIS or BLOCKED.
3. Record corrections and meaningful findings in the appropriate `.rehamr/` evidence file when that workspace exists.
4. Report what changed classification and what evidence is still needed.

Refuse to present guessed addresses, layouts, behavior, or symbol names as confirmed.
