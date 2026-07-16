# Legacy Feature Holding Pen

Everything in this document is intentionally **excluded** from the barebones baseline. Exclusion is not a judgment that the feature is bad; it prevents feature integration from obscuring baseline regressions.

| Feature family observed in the Legacy reference | Baseline status | Earliest integration stage |
|---|---|---|
| classifier | excluded | after Stage C, when a verified owner/use case is defined |
| doctor/environment diagnostics | excluded | Stage D or later |
| project memory/workspace state | Stage D foundation accepted: secure optional state loading and refresh are verified; initialization/status commands and RE templates remain deferred | Stage D foundation accepted; Stage F/G for template/command exposure |
| expanded agent-loop behavior | accepted: Stage E verified the required Legacy policies as equivalent or improved in the current backend; no redundant port was needed | Stage E accepted at `b4fee3c` |
| `repomixr` | accepted with improved strict public-GitHub, bounded, protected-cache behavior | Stage F accepted at `d7286be` |
| `recomp_reference` | accepted with improved public-network-only, bounded, protected-cache behavior | Stage F accepted at `d7286be` |
| extra RE-specific built-in tools | not applicable: direct Legacy inventory found no additional non-MCP built-ins | Stage F audit |
| extended slash-command surface | excluded | Stage G |
| built-in skills and skill loading | excluded; Legacy format is not the target | Stage G using the Agent Skills standard |
| MCP clients, servers, configuration, and tool exposure | excluded | Stage H |
| updater/self-reexec/release delivery | excluded from reconstruction baseline | Stage I only if still desired |
| installers and promotional assets | excluded | not required for core reconstruction |

## Porting rule

A row leaves the holding pen only when:

1. its required capability contract is verified against relevant Legacy source, tests, docs, and runtime evidence when available;
2. its target package owner is documented;
3. the baseline gate is accepted;
4. Stage C separation is complete;
5. focused tests exist before TUI integration;
6. docs and parity evidence are updated in the same change.

Do not restore a Legacy package merely to make old tests compile. Port required behavior into the current architecture.

The implementation may be cleaner, safer, or standards-based rather than one-to-one. Record whether the result is `equivalent`, `improved`, `intentionally changed`, `not applicable`, `blocked`, or `unverified`.

For skills specifically, use `../roadmap/agent-skills-standard.md`; do not restore the Legacy skill loader or old skill Markdown format by default.
