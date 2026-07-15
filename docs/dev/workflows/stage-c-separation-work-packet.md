# Stage C Separation Work Packet

## Outcome

RecompHamr application lifecycle and agent-turn orchestration are owned outside presentation, while the accepted TUI layout, interaction behavior, dependency versions, tools, transport, configuration contract, and runtime results remain equivalent.

## In scope

- `internal/app` composition and application lifecycle.
- `internal/agent` turn lifecycle, context coordination, streaming, tool sequencing, cancellation, and loop policy.
- Typed contracts between application/runtime ownership and `internal/tui`.
- Entrypoint reduction, equivalence tests, architecture checks, documentation, and runtime acceptance.

## Out of scope

- Legacy capabilities, MCP, skills, new commands, TUI redesign, Charm dependency upgrades, provider changes, or tool behavior changes.

## Authorities

- Root and TUI `AGENTS.md` instructions.
- `../architecture/current-baseline.md` and `../architecture/target-separation.md`.
- Accepted Stage A behavioral inventory and runtime evidence.
- Current source/tests and Bubble Tea v1.2.4 behavior.

## Evidence before editing

- `cmd/recomphamr` currently bootstraps configuration, applies environment overrides, opens TUI debug logging, creates the LLM client and TUI model, constructs Bubble Tea, clears the inline terminal, and runs the program.
- `internal/tui.Model` currently owns configuration reload/persistence, client replacement, context packing, chat streaming, tool execution, cancellation, and agent-loop nudges in addition to presentation.
- The accepted state-aware TUI harness can reproduce startup, resize, exit, restoration, model streaming, cancellation, and recovery without production test hooks.

## Implementation approach

Move ownership in narrow slices with behavior-preserving adapters. First extract startup composition into `internal/app` and leave the TUI model contract unchanged. Then extract turn orchestration into `internal/agent`, connect it through typed frontend events/intents, and finally remove remaining persistence/network/process ownership from `internal/tui`.

## Behavioral surface inventory

Each slice updates the active inventory before closure. Stage C must preserve every Stage A row and add architecture/lifecycle rows for composition, intent delivery, exactly-once actions, stream/tool ordering, cancellation, persistence, and frontend deletion boundaries.

## Verification

- Focused tests for `cmd/recomphamr`, `internal/app`, `internal/agent`, and affected TUI adapters.
- Architecture checks proving backend packages do not import presentation except the application composition root.
- `pwsh -NoProfile -File ./scripts/verify.ps1` at 100% statements.
- State-aware Windows Terminal smoke and real-model cancellation scenarios when affected slices reach runtime boundaries.

## Documentation impact

Update package/exported-symbol Go documentation, architecture diagrams, behavioral inventory, documentation map/contract, decisions, and runtime evidence for each completed slice.

## Security impact

Preserve current-user-only private state, secret non-disclosure, bounded tool execution, cancellation cleanup, log protection, and provider authentication behavior. Typed boundaries must not expose keys or raw debug content to presentation.

## Stop condition

Deleting `internal/tui` from a separated build removes presentation only: application lifecycle, model orchestration, configuration persistence, tool execution, and cancellation remain buildable and testable behind frontend contracts. The canonical gate and required Windows Terminal scenarios pass with unchanged accepted layout and behavior.

## Completion evidence

- Changed: open.
- Documented: open.
- Verified: open.
- Coverage: open.
- Security: open.
- Evidence: open.
- Known limits: Stage C remains open until all ownership slices meet the stop condition.

## Slice 1 — application composition

- Changed: moved configuration bootstrap, environment override, client/frontend construction, debug-log lifecycle, inline startup, and Bubble Tea lifecycle from `cmd/recomphamr` to `internal/app`; the entrypoint now parses process arguments and delegates.
- Documented: transitional architecture, decision D-011, behavioral row `APP-01`, and this slice record.
- Verified: focused command/app tests at 100%, canonical gate at 100%, and the state-aware Windows Terminal startup/resize/exit/restored-shell scenario against the exact build.
- Security: private logging and key resolution retain the same implementations and lifetime; no secret crosses a new presentation contract in this slice.
- Evidence: local harness report `E:\ReProject\StageA-Acceptance\StageC-App-Slice\report.json` and reviewed standard/constrained screenshots remain outside the repository.
- Known limits: turn orchestration, tool execution, config persistence, and client replacement still reside in `internal/tui`; later slices must move them without behavior changes.
