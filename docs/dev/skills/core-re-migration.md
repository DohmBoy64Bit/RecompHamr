# `core-re` Skill Migration Record

## Sources and authorities

- Legacy source read completely: `RecompHamr-Legacy-main/internal/skills/core-re.md`; it has no linked supporting resources.
- The complete current Agent Skills specification, quickstart, best-practices, description-optimization, evaluation, and script-guidance authority set, plus the official Codex skills guidance, was read in this Stage G task before this individual edit as required by the repository migration policy.

## Intent and boundary

Trigger for evidence-first startup, triage, correction, or closure of reverse engineering, decompilation, and static-recompilation work on unfamiliar artifacts. Do not trigger for ordinary software development without a binary-analysis claim, or instead of a more specific platform workflow once that workflow is known.

## Disposition

**Improved.** The migration preserves inspect-first work, explicit evidence classes, minimal changes, falsifiable verification, unknown-is-unknown discipline, repeated-failure strategy changes, evidence workspace updates, and accurate closure. It removes the obsolete `bash` assumption, the hard dependency on activating `evidence-mode`, ambiguous state-file paths, and blanket instructions to copy raw command output. It adds concrete falsification language, secret/data-minimization guidance, and clearer non-trigger boundaries.

Final skill: `internal/skills/builtin/core-re/SKILL.md`. No scripts, references, or assets are justified; the workflow is concise instruction rather than deterministic executable behavior.

## Evaluation

`internal/skills/builtin/core-re/evals/evals.json` defines eight realistic positive and eight adjacent negative trigger cases plus three output-quality cases covering weak evidence, runtime divergence, and missing artifacts. Specification parsing and client activation are covered by repository automated tests. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts.
