# file-format-reversing

Use this skill when reverse-engineering unknown binary file formats — archives,
assets, maps, scripts, configs, model formats, texture formats, audio banks,
save files, or any custom binary/text container.

## Goal

Map unknown binary structures through evidence-backed field discovery — every
claimed field must have an offset, observed value, and supporting evidence.
Unknown bytes stay unknown until proven.

## Rules

1. **Every claimed field needs evidence.** An offset, sample value, observed
   pattern, code reference, tool output, or repeated behavior across samples.
2. **Unknown bytes stay unknown.** Do not name fields by vibes. Use
   `unknown_XX`, `tentative_*`, or HYPOTHESIS labels until evidence supports
   a stronger name.
3. **Validate against multiple samples.** A single-sample theory is a
   hypothesis — two or more samples with consistent behavior confirms it.
4. **Write a tiny parser before a full editor.** Prove you understand the
   format by dumping its structure before attempting to modify it.
5. **Keep failures documented.** A parser failure on a specific sample is
   evidence of an untested format variant — record it, don't hide it.
6. **Use real sample files where possible.** Synthetic samples test your
   hypothesis, not the format.

## Workflow

1. **Inventory samples.** Collect file hashes, sizes, and sources for all
   available samples of the format.
2. **Identify structure.** Look for: magic bytes, endian clues, size fields,
   offsets, counts, tables, compression signatures, and cross-references.
3. **Write a tiny parser/dumper.** The smallest program that reads the format
   and dumps its structure to human-readable output. Prove you understand it.
4. **Validate against multiple samples.** Run the parser on every available
   sample. Record successes and failures.
5. **Document everything.** Keep unknown ranges, edge cases, and failed samples
   documented — these are evidence, not mistakes.

## Required Output

Use the evidence taxonomy from `/skill evidence-mode`:

```md
## CONFIRMED
- <field>: offset <N>, type <T>, observed value <V> in samples <A,B,C>
- Source: parser output / hex dump / code reference

## HYPOTHESIS
- <field>: possibly <purpose> based on <pattern>
- Promotion requires: <what evidence would confirm>

## UNKNOWN RANGES
- offset <start>-<end>: unparsed bytes, observed values <X>

## FAILURES
- sample <name>: parser error at offset <N> — <description>
```

## Evidence / Artifact Targets

Required:
- `.rehamr/formats/inventory.md` — format catalog with sample hashes and sizes
- `.rehamr/formats/hypotheses.md` — unconfirmed field theories with promotion criteria
- `.rehamr/formats/parsers/` — parser/dumper source code

Optional:
- `.rehamr/formats/samples.md` — sample file inventory
- `.rehamr/formats/tests/` — parser test cases and sample files
- `REPHAMR_STATE.md` — format discovery phase, confirmed fields, active hypotheses

## Stop Conditions

Stop and gather better evidence when:
- You are about to name a field without offset, sample value, or code reference
- Your hypothesis works on one sample but fails on another — document the difference,
  do not force the hypothesis
- A field's interpretation depends on guessing its purpose without behavioral proof
- You are about to claim a format is "fully understood" with unparsed bytes remaining

## Session Close

1. Update `.rehamr/formats/inventory.md` with confirmed field discoveries.
2. Update `.rehamr/formats/hypotheses.md` with promoted/demoted hypotheses.
3. Report: confirmed fields, new hypotheses, failed samples, next 3 fields to investigate.
