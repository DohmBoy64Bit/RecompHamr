---
name: build-fix-loop
description: Diagnose and repair a reproducible failing build, failing test or stack trace, code-generation error, generated recomp code that no longer compiles, or decomp/recomp failure after regeneration through one-change iterations. Use when an exact command currently fails and evidence is needed to find the earliest actionable source or metadata cause; do not use for speculative refactoring, performance tuning, or a build that already passes.
compatibility: RecompHamr with the failing project's documented build tools.
---
# Build failure loop

Preserve the exact failing command, working directory, exit status, and the smallest relevant error chain. Identify the earliest actionable failure, not merely the final summary line. Inspect the referenced source, configuration, metadata, generated-input, or dependency boundary before editing.

Iterate with one causal change:

1. State the current root-cause hypothesis and evidence.
2. Change the smallest source-of-truth input that can test it.
3. Re-run the exact failed command in the same relevant environment.
4. Compare the failure identity, not just whether output changed.
5. Keep the change only if the evidence advances or resolves the failure.

Do not hand-edit generated or recompiled output when configuration, metadata, generator logic, or source input owns it. Do not automatically revert unrelated user changes; isolate your edit and explain any overlap. If the same causal attempt produces the same failure twice, stop repeating it and gather different evidence. If the failure moves to an unrelated layer, close the current diagnosis and record the new blocker separately.

When the failure already names an unavailable dependency, proprietary SDK, permission, or external service that cannot be supplied in the current environment, record that exact requirement as **BLOCKED** immediately. State the external input or environment needed and any narrower checks that remain possible; do not conceal the blocker behind a generic request for logs or speculate about installing a substitute.

Success requires the originally failing command to pass, plus any narrower regression check needed for the changed contract. A successful compile alone does not prove runtime or binary equivalence. If asked to call a passing build fully correct, refuse the overclaim: report only that the exact build passed and name runtime, binary-equivalence, or other acceptance checks that remain unverified.

Report:

- **Changed** — exact files and causal reason.
- **Verified** — exact command, exit result, and relevant observation.
- **Remaining** — only reproduced failures, named blockers, or explicitly unverified acceptance.

Store large logs in the evidence workspace only when useful, bounded, and free of secrets or unrelated user data.
