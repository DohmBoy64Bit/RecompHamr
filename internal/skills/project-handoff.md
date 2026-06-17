# project-handoff

Use this skill when preparing context for another AI, another developer, or a
future session — at milestones, before stepping away from a multi-session
project, or when handing off between tools.

## Goal

Produce a compact, evidence-backed handoff document that lets the next agent
or developer resume work without re-discovering project state.

## Rules

1. **Write for the next person.** Not a log of what you did — a map of what
   they need to know. Include paths, commands, and evidence, not narration.
2. **Separate facts from guesses.** Put uncertain items under HYPOTHESIS or
   TODO. CONFIRMED items must have a source citation.
3. **Be specific.** "Fixed the crash" is useless. "Fixed ICALL crash in
   sub_001B4170 by adding dispatch entry in recomp_manual.c" is useful.
4. **Keep it compact.** The next agent has limited context. 50-80 lines is
   the sweet spot — enough to resume, not enough to overwhelm.
5. **Include exact commands that work.** Verbatim build commands, wrapper
   invocations, and tool calls — never paraphrase from memory.

## Required Output

Write to `REPHAMR_STATE.md` or a dedicated handoff file:

```md
## Project Goal
- <one sentence describing the project's purpose>

## Current State
- **Track / Phase:** <e.g., N64 Track A Phase 3>
- **Confirmed findings:** <what is proven, with evidence sources>
- **Active hypotheses:** <what is suspected but unproven>

## Toolchain & Environment
- <compiler, SDK, tool versions, build directory, active MCP servers>
- <exact build command that works>
- <exact run command that works>

## Recent Actions
- **Changed:** <file paths + short reason>
- **Verified:** <command + result>
- **Evidence updated:** <file paths>

## Open Blockers
| Blocker | Status | Evidence | Next Step |
|---|---|---|---|
| <description> | BLOCKED/IN PROGRESS | <source> | <action> |

## Next 3 Concrete Steps
1. <action> — expected evidence: <what proves completion>
2. <action> — expected evidence: <what proves completion>
3. <action> — expected evidence: <what proves completion>
```

## Evidence / Artifact Targets

- `REPHAMR_STATE.md` — primary handoff target (persistent memory)
- `.rehamr/CHANGELOG.md` — session-by-session change log (when `/init-re` has run)

## Stop Conditions

Do not complete the handoff until:
- Every CONFIRMED claim has a source citation
- Every command listed has been run successfully at least once
- Every blocker has a concrete next step (not "investigate" — what tool,
  what address, what expected evidence)
- HYPOTHESIS items are clearly separated from CONFIRMED facts

## Session Close

1. Write the handoff document to `REPHAMR_STATE.md`.
2. Verify every command and file path in the handoff is accurate.
3. Report: handoff written, blocker count, next 3 steps ready for resume.
