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

## Slice 2 — agent-turn orchestration

### Outcome

`internal/agent` owns context coordination, model streaming, sequential tool dispatch, cancellation, token accounting, and loop policy. `internal/tui` retains rendering, presentation state, input translation, and queued-prompt editing while preserving the accepted layout and interaction contract.

### In scope

- Typed turn intents, immutable runtime snapshots, and ordered agent events.
- Stable turn and round identity, conversation history, context packing, streaming, tool sequencing, cancellation, accounting, and existing loop nudges.
- Application composition of the agent runtime and the Bubble Tea presentation adapter.
- Focused equivalence tests, architecture enforcement, documentation, and exact-build runtime acceptance.

### Out of scope

Configuration-format changes, configuration-persistence extraction, new providers, tool behavior changes, Legacy capabilities, MCP, skills, new commands, TUI redesign, and Charm dependency upgrades.

### Evidence before editing

- `internal/tui/model.go` currently owns the per-turn context, model-facing history, stream channel, pending tool queue, token/timing state, live context hints, and all loop-policy latches.
- `internal/tui/commands.go` currently reads model events and executes tools through Bubble Tea commands, tagging work with channels or context pointers to reject stale results.
- `state_machine_comprehensive_test.go`, `baseline_comprehensive_test.go`, and `queue_test.go` encode the accepted ordering, cancellation, nudge, queue, accounting, and stale-event behavior.
- The existing runtime harness proves streaming cancellation and recovery, but does not by itself prove multi-tool ordering or cancellation during tool execution.

### Implementation approach

Move ownership incrementally behind a typed runtime composed by `internal/app`. Stable turn and round identifiers replace channel/context identity at the presentation boundary. The agent owns asynchronous model/tool work and emits immutable presentation facts; the TUI adapter maps those facts into the existing Bubble Tea state without importing agent policy or executing side effects.

### Verification

- Focused contract tests for `internal/agent`, `internal/app`, and the TUI adapter at 100% statements.
- Architecture checks forbidding presentation imports from the agent and direct model/tool orchestration from the TUI.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` at exactly 100% statements.
- Exact-build Windows Terminal streaming/cancellation recovery plus a clean-log multi-tool and cancelled-PowerShell scenario.

### Documentation impact

Update the current architecture, target architecture as needed, decisions, active Stage C behavioral inventory, documentation map/contract, package and exported-symbol Go documentation, and this completion record in the same changes as ownership moves.

### Security impact

Keep credentials, contexts, private reasoning, and unrestricted tool arguments outside presentation snapshots. Preserve bounded tool execution, current cancellation/process cleanup, sanitized private logging, and provider authentication behavior.

### Stop condition

The TUI no longer opens model streams, executes tools, stores turn contexts, packs model history, or decides loop policy. All retained behavior has complete statement, behavioral-surface, documentation, and required runtime evidence, with no in-scope blocked or unverified row.

### Completion evidence

- Changed: open.
- Documented: this work packet and the active Stage C inventory establish the pre-edit contract.
- Verified: open.
- Coverage: open.
- Security: open.
- Evidence: source and accepted tests inventoried; implementation and runtime evidence remain open.
- Known limits: configuration persistence and model-client replacement remain later Stage C ownership slices except for the atomic runtime handoff required by this slice.

#### Checkpoint 2A — pure request and policy contracts

- Changed: added `internal/agent` package ownership for context-pack request construction, the four ordered tool definitions, assistant-finish classification, tool-target/failure classification, and the exact existing nudge texts and thresholds; the TUI delegates through temporary compatibility helpers.
- Documented: current transitional architecture and this work packet now distinguish pure agent ownership from the still-open mutable lifecycle ownership.
- Verified: focused `internal/agent` and `internal/tui` tests pass; `internal/agent` reports 100.0% statement coverage.
- Coverage: behavioral rows remain unverified until the asynchronous lifecycle, adapter, canonical gate, and runtime evidence are complete.
- Security: the extracted pure contracts contain no credentials, contexts, raw debug data, filesystem access, process execution, or network execution.
- Evidence: production TUI paths call `agent.BuildMessages`, `agent.Tools`, and agent policy helpers; focused tests cover every new agent statement.
- Known limits: stream reading, tool execution, context cancellation, accounting, policy latches, and model-facing history remain in `internal/tui` for the next checkpoint.

#### Checkpoint 2B — turn root and stable identity

- Changed: moved model-facing history, begin/end/reset context lifecycle, cancellation ownership, and stable monotonically increasing turn identity into `agent.TurnState`; accepted tool results now match `agent.TurnID` instead of exposing context identity to presentation.
- Documented: current architecture and this checkpoint distinguish the agent-owned turn root from the still-open stream/tool reducer.
- Verified: focused `internal/agent` and `internal/tui` tests pass, including context replacement cancellation, idempotent end/reset, accepted result handling, and stale result rejection.
- Coverage: `internal/agent` remains at 100.0% focused statement coverage; behavioral rows remain unverified until the complete adapter and runtime evidence exist.
- Security: contexts and cancel functions remain opaque runtime values and are neither logged nor included in presentation messages; `TurnID` carries no secret or user content.
- Evidence: production submit/end/clear/result paths use `agent.TurnState` and `agent.TurnID`.
- Known limits: stream channels, pending tool calls, accounting, phase transitions, and loop-policy latch state still reside in `internal/tui`.

#### Checkpoint 2C — model stream and accounting reducer

- Changed: moved model-round startup, opaque stream reading, stable round identity, phase/retry/connectivity transitions, pending-call collection, live context hints, token accounting, interrupted-turn finalization, and transport-event reduction into `internal/agent`; the TUI applies typed content/reasoning/flush/retry/done/error effects.
- Documented: current architecture and this checkpoint record the exact remaining close/tool/policy adapter debt.
- Verified: focused agent reducer tests cover every event kind, unknown events, malformed nil tool-call safety, retry clearing, content/reasoning/tool-argument accounting, authoritative and estimated usage, context hints, connection failures, opaque event/close delivery, stream replacement, and session finalization; retained TUI state/render tests pass.
- Coverage: `internal/agent` passes at 100.0% focused statement coverage; full behavioral rows remain unverified pending tool-loop extraction and exact-build runtime acceptance.
- Security: typed `StreamEffect` excludes contexts, credentials, and resolved tool arguments; presentation holds only an opaque stream reader and stable identifiers.
- Evidence: production chat startup and event handling call `StreamState.StartRound`, `Stream.Read`, `StreamState.Apply`, and `StreamState.Finalize`.
- Known limits: sequential tool execution, stream-close finish policy, policy latches, error diagnostic mapping, and app-composed frontend contracts remain open.

#### Checkpoint 2D — sequential tools and close policy

- Changed: moved pending-call popping, bounded inline status construction, actual local-tool execution, strict result pairing, same-target failure state, runaway/empty/verification latches, post-queue nudge injection, and stream-close decisions into `internal/agent`; TUI receives opaque tool work and typed result/close effects.
- Documented: current architecture and this checkpoint record that production TUI code no longer imports or invokes `internal/tools`.
- Verified: deterministic agent tests prove emission-order execution, complete sibling pairing before re-entry or nudges, cancellation through the root context, stale cancelled-result rejection after a fresh turn, failure and runaway thresholds, no nudge before queue drain, one-shot latches, empty rearm after tool progress, empty stall, leaked-call stop, verification re-prompt, honest `unverified` suppression, and clean finish.
- Coverage: focused `internal/agent` and retained TUI tests pass at 100.0% agent statements; repository and runtime closure remain pending the final adapter checkpoint.
- Security: presentation receives bounded inline status and tool result messages but no executable call arguments or process handle; cancellation retains the existing tool cleanup implementation.
- Evidence: production dispatch/result/close paths call `LoopState.NextTool`, `ToolWork.Run`, `LoopState.ApplyToolResult`, and `LoopState.DecideClose`.
- Known limits: app-composed typed frontend ownership, provider diagnostic mapping, agent logging, architecture enforcement, and new real-model tool-loop runtime evidence remain open.

#### Checkpoint 2E — application-composed runtime and diagnostics

- Changed: added a typed `agent.Runtime` aggregate; `internal/app` now constructs it with the selected model client and local-tool executor before frontend construction, and provider-specific stream/probe error classification moved from TUI helpers into `internal/agent`.
- Documented: the current architecture and decision D-012 identify this as a transitional composition checkpoint, not the final encapsulated frontend boundary.
- Verified: focused `internal/agent`, `internal/app`, and `internal/tui` tests prove dependency preservation, idle runtime initialization, unchanged frontend construction, and every retained diagnostic branch.
- Coverage: focused packages pass; the canonical repository gate and runtime evidence are required before this checkpoint is committed and before behavioral rows can be verified.
- Security: the runtime constructor receives the already-resolved client and executor without exposing credentials or tool arguments in a new presentation fact; diagnostic strings retain only the previously accepted profile name, endpoint, and error detail.
- Evidence: production application composition calls `agent.NewRuntime` exactly once and passes the result to `tui.New`; TUI diagnostic adapters call `agent.StreamErrorMessage` and `agent.ProbeErrorMessage`.
- Known limits: the Bubble Tea value adapter still unpacks mutable runtime components, orchestration logging is still emitted by TUI code, typed delivery state remains wider than the final snapshot/event contract, and exact-build tool-loop runtime evidence remains open.
