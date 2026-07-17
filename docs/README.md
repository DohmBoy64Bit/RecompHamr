# Documentation Map

The documentation is intentionally split into user-facing material and durable developer/project memory.

The machine-readable required-document/content contract is [`documentation-contract.json`](documentation-contract.json).

Repository-wide always-on instructions live in [`../AGENTS.md`](../AGENTS.md). Codex-specific local instructions also exist in this documentation subtree and the TUI subtree.

## User documentation

- [`user/configuration.md`](user/configuration.md)
- [`user/workspace.md`](user/workspace.md)
- [`user/tools.md`](user/tools.md)
- [`user/commands-and-skills.md`](user/commands-and-skills.md)

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
- [`dev/verification/stage-c-behavioral-surface.md`](dev/verification/stage-c-behavioral-surface.md) — accepted separation-of-concerns ownership and equivalence inventory
- [`dev/verification/stage-d-behavioral-surface.md`](dev/verification/stage-d-behavioral-surface.md) — accepted workspace/configuration foundation and Legacy parity inventory
- [`dev/verification/stage-e-behavioral-surface.md`](dev/verification/stage-e-behavioral-surface.md) — accepted agent/runtime parity audit and evidence map
- [`dev/verification/stage-f-behavioral-surface.md`](dev/verification/stage-f-behavioral-surface.md) — accepted RecompHamr tool integration and security evidence map
- [`dev/verification/stage-g-behavioral-surface.md`](dev/verification/stage-g-behavioral-surface.md) — active command-registry and standards-based Agent Skills inventory
- [`dev/verification/stage-a-runtime-acceptance.md`](dev/verification/stage-a-runtime-acceptance.md) — sanitized Windows Terminal, LM Studio, tool, cancellation, security, and screenshot evidence

### Architecture

- [`dev/architecture/current-baseline.md`](dev/architecture/current-baseline.md)
- [`dev/architecture/target-separation.md`](dev/architecture/target-separation.md)

### Workflows

- [`dev/workflows/baseline-gate.md`](dev/workflows/baseline-gate.md)
- [`dev/workflows/change-control.md`](dev/workflows/change-control.md)
- [`dev/workflows/work-packet-template.md`](dev/workflows/work-packet-template.md)
- [`dev/workflows/stage-a-acceptance-work-packet.md`](dev/workflows/stage-a-acceptance-work-packet.md)
- [`dev/workflows/tui-acceptance-harness.md`](dev/workflows/tui-acceptance-harness.md) — state-aware Windows Terminal scenario runner, evidence channels, security, and usage
- [`dev/workflows/tui-acceptance-harness-work-packet.md`](dev/workflows/tui-acceptance-harness-work-packet.md)
- [`dev/workflows/stage-c-separation-work-packet.md`](dev/workflows/stage-c-separation-work-packet.md) — accepted application/agent/presentation ownership extraction packet
- [`dev/workflows/stage-d-workspace-configuration-work-packet.md`](dev/workflows/stage-d-workspace-configuration-work-packet.md) — accepted Stage D scope, evidence, security, verification, and closure record
- [`dev/workflows/stage-e-agent-runtime-work-packet.md`](dev/workflows/stage-e-agent-runtime-work-packet.md) — accepted Stage E scope, Legacy evidence audit, verification, and closure record
- [`dev/workflows/stage-f-recomp-tools-work-packet.md`](dev/workflows/stage-f-recomp-tools-work-packet.md) — accepted Stage F tool scope, Legacy evidence, security, verification, and closure record
- [`dev/workflows/stage-g-commands-skills-work-packet.md`](dev/workflows/stage-g-commands-skills-work-packet.md) — active Stage G scope, authorities, command dispositions, skill evaluation, security, and closure record
- [`dev/workflows/default-gemma-profile-work-packet.md`](dev/workflows/default-gemma-profile-work-packet.md) — fresh-project LM Studio Gemma default change
- [`dev/workflows/default-devstral-profile-work-packet.md`](dev/workflows/default-devstral-profile-work-packet.md) — current Devstral default with retained Gemma profile

### Roadmap

- [`dev/roadmap/integration-order.md`](dev/roadmap/integration-order.md)
- [`dev/roadmap/legacy-parity-policy.md`](dev/roadmap/legacy-parity-policy.md) — behavior/capability parity with evidence-backed modernization allowed
- [`dev/roadmap/agent-skills-standard.md`](dev/roadmap/agent-skills-standard.md) — mandatory future Agent Skills client and per-skill migration policy

### Reference

- [`dev/reference/source-inputs.md`](dev/reference/source-inputs.md)
- [`dev/reference/strip-manifest.md`](dev/reference/strip-manifest.md)
- [`dev/reference/legacy-feature-holding-pen.md`](dev/reference/legacy-feature-holding-pen.md)
- [`dev/reference/upstream-85409d1-tree.sha256`](dev/reference/upstream-85409d1-tree.sha256)

### Skill migrations

- [`dev/skills/core-re-migration.md`](dev/skills/core-re-migration.md) — individual standards-based migration and still-open evaluation evidence
- [`dev/skills/evidence-mode-migration.md`](dev/skills/evidence-mode-migration.md) — evidence-classification migration and still-open evaluation evidence
- [`dev/skills/build-fix-loop-migration.md`](dev/skills/build-fix-loop-migration.md) — build-diagnosis migration and still-open evaluation evidence
- [`dev/skills/function-discovery-migration.md`](dev/skills/function-discovery-migration.md) — binary function-inventory migration and still-open evaluation evidence
- [`dev/skills/file-format-reversing-migration.md`](dev/skills/file-format-reversing-migration.md) — binary-format migration and still-open evaluation evidence
- [`dev/skills/project-handoff-migration.md`](dev/skills/project-handoff-migration.md) — resumable project-state migration and still-open evaluation evidence
- [`dev/skills/recomp-foundations-disposition.md`](dev/skills/recomp-foundations-disposition.md) — evidence-backed rejection of the stale static link-router format
- [`dev/skills/cdb-debug-migration.md`](dev/skills/cdb-debug-migration.md) — CDB host-debugging migration and still-open runtime/evaluation evidence
- [`dev/skills/gb-recomp-migration.md`](dev/skills/gb-recomp-migration.md) — Game Boy static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/gc-decomp-migration.md`](dev/skills/gc-decomp-migration.md) — GameCube static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/gen-decomp-migration.md`](dev/skills/gen-decomp-migration.md) — Genesis/Mega Drive decompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/bizhawk-stage-h-disposition.md`](dev/skills/bizhawk-stage-h-disposition.md) — deliberate deferral of the MCP BizHawk bridge to Stage H
- [`dev/skills/ghidra-mcp-stage-h-disposition.md`](dev/skills/ghidra-mcp-stage-h-disposition.md) — deliberate deferral of the Ghidra MCP integration to Stage H
- [`dev/skills/mcp-pine-stage-h-disposition.md`](dev/skills/mcp-pine-stage-h-disposition.md) — deliberate deferral of the RPCS3 PINE MCP bridge to Stage H
- [`dev/skills/imhex-migration.md`](dev/skills/imhex-migration.md) — interactive ImHex Pattern Language migration and still-open runtime/evaluation evidence
- [`dev/skills/n64-debug-mcp-stage-h-disposition.md`](dev/skills/n64-debug-mcp-stage-h-disposition.md) — deliberate deferral of the N64 runtime MCP bridge to Stage H
- [`dev/skills/n64-decomp-migration.md`](dev/skills/n64-decomp-migration.md) — N64 matching/static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/objdiff-migration.md`](dev/skills/objdiff-migration.md) — objective object-code matching migration and still-open runtime/evaluation evidence
- [`dev/skills/pcrecomp-migration.md`](dev/skills/pcrecomp-migration.md) — legacy PC static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/pcsx2-stage-h-disposition.md`](dev/skills/pcsx2-stage-h-disposition.md) — deliberate deferral of the PCSX2 runtime MCP bridge to Stage H
- [`dev/skills/ps2recomp-migration.md`](dev/skills/ps2recomp-migration.md) — PlayStation 2 static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/ps3recomp-migration.md`](dev/skills/ps3recomp-migration.md) — PlayStation 3 static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/sega2asm-merge-disposition.md`](dev/skills/sega2asm-merge-disposition.md) — merge of the tool-neutral Sega splitter workflow into `gen-decomp`, with MCP deferred
- [`dev/skills/snesrecomp-migration.md`](dev/skills/snesrecomp-migration.md) — SNES static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/vb-decomp-migration.md`](dev/skills/vb-decomp-migration.md) — Virtual Boy static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/windows-game-decomp-migration.md`](dev/skills/windows-game-decomp-migration.md) — Windows game reconstruction migration and still-open runtime/evaluation evidence
- [`dev/skills/xbox360-decomp-migration.md`](dev/skills/xbox360-decomp-migration.md) — Xbox 360 decompilation/static-recompilation migration and still-open runtime/evaluation evidence
- [`dev/skills/xboxrecomp-migration.md`](dev/skills/xboxrecomp-migration.md) — original-Xbox static-recompilation migration and still-open runtime/evaluation evidence
