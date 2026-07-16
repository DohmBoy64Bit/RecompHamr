# Scope Control

The accepted Stage A/Stage C gates and the active Stage D work packet define the current project boundary.

## Allowed now

- implement the workspace and configuration foundations explicitly listed in the active Stage D packet;
- preserve or improve verified Legacy behavior within its current architecture owner;
- improve tests, verification, rules, or documentation without changing the frozen TUI contract;
- fix evidence-backed defects in accepted behavior;
- prepare durable future-stage policy that does not add runtime placeholders or prematurely implement blocked subsystems.

## Still blocked by integration order

- MCP;
- skills runtime support;
- reverse-engineering helper tools;
- diagnostics or project-workspace commands not authorized by the Stage D packet;
- command-surface expansion;
- TUI redesign;
- Bubble Tea framework migration.

Blocked work belongs in the roadmap, policy, or holding pen, not in placeholder production code. Stage D optional workspace-state loading is the only active project-memory surface.

## Future parity scope

Once feature integration begins, scope is defined by the verified capability contract, not by a requirement to copy Legacy code one-for-one.

A task may improve or replace Legacy internals when:

- the required compatibility boundary is explicit;
- the new owner fits the current architecture;
- the improvement is documented before closure;
- regression and migration risks are tested;
- the parity disposition is explicit.

See `../roadmap/legacy-parity-policy.md`.
