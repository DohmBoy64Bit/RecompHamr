# Stage F RecompHamr Tools Work Packet

## Outcome

Add the two verified Legacy reverse-engineering support tools, `repomixr` and `recomp_reference`, through an application-configured `internal/tools` owner. Preserve the accepted four tools and agent/TUI behavior while replacing unsafe Legacy internals with deterministic, cancellable, bounded, current-user-protected cache operations.

## In scope

- `repomixr`: strict public GitHub repository identification, shallow branch/tag clone, deterministic bounded UTF-8 source selection, optional transforms, safe XML packing, optional `.rehamr/repomix-instruction.md`, and a durable cache beneath `.rehamr/repos`.
- `recomp_reference`: HTTP(S) reference retrieval, readable HTML-to-text reduction, bounded non-HTML capture, deterministic cache naming, 24-hour freshness, redirects, cancellation, and a durable cache beneath `.rehamr/reference`.
- An immutable application-configured tool set/executor so cache roots, Git/process execution, HTTP transport, clock, and limits are owned below presentation and are injectable in tests.
- Model schemas, request ordering, inline statuses, result/error contracts, architecture rules, user documentation, behavioral inventory, and exact-build runtime evidence.
- Evidence-backed improvements to Legacy security, correctness, cancellation, bounds, and testability while retaining the required capability.

## Out of scope

- Bash restoration; the accepted native `powershell` tool remains the cross-platform shell contract.
- Ghidra/emulator/decompiler tools, MCP configuration or execution, and unknown-tool fallbacks (Stage H).
- Skills, skill-template classification, slash commands, `/doctor`, `/init-re`, `/status-re`, or creation of `repomix-instruction.md` (Stage G).
- New TUI controls, layout changes, dependency upgrades, hosted services, repository authentication, arbitrary Git remotes, or general-purpose web browsing.
- Automatic cache eviction, background refresh, repository updates, or migration of unsafe Legacy cache contents.

## Authorities

- Root and documentation `AGENTS.md`; engineering workflow, evidence, definition-of-done, scope-control, architecture, and Legacy parity policy.
- Accepted Stage A/C/D/E source, tests, inventories, and runtime evidence.
- Current `internal/tools`, `internal/agent`, `internal/app`, `internal/config`, `internal/workspace`, and their tests.
- Legacy `internal/tools/repomixr.go`, `recomp_reference.go`, their tests, application wiring, schemas, README, and tool documentation as behavioral evidence rather than architecture authority.

## Legacy evidence and dispositions before editing

- Legacy exposes exactly two non-baseline built-ins: `repomixr` and `recomp_reference`. MCP tools belong to Stage H; skills and their commands belong to Stage G.
- `repomixr` requires `repo_url`; defaults `branch` to `main`; accepts four optional transforms; clones a shallow GitHub branch/tag; includes sorted UTF-8 files smaller than 512 KiB while excluding `.git` and common binary/archive/media extensions; writes an XML tree and file bodies; optionally appends `.rehamr/repomix-instruction.md`; and returns a path/count/size summary.
- `recomp_reference` requires `url`; fetches with a 15-second bound; reduces HTML while excluding non-content elements; captures at most 512 KiB of other bodies; caches a source/timestamp-prefixed text file for 24 hours; and returns its path.
- The Legacy global cache variables, live-network tests, permissive GitHub parsing, unescaped XML attributes and CDATA terminators, unbounded HTML parsing, non-atomic writes, link-following cache paths, and process/network boundaries are not compatibility requirements. They are defects or incidental internals and will be improved.

## Implementation approach

Use an adapted implementation behind a configured `tools.Set` rather than copying globals. Core `internal/app` supplies the protected `.rehamr` root when composing the agent executor. `internal/agent` continues to own request ordering and turn lifecycle, and the TUI continues to receive only display-safe statuses/events.

Validate all external identifiers before filesystem or network work. Refuse credentials, queries/fragments, ambiguous GitHub paths, unsafe cache components, links/reparse points, non-regular cache entries, oversized or invalid content, unsafe redirects, and destinations outside the configured cache. Write cache artifacts through same-directory atomic replacement and apply POSIX owner modes or a protected current-user Windows DACL. Bound clone output, file count, total packed bytes, response bytes, redirects, and execution time. Preserve cancellation through both process and HTTP work.

## Behavioral surface inventory

The active inventory is [`../verification/stage-f-behavioral-surface.md`](../verification/stage-f-behavioral-surface.md). `TOOLSET-01`, `REPOMIX-01`, and `REFERENCE-01` begin unverified and no production capability is accepted merely because a similarly named Legacy function exists.

## Verification

- Focused `internal/tools`, `internal/agent`, and `internal/app` tests at exactly 100.0% statements.
- Deterministic fake Git/process, filesystem-failure, HTTP/redirect/DNS-policy, clock, timeout, cancellation, Unicode, XML, atomicity, permission/ACL, link/reparse, output-bound, and stale/fresh-cache tests.
- Architecture checks proving only application composition configures concrete tool capabilities; frontend/TUI remain backend-neutral; unknown tools still fail closed; Stage G/H owners are absent.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` with exactly 100.0% repository coverage and complete documentation/link checks.
- Exact-build Windows Terminal use with LM Studio `google/gemma-4-12b-qat`: retain the four-tool/cancellation scenario and add sanitized real `repomixr` plus controlled local-reference evidence without secrets or private data.
- Independent passing `windows-latest` and `ubuntu-latest` jobs for the accepted commit.

## Documentation impact

Maintain this packet, the Stage F inventory, documentation map/contract, AGENTS stage state, scope, architecture, integration order, holding pen, decisions, baseline status, README, and `docs/user/tools.md`. Document every schema, side effect, cache path, bound, dependency, error class, cancellation behavior, trust boundary, and deliberate Legacy change. Update Go package/exported-symbol documentation with the implementation.

## Security impact

Both tools add explicit network and persistent-cache authority to model-requested actions. They run with the current user's permissions and must be described as untrusted external-content ingestion. `repomixr` must not accept credentials or arbitrary remotes. `recomp_reference` must prevent model-driven access to loopback, link-local, private, unspecified, multicast, and other non-public destinations, including redirect and DNS-rebinding paths; HTTPS is preferred and HTTP is limited to verified public destinations. Neither tool may disclose response bodies, credentials, private paths beyond its own returned cache path, or raw process diagnostics outside bounded tool results. Cache paths must remain current-user-only and link-safe.

## Stop condition

Stage F closes only when both tools are real and application-configured; every Legacy contract has an explicit parity disposition; all new and retained tool surfaces have complete applicable behavioral coverage and meaningful documentation; caches are bounded, atomic, link-safe, and current-user-protected on POSIX and Windows; cancellation and stale work remain correct; canonical coverage is exactly 100%; exact-build runtime evidence passes; both CI platforms pass; no Stage F row is blocked or unverified; and no Stage G/H behavior entered production.

## Completion evidence

- Changed: work packet and initial inventory only; no production behavior yet.
- Documented: exact Stage F scope, verified Legacy surface, intended improvements, exclusions, verification, security, and closure criteria.
- Verified: direct Legacy source/tests/docs and current ownership inspection identify exactly two in-scope tool families.
- Coverage: accepted Stage E is the starting baseline; all Stage F rows are initially unverified.
- Security: the packet refuses the Legacy global and permissive network/filesystem design before implementation.
- Known limits: implementation, focused/canonical verification, runtime acceptance, CI, and final dispositions remain open.

### Checkpoint F1B — verified contracts and configured boundary

- Changed: added immutable application-configured `tools.Set`; core app supplies the protected `.rehamr` root and injects the executor; agent requests now expose six stable schemas.
- Disposition: application ownership is improved over Legacy globals. Unknown tools still fail closed and presentation remains unchanged.
- Verified: focused app/agent tests and architecture checks pass; only core app constructs the concrete tool set.
- Security: cache/network/process capabilities remain below frontend and TUI.

### Checkpoint F1C — tool implementation

- Changed: implemented strict public-GitHub `repomixr` and public-network-only `recomp_reference` with cancellation, deterministic bounds, safe XML/CDATA, HTML reduction, query redaction, URL hashing, link refusal, atomic writes, and platform-native private-path protection.
- Disposition: both Legacy capabilities are improved. Legacy schemas and useful output/cache behavior are retained; permissive URLs, globals, unsafe XML, unbounded HTML, weak tests, and non-atomic/unprotected cache writes are intentionally replaced.
- Verified: `internal/tools` is exactly 100.0% statement-covered across success, failure, malformed, boundary, cancellation, timeout, cleanup, platform, Unicode, persistence, concurrency, and security surfaces. The first repository canonical gate passes at 100.0% (`2587/2587` statements).
- Known limits: exact-build runtime, final documentation audit, CI, and closure remain open.
