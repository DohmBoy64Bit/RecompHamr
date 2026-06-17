# build-fix-loop

Build failure iteration skill. Load when the project doesn't compile, tests
fail, or generated code is broken.


## Kickoff

`/skill build-fix-loop` — then "The build is failing with [error]. Diagnose and fix."

## What it teaches

- Capture the exact command and full error — not just the last line
- Identify the earliest root failure, not surface symptoms
- Make one focused fix, re-run the exact command, verify
- Stop only when verified or a new unrelated blocker is reached
- Never claim success unless the command actually passes

## When to use

During any recompilation, decompilation, or build iteration cycle. Pairs well
with `core-re` for project structure context and `evidence-mode` for tracking
what you've proven works.
