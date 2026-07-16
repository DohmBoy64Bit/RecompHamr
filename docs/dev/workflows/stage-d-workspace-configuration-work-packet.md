# Stage D Workspace and Configuration Work Packet

## Outcome

RecompHamr gains secure, application-owned workspace and configuration foundations that can support later capability integration without weakening the accepted Stage C boundary or changing the frozen terminal interface. The first slice gives the application a canonical workspace identity and a bounded persistent-state input that is refreshed for model turns while absence remains a normal no-op.

## In scope

- A backend workspace owner rooted at the canonical absolute project directory.
- Detection and bounded reading of `.rehamr/REPHAMR_STATE.md` as optional project state.
- System-prompt composition and refresh owned outside presentation.
- Symlink/reparse refusal, regular-file validation, size limits, Unicode preservation, deterministic errors, and secret-safe diagnostics.
- Complete package/exported-symbol documentation, behavioral inventory, architecture enforcement, tests, runtime evidence, and Stage D parity records.
- Audit of remaining configuration and diagnostic foundations to schedule later Stage D slices.

## Out of scope

- New slash commands, including Legacy `/init-re`, `/status-re`, and `/doctor`; command expansion remains Stage G.
- Reverse-engineering workspace templates, ledgers, recomp/decomp directories, `repomixr`, or other Stage F tool artifacts.
- `skills/`, skill activation, or skill prompt injection; those remain Stage G and must follow the Agent Skills standard.
- `mcp.json`, MCP environment diagnostics, server lifecycle, or MCP tools; those remain Stage H.
- Agent-loop policy additions, TUI layout or interaction changes, Charm upgrades, Legacy package copying, updater/installers, or Stage E and later capability work.

## Authorities

- Root/documentation/TUI `AGENTS.md` rules and the accepted Stage C architecture.
- Engineering workflow, behavioral-surface, evidence, scope-control, and definition-of-done contracts.
- Integration order, Legacy parity policy, and Legacy feature holding pen.
- Current `internal/app`, `internal/config`, `internal/session`, `internal/agent`, and `internal/frontend` source/tests.
- Exact Legacy project/config/TUI source and tests plus `docs/memory.md` and `docs/doctor.md` as historical behavioral evidence only.

## Legacy evidence before editing

- Legacy reads `.rehamr/REPHAMR_STATE.md` when present and non-empty, inserts it under `## Persistent Memory`, and otherwise continues without error.
- Legacy reconstructs that system prompt after `/clear`, so later turns can observe external state-file edits.
- Legacy `/init-re` creates a mixed RE workspace at `.rehamr/`, does not overwrite existing files, and uses date-bearing templates.
- The Legacy initializer also creates Stage F/G/H artifacts (`recomp`, `decomp`, `skills`, and `mcp.json`) and requests `0755` directories plus `0644` files. A wholesale port would violate current stage boundaries and the accepted current-user-only protection of `.rehamr`.
- Legacy status output reads a fixed file list, truncates each entry at 1,800 bytes without splitting UTF-8, and marks missing entries. That presentation command remains deferred, but bounded Unicode-safe reading is retained as useful contract evidence.
- Legacy doctor combines general system/config diagnostics with future MCP, skill, cached-repository, and PCRECOMP concerns. It is not a coherent Stage D port and remains decomposed pending the later-stage owners.
- Current RecompHamr already creates `.rehamr` for protected configuration/history/logging, so directory existence cannot serve as “RE workspace initialized” identity.

## Implementation approach

Use an adapted, security-improved compatibility contract rather than a direct port. Legacy structure may be redesigned or optimized whenever the current implementation is cleaner, safer, simpler, or faster, but every verified required observable behavior remains equivalent unless the user explicitly approves a tested and documented intentional change. Add a small backend workspace package that owns canonical project identity and optional state loading. Application composition uses it to build model-facing system prompts; the controller refreshes through an injected prompt source at the lifecycle points proven necessary by tests. Presentation receives no state contents, paths, file handles, or errors.

Do not add an initializer until its user entry point and the exact subset of templates are authorized. Do not create dormant APIs for later tools. Any future initializer must distinguish RecompHamr private state from an initialized RE evidence workspace and must inherit the platform-native private-path protections already enforced for `.rehamr`.

## Behavioral surface inventory

The active rows live in [`../verification/stage-d-behavioral-surface.md`](../verification/stage-d-behavioral-surface.md). The first slice cannot close until every applicable category for `WORKSPACE-01`, `MEMORY-01`, and `APP-03` is verified. Remaining Legacy families stay explicitly deferred rather than being represented as implemented.

## Verification

- Focused workspace/app/controller/agent tests at exactly 100% statements.
- Security tests for missing, empty, oversized, non-regular, symlink/reparse, inaccessible, changing, and Unicode state inputs on applicable platforms.
- Architecture checks proving workspace filesystem ownership does not enter TUI/frontend/agent transport packages and state contents do not cross the frontend contract.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` with exactly 100.0% repository statement coverage.
- Exact-build Windows Terminal runs with LM Studio `google/gemma-4-12b-qat` proving absent-state equivalence and present/changed/cleared state behavior without exposing state contents in sanitized reports.
- Independent `windows-latest` and `ubuntu-latest` GitHub Actions results before slice acceptance.

## Documentation impact

Add this packet and the Stage D inventory to the documentation map/contract. Update current architecture, decisions, user workspace/configuration documentation, package/exported-symbol Go documentation, integration status, holding-pen/parity dispositions, and runtime evidence as the implementation becomes factual. Existing command help remains unchanged in the first slice.

## Security impact

Persistent state is model-facing untrusted project content. It must never be treated as an instruction authority above the embedded system contract, printed in logs/reports, or exposed through frontend snapshots/events. Reads must stay beneath the canonical private `.rehamr` root, refuse links/reparse points and non-regular files, be bounded before allocation, preserve valid Unicode bytes, and fail closed without disclosing file contents. Existing POSIX owner-only modes and Windows current-user-only DACL guarantees remain unchanged.

## Stop condition

The first slice closes only when application-owned workspace identity and optional persistent-state refresh are implemented without TUI/backend-boundary regression; all three active rows are verified; statement, behavioral, and meaningful documentation coverage are 100%; security/platform cases pass; exact-build Windows evidence and both CI platforms pass; and no Stage E/F/G/H surface has entered production.

Stage D as a whole remains open until later work packets disposition the remaining workspace/configuration/diagnostic foundations with no blocked or unverified Stage D row.

## Completion evidence

- Changed: work packet and evidence freeze only; production behavior is unchanged.
- Documented: initial Stage D scope, Legacy decomposition, security boundary, verification, and behavioral rows.
- Verified: source/reference audit complete; implementation and runtime evidence remain open.
- Coverage: accepted Stage C coverage remains the baseline; Stage D rows are initially unverified.
- Security: unsafe Legacy permissions and mixed-stage artifacts are explicitly rejected from the direct-port scope.
- Evidence: current and Legacy source/tests/docs named above.
- Known limits: all first-slice implementation, canonical verification, runtime acceptance, CI, and later Stage D scheduling remain open.

### Checkpoint D1B — secure workspace identity and state reader

- Changed: added immutable `internal/workspace` ownership for canonical absolute project identity and optional `REPHAMR_STATE.md` reads. The reader is bounded at 64 KiB before and during allocation, preserves valid UTF-8, is concurrent-safe, tightens private protection, and rejects links/reparse points, non-regular files, replacement races, invalid UTF-8, and filesystem/security failures.
- Documented: package/exported-symbol contracts, D-016, active architecture, workspace user contract, configuration security, holding-pen status, and behavioral rows agree with the implementation.
- Verified: focused workspace tests exercise every statement and applicable success/failure/malformed/boundary/platform-simulated/Unicode/persistence/concurrency/security branch at 100.0% statements.
- Coverage: Stage D behavioral rows remain unverified until application integration, canonical verification, runtime evidence, and CI complete.
- Security: file contents never appear in errors; the path is fixed beneath the private `.rehamr` root and verified before reading.
- Evidence: current source/tests and strengthened architecture checks.
- Known limits: exact-build runtime and cross-platform CI remain open.

### Checkpoint D1C — application prompt refresh

- Changed: core application composition is the only production workspace consumer and supplies a refreshable system-prompt function to the controller. Startup accounting and every model round refresh optional state. Missing, empty, or securely rejected state falls back silently to the unchanged embedded prompt plus working-directory anchor.
- Documented: D-016 and current architecture record lower-trust prompt placement and the app-only capability edge. The frontend/TUI contract and command help are unchanged.
- Verified: focused app/controller/TUI tests pass at 100.0% statements and prove changing state refresh, silent unsafe-state fallback, bootstrap/round refresh, and unchanged presentation contracts.
- Coverage: runtime-dependent Stage D rows remain unverified pending canonical and exact-build acceptance.
- Security: state content, path capabilities, errors, and file handles never enter frontend snapshots/events or orchestration logs.
- Evidence: app composition, controller prompt-source, workspace prompt, and agent request contract tests.
- Known limits: canonical gate, runtime acceptance, CI, and slice closure remain open.

### Checkpoint D1D — architecture and documentation boundary

- Changed: architecture enforcement makes workspace an app-only backend owner, excludes it from TUI/frontend/entrypoint imports, and includes it in the positive deletion graph.
- Documented: current architecture, D-016, user workspace/configuration security, documentation map/contract, parity holding pen, work packet, and behavioral inventory are synchronized.
- Verified: focused architecture and affected package checks pass; canonical verification remains required before checkpoint acceptance.
- Coverage: meaningful package/exported-symbol and user/developer contract documentation is present; final docscheck/link/canonical results remain open.
- Security: later-stage initializer, doctor, RE templates, skills, MCP, and tool artifacts remain absent.
- Evidence: architecture script and documentation contract.
- Known limits: runtime, CI, and final first-slice acceptance remain open.
