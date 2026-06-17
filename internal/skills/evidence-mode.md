# evidence-mode

Use this skill when making claims, naming symbols, documenting findings, or
updating evidence files — any time you are about to assert something as fact.

## Goal

Enforce strict evidence classification: separate confirmed facts from
hypotheses, prevent hallucinated claims, and require citation for every
assertion that enters project documentation.

## Rules

1. **Classify every finding.** Use the taxonomy below — never present an
   unclassified statement as fact.
2. **Never rename functions, structs, fields, assets, or binary sections
   based on vibes.** A name requires disassembly, xref, string reference,
   or symbol table evidence.
3. **Preserve existing evidence and notes** unless a stronger source proves
   them wrong. Overwriting weak evidence with a guess is worse than leaving
   the weak evidence in place.
4. **Include exact paths, commands, offsets, hashes, symbols, or log
   snippets** when they are the basis for a claim. A claim without a
   citation is not evidence.
5. **Do not add claims to confirmed documentation** unless they are directly
   supported by source code, build output, tool output, logs, or
   reproducible commands.

## Workflow

1. Before making a claim, classify it: CONFIRMED, HYPOTHESIS, TODO, or BLOCKED.
2. If CONFIRMED, include the exact source: command, file path, offset, hash,
   tool output, or log snippet that proves it.
3. If HYPOTHESIS, document what narrow evidence would promote it to CONFIRMED.
4. If TODO, document the specific next action and expected evidence.
5. If BLOCKED, document the missing tool, file, dependency, sample, or
   user decision preventing progress.

## Required Output

```md
## CONFIRMED
- <evidence-backed facts with source citation>
- Source: <command / file path / hash / offset / tool output>

## HYPOTHESIS
- <plausible but unproven idea>
- Promotion requires: <what evidence would confirm this>

## TODO
- <next evidence-gathering or implementation step>
- Expected evidence: <what output would prove completion>

## BLOCKED
- <missing file, tool, build dependency, sample, or user-provided artifact>
- Resolution: <what must happen to unblock>
```

## Evidence / Artifact Targets

- `REPHAMR_STATE.md` — updated with CONFIRMED facts and evidence citations
- `.rehamr/evidence/` — command outputs, tool logs, disassembly excerpts
- `.rehamr/functions/inventory.csv` — function classifications with evidence sources
- `.rehamr/CHANGELOG.md` — meaningful evidence-backed changes (when `/init-re` has run)

## Stop Conditions

Stop and refuse to write when:
- You are about to write a claim as CONFIRMED without a source citation
- You are about to rename a symbol without disassembly, xref, or string
  reference evidence
- You are about to overwrite existing evidence with a weaker claim
- A claim depends on guessing binary behavior you have not inspected
- You cannot provide exact paths, offsets, or commands for a claim

## Session Close

1. Verify all CONFIRMED claims have source citations.
2. Update `REPHAMR_STATE.md` or evidence files with evidence-backed facts only.
3. Report: verified facts, promoted hypotheses, remaining TODO/BLOCKED items,
   next 3 evidence-gathering steps.
