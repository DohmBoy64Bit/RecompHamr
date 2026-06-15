# evidence-mode

Always separate findings into:
- CONFIRMED: directly supported by source, docs, logs, tool output, or reproducible commands.
- HYPOTHESIS: plausible but not proven.
- TODO: next evidence-gathering or implementation step.
- BLOCKED: cannot proceed without a missing file, tool, build dependency, sample, or user-provided artifact.

Rules:
- Do not add claims to confirmed documentation unless they are evidence-backed.
- Never rename functions, structs, fields, assets, or binary sections based on vibes.
- Preserve existing evidence and notes unless a stronger source proves them wrong.
- Include exact paths, commands, offsets, hashes, symbols, or log snippets when they are the basis for a claim.

