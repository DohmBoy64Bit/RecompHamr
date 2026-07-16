# Stage E Agent and Runtime Work Packet

## Outcome

Determine the complete evidence-backed Legacy agent/runtime capability contract, compare it with the accepted Stage C backend, and implement only genuine gaps behind the existing application-owned boundary. Stage E closes with equivalent or improved agent policy, context management, and model-stream resilience without changing the frozen TUI or importing later-stage feature families.

## In scope

- Agent turn identity, sequential tool pairing, cancellation, stale-result rejection, final accounting, and deterministic loop policy.
- Repeated-failure, runaway, empty-reply, tool-call-leak, and substantial-turn verification behavior.
- Context budgeting, Unicode-safe tool-output truncation, valid tool-call grouping, original-task anchoring, and strict-backend-safe nudge packing.
- Reasoning streaming, structured tool fragments, reasoning/tool fallback, retry/backoff, cancellation, idle timeout, context-window probes, and transport error reduction as relied on by the agent runtime.
- A source/test/runtime traceability audit that may correctly conclude current behavior is already equivalent or improved.
- Focused regression work only where the audit proves a missing or defective contract.

## Out of scope

- New tools, reverse-engineering helpers, `repomixr`, or `recomp_reference` (Stage F).
- Slash commands, skill discovery/loading, skill-template classification, or classifier UI (Stage G).
- MCP configuration, transports, tools, or lifecycle (Stage H).
- Hosted-service budget/quota behavior removed in Stage A.
- TUI layout, interaction, command, persistence, or dependency changes.
- Speculative autonomy, hard-yield controls, parallel tool execution, or policy redesign not supported by verified evidence.

## Authorities

- Root and documentation `AGENTS.md`, engineering workflow, evidence, behavioral-surface, definition-of-done, and scope-control rules.
- Accepted Stage A/C/D inventories, Stage C work packet, current architecture, integration order, and Legacy parity policy.
- Current `internal/agent`, `internal/ctx`, `internal/llm`, `internal/app/controller`, and their complete tests.
- Legacy `internal/tui/model.go`, `internal/tui/commands.go`, `internal/ctx`, `internal/llm`, and relevant tests as historical behavioral evidence.

## Legacy evidence before editing

- Legacy keeps one cancellation root across every model and tool round, executes emitted tool calls sequentially, pairs every accepted result, rejects stale work, and removes a running-tool goal on cancellation so it cannot be reissued.
- Four soft policy backstops are bounded rather than hard stops: five same-target failures, 75 tool calls, one empty-round retry, and a verification re-ground after eight tool calls. Raw textual `<tool_call>` leakage ends stopped instead of being executed.
- Legacy context packing truncates large tool output with head/tail preservation, maintains Unicode validity, rejects orphan/dangling/partial tool groups, keeps the newest complete group when over budget, anchors a real user task, and demotes historical system nudges for strict OpenAI-compatible servers.
- Legacy streams content, reasoning, and fragmented tool calls; probes context size; retries selected transient failures; falls back once and sticks when a server rejects `reasoning_effort`; and uses an activity-reset idle watchdog rather than a total turn deadline.
- Direct inspection shows these contracts already have current owners and accepted tests after Stage C. Current context and transport implementations contain additional defensive structure and injected deterministic boundaries; similarity alone is not the acceptance proof.
- Legacy `internal/classifier` classifies skill document templates and is invoked from skill-related slash commands. It has no model-turn decision role and remains Stage G evidence, not Stage E production scope.

## Implementation approach

Use an audit-first adapted-compatibility approach. Build a line-of-evidence matrix from Legacy contract to current owner and exact tests. If source and existing tests prove a contract completely, record `equivalent` or `improved` without rewriting it. If a gap is found, add the smallest backend-only correction and regression test at the owning boundary. Do not move behavior back into presentation and do not add an API merely to resemble Legacy structure.

## Behavioral surface inventory

The active inventory is [`../verification/stage-e-behavioral-surface.md`](../verification/stage-e-behavioral-surface.md). `AGENT-03`, `CTX-02`, and `LLM-02` begin unverified and cannot close from source resemblance alone.

## Verification

- Contract-level source/test traceability for every Legacy agent/runtime branch in scope.
- Focused `internal/agent`, `internal/ctx`, `internal/llm`, and controller tests at exactly 100% statements.
- Architecture checks proving policy, history, raw transport, tool arguments/results, contexts, and cancellation capabilities remain outside TUI/frontend.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` at exactly 100.0% repository statement coverage.
- Exact-build Windows Terminal runs with LM Studio `google/gemma-4-12b-qat` for real streaming, ordered tools, stream/tool cancellation, stale-result recovery, clean exit, and restored shell if production behavior changes or closure evidence requires reconfirmation.
- Independent `windows-latest` and `ubuntu-latest` CI results for the accepted commit.

## Documentation impact

Maintain this packet and the Stage E inventory, then synchronize AGENTS, architecture, decisions, integration status, holding-pen dispositions, package/exported-symbol Go docs, and any genuinely affected user contract. No command/help documentation changes unless separately authorized in Stage G.

## Security impact

Agent history, reasoning, raw model events, tool arguments/results, contexts, cancel functions, observer output, endpoint clients, and process capabilities remain backend-private. Policy notes are model-facing lower-authority application content and must not be misrepresented as user messages. Cancellation must terminate active work and prevent stale or cancelled side effects from entering later turns. Transport errors and logs must not disclose credentials or raw private context.

## Stop condition

Stage E closes only when every in-scope Legacy agent/runtime contract is dispositioned with exact source and test evidence; genuine gaps are corrected without TUI or later-stage drift; `AGENT-03`, `CTX-02`, and `LLM-02` have no blocked or unverified category; focused and canonical coverage are exactly 100%; documentation is complete and factual; required runtime evidence and both CI platforms pass; and the skills classifier remains deferred to Stage G.

## Completion evidence

- Changed: work packet and initial evidence inventory only; no production behavior changed.
- Documented: Stage E scope, ownership, Legacy evidence, later-stage exclusions, security boundary, verification, and stop condition.
- Verified: initial source inspection identifies current owners and likely existing parity; the branch-by-branch traceability audit remains open.
- Coverage: accepted Stage C/D evidence is the baseline; all Stage E rows begin unverified.
- Security: no new capability or exposure is authorized by this checkpoint.
- Evidence: current and Legacy files named above plus accepted Stage C agent rows.
- Known limits: production equivalence, focused/canonical reruns, runtime reconfirmation, CI, and final parity dispositions remain open.
