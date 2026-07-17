---
name: project-handoff
description: Prepare a compact, evidence-backed handoff for another developer, agent, tool, or future session on a multi-step reverse-engineering or recompilation project. Use at a milestone, context boundary, or pause when the next worker must resume accurately; do not use for a chronological activity log, release notes, or a casual status message with no continuation need.
compatibility: RecompHamr evidence workspace is recommended but not required.
---
# Project handoff

Write a resume map, not a diary. Verify current repository state, paths, commands, artifacts, and blockers immediately before writing; do not reconstruct them from memory. Keep the handoff compact enough to load as working context, but let complexity determine length rather than padding or an arbitrary line target.

Include:

1. **Goal and scope** — one-sentence outcome, current track/phase, and explicit out-of-scope work.
2. **Current state** — confirmed findings with precise sources; hypotheses with promotion/falsification evidence.
3. **Environment** — relevant OS, tool/compiler/runtime versions, target identity, working/build directories, and external services actually required. Never include keys or secret values.
4. **Working commands** — exact command, directory, and last observed result. Label stale or unrerun commands as unverified.
5. **Changed and retained artifacts** — exact paths, ownership, generated status, and why they matter; distinguish uncommitted user changes.
6. **Blockers and risks** — evidence, impact, attempted approaches when useful, and a concrete unblocking input or next query.
7. **Next steps** — ordered evidence-producing actions with an observable completion condition.

Do not claim a fix or parity from an unrun command. Do not paste bulky logs; link or name bounded evidence artifacts and quote only the decisive observation. Separate CONFIRMED, HYPOTHESIS, TODO, BLOCKED, and UNVERIFIED facts.

If a prior handoff says a check passed but it was not rerun after the latest relevant edit, immediately replace that stale success with **UNVERIFIED** and name the edit, exact command, and rerun needed. Do not wait for more context before correcting the claim.

Treat unrelated worktree edits as user-owned retained artifacts: identify them separately when their paths are known, do not claim, alter, revert, or silently ignore them, and scope next steps around the current task. When repository details are unavailable, still produce a minimal handoff from the facts given, label missing paths/commands **UNVERIFIED** or **BLOCKED**, and request them as next steps. Never emit private deliberation, a repeated self-dialogue, or placeholder success.

Use `.rehamr/REPHAMR_STATE.md` for the current compact persistent state when appropriate and `.rehamr/CHANGELOG.md` for historical milestones. Preserve existing useful state and update contradictions explicitly.

Before closing, re-open the handoff and verify every referenced path, current branch/commit when relevant, command status, blocker, and next-step dependency.
