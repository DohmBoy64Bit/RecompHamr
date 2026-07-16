# Barebones Baseline Status

## Gate

**ACCEPTED — STAGE A BASELINE CLOSED**

Stage A was accepted on 2026-07-15. The accepted implementation commit is `71f70e922ff4cf5c5e05fd2b8325610425365b02`; the acceptance-record commit is the commit containing this file.

## Acceptance evidence

- Canonical local gate: `pwsh -NoProfile -File ./scripts/verify.ps1` on Go 1.26.4, including 100% statement coverage, documentation, links, architecture, formatting, build, and CLI smoke checks.
- Dual-platform CI: GitHub Actions run `29423037990` on `windows-latest` and `ubuntu-latest`.
- Behavioral and documentation audit: [`stage-a-behavioral-surface.md`](stage-a-behavioral-surface.md), with no blocked or unverified row.
- Runtime record: [`stage-a-runtime-acceptance.md`](stage-a-runtime-acceptance.md).
- Visual evidence: [`evidence/stage-a-wide-120x36.png`](evidence/stage-a-wide-120x36.png), [`evidence/stage-a-80x24.png`](evidence/stage-a-80x24.png), and [`evidence/stage-a-constrained-50x16.png`](evidence/stage-a-constrained-50x16.png).

## Accepted scope

- Fresh-source stripped baseline at upstream lineage `85409d167b97bec64ee330d51872d358d3ce2d57`.
- Frozen Bubble Tea v1 screen composition and dependency versions.
- Local OpenAI-compatible transport, four tools, configuration/history/log persistence, cancellation, resize, and terminal restoration.
- POSIX owner-only modes and Windows current-user-only DACL protection for private state.
- No Legacy features, MCP, skills, reverse-engineering tools, new commands, TUI redesign, or Stage C extraction.

## Next allowed work

Stage C separation was accepted at implementation commit `72e6b43215cc14f91eb6547de15a7386bc77b927`. Stage D workspace and configuration foundations are active under the Stage D work packet; feature integration remains governed by the documented integration order and change-control gates.
