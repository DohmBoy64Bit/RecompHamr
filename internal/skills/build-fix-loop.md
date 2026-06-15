# build-fix-loop

Use this skill when a project does not build, tests fail, or generated/recompiled code is broken.

Loop:
1. Capture the exact command and full relevant error.
2. Identify the earliest/root failure, not just the last printed line.
3. Inspect the owning file/config/script.
4. Make one focused fix.
5. Re-run the exact failed command.
6. Stop when verified or when a new unrelated blocker is reached.

Report format:
- Changed: paths and short reason.
- Verified: command and result.
- Remaining: only real blockers or next errors.

Never claim the build is fixed unless the command succeeds.

