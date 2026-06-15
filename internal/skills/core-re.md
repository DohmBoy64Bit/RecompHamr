# core-re

Use this skill for general reverse-engineering and recompilation-adjacent software work.

Default workflow:
1. Inspect repository shape with targeted commands.
2. Identify source-of-truth files: README, build files, scripts, docs, logs, configs, generated artifacts.
3. Capture evidence before changing behavior.
4. Make the smallest useful edit.
5. Run the narrowest useful verification.
6. Update `.rehamr/CHANGELOG.md` or `.rehamr/EVIDENCE.md` if `/init-re` has been run.

Do not guess binary behavior. If the task depends on binary facts, ask tools to inspect the binary or mark the point as HYPOTHESIS.

