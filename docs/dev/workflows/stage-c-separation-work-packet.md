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

- Changed: accepted. Turn lifecycle, model streaming, sequential tools, cancellation, accounting, loop policy, diagnostics, and causal logging are agent-owned behind the app-composed runtime façade; cancellation during a running tool removes the incomplete model-facing goal so it cannot be reissued.
- Documented: current/target architecture, decisions D-012/D-013, active Stage C inventory, harness contract, package/exported-symbol documentation, and checkpoints 2A through 2L agree with the accepted implementation.
- Verified: focused agent/app/TUI/logging tests, architecture enforcement, the canonical repository gate, and the strengthened exact-build Windows Terminal Gemma scenario all pass.
- Coverage: exactly 100.0% statements (2028/2028), every Slice 2 behavioral row verified, and affected documentation audited with no stale open claim.
- Security: contexts, cancellation functions, credentials, reasoning, raw tool arguments/results, and log handles remain outside presentation snapshots; cancelled process cleanup, absent side effects, stale-result rejection, cancelled-goal non-reissue, and report non-disclosure are proven.
- Evidence: commit containing this record; `E:\ReProject\StageA-Acceptance\StageC-Agent-Slice2\report.json` and reviewed `tools-complete.png`, `powershell-running.png`, and `recovered.png` from 2026-07-15 against `google/gemma-4-12b-qat` at 16,384 context tokens.
- Known limits: configuration persistence and model-client replacement remain later Stage C ownership slices and are not Slice 2 debt.

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

#### Checkpoint 2F — single state root and opaque deliveries

- Changed: the app-composed runtime now allocates the single turn, stream, and loop state roots shared by copied Bubble Tea models; request packing and model-round opening execute atomically inside `agent.Runtime.StartRound`, which returns only content-free packing facts; stream delivery internals are opaque and stable turn/round validation plus raw-event reduction occur inside `StreamState.ApplyDelivery`.
- Documented: the current architecture identifies the exact remaining mutable-pointer, logging, and deletion-boundary debt without treating this checkpoint as Slice 2 closure.
- Verified: focused agent/app/TUI tests cover runtime pointer sharing, packed-round dependencies and summary facts, current/stale/closed delivery handling, and retained adapter rendering; the strengthened architecture check forbids presentation from importing tools, constructing agent roots/executors, or classifying provider errors.
- Coverage: focused packages pass; the canonical repository gate is required before checkpoint commit, and Stage C behavioral rows remain unverified pending runtime evidence and final closure.
- Security: model-facing packed messages, raw transport events, stable round identity, contexts, credentials, and resolved tool arguments do not cross the production presentation delivery boundary; request summaries contain counts only.
- Evidence: production startup constructs one `agent.Runtime`; production chat startup calls `Runtime.StartRound`; stream messages carry opaque `StreamDelivery` values reduced by `ApplyDelivery`.
- Known limits: the adapter still retains pointers to agent-owned mutable state for Bubble Tea compatibility, agent-event logging is still emitted through TUI logging calls, configuration/client/persistence ownership is a later Stage C slice, and the required real-model multi-tool/cancel runtime record remains open.

#### Checkpoint 2G — cancellation regression and runtime scenario

- Changed: tool-result acceptance now requires both matching stable identity and an active turn, preserving the pre-extraction rule that a result delivered after cancellation cannot re-enter chat before a replacement turn exists; the state-aware harness adds exact event-count and absent-file assertions plus a real-model ordered-tool/cancel/recovery scenario.
- Documented: the harness contract names the new scenario and its non-disclosure/evidence semantics; this packet records why the lifecycle check is required in addition to `TurnID`.
- Verified: deterministic agent tests cover cancellation delivery both immediately after turn end and after a fresh turn; the canonical gate validates all three scenario schemas and passes every automated check. Real execution is currently unverified because `http://localhost:1234/v1/models` was not listening when checked.
- Coverage: the canonical repository profile passes at exactly 100.0% (1988/1988 statements); behavioral rows remain unverified until the exact-build real-model scenario runs successfully.
- Security: the scenario uses only synthetic prompts and disposable relative fixture paths; reports contain event categories, never bodies, and the protected debug log is not copied to artifacts.
- Evidence: `agent-tool-loop-cancel.json` waits for explicit bounded `tool_start` records, requires exactly paired successful results, asserts the cancelled result count does not advance, proves the delayed side-effect file is absent, and verifies recovery plus restored shell control.
- Known limits: exact-build Windows Terminal execution requires LM Studio to listen on the configured local endpoint with the acceptance model loaded.

#### Checkpoint 2H — private cancellation capability

- Changed: `TurnState` no longer exports its `context.Context` or `CancelFunc`; model-round and tool-work construction consume those capabilities only inside `internal/agent`, while TUI cancellation checks the typed active lifecycle state and ends the turn through agent methods.
- Documented: current architecture and the architecture guard distinguish a presentation-visible lifecycle fact from the private cancellation capability that enforces it.
- Verified: focused lifecycle, stream, loop, cancellation, queue, and TUI state-machine tests pass; after clearing only the rebuildable Go build cache to recover linker space, the canonical gate passes every automated check and validates all three scenarios.
- Coverage: the canonical repository profile passes at exactly 100.0% (1988/1988 statements); real-model behavioral evidence remains blocked independently of automated coverage.
- Security: contexts and cancellation functions cannot be read, stored, logged, or invoked directly by presentation; opaque tool work remains the only carrier into tool execution.
- Evidence: production TUI code contains no turn-context or cancel-function access, and the architecture check rejects either access pattern if reintroduced.
- Known limits: exact-build runtime evidence remains blocked after three LM Studio load attempts failed for `google/gemma-4-12b-qat`: automatic settings failed at 52%, 4,096 context/50% GPU failed at 45%, and 2,048 context/full GPU failed at 53%, each with LM Studio's unknown model-load error. Agent-event logging and broader application/configuration ownership remain later Stage C slices.

#### Checkpoint 2I — runtime façade and causal observation

- Changed: production TUI turn/stream/tool transitions now pass through the app-composed `agent.Runtime` façade; agent orchestration emits request, retry, reasoning, assistant, accounting, pressure, error, tool, nudge, leak, cancellation, and turn records at their owning transitions. The protected file sink moved from presentation into `internal/logging`, with lifecycle and observer injection owned by `internal/app`.
- Documented: current architecture, decision D-013, the active behavioral inventory, and this packet distinguish causal runtime observation from presentation rendering.
- Verified: focused agent/runtime-observer, logging lifecycle/security, app composition, and TUI adapter suites pass. Architecture enforcement forbids TUI imports of tools/logging, direct agent-root construction, provider-error policy, and cancellation capabilities; history/configuration filesystem ownership remains explicitly outside this slice.
- Coverage: focused `internal/agent` and `internal/logging` profiles and the canonical repository profile pass at exactly 100.0% (2011/2011 statements).
- Security: prompts, reasoning, tool arguments, and results remain confined to the owner-protected log; acceptance reports consume category names only. Presentation receives no log handle, observer, credentials, context, cancel function, or resolved tool arguments.
- Evidence: production stream and tool messages call `agent.Runtime.ApplyDelivery`, `ApplyToolResult`, `NextTool`, and `DecideClose`; application startup calls `logging.Open`, injects `logging.NewObserver`, and defers `logging.Close`.
- Known limits: exact-build Gemma runtime execution remains blocked by the recorded LM Studio load failure; broader configuration/client/persistence extraction is outside Slice 2 and remains later Stage C work.

#### Checkpoint 2J — immutable presentation snapshots

- Changed: production TUI phase, connectivity, retry, token, context-hint, active-turn, and current-stream access now goes through `agent.Runtime` methods; rendering consumes an immutable `agent.Snapshot`. Probe and ping results update runtime facts through typed setters, and leaked-tool presentation follows the typed close reason rather than reading model history.
- Documented: current architecture and this checkpoint distinguish the opaque app-composed runtime handle from its immutable presentation facts.
- Verified: focused snapshot/lifecycle tests cover idle/active, connection, positive/absent context hints, current stream, accounting, and reset behavior; retained TUI input, probe, popover, render, queue, resize, and state-machine tests pass. Architecture enforcement rejects production TUI reads of mutable turn, stream, or loop component fields.
- Coverage: focused `internal/agent` and affected TUI suites and the canonical repository profile pass at exactly 100.0% (2017/2017 statements).
- Security: snapshots exclude history, contexts, cancellation, streams, tool calls/arguments, credentials, reasoning, and observer state. The only stream accessor returns the existing opaque reader needed for Bubble Tea scheduling.
- Evidence: production rendering calls `Runtime.Snapshot`; input calls `Runtime.Active`; probe/context paths call typed runtime methods; no production TUI source matches `m.turn.`, `m.runtime.`, or `m.loop.`.
- Known limits: exact-build Gemma runtime evidence remains open; the model is now loaded, but the Codex desktop host denied foreground activation of the bound Windows Terminal window before input delivery. Configuration/client/persistence extraction is explicitly outside Slice 2.

#### Checkpoint 2K — deletion-boundary and runtime audit

- Changed: no production behavior changed. The final source audit confirmed that TUI orchestration references are confined to the app-composed runtime façade; retained state aliases exist only as focused test-observation seams, and the agent-owned cancellation capability remains private.
- Documented: current architecture and this packet now replace the obsolete Gemma-load blocker with the directly observed Windows foreground-activation blocker.
- Verified: the canonical gate passes on the exact build. Source searches found no production TUI imports of tools/logging, mutable agent-component reads, raw transport handling, direct tool execution, cancellation capabilities, or loop-policy decisions.
- Coverage: the canonical repository profile remains exactly 100.0% (2017/2017 statements); behavioral rows remain `unverified` until runtime acceptance completes.
- Security: three acceptance launches failed before any synthetic prompt was delivered. Reports remain sanitized, the private log was not copied, and no fixture tool or process side effect ran.
- Evidence: LM Studio reports `google/gemma-4-12b-qat` loaded at 16,384 tokens on port 1234. Normal foreground activation with verification retries, attached-input activation, and direct window switching were each denied by the Codex desktop host; the sanitized report records failure at the first `type_text` step.
- Known limits: run the committed `agent-tool-loop-cancel.json` scenario from an interactive desktop PowerShell/Windows Terminal session against the same build, then review its report and screenshots before changing `AGENT-01`, `AGENT-02`, or `TUI-04` to verified.

#### Checkpoint 2L — cancellation non-reissue and runtime acceptance

- Changed: runtime review exposed that rejecting a cancelled tool result was insufficient: the unresolved user instruction remained in model history and Gemma reissued it during recovery. `Runtime.CancelTurn` now discards only the incomplete model-facing goal when cancellation interrupts `PhaseRunning`; streaming cancellations retain their prior history. The harness now verifies foreground ownership after a bounded Alt activation fallback and asserts tool counts again after recovery.
- Documented: current architecture, the active inventory, harness contract, and Slice 2 completion evidence record the corrected cancellation invariant and exact artifacts.
- Verified: `TestRuntimeCancelTurnDropsOnlyRunningToolGoal` proves running-tool removal and non-tool retention. Focused agent/TUI tests, scenario validation, and the canonical gate pass. The strengthened real-model scenario passes in Windows Terminal: four ordered fixture calls/results, a fifth cancelled PowerShell start, exactly four total accepted results, no reissued sixth tool, no side-effect file, `RECOVERED`, clean exit, and restored shell.
- Coverage: canonical atomic profile is exactly 100.0% (2028/2028 statements); `AGENT-01`, `AGENT-02`, and `TUI-04` are verified with source, tests, and runtime evidence.
- Security: cancellation removes the unresolved side-effect instruction from model-facing history without deleting visible transcript or prompt recall; no prompt, key, raw log, or unrelated data was copied to the report or repository.
- Evidence: sanitized report and three reviewed screenshots at `E:\ReProject\StageA-Acceptance\StageC-Agent-Slice2`, generated 2026-07-15 with LM Studio `google/gemma-4-12b-qat` and the exact canonical build.
- Known limits: none within Slice 2. Later Stage C slices still own configuration/client/persistence extraction.

## Slice 3 — session, configuration, and persistence ownership

### Outcome

An application-composed runtime outside `internal/tui` owns configuration reload and profile activation, concrete model-client construction/replacement, backend reachability/probing, and prompt-history filesystem persistence. The TUI consumes immutable non-secret snapshots, emits typed intents, and preserves the accepted `/models`, `/clear`, startup, footer, popover, history, and error behavior exactly.

### In scope

- A neutral `internal/session` runtime composed by `internal/app`, avoiding an `app`/presentation import cycle.
- Immutable active-profile/model-list snapshots that exclude resolved keys.
- Existing config bootstrap reload semantics, process-only URL override preservation, atomic active-profile persistence, and concrete client replacement when the resolved URL/model/key triple changes.
- Startup reachability/probe behavior, keyed-profile activation probes, context-window results, stale-result identity, and accepted diagnostic strings.
- Existing `.rehamr/history` load, append, trim, malformed/oversized-entry recovery, security tightening, concurrency, clear, and failure behavior.
- Focused equivalence tests, architecture enforcement, synchronized documentation, canonical verification, and exact-build Windows Terminal profile/persistence acceptance.

### Out of scope

Configuration schema/default changes, new profile commands, provider or transport changes, agent-loop changes, Legacy capabilities, MCP, skills, TUI redesign, Charm upgrades, and changes to user-owned config/history contents outside disposable tests.

### Evidence before editing

- `internal/tui.Model` stores `*config.Config` and `*llm.Client`; rendering, popovers, request startup, and diagnostics read those concrete values.
- `internal/tui/slash.go` re-runs `config.Bootstrap`, preserves `URLOverride`, persists `/models <name>`, compares resolved endpoint/model/key triples, and constructs replacement clients.
- `internal/tui/model.go` and `probe.go` execute `provider.Reachable` and `llm.Client.Probe` through Bubble Tea commands and apply profile-tagged results.
- `internal/tui/history_store.go` directly opens, scans, locks, trims, secures, rewrites, and removes `.rehamr/history`; `/clear` and submit call it directly.
- Retained interaction, state-machine, history-store, configuration, rendering, and acceptance tests encode the required observable contracts.

### Implementation approach

Move one owner at a time behind an app-composed runtime. First relocate the history implementation without changing its format or failure policy. Then move config/profile/client state as one atomic service so a persisted selection and the client used by the agent cannot diverge. Finally move reachability/probe execution and replace presentation reads with immutable non-secret snapshots. Temporary adapters are allowed only between passing checkpoints and must be deleted before closure.

### Verification

- Focused `internal/session`, `internal/app`, `internal/config`, `internal/agent`, and TUI adapter tests at 100% statements.
- Architecture checks forbidding production TUI imports of `internal/config`, `internal/llm`, and `internal/provider`, direct config/history filesystem calls, and concrete client construction.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` at exactly 100% statements.
- Exact-build Windows Terminal evidence for startup profile, `/models` listing/switch, keyed or keyless activation result as applicable, persisted selection after restart, history recall/clear, real model response, screenshots, clean exit, and restored shell.

### Documentation impact

Update current architecture, decisions if the new ownership boundary is durable, active Stage C inventory, harness/scenario documentation, package/exported-symbol Go documentation, user configuration/history contracts if behavior requires clarification, and this completion record in the same checkpoints as implementation.

### Security impact

Resolved credentials must remain inside the session/client boundary and never enter snapshots, messages, reports, or screenshots. Preserve Windows current-user-only DACLs, POSIX owner-only modes, symlink/reparse refusal, atomic config persistence, private history permissions, process-only URL overrides, and sanitized runtime evidence.

### Stop condition

Production `internal/tui` no longer imports config, LLM, or provider packages; does not bootstrap, save, reload, or resolve configuration; does not construct or replace clients; does not execute backend probes; and performs no history filesystem I/O. All retained Slice 3 rows are verified with exact tests, documentation, canonical coverage, and required runtime evidence.

### Completion evidence

- Changed: accepted. Prompt-history persistence, configuration reload/profile activation, concrete model-client lifecycle, and backend reachability/probe execution are owned by the app-composed `internal/session` runtime; presentation consumes only non-secret snapshots and opaque work.
- Documented: current architecture, decision D-014, active behavioral inventory, package/exported-symbol contracts, architecture enforcement, and checkpoints 3A/3B agree with the accepted ownership boundary.
- Verified: focused session/app/TUI/config/LLM/provider tests, architecture enforcement, the canonical repository gate, and the complete user-observed Windows Terminal checklist pass against implementation commit `39fb263`.
- Coverage: exactly 100.0% repository statement coverage, every Slice 3 behavioral row verified, and affected documentation audited with no stale open claim.
- Security: resolved credentials remain inside configuration/client construction; snapshots expose only keyed booleans; private history/config enforcement, atomic persistence, rollback, override handling, and secret non-disclosure are preserved.
- Evidence: source and automated evidence at `39fb263`; on 2026-07-15 the user confirmed all 12 manual checks passed with LM Studio `google/gemma-4-12b-qat`, including startup/connectivity, model list/switch, real response, persisted restart selection, history recall/clear, 120×36/80×24/50×16 rendering, clean Ctrl+D exit, and restored PowerShell input. No screenshot files were supplied or claimed as retained artifacts.
- Known limits: none within Slice 3. Later Stage C work remains governed by the overall separation stop condition.

#### Checkpoint 3A — prompt-history persistence

- Changed: moved the retained `.rehamr/history` implementation from `internal/tui` into `internal/session.History`; `internal/app` constructs the store, and presentation maps typed strings to its existing prompt-entry state while preserving submit and `/clear` behavior.
- Documented: current architecture, package/exported-symbol documentation, architecture enforcement, and this checkpoint identify session ownership without changing the user-facing format or policy.
- Verified: focused session/app/TUI tests pass; relocated tests cover missing/unreadable files, Unicode and multiline round trips, malformed-line skipping, empty/oversized/scanner-incompatible omission, exact cap/trim direction, concurrent append preservation, protection/open/write/close/count/remove failures, idempotent clear, and adapter startup/submit/clear paths.
- Coverage: `internal/session` passes at exactly 100.0% focused statement coverage; repository coverage and Slice 3 behavioral rows remain open until later checkpoints and runtime evidence complete.
- Security: path protection still delegates to `config.RestrictPrivatePath`; the app injects only an immutable directory-backed store, and no history content is added to snapshots, logs, reports, or config state.
- Evidence: production TUI has no history filesystem helper or direct file operation for recall; architecture enforcement rejects the removed helper names and prevents `internal/session` from importing presentation.
- Known limits: config/profile/client and backend probe ownership remain in TUI; `SESSION-03` remains `unverified` until exact-build restart recall and `/clear` evidence is recorded.

#### Checkpoint 3B — configuration, client, and probe runtime

- Changed: added one app-composed `internal/session.Runtime` that owns configuration reload, profile activation/persistence, concrete client construction/replacement, process-only URL override retention, prompt history, and captured reachability/authenticated-probe work. The agent uses the stable session chat contract while presentation consumes immutable snapshots and opaque work.
- Documented: current architecture, decision D-014, architecture enforcement, active inventory, package/exported-symbol contracts, and this checkpoint agree on the new owner and non-secret boundary.
- Verified: focused session/app/TUI tests pass at exactly 100.0% statements; tests cover snapshot secrecy, successful/failed activation, save rollback, unchanged/changed/failed reload, URL override, real chat/probe transport, captured reachability identity, history delegation, stale UI results, keyed/keyless startup, lists, switching, reload diagnostics, and rendering adapters.
- Coverage: focused and canonical repository gates pass at exactly 100.0% statements; Slice 3 behavioral rows are verified by the completed runtime checklist.
- Security: resolved keys remain private to config/client construction and are represented only as booleans in snapshots; save rollback and existing private-path enforcement are preserved.
- Evidence: production TUI has no config/LLM/provider import, client construction, probe transport, or persistence filesystem call; architecture enforcement rejects their return.
- Known limits: none within this checkpoint; manual evidence is user-observed and no screenshot files were supplied or recorded.

## Slice 4 — typed frontend boundary and Stage C closure

### Outcome

Core application orchestration communicates with presentation only through backend-neutral typed frontend contracts. `internal/tui` owns Bubble Tea rendering, local editor/popover/queue state, and input-to-intent translation; an application-owned controller owns every agent/session action and asynchronous completion. A thin application terminal adapter is the sole concrete TUI wiring edge.

### In scope

- Backend-neutral intents, immutable snapshots, ordered presentation events, and opaque asynchronous work/completions in `internal/frontend`.
- An application controller for startup/history, configuration/profile/probe lifecycle, turn/stream/tool sequencing, cancellation, reset, final accounting, exactly-once completion, and stale-work cleanup.
- Removal of production TUI dependencies on agent, session, context, configuration, transport, provider, tools, and logging packages.
- Isolation of Bubble Tea and concrete TUI construction in `internal/app/terminal`, with core `internal/app` independently buildable/testable through frontend contracts.
- Architecture enforcement, focused equivalence tests, documentation, canonical verification, exact-build Windows Terminal acceptance, and final Stage C closure.

### Out of scope

Visible layout or interaction changes, command additions, agent-policy changes, persistence-format changes, provider/tool changes, dependency upgrades, Legacy capabilities, MCP, skills, and Stage D feature work.

### Authorities

- Root, documentation, and TUI `AGENTS.md` instructions.
- Current and target architecture, decisions D-011 through D-014, accepted Slices 1 through 3, Stage A inventory/runtime evidence, current source/tests, and Bubble Tea v1.2.4 behavior.

### Evidence before editing

- Core `internal/app` imports `internal/tui` and Bubble Tea, constructs `tui.New`, delegates command help to `tui.PrintHelp`, and runs the concrete Tea program.
- Production TUI stores `agent.Runtime` and `*session.Runtime`, invokes their lifecycle methods, schedules their opaque stream/tool/probe/reachability work, and contains transitional mutable agent aliases used by tests.
- TUI-local queueing, prompt editing, popovers, transcript formatting, resize, footer timing, and rendering are presentation responsibilities that must remain local and byte/behavior equivalent.
- The accepted automated and Windows evidence for Slices 1 through 3 is the equivalence baseline; no Legacy implementation is an architectural authority.

### Implementation approach

Introduce contracts and an app-owned controller beside the accepted path, migrate session actions, then agent actions, and delete compatibility seams only after focused equivalence passes. Isolate the unavoidable concrete terminal wiring in an application subpackage so deleting the adapter removes the runnable terminal target while core application/backend behavior remains buildable and testable.

### Verification

- Focused frontend/controller/app/TUI tests at exactly 100% statements, covering every intent, event, work/completion, duplicate/stale/malformed input, cancellation, timeout, cleanup, persistence, security, and retained adapter behavior.
- Architecture checks forbidding core app-to-TUI/Bubble Tea imports, TUI-to-backend imports/lifecycle calls, frontend-to-backend/Bubble Tea imports, backend-to-TUI imports, and entrypoint bypass.
- A positive core-app/backend build/test that excludes the concrete TUI adapter.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` at exactly 100.0% statements.
- Exact-build Windows Terminal runs for the three committed harness scenarios plus profile/history persistence, multiline/queue, representative sizes, clean exit, and restored shell.

### Documentation impact

Update package/exported-symbol Go documentation, decision D-015, current/target architecture, active behavioral inventory, this packet, architecture enforcement, and runtime evidence. Update the documentation map/contract or user help only if their durable contracts actually change.

### Security impact

Frontend contracts must exclude credentials, model history, reasoning, contexts/cancel functions, raw transport events, tool arguments/results, process handles, log handles, and raw log content. Opaque work must retain captured endpoint/client/turn/round identity, bounded execution, cancellation cleanup, and stale-result rejection.

### Stop condition

Production TUI imports only `internal/frontend` among project runtime packages and emits typed intents without invoking backend lifecycle methods. Core `internal/app` imports neither TUI nor Bubble Tea and is buildable/testable with backend packages when the terminal adapter is excluded. Every Stage C row is verified, the canonical gate passes at 100%, required Windows evidence passes against the exact accepted build, and no compatibility adapter remains.

### Completion evidence

- Changed: open.
- Documented: pre-edit contract and inventory recorded.
- Verified: open.
- Coverage: open.
- Security: open.
- Evidence: accepted Slices 1 through 3 and current source/tests frozen as the equivalence baseline.
- Known limits: all Slice 4 implementation, deletion-boundary proof, and runtime acceptance remain open.
