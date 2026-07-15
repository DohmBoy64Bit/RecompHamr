# Documentation Map

The documentation is intentionally split into user-facing material and durable developer/project memory.

The machine-readable required-document/content contract is [`documentation-contract.json`](documentation-contract.json).

Repository-wide always-on instructions live in [`../AGENTS.md`](../AGENTS.md). Codex-specific local instructions also exist in this documentation subtree and the TUI subtree.

## User documentation

- [`user/configuration.md`](user/configuration.md)
- [`user/tools.md`](user/tools.md)

## Developer documentation

### Governance

- [`dev/governance/scope-control.md`](dev/governance/scope-control.md)
- [`dev/governance/evidence-policy.md`](dev/governance/evidence-policy.md)
- [`dev/governance/definition-of-done.md`](dev/governance/definition-of-done.md)
- [`dev/governance/engineering-workflow-rules.md`](dev/governance/engineering-workflow-rules.md) — detailed project-wide engineering policy routed from the compact root `AGENTS.md`

### Project memory

- [`dev/memory/repository-lineage.md`](dev/memory/repository-lineage.md)
- [`dev/memory/decisions.md`](dev/memory/decisions.md)

### Verification

- [`dev/verification/baseline-status.md`](dev/verification/baseline-status.md)
- [`dev/verification/verification-contract.md`](dev/verification/verification-contract.md) — documentation contract, 100% statement coverage, behavioral-surface coverage, build, and smoke-test gates
- [`dev/verification/behavioral-surface-coverage.md`](dev/verification/behavioral-surface-coverage.md) — 100% old-and-new behavioral surface coverage and meaningful documentation coverage contract
- [`dev/verification/stage-a-behavioral-surface.md`](dev/verification/stage-a-behavioral-surface.md) — active Stage A surface inventory and acceptance evidence map
- [`dev/verification/stage-a-runtime-acceptance.md`](dev/verification/stage-a-runtime-acceptance.md) — sanitized Windows Terminal, LM Studio, tool, cancellation, security, and screenshot evidence

### Architecture

- [`dev/architecture/current-baseline.md`](dev/architecture/current-baseline.md)
- [`dev/architecture/target-separation.md`](dev/architecture/target-separation.md)

### Workflows

- [`dev/workflows/baseline-gate.md`](dev/workflows/baseline-gate.md)
- [`dev/workflows/change-control.md`](dev/workflows/change-control.md)
- [`dev/workflows/work-packet-template.md`](dev/workflows/work-packet-template.md)
- [`dev/workflows/stage-a-acceptance-work-packet.md`](dev/workflows/stage-a-acceptance-work-packet.md)

### Roadmap

- [`dev/roadmap/integration-order.md`](dev/roadmap/integration-order.md)
- [`dev/roadmap/legacy-parity-policy.md`](dev/roadmap/legacy-parity-policy.md) — behavior/capability parity with evidence-backed modernization allowed
- [`dev/roadmap/agent-skills-standard.md`](dev/roadmap/agent-skills-standard.md) — mandatory future Agent Skills client and per-skill migration policy

### Reference

- [`dev/reference/source-inputs.md`](dev/reference/source-inputs.md)
- [`dev/reference/strip-manifest.md`](dev/reference/strip-manifest.md)
- [`dev/reference/legacy-feature-holding-pen.md`](dev/reference/legacy-feature-holding-pen.md)
- [`dev/reference/upstream-85409d1-tree.sha256`](dev/reference/upstream-85409d1-tree.sha256)
