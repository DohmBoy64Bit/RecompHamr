# State-Aware TUI Acceptance Harness

The Windows-only harness at `scripts/acceptance/Invoke-TuiAcceptance.ps1` drives the compiled RecompHamr application through a real Windows Terminal window. It complements Go state/render tests; it does not replace them or add test hooks to production code.

## Evidence channels

The harness combines three independent channels:

1. **Control** — a newly created Windows Terminal top-level window is identified by its Win32 handle. Named keys, clipboard paste, focus, movement, and resize actions target only that handle.
2. **State** — steps wait on category-only records from `.rehamr/log.txt`, such as `session`, `request`, `assistant`, `cancel`, and `turn_end`. Fixed sleeps are used only where visual settling is the behavior under observation.
3. **Observation** — exact-window screenshots, file/hash/content assertions, event ordering, window lifecycle, and a command executed after Ctrl+D independently prove visible and external effects.

Reports contain scenario/step labels, durations, terminal size, the last event categories, and failure text. They never copy event bodies, prompts, model responses, configuration values, or debug logs.

## Running a scenario

Build first, prepare a disposable workspace with logging enabled, and keep credentials out of the repository. Then run:

```powershell
pwsh -NoProfile -File ./scripts/acceptance/Invoke-TuiAcceptance.ps1 `
  -ScenarioPath ./scripts/acceptance/scenarios/smoke.json `
  -AppPath ./dist/recomphamr.exe `
  -Workspace E:\Disposable\RecompHamr-Acceptance `
  -ArtifactDirectory E:\Disposable\RecompHamr-Acceptance\artifacts
```

Use `-ValidateOnly` to check a scenario on any platform without opening a terminal. The canonical repository gate validates every committed scenario.

The `smoke.json` scenario proves startup, event readiness, pinned screenshots, resize handling, Ctrl+D terminal restoration, a command in the restored PowerShell shell, cleanup, and window closure. `model-stream-cancel.json` additionally exercises a real model turn, streaming screenshots, cancellation, recovery ordering, and clean exit. `agent-tool-loop-cancel.json` requires exactly four ordered fixture tool starts/results, verifies both files, cancels a long PowerShell call, proves its result and side effect stay absent, then proves recovery and terminal restoration. Real-model scenario endpoints must already be running.

## Scenario contract

Each JSON file has schema `version: 1`, a unique name, application/workspace/artifact/debug-log paths, an initial terminal size, and ordered steps with unique labels.

Supported steps:

- `launch` — open a new Windows Terminal window and bind its new top-level handle;
- `wait_event` — poll the private debug log until a category reaches a required count;
- `assert_event_count` — require an exact category count at a causal checkpoint, including proof that a cancelled stale result was not accepted;
- `assert_event_sequence` — require a category subsequence without reading or reporting bodies;
- `type_text` — paste Unicode/multiline text while restoring the previous clipboard afterward;
- `key` — send an allow-listed terminal key (`enter`, `tab`, `escape`, arrows, page keys, Ctrl+C, Ctrl+D, or Alt+F4);
- `resize` and `screenshot` — resize the bound window and capture only its rectangle;
- `assert_file` and `remove_file` — wait for and verify an isolated fixture by exact content or SHA-256, or require that a cancelled side effect is absent, then clean it up;
- `sleep` — a deliberate, labeled visual-settle delay;
- `close_window` — close the bound terminal and verify that its handle disappears.

Every step has a timeout where external progress is possible. On failure the harness writes `failure.png` when a window exists and always writes a sanitized `report.json` before failing the process.

## Security and limitations

- Run only in a disposable project directory. Model-requested tools retain the current user's permissions.
- Debug logging must be enabled, but the raw log remains in the protected `.rehamr` directory and is never copied to artifacts.
- Scenario files are source-controlled; use only synthetic prompts and never embed keys, private paths, or user data.
- Screenshot contents are evidence and require review before publication.
- Execution supports Windows Terminal on Windows. Validation is cross-platform. Pixel resize is calibrated from the current window and is intended for responsive-layout evidence, while initial `--size` provides the pinned cell dimensions.
