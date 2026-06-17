# core-re

Use this skill when starting any reverse-engineering or recompilation task
on an unfamiliar codebase — before loading platform-specific skills.

## Goal

Establish an evidence-first workflow: inspect before acting, verify before
claiming, and never guess binary behavior without tool-backed evidence.

## Rules

1. **Inspect before editing.** Identify source-of-truth files first: README,
   build files, scripts, docs, logs, configs, and generated artifacts.
2. **Capture evidence before changing behavior.** Record what you see before
   making modifications — commands, tool output, file state.
3. **Make the smallest useful edit.** One focused change per iteration.
4. **Run the narrowest useful verification.** The check must fail when the
   thing is broken.
5. **Never guess binary behavior.** If the task depends on binary facts, use
   tools to inspect the binary or mark the point as HYPOTHESIS. Unknown
   bytes, unknown functions, and unknown structures stay unknown until
   evidence supports a name.
6. **Update evidence files.** Write to `.rehamr/CHANGELOG.md` or
   `.rehamr/EVIDENCE.md` when `/init-re` has been run.

## Workflow

1. Inspect repository shape with targeted commands (`ls`, `read_file` on key
   files, `bash` for tool output).
2. Identify source-of-truth files and record paths in `REPHAMR_STATE.md`.
3. Capture evidence before changing behavior — commands, output, file hashes.
4. Make the smallest useful edit that moves the task forward.
5. Run the narrowest useful verification — the check must fail when broken.
6. Update evidence files or state file after meaningful changes.

## Required Output

Classify every finding using the evidence taxonomy from `/skill evidence-mode`:

```md
## CONFIRMED
- <evidence-backed facts with source citation>

## HYPOTHESIS
- <plausible but unproven ideas>

## TODO
- <next evidence-gathering or implementation steps>

## BLOCKED
- <missing tools, files, permissions, or user-provided artifacts>
```

## Evidence / Artifact Targets

- `REPHAMR_STATE.md` — current phase, active blocker, learned patterns
- `.rehamr/CHANGELOG.md` — meaningful changes per session (if `/init-re` has run)
- `.rehamr/evidence/` — command outputs, tool logs, confirmed facts

## Stop Conditions

Stop and gather better evidence when:
- A claim depends on binary behavior you have not inspected
- You are about to rename a function, struct, or symbol without disassembly/xref proof
- The same edit or command fails the same way twice — the approach is wrong, not
  your luck. Change strategy.
- You cannot determine the source of truth for a file or configuration

## Session Close

1. Update `REPHAMR_STATE.md` or `.rehamr/CHANGELOG.md` with evidence-backed facts.
2. Report changed files, verified commands, remaining blockers, and next 3
   concrete steps.
