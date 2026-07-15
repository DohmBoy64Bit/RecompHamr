# Barebones Strip Manifest

This manifest describes intentional differences from the pinned upstream CodeHamr tree.

## Removed directories

- `.devcontainer/`
- `example/`
- `internal/cloud/`
- `internal/update/`
- `cmd/codehamr/` after the RecompHamr entrypoint was created

## Removed top-level files

- `Dockerfile`
- `.goreleaser.yaml`
- `install.cmd`
- `install.sh`
- `codehamr.gif`

## Removed workflow

- upstream release/publish automation

A verification-only workflow remains.

## Removed tool implementation

- `bash` tool implementation and platform helpers
- bash-specific tests

The baseline command tool is now `powershell`, with `read_file`, `write_file`, and `edit_file` retained.

## Hosted-service removal

Removed from active code:

- HamrPass profiles and slash command;
- CodeHamr hosted endpoint assumptions;
- budget/quota headers and status rendering;
- hosted-service-specific auth/depletion behavior.

Generic bearer authentication and generic OpenAI-compatible endpoint errors remain in `internal/provider`.

## Product naming

Active product surfaces use:

- executable: `recomphamr`
- config directory: `.rehamr`
- environment prefix: `RECOMPHAMR_`
- module path: `github.com/DohmBoy64Bit/RecompHamr`

Upstream names remain only where legally or historically required, such as lineage documentation and the MIT license.

## TUI policy

No framework upgrade was applied. The pinned versions remain:

- Bubble Tea `v1.2.4`
- Bubbles `v0.20.0`
- Lip Gloss `v1.0.0`
- Glamour `v0.8.0`

The TUI is not being redesigned. Only product branding, PowerShell naming, and removal of hosted-service-only UI state are intentional Stage A differences.

## Tests

Hosted-service- and bash-specific tests were removed with the features they exercised. Focused retained tests remain. The very large upstream monolithic `internal/tui/tui_test.go` was not carried wholesale because it contains extensive tests for removed HamrPass, cloud-budget, and bash behavior mixed with general TUI tests. This is a documented test-coverage debt: general retained behavior must be recovered as focused tests before baseline closure rather than pretending those removed tests still apply unchanged.

## Verification hardening added by the template

The stripped baseline adds verification-only infrastructure that was not part of the pinned upstream runtime:

- `cmd/docscheck` and `docs/documentation-contract.json` for required durable documentation and mandatory content terms;
- `cmd/coveragecheck` for exact 100% Go statement-coverage enforcement;
- `scripts/check-coverage.ps1` to generate a temporary atomic coverage profile and enforce it;
- the existing Markdown-link, baseline-policy, architecture, formatting, build, and CLI smoke checks remain part of the canonical gate.

These commands are repository verification tools. They are not part of the interactive RecompHamr runtime feature surface.
