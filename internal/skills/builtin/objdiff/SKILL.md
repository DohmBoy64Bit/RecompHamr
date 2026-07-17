---
name: objdiff
description: Configure and interpret encounter/objdiff-style object-code comparisons for matching decompilation, including unit selection, compiler/build identity, mismatch classification, and progress evidence. Use when a project explicitly uses objdiff or object-level matching; do not use for runtime parity, source-only review, or projects with a different authoritative matcher.
compatibility: Requires the project's pinned objdiff-compatible tool, configuration, target/base artifacts, and reproducible build; commands and schemas must be verified against that version.
---
# Objective object-code matching

First inspect the project's matching contract, pinned tool/version, configuration, build command, compiler/linker identity, target provenance, and generated artifact paths. Do not invent `objdiff` commands, configuration fields, supported architectures, or status semantics from another version.

Before interpreting a result, prove both comparison inputs:

1. The target object or extracted code/data corresponds to the exact original build, address/unit, section layout, and expected relocation treatment.
2. The base object was freshly built from the current source with the pinned compiler, flags, environment, and build graph.
3. The matcher configuration maps the intended unit and symbols, and its normalization settings match the project's accepted definition of “matching.”

Run the smallest relevant unit through the project's documented non-interactive comparison path. Classify outcomes separately:

- **match** — the tool's defined comparison passed for that identified unit and configuration;
- **mismatch** — retain the structured instruction/data/relocation differences;
- **build failure** — no comparison conclusion is available;
- **missing/stale input** — repair provenance or configuration before comparing;
- **inconclusive** — tool/config/version/output integrity is uncertain.

For a mismatch, determine whether divergence comes from source shape, compiler or flags, ABI/calling convention, register allocation/scheduling, constants, sections/alignment, relocations, symbols, stale generated inputs, or an incorrect function/data boundary. Change one supported variable at a time and rerun the same unit.

A unit match proves only that unit under the recorded comparison definition. Project completion additionally requires coverage of every required unit/data surface and any separate link/ROM/runtime contracts. Percentages are inventory metrics; expose excluded, missing, failed, and unverified units beside the numerator and denominator.

Preserve machine-readable output when available, but bound logs and avoid proprietary object contents or source snippets in shared reports. Record tool/config hashes, target/base identities, build command and result, unit, comparison status, decisive differences, change made, and remaining uncertainty. Stage G assumes no MCP server and never installs or launches objdiff merely because this skill was activated.
