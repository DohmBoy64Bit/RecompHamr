# Integration Order

## Stage A — Fresh barebones upstream strip

Accepted.

Prove the stripped CodeHamr-derived application with the inherited TUI and four retained tools.

## Stage B — Baseline acceptance

Pass automated checks and complete real Windows/TUI/manual acceptance.

## Stage C — Separation of concerns

Extract application orchestration and agent-loop ownership from the TUI without changing accepted behavior or layout.

Accepted at implementation commit `72e6b43215cc14f91eb6547de15a7386bc77b927`.

Required outcome:

- presentation does not execute tools;
- presentation does not persist config;
- presentation does not own LLM/client lifecycle;
- future extension systems have backend owners outside presentation;
- runtime behavior remains equivalent to the accepted baseline.

## Stage D — Workspace and configuration foundations

Integrate verified RecompHamr capability contracts that belong to configuration and workspace ownership.

Accepted at implementation commit `449a83cb379e79fd84b817b0e95f63de7472578a`. Secure application-owned workspace identity, bounded persistent-state loading, and the existing configuration foundations are verified without adding later-stage commands, tools, skills, or MCP configuration.

Use `legacy-parity-policy.md`: preserve required behavior and compatibility, but improve internal architecture when the new design is better supported by evidence.

## Stage E — Agent/runtime capabilities

Integrate verified agent-loop capabilities below the TUI.

Active under `../workflows/stage-e-agent-runtime-work-packet.md`. The first checkpoint audits the complete Legacy agent/runtime contract against the accepted Stage C backend before authorizing any production delta.

Do not copy Legacy orchestration into presentation merely to achieve superficial parity.

## Stage F — RecompHamr tools

Add reverse-engineering/recompilation tool families through typed tool contracts, tests, security boundaries, and docs.

## Stage G — Commands and Agent Skills

Add command registries and standards-based skills support after their backend owners exist.

Skills requirements:

- implement the client lifecycle according to `agent-skills-standard.md`;
- read the mandatory current Agent Skills client-implementation guide before implementation;
- convert Legacy skills individually to the current Agent Skills specification;
- use progressive disclosure rather than eagerly loading all skill bodies;
- validate and evaluate each migrated skill before closure;
- allow evidence-backed improvement of Legacy skill structure, descriptions, scripts, and workflows instead of requiring one-to-one copies.

## Stage H — MCP

Add lifecycle and transports outside the TUI, then expose presentation state through typed contracts.

## Stage I — Final presentation integration

Expose completed backend capabilities through the inherited TUI without changing its fundamental layout unless a later explicit design decision lifts the freeze.

## Stage J — Final parity, optimization audit, and release

Audit source, docs, tests, runtime evidence, and release artifacts against the verified feature references.

The final parity audit must distinguish:

- equivalent behavior;
- improved behavior;
- intentionally changed behavior;
- not-applicable Legacy behavior;
- blocked or unverified behavior.

Source-code similarity is not a release criterion.
