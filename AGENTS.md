# RecompHamr Repository Instructions

## Mission and current stage

RecompHamr is being rebuilt from the pinned stripped CodeHamr baseline, not from RecompHamr-Legacy.

Order of work:

1. prove the barebones baseline;
2. separate application/runtime ownership from presentation without changing accepted TUI behavior or layout;
3. integrate RecompHamr capabilities one coherent subsystem at a time;
4. verify each capability against evidence, tests, docs, and runtime behavior.

The repository is currently **Stage A: fresh-source barebones baseline**. Until the baseline gate is accepted, do not add Legacy feature families, MCP, skills, new command families, project memory, diagnostics systems, reverse-engineering helpers, or a TUI redesign.

## Codex reading route

Codex reads this file automatically. Keep this file as the always-on rule set and load detailed policy only when relevant.

Before any non-trivial change, read:

1. `docs/dev/verification/baseline-status.md`
2. `docs/dev/workflows/baseline-gate.md`
3. `docs/dev/governance/engineering-workflow-rules.md`
4. the owning source files and focused tests

Then read the task-specific authority:

- architecture/separation work: `docs/dev/architecture/current-baseline.md` and `docs/dev/architecture/target-separation.md`
- Legacy feature work: `docs/dev/roadmap/legacy-parity-policy.md` and `docs/dev/reference/legacy-feature-holding-pen.md`
- skill client or skill migration work: `docs/dev/roadmap/agent-skills-standard.md`
- documentation work: follow `docs/AGENTS.md`
- TUI work: follow `internal/tui/AGENTS.md`

Do not bulk-read unrelated documents merely to satisfy a ritual. Read all authorities required by the active work packet, and inspect every source/test file necessary to make the task evidence-complete.

## Stage A freeze

Until baseline acceptance:

- Do not upgrade Bubble Tea, Bubbles, Lip Gloss, or Glamour.
- Do not redesign, rearrange, modernize, restyle, or recompose the TUI.
- Preserve the inherited model/update/view mechanics, inline terminal rendering, composer, transcript behavior, resize behavior, model picker, cancellation, exit behavior, layout composition, and spacing.
- Do not add speculative abstractions for future features.
- Do not add TODO-only behavior, placeholders, fake success states, hidden fallbacks, or unsupported branches presented as finished.

Branding changes and removal of hosted-service-only UI do not authorize a layout redesign.

## Work packet

Every non-trivial task must establish:

- **Outcome** — observable result required.
- **In scope** — behavior and files intentionally changed.
- **Out of scope** — nearby work that must remain untouched.
- **Authorities** — repository docs, source references, and current external docs that govern the task.
- **Evidence** — source, tests, runtime observations, or reference behavior used.
- **Verification** — focused and canonical checks capable of falsifying the result.
- **Documentation impact** — docs that must change, or evidence that none are affected.
- **Security impact** — trust, permissions, secrets, process, filesystem, or network consequences.
- **Stop condition** — concrete closure condition.

Do not widen scope because adjacent work is convenient.

## Truth and evidence

- Ground factual claims in source inspection, automated verification, runtime observation, or authoritative documentation.
- For important completion claims, seek three evidence classes where applicable: source, tests/automation, and runtime/manual observation.
- Use `unverified:`, `blocked:`, and `unsupported:` precisely when evidence is incomplete.
- A build does not prove behavior. A screenshot does not prove backend correctness. Existing Legacy code does not prove that the same architecture should be copied.
- It is valid to conclude that existing code is already correct.

## Legacy parity is behavioral, not structural

RecompHamr-Legacy is a feature and behavior reference, not a base tree and not an architecture mandate.

When integrating a Legacy capability:

- preserve the required user-visible capability and compatibility contract;
- do **not** copy code, package boundaries, APIs, commands, or internal architecture 1:1 unless evidence shows that is the best design;
- prefer a cleaner, safer, simpler, more testable implementation that fits the current architecture;
- fix known defects and remove obsolete technical debt when doing so does not violate a required compatibility contract;
- record whether the result is `equivalent`, `improved`, `intentionally changed`, `not applicable`, `blocked`, or `unverified`;
- prove the chosen result with tests, docs, and runtime evidence before closing the parity row.

See `docs/dev/roadmap/legacy-parity-policy.md`.

## Agent Skills rule

Do not implement the skills subsystem before Stage G. When Stage G begins, the implementation must follow the current Agent Skills standard rather than mechanically restoring the Legacy loader.

Before implementing RecompHamr skills-client support, reading the complete current `https://agentskills.io/client-implementation/adding-skills-support` guide is mandatory.

Before editing each individual Legacy skill, follow the mandatory current Agent Skills authority-set reading rule in `docs/dev/roadmap/agent-skills-standard.md`. Every migrated skill must individually comply with that standard and pass its own validation and evaluation gates. Do not bulk rename old Markdown files and call the migration complete.

## Architecture direction

Stage A temporarily preserves some inherited TUI/runtime coupling so the stripped baseline can be proven first. After acceptance, establish these ownership boundaries before Legacy feature ports:

- `cmd/recomphamr`: process entrypoint only;
- `internal/app`: composition and application orchestration;
- `internal/tui`: rendering, presentation state, and input-to-intent translation only;
- `internal/agent`: agent-turn lifecycle and tool-loop policy;
- `internal/llm`: provider-neutral transport;
- `internal/tools`: tool contracts and execution;
- `internal/config`: configuration persistence and validation;
- future skills subsystem: discovery, parsing, precedence, activation, diagnostics, and evaluation outside the TUI.

Do not claim separation of concerns is complete while presentation code directly executes tools or owns provider, persistence, or skills lifecycle behavior.

## Change discipline

- Inspect `git status` before editing and preserve unrelated user changes.
- Prefer focused, coherent edits over rewrites.
- Preserve accepted observable behavior unless the active work packet explicitly changes it.
- Never weaken or delete meaningful tests merely to make verification pass.
- Every bug fix requires a regression test when the behavior is testable.
- Maintain **100% behavioral surface coverage** across retained upstream behavior, modified behavior, replacements, Legacy parity work, and newly added behavior. No old surface is grandfathered and no new surface is exempt.
- Maintain **100% meaningful documentation coverage**: every Go package and exported symbol must have appropriate Go documentation, and every relied-on user, integration, configuration, persistence, security, lifecycle, and extension contract must be documented. Do not add useless comments to trivial private locals merely to inflate a metric.
- Keep a behavioral-surface inventory that maps every surface to implementation, applicable success/failure/boundary/security tests, documentation, and verification evidence.
- Keep implementation, tests, help, and durable docs synchronized.
- Do not claim completion while required verification is unavailable; report the exact remaining evidence.

## Required gate

On Go 1.26+ with PowerShell available:

```powershell
pwsh -NoProfile -File ./scripts/verify.ps1
```

This is the canonical automated gate. It includes baseline policy, documentation-contract and link checks, architecture checks, formatting, tests with strict **100% statement coverage**, build, and CLI smoke verification.

Passing statement coverage is necessary but not sufficient. Phase/task closure also requires **100% behavioral surface coverage** and **100% meaningful documentation coverage** as defined in `docs/dev/verification/behavioral-surface-coverage.md` and the engineering workflow rules.

Manual TUI/runtime acceptance remains separate where required.
