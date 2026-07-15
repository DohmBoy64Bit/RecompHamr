# TUI Acceptance Harness Work Packet

## Outcome

A reusable scenario runner controls a real RecompHamr Windows Terminal session, waits on structured lifecycle evidence, captures named screenshots, verifies external effects, fails with sanitized diagnostics, and produces a machine-readable report.

## In scope

- Windows Terminal launch, handle discovery, focus, input, paste, resize, screenshot, and closure.
- Category-only debug-event waits and sequence assertions.
- Disposable fixture assertions, shell-restoration proof, scenario validation, examples, and documentation.

## Out of scope

- Changes to the frozen RecompHamr TUI or runtime behavior.
- OCR as a correctness oracle, arbitrary desktop automation, model installation, credential management, CI execution of interactive Windows Terminal, or replacement of Go unit/render tests.

## Authorities

Root and TUI `AGENTS.md`, the accepted baseline status, engineering workflow rules, target architecture, and the TUI design skill's layered testing guidance.

## Evidence before editing

Stage A used an inline `WScript.Shell.SendKeys` workflow plus a separate screenshot helper. The debug log already provides timestamped lifecycle categories, and the accepted configuration contract protects that raw log as sensitive owner-only state.

## Implementation approach

Keep automation under `scripts/acceptance`, outside production packages. Use a declarative JSON scenario, Win32 top-level-window identity, category-only state polling, allow-listed input, exact-window observation, bounded waits, and sanitized failure artifacts.

## Behavioral surface inventory

The harness surface is recorded as `AUT-01` in `../verification/stage-a-behavioral-surface.md`. Validation covers schema failure and cross-platform parsing; a real smoke run covers window, event, resize, screenshot, terminal restoration, filesystem, cleanup, and lifecycle behavior.

## Verification

- `pwsh -NoProfile -File ./scripts/check-tui-acceptance.ps1`
- Real Windows smoke scenario against the verified executable and disposable Stage A workspace.
- `pwsh -NoProfile -File ./scripts/verify.ps1`

## Documentation impact

Add the harness contract, examples, work packet, documentation-map entries, and machine-readable required-document terms.

## Security impact

The harness reads a sensitive debug log but extracts only category headers. It restores the clipboard, allow-lists key chords, confines file assertions to the selected disposable workspace, never copies raw logs, and warns that screenshots require review.

## Stop condition

Both example scenarios validate, the real smoke scenario proves state-aware startup through restored-shell closure, the canonical gate passes on both supported platforms, and documentation/behavioral inventory are complete.

## Completion evidence

- Changed: declarative scenario runner, Win32 window binding, state waits, allow-listed control, screenshots, fixture assertions, recovery ordering, restoration probe, and bounded cleanup.
- Documented: evidence channels, scenario contract, usage, security boundary, limitations, work packet, map, and behavioral inventory.
- Verified: two scenarios validate; malformed schema fails; real Windows smoke passed from state-aware startup through restored-shell command and confirmed window closure; canonical repository gate passed.
- Coverage: Go remains at 100% statements; `AUT-01` maps the complete new behavioral surface; durable docs and machine-readable terms cover the harness contract.
- Security: category-only log extraction, clipboard restoration, contained fixture/artifact paths, synthetic committed prompts, sanitized reports, and failure-window cleanup.
- Evidence: `E:\ReProject\StageA-Acceptance\Harness-Smoke-5\report.json` and its two locally reviewed screenshots; these runtime artifacts remain outside the repository.
- Known limits: execution is Windows Terminal/Windows only; interactive CI is intentionally unsupported; responsive resize uses calibrated pixels after the initial exact cell-size launch; OCR is not a correctness oracle.
