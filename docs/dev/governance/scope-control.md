# Scope Control

The baseline gate is the active project boundary.

## Allowed now

- fix defects introduced by stripping;
- restore behavior that direct upstream evidence proves was accidentally broken;
- improve tests, verification, rules, or documentation without changing the frozen TUI contract;
- make the four retained tools correct and secure;
- prepare durable future-stage policy that does not add runtime placeholders or prematurely implement blocked subsystems.

## Blocked until baseline acceptance

- RecompHamr-Legacy feature integration;
- MCP;
- skills runtime support;
- reverse-engineering helper tools;
- project memory and diagnostics systems;
- command-surface expansion;
- TUI redesign;
- Bubble Tea framework migration.

Blocked work belongs in the roadmap, policy, or holding pen, not in placeholder production code.

## Future parity scope

Once feature integration begins, scope is defined by the verified capability contract, not by a requirement to copy Legacy code one-for-one.

A task may improve or replace Legacy internals when:

- the required compatibility boundary is explicit;
- the new owner fits the current architecture;
- the improvement is documented before closure;
- regression and migration risks are tested;
- the parity disposition is explicit.

See `../roadmap/legacy-parity-policy.md`.
