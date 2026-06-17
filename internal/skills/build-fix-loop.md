# build-fix-loop

Use this skill when a project does not build, tests fail, generated/recompiled
code is broken, or you are iterating on build errors in a recompilation project.

## Goal

Resolve build failures through root-cause diagnosis and single-change iteration —
never patch symptoms, never claim success without command output proving it.

## Rules

1. **Capture the exact command and full relevant error.** The failure context
   matters — copy verbatim, do not summarize.
2. **Identify the root failure.** The earliest/innermost error is the real
   cause, not the last printed line. Trace the error chain upstream.
3. **Make one focused fix per iteration.** Multiple simultaneous changes make
   root cause unclear. If the fix doesn't work, revert before trying a
   different approach.
4. **Re-run the exact failed command.** Changing the command between attempts
   hides whether the fix worked.
5. **Stop when verified.** The build exits 0, the test passes, or a new
   unrelated blocker is reached. If the same fix fails the same way twice,
   the approach is wrong — change strategy, do not retry.
6. **Never claim the build is fixed unless the command succeeds.** Confidence
   is not proof. Output is proof.

## Workflow

```
Capture error → identify root cause → inspect source file/config →
make one fix → re-run command → verify output →
if fixed: document; if same error: change strategy; if new error: restart
```

## Required Output

After each build-fix cycle, report:

```md
## Changed
- <file path>: <short reason for the change>

## Verified
- <command>: <result — exit code, key output, or error message>

## Remaining
- <real blockers or next errors only — no speculation>
```

## Evidence / Artifact Targets

- `REPHAMR_STATE.md` — updated with build status, working commands, error patterns
- `.rehamr/evidence/` — full build output, error traces, verified command logs
- `logs/build_*.txt` — raw build output if large

## Stop Conditions

Stop and change strategy when:
- The same edit or command fails the same way twice — the approach is wrong,
  not your luck
- The error points to a layer you cannot fix (missing SDK, missing tool,
  user decision needed) — document as BLOCKED
- A new unrelated error appears that is outside the current fix scope —
  document and restart the loop from the new error
- You are about to make a change in generated code (`recompiled/`, `gen/`,
  `runner/`, `asm/*.s`) — the fix belongs in config, metadata, or the tool

## Session Close

1. Update `REPHAMR_STATE.md` with verified commands that work.
2. Document error patterns learned (format: "X causes Y, fix with Z").
3. Report: changed files, verified commands, remaining blockers, next 3 steps.
