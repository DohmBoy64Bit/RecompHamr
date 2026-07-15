# RecompHamr Fresh Barebones Template

This repository is a **fresh-source reconstruction** of the stripped RecompHamr baseline.

It starts from the official `codehamr/codehamr` source line, pinned to:

```text
85409d167b97bec64ee330d51872d358d3ce2d57
```

The supplied RecompHamr-Legacy repository was **not used as the starting code tree**. It is reserved for later feature-by-feature behavioral reference after the barebones baseline is accepted.

## Purpose

The order of work is intentionally strict:

1. strip CodeHamr to the maintainable local application core;
2. preserve the inherited TUI framework, layout, and interaction model;
3. prove the barebones baseline still works;
4. separate presentation from orchestration without changing observable behavior;
5. only then integrate verified RecompHamr capabilities.

There is **no Bubble Tea v2 migration** in this template and no planned TUI redesign. The inherited dependency versions are frozen until an explicit future decision changes that policy.

## Barebones runtime surface

The current template retains:

- the inherited Bubble Tea v1 TUI and interaction model;
- generic OpenAI-compatible model profiles;
- conversation/context packing;
- history persistence;
- streaming and cancellation;
- model switching;
- `powershell`;
- `read_file`;
- `write_file`;
- `edit_file`.

The current template removes:

- HamrPass and CodeHamr-hosted service behavior;
- `internal/cloud`;
- self-update logic and re-exec support;
- installers, promotional media, demo material, and upstream release automation;
- the inherited `bash` tool and its WSL/POSIX-shell requirement;
- MCP, skills, RecompHamr reverse-engineering tools, project memory, doctor flows, classifiers, and other Legacy feature families.

Those RecompHamr capabilities are **not placeholders**. They are simply not part of the accepted barebones baseline yet.

## TUI freeze

Until the baseline gate is accepted:

- do not upgrade Bubble Tea, Bubbles, Lip Gloss, or Glamour;
- do not rearrange the layout;
- do not redesign the composer, transcript, status line, model picker, resize behavior, or cancellation behavior;
- do not mix separation-of-concerns refactoring with feature integration.

Brand text may say RecompHamr and hosted-service-only UI paths may be removed. Those are not permission to redesign the screen.

## Build

Requirements:

- Go 1.26+
- PowerShell 7 (`pwsh`) recommended; Windows PowerShell is used as a fallback on Windows

```powershell
pwsh -NoProfile -File ./scripts/verify.ps1
```

The canonical verification gate requires the durable documentation contract, valid relative documentation links, architecture and formatting checks, passing Go tests, and 100% statement coverage before the build and CLI smoke test may pass.

Or:

```powershell
go run ./cmd/recomphamr
```

Configuration is stored in:

```text
.rehamr/config.yaml
```

## Repository map

```text
cmd/recomphamr/        process entrypoint
cmd/docscheck/         durable documentation contract checker
cmd/coveragecheck/     strict 100% statement coverage checker
internal/config/       configuration persistence
internal/ctx/          context packing
internal/llm/          OpenAI-compatible transport
internal/provider/     generic endpoint/auth error helpers
internal/tools/        four barebones tools
internal/tui/          inherited presentation and current Stage A orchestration
docs/user/             user-facing baseline docs
docs/dev/              governance, memory, verification, architecture, workflows, roadmap
scripts/               reproducible baseline checks
```

Stage A intentionally keeps some inherited orchestration inside `internal/tui`. That is documented temporary debt. The next architectural stage, after baseline acceptance, moves runtime orchestration behind application contracts while keeping the TUI visually and behaviorally unchanged.

## Start here

Read in this order before changing code:

1. [`AGENTS.md`](AGENTS.md)
2. [`docs/README.md`](docs/README.md)
3. [`docs/dev/verification/baseline-status.md`](docs/dev/verification/baseline-status.md)
4. [`docs/dev/workflows/baseline-gate.md`](docs/dev/workflows/baseline-gate.md)
5. [`docs/dev/architecture/current-baseline.md`](docs/dev/architecture/current-baseline.md)
6. [`docs/dev/architecture/target-separation.md`](docs/dev/architecture/target-separation.md)

## Provenance

The exact upstream reconstruction and strip scope are recorded in:

- [`docs/dev/memory/repository-lineage.md`](docs/dev/memory/repository-lineage.md)
- [`docs/dev/reference/upstream-85409d1-tree.sha256`](docs/dev/reference/upstream-85409d1-tree.sha256)
- [`docs/dev/reference/strip-manifest.md`](docs/dev/reference/strip-manifest.md)
- [`docs/dev/reference/source-inputs.md`](docs/dev/reference/source-inputs.md)

The upstream MIT license is retained in [`LICENSE`](LICENSE).
