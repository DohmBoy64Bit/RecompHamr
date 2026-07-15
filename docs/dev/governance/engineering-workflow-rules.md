# Engineering Workflow Rules

## Purpose

These rules adapt the project's strict engineering policy to the current RecompHamr barebones template. They apply to implementation, correction, audit, migration, parity work, documentation, and release work.

The repository stage and approved work packet define scope. Do not silently reinterpret a task, mix future stages into the active stage, or present deferred behavior as finished.

## Authority and instruction hierarchy

Use the closest applicable repository instruction file and the active work packet.

The root `AGENTS.md` is the compact always-on contract. Detailed policy lives in durable docs so Codex does not spend its project-instruction budget on rules irrelevant to the current task. More specific `AGENTS.md` files may add local rules for a subtree.

When sources disagree, use this order unless the task explicitly changes it:

1. current user direction;
2. active repository instructions and stage constraints;
3. current implementation and tests;
4. current authoritative external documentation for changed dependencies or standards;
5. historical plans and Legacy references.

Historical documents do not override current code or an explicitly superseding decision.

## Mandatory context refresh

### At the start of every stage or major workstream

Read:

- `AGENTS.md`;
- current baseline/status and roadmap documents;
- scope, evidence, definition-of-done, and architecture documents;
- the active work packet;
- all current authoritative external documentation that governs the planned implementation.

### Before each non-trivial task

Read the root instructions, current status, the active work packet, and all relevant source/tests/docs for the owning subsystem.

Do not reread every repository document for every small edit. Context loading must be complete for the task but targeted enough to avoid burying the governing rules in irrelevant material.

### After a completed stage

Refresh the stage/status, architecture, parity, verification, and affected user/developer docs. Record what changed and what remains open.

## Work packets and stage discipline

Every non-trivial task or phase requires a durable or clearly recorded packet containing:

- Outcome
- In-scope work
- Explicitly out-of-scope work
- Authorities
- Required evidence
- Verification commands
- Documentation impact
- Security impact
- Stop condition

Complete ordered stages in order unless the active roadmap explicitly permits otherwise. Preserve superseded decisions as history and mark them clearly rather than deleting evidence.

Do not stop at arbitrary implementation slices when the scoped task can be completed safely. Stop for a real blocker, required user evidence, exhausted safe approaches, or an explicit direction change.

## Truthfulness and evidence

Every material factual claim must be grounded in at least one of:

- direct source inspection;
- automated verification;
- runtime observation;
- authoritative external documentation.

For high-impact completion claims, use multiple independent evidence classes where practical. The strongest common pattern is:

1. source inspection;
2. focused and canonical automated checks;
3. runtime or manual observation.

Use these labels precisely:

- `verified:` reproduced by stated evidence;
- `unverified:` required evidence has not been executed;
- `blocked:` a named external input or environment prevents the check;
- `unsupported:` intentionally outside the current supported contract.

Never invent APIs, parity, metrics, timings, release status, or success evidence. Code existing or compiling is not proof that a feature works end to end.

## No placeholders or fake completion

Do not present as finished:

- TODO-only files;
- placeholder implementations;
- mock production success paths;
- speculative APIs wired into production paths;
- silent unsupported branches;
- hidden deferred behavior;
- success states not backed by real execution.

A phase remains open until its scoped behavior is real and the required evidence exists.

## Documentation contract

Maintain **100% meaningful documentation coverage** for every retained, modified, replaced, parity, and newly added surface. Documentation coverage is a required contract, not a best-effort target.

At minimum:

- every Go package must have package documentation;
- every exported function, method, type, constant, and variable must have an appropriate Go doc comment;
- every public or cross-package contract must document its purpose, inputs, outputs, errors, lifecycle, and important invariants as applicable;
- architecture boundaries and non-obvious invariants must be documented;
- commands, arguments, options, keybindings, and side effects must be documented;
- tools, schemas, inputs, outputs, errors, permissions, cancellation, and examples must be documented;
- configuration files, keys, defaults, validation, precedence, environment variables, and secret handling must be documented;
- persistent files, logs, history, caches, generated artifacts, migration behavior, and lifecycle must be documented;
- user-visible states, warnings, failures, unsupported states, and blocked states must be documented;
- protocols, external integrations, trust boundaries, and startup/shutdown behavior must be documented;
- skills, extension points, discovery, precedence, activation, and diagnostics must be documented.

This is the Go equivalent of complete docstring coverage for the symbols and contracts that require durable explanation. Do not add useless comments to trivial private locals or obvious one-line implementation details merely to inflate a metric. Unexported helpers require comments when their intent, invariant, risk, protocol role, or tradeoff is not obvious from idiomatic code.

Update affected user docs, developer docs, help, examples, behavioral-surface evidence, parity/traceability evidence, and indexes in the same change as behavior. A surface is not complete while its required documentation is missing, stale, or contradicted by implementation.

## Documentation drift gate

Before a major behavior-changing phase, capture the identity of relevant governing docs. After implementation, verify whether each affected document changed appropriately.

A behavior change with no documentation update requires evidence that the documentation contract truly did not change. Avoid self-referential hash loops; use commit identity or an external evidence record for final immutable status.

## Testing and verification

Maintain two independent coverage requirements:

1. **100% statement coverage** for instrumented Go statements once the repository coverage gate applies.
2. **100% behavioral surface coverage** across every retained upstream surface, modified surface, replacement, Legacy parity surface, optimization, refactor with observable consequences, and newly added surface.

Statement coverage does not prove behavioral coverage and behavioral coverage does not waive statement coverage. Both are mandatory.

Every behavioral surface must be inventoried and mapped to:

- implementation ownership;
- at least one test that proves the required contract;
- every applicable success, failure, malformed-input, boundary, cancellation, timeout, cleanup, platform, Unicode, persistence, migration, compatibility, concurrency, and security case;
- required documentation;
- verification evidence and status.

A category may be marked not applicable only with a concrete reason. No retained or historical behavior is grandfathered merely because it predates the current test policy, and no new or improved behavior is exempt because it was introduced during modernization.

Additional rules:

- Do not weaken, skip, exclude, or delete meaningful tests merely to satisfy coverage.
- Every reproducible bug fix must add a regression test that would fail without the fix when practical.
- Use deterministic fakes for external systems to verify orchestration; never let a fake become a production success path.
- Prefer contract-level tests over tests coupled to private implementation details.
- Run focused checks while implementing and the canonical full gate before closure.
- Run formatting, architecture, documentation, statement coverage, behavioral-surface review, build, and runtime checks required by the active stage.

The behavioral-surface contract is defined in `../verification/behavioral-surface-coverage.md`. If a check or required observation cannot run, report exactly what remains unverified and why.

## Architecture and separation of concerns

Presentation owns:

- display state;
- rendering;
- input interpretation;
- emitted user intent.

Application/runtime services own:

- side effects;
- configuration persistence;
- filesystem access;
- process execution;
- networking;
- model calls;
- tool dispatch;
- skills lifecycle;
- protocol lifecycle;
- durable state;
- security enforcement.

Use typed contracts across boundaries. Add abstractions only when they remove real complexity, establish a required ownership boundary, or match an evidence-backed architecture.

The Stage A exception is deliberate: inherited TUI/runtime coupling remains temporarily until the baseline is accepted. Do not expand that debt, and do not claim it is already resolved.

## Terminal interface rules

Until the TUI freeze is explicitly lifted:

- preserve the inherited layout and interaction contract;
- do not upgrade the terminal UI framework stack;
- use the declared framework version's official behavior as authority for changes;
- keep rendering side-effect free where currently practical and move side effects out during Stage C;
- measure display width correctly for Unicode, combining marks, CJK, wrapping, and long values when touching layout logic;
- communicate important state with text/symbols as well as color;
- do not fabricate token counts, costs, progress, connection state, or private reasoning;
- treat third-party interfaces as inspiration, not source or a one-to-one visual target.

Automated render tests are regression evidence. Manual target-terminal acceptance is required when the active gate calls for visual approval.

## Security and configuration

- Document permission and trust boundaries for side effects.
- Never print, log, commit, or expose secret values.
- Test traversal, malformed configuration, unsafe permissions, redaction, cancellation, process cleanup, output bounds, and transport failures where relevant.
- Prefer structured parsers and explicit validation for structured configuration.
- Make defaults explicit, secure, documented, and tested.
- Persistent processes and integrations require lifecycle, cancellation, recovery, and shutdown documentation and tests.

## Legacy feature parity and modernization

Parity means satisfying the required capability contract, not reproducing Legacy source code.

For every Legacy feature family:

1. inventory observed behavior, inputs, outputs, errors, persisted data, commands, and tests;
2. separate required compatibility from incidental implementation detail and historical defects;
3. select the correct owner in the current architecture;
4. design the smallest coherent modern implementation;
5. preserve compatibility where required, but improve architecture, safety, testability, performance, or usability when evidence supports it;
6. inventory every retained, replaced, intentionally changed, and newly introduced behavioral surface for the feature;
7. write tests against the required contract, not against private Legacy structure, until the feature reaches 100% behavioral surface coverage;
8. document every required surface and all approved behavior changes;
9. connect presentation only after backend behavior is verified;
10. record the parity disposition and evidence.

Allowed dispositions:

- `equivalent` — required behavior preserved;
- `improved` — required capability preserved with a documented upgrade;
- `intentionally changed` — old behavior deliberately replaced with an approved contract;
- `not applicable` — obsolete or irrelevant behavior with evidence;
- `blocked` — named external blocker;
- `unverified` — implementation or evidence incomplete.

Do not copy obsolete technical debt, bugs, package boundaries, APIs, or UI composition merely because Legacy used them. Do not call an incompatible rewrite parity without explicitly documenting the changed contract.

## Agent Skills migration

Stage G skills work is governed by `docs/dev/roadmap/agent-skills-standard.md`.

The Legacy skills system is reference material only. RecompHamr's future skills client and every migrated skill must follow the current Agent Skills standard and RecompHamr's current architecture.

No bulk conversion is complete until every individual skill has:

- specification-valid structure;
- focused `SKILL.md` instructions;
- a tested description and trigger boundary;
- supporting resources only when needed;
- deterministic scripts only when justified;
- per-skill evals and evidence;
- documentation and diagnostics.

## Source control and releases

- Preserve unrelated user and dirty-worktree changes.
- Use non-destructive Git operations.
- Do not reset, revert, or overwrite work you did not create.
- Commit only after required verification passes unless an explicitly requested checkpoint documents incomplete status.
- Push only when requested or required by an approved phase.
- Release evidence must distinguish local build artifacts from externally published artifacts.

## Blocker discipline

After three genuinely different failed strategies against the same blocker, record:

- the blocker;
- attempts and evidence;
- why each failed;
- the exact external input or environment needed.

Do not relabel difficult work as blocked while meaningful evidence-based progress remains possible.

## Required completion report

A completed phase reports:

- **Changed** — implementation and behavior;
- **Documented** — affected user/developer/help/parity material;
- **Verified** — exact commands and observed results;
- **Coverage** — statement coverage, behavioral-surface coverage, and meaningful documentation coverage results;
- **Security** — permissions, trust, redaction, and regressions;
- **Evidence** — runtime observations, hashes, artifacts, commits, or links as applicable;
- **Known limits** — only evidence-backed unsupported, blocked, or unverified behavior.

Do not declare completion until implementation, tests, documentation, runtime evidence required by the phase, and parity/traceability status agree.
