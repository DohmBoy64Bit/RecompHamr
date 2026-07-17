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

Accepted at implementation/parity-audit commit `b4fee3c02b6178aa9e2d1e2a7cdf843b13281355`. The complete Legacy agent/runtime contract was audited against the accepted Stage C backend; no genuine production gap was found, and equivalent or improved behavior was verified without a redundant rewrite.

Do not copy Legacy orchestration into presentation merely to achieve superficial parity.

## Stage F — RecompHamr tools

Add reverse-engineering/recompilation tool families through typed tool contracts, tests, security boundaries, and docs.

Accepted at implementation commit `d7286bee7a00b0debacc4708e9b5807550e3b7ba`. Direct Legacy inspection limited the stage to `repomixr`, `recomp_reference`, and their application-owned typed tool-set boundary; both capabilities were implemented with improved security and verified without adding Stage G commands/skills or Stage H MCP tools.

## Stage G — Commands and Agent Skills

Add command registries and standards-based skills support after their backend owners exist.

Stage G is active. Its evidence-backed packet is established and the mandatory current Agent Skills authorities were read before implementation and individual migration work. The application-owned registry/client, 21 bundled migrations, direct case review, exact-build Devstral acceptance, and 3183/3183 canonical gate are complete; dual-platform CI and accepted-commit closure remain open.

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
