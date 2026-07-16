# Decisions

## D-001 — Fresh upstream source is the code base

RecompHamr-Legacy is not the reconstruction base. The stripped template starts from pinned CodeHamr upstream source.

## D-002 — No TUI framework upgrade

Bubble Tea, Bubbles, Lip Gloss, and Glamour remain at the pinned upstream versions. The original Bubble Tea v2 migration phase is removed from the active plan.

## D-003 — No TUI redesign during reconstruction

The inherited screen composition and interaction contract are frozen through baseline acceptance and separation work.

## D-004 — Baseline before separation

Do not refactor orchestration ownership until the stripped application is proven to work. This avoids mixing strip regressions with architectural refactoring.

## D-005 — Separation before RecompHamr feature integration

After baseline acceptance, move runtime orchestration out of the TUI before adding Legacy feature families.

## D-006 — Windows-first command execution

The retained command tool is `powershell`, not `bash`. Generic OpenAI-compatible model support remains provider-neutral.

## D-007 — Verification gates are executable contracts

The canonical verification pipeline keeps the existing baseline, architecture, formatting, build, smoke, and Markdown-link checks and adds two explicit Go verification commands:

- `cmd/docscheck` enforces required durable documentation and mandatory contract terms defined in `docs/documentation-contract.json`;
- `cmd/coveragecheck` rejects any Go coverage profile below 100% statement coverage.

The documentation contract is stored outside active Go source so it can name removed subsystems without weakening the forbidden-reference scan over executable code.

## D-008 — Codex instructions are layered and routed

Keep the root `AGENTS.md` compact enough to act as the always-on repository contract. Put detailed workflow, documentation, TUI, parity, and skills rules in durable or subtree-specific instructions so Codex loads the closest relevant guidance without bloating every task.

## D-009 — Legacy parity permits evidence-backed modernization

RecompHamr-Legacy defines feature evidence and required compatibility, not internal architecture. A feature may be adapted, rewritten, fixed, simplified, or upgraded when the required capability is preserved or an intentional contract change is explicitly approved, tested, and documented.

## D-010 — Future skills use the Agent Skills standard

Stage G must implement standards-based Agent Skills support rather than restore the Legacy skills system one-to-one. The Agent Skills client-implementation guide is mandatory before client work, and every migrated Legacy skill must be converted and evaluated under the current Agent Skills authority set documented in `docs/dev/roadmap/agent-skills-standard.md`.

## D-011 — Stage C is ownership movement, not a rewrite

Separation preserves every accepted Stage A behavior. Move existing responsibilities behind narrow typed boundaries in small slices; do not redesign the TUI, replace the agent policy, alter persistence, or opportunistically modernize behavior. Slice 1 makes `internal/app` the composition root while keeping the existing TUI model intact.

## D-012 — Application composition constructs the agent runtime

`internal/app` constructs the agent runtime and injects its model and tool dependencies before selecting the concrete frontend. The runtime aggregate is a transitional mechanical boundary: it preserves the Bubble Tea value-model behavior while later checkpoints encapsulate mutable orchestration state and narrow presentation access to typed events and immutable facts.

## D-013 — Agent orchestration emits causal records

`internal/agent` emits stream, tool, cancellation, policy, and turn records at the state transition that owns them. `internal/app` injects the observer and owns its lifecycle through `internal/logging`; presentation applies typed effects but does not reconstruct orchestration events or open the protected log file.

## D-014 — Session identity is stable while concrete clients are replaceable

`internal/app` composes one stable `internal/session.Runtime` and supplies it to both the agent chat boundary and presentation adapter. The session runtime owns mutable configuration and concrete client replacement; asynchronous reachability and authenticated probes capture their endpoint/client at dispatch time so late results preserve the accepted stale-result identity. Presentation receives only non-secret immutable facts and opaque work.

## D-015 — The frontend boundary is neutral and terminal wiring is isolated

`internal/frontend` defines the only project-runtime contract imported by production TUI code. An application-owned controller translates typed intents, owns asynchronous work identity and backend lifecycle, and emits immutable display-safe snapshots and ordered semantic events. Core `internal/app` exposes only that neutral controller and cleanup lifetime; `internal/app/terminal` is the sole concrete TUI and Bubble Tea wiring edge. This preserves the runnable terminal while allowing application and backend packages to build and test without the presentation implementation.

## D-016 — Workspace state is optional lower-trust model context

Stage D preserves the Legacy-compatible `.rehamr/REPHAMR_STATE.md` filename and optional persistent-context behavior without porting the mixed Legacy initializer. `internal/workspace` owns canonical identity and bounded secure reads; core application composition is its only consumer. State is labeled untrusted project-maintained context beneath the embedded application contract, refreshes for each model round, and is silently omitted on absence or secure-read failure. Commands, RE templates, tools, skills, and MCP configuration remain in their authorized later stages.

## D-017 — Stage E parity is satisfied by the accepted backend, not a rewrite

The complete Legacy agent/runtime audit maps the required turn policy, context packing, and model-stream resilience contracts to the accepted Stage C backend. Where current behavior is equivalent, it remains untouched; where it is safer or more robust, the improvement is retained and documented. Stage E does not recreate Legacy TUI orchestration or add parallel abstractions merely to produce a code delta. The Legacy skills-template classifier remains Stage G because it classifies skill documents and is not an agent-turn decision service.

## D-018 — Stage F tools use configured capabilities and public-only ingestion

Stage F preserves the useful `repomixr` and `recomp_reference` contracts without restoring Legacy global directories or permissive I/O. Core application composition supplies an immutable tool set with protected cache identity. Repository import is limited to strict public GitHub HTTPS URLs; reference retrieval revalidates redirects and resolved addresses and refuses non-public networks. Both paths are cancellable, bounded, link-safe, atomically persisted, and treated as untrusted external content. This is improved behavioral parity, not a source-level port.
