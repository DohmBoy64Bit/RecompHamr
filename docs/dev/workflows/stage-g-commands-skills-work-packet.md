# Stage G Commands and Agent Skills Work Packet

## Outcome

Add one application-owned slash-command registry and a standards-based Agent Skills client with progressive disclosure. Preserve the accepted terminal layout and all Stage A/C/D/E/F behavior while moving command semantics, skill discovery, validation, activation, diagnostics, and durable configuration below presentation.

## In scope

- A typed command registry shared by application dispatch, terminal help, and TUI completion without duplicating names, descriptions, arguments, or availability.
- Standards-based skill discovery, parsing, validation, deterministic precedence, trust gating, catalog generation, activation, relative-resource resolution, diagnostics, enable/disable configuration, cancellation, and testability without the TUI.
- Project and user `.agents/skills/` interoperability plus a documented RecompHamr-native location only if it adds a distinct supported contract.
- Explicit `/skills` catalog and `/skill <name>` activation commands, and evidence-backed Stage G dispositions for the Legacy `/skill-audit`, `/skill-new`, `/init-re`, `/status-re`, `/doctor`, and `/help` commands.
- Individual review, migration, validation, trigger evaluation, output-quality evaluation, and manual review of each approved Legacy skill. Each skill gets its own migration record and may be improved, split, merged, or rejected when evidence supports that disposition.

## Out of scope

- `/mcp`, MCP discovery, transports, servers, tools, or configuration; these remain Stage H.
- TUI layout redesign, Charm dependency upgrades, hosted services, plugin distribution, organization registries, or automatic remote skill installation.
- Mechanical restoration of Legacy flat `.md` skills, eager injection of every skill body, the Legacy global loader, or TUI-owned filesystem/network/process work.
- Treating a parsed skill, passing statement coverage, or a picker entry as proof that a skill adds value.

## Authorities

- Root and documentation `AGENTS.md`; engineering workflow, evidence, definition-of-done, scope-control, current architecture, Legacy parity policy, and Agent Skills migration policy.
- Complete current Agent Skills client guide and specification; current skill quickstart, best practices, description optimization, evaluation, and script guidance; current official Codex skill documentation.
- Accepted Stage A/C/D/E/F source, tests, inventories, runtime evidence, and the current frontend/controller/TUI boundary.
- Legacy `internal/skills`, `internal/tui/slash.go`, classifier, project, doctor, configuration, tests, documentation, and every Legacy skill file as behavioral evidence only.

The mandatory current external authority set was read in full in this Stage G task before this packet was created. It must be refreshed if it changes, and the complete creator authority set must be recorded again before editing each individual skill as required by [`../roadmap/agent-skills-standard.md`](../roadmap/agent-skills-standard.md).

## Legacy command dispositions before production editing

| Legacy command | Stage G starting disposition |
|---|---|
| `/clear`, `/models` | retained accepted commands; migrate metadata/dispatch to the typed registry without behavior change |
| `/skills`, `/skill` | replace the Legacy flat-file behavior with standards-based catalog and explicit activation contracts |
| `/skill-audit`, `/skill-new` | do not restore the obsolete classifier/fetch-and-wrap workflow; evaluate a standards-valid authoring/evaluation replacement before any command is approved |
| `/init-re`, `/status-re` | evaluate against the accepted Stage D workspace owner and current user contract; no TUI filesystem ownership |
| `/doctor` | evaluate as application-owned diagnostics with no secret disclosure or Stage H probing |
| `/help` | derive from the single typed registry if retained; do not duplicate a static list |
| `/mcp` | deferred to Stage H |

All rows remain `unverified` until direct Legacy evidence, current implementation, tests, documentation, and required runtime evidence agree.

## Implementation approach

Create a backend-neutral command description/intent contract and an application-owned dispatcher. Presentation may filter and render immutable command facts and emit typed command intents; it may not execute command side effects. Introduce an `internal/skills` owner whose discovery catalog stores metadata and locations only. Load and wrap instructions only on activation, enumerate but do not eagerly read resources, preserve activated instructions across context packing, and deduplicate activation for the conversation.

Discovery will be bounded, deterministic, link/reparse-aware, and trust-aware. RecompHamr-authored skills are validated strictly against the current specification. If third-party malformed-but-recoverable compatibility is supported, it will be explicit, diagnostic, tested, and never weaken built-in validation. Collision, skip, invalid, shadow, truncation, trust, and disabled-state facts will be display-safe and secret-free.

## Behavioral surface inventory

The active inventory is [`../verification/stage-g-behavioral-surface.md`](../verification/stage-g-behavioral-surface.md). Every row starts unverified and must map implementation, all applicable behavioral categories, exact tests, durable docs, runtime evidence, and a final parity disposition.

## Verification

- Focused command, skills, app/controller, agent/context, configuration, workspace, and TUI tests at exactly 100.0% statements.
- Parser/discovery/precedence/trust/link/reparse/race/Unicode/size/resource/configuration/activation/deduplication/context/diagnostic/cancellation/security tests on Windows and POSIX branches.
- Registry tests proving one canonical ordered source drives help, completion, argument choices, dispatch, availability, and unknown-command behavior.
- Per-skill specification validation plus direct agent review of every realistic positive/negative trigger and output-contract case from each co-located `evals/evals.json`, with a sanitized human-review artifact. Per the user's 2026-07-17 direction, the automated model-evaluation runner is not the acceptance authority and is not modified for this review.
- Architecture checks proving `internal/tui` owns no command, skill, filesystem, process, network, persistence, trust, or activation lifecycle.
- Canonical `pwsh -NoProfile -File ./scripts/verify.ps1` with exactly 100.0% repository statement coverage, complete docs/links/architecture/build/smoke gates, and independent Windows/Ubuntu CI.
- Exact-build Windows Terminal runtime with LM Studio `mistralai/devstral-small-2-2512`, while retaining the Gemma profile, including retained commands, catalog/explicit activation, implicit model-driven activation, resource loading, invalid/shadowed/disabled diagnostics, cancellation/stale-result recovery, representative sizes, clean exit, and terminal restoration.

## Documentation impact

Maintain this packet, Stage G inventory, documentation map/contract, architecture, decisions, integration order, holding pen, baseline status, README/help, user command/skill/configuration/workspace contracts, and all affected Go package/exported-symbol documentation. Each migrated skill must retain its own source/authority/disposition/evaluation record.

## Security impact

Project skills are untrusted repository-provided instructions and require an explicit trust policy before catalog disclosure or activation. Discovery and resources must remain within approved roots, refuse unsafe links/reparse points and replacement races, bound file count/depth/bytes, avoid secrets in catalogs/diagnostics/logs, and never execute scripts merely by discovering or activating a skill. Script execution remains subject to existing model tool permissions, cancellation, time/output bounds, and user authority. Remote fetching or installation is unsupported in Stage G unless separately approved.

## Stop condition

Stage G closes only when the command registry and skills client are application-owned and real; all approved Legacy commands and skills have explicit final dispositions; every retained, replaced, new, and intentionally omitted surface has complete behavioral and meaningful documentation coverage; every migrated skill passes current specification validation and its own trigger/output/manual evaluation gates; the canonical gate is exactly 100%; exact-build runtime evidence passes; both CI platforms pass; no Stage G row is blocked or unverified; and no Stage H behavior entered production.

## Initial evidence record

- verified: complete current Agent Skills client, specification, creator, evaluation, scripts, and official Codex skill authorities were read in this task.
- verified: current source exposes only `/clear` and `/models`; presentation still owns their concrete dispatch table while neutral help metadata is duplicated in `internal/frontend`.
- verified: Legacy exposes eleven commands, of which `/mcp` is Stage H, and uses a global flat-Markdown loader with custom files overriding embedded files.
- unverified: all Stage G implementation, focused/canonical verification, skill evaluations, runtime acceptance, CI, and final dispositions.

## Accepted implementation evidence

- verified by source and focused tests: `internal/frontend` now provides one typed ordered registry for retained commands plus `/skills`, `/skill`, `/init-re`, `/status-re`, and `/help`; the TUI maps typed command kinds to presentation handlers and does not own their backend side effects.
- verified by source and focused tests: `internal/skills` performs bounded immediate-child `.agents/skills` discovery, strict current-spec metadata validation, project-over-user precedence, explicit project trust, configured disable filtering, display-safe diagnostics, activation-time revalidation, concurrency-safe deduplication, and reload-time catalog replacement.
- verified by source and focused tests: model disclosure is progressive and bounded: an 8 KiB metadata catalog, on-demand instruction activation, enumerated resource names, and exact active-resource reads capped at 2 MiB with link, regular-file, and replacement-race checks. Activated system content above the fixed reservation reduces history budget.
- verified by source and focused tests: `/init-re` and `/status-re` preserve the useful Legacy evidence-workspace capability with current-user-only protection, link refusal, no overwrite, bounded UTF-8 status, and deliberate exclusion of Stage H `mcp.json` and obsolete flat skills.
- intentionally changed: `/skill-audit` and `/skill-new` are not production commands. Current specification validation and mandatory per-skill trigger/output evaluation replace the heuristic classifier and remote fetch-and-wrap workflow.
- not applicable in Stage G: Legacy `/doctor` synchronously combined local processes, endpoint networking, secret-adjacent environment output, and MCP status. A future diagnostics capability requires its own non-secret asynchronous contract; Stage H owns MCP observations.
- verified: the canonical gate reached 100.0% at 2949/2949 before the evidence-workspace command slice. The first post-slice run correctly failed at 99.5%; focused controller and workspace profiles were then restored to 100.0%. A new full canonical run is still required.
- verified by complete source-by-source review and structural tests: all 28 Legacy skill Markdown files now have explicit individual dispositions. Twenty-one are standards-valid bundled migrations with eight positive triggers, eight negative triggers, and three output cases each; `sega2asm` is merged into `gen-decomp`; `recomp-foundations` is rejected as a stale static link router; and the five MCP-only guides (`bizhawk`, `ghidra-mcp`, `mcp-pine`, `n64-debug-mcp`, and `pcsx2`) are deferred to Stage H.
- verified by direct review: all 21 bundled migrations passed 8/8 positive triggers, 8/8 adjacent negative triggers, and 3/3 output-contract cases. The sanitized review is retained at `E:\ReProject\StageG-Manual-Evaluation-Devstral\human-review.md`; it explicitly does not claim 399 generated model responses.
- verified by exact-build Windows Terminal runtime: Devstral was active; Gemma remained selectable; `/help`, `/models`, `/skills`, `/skill`, `/clear`, `/init-re`, and `/status-re` worked; explicit and implicit activation both loaded an exact resource; 120x36, 80x24, and 50x16 were reviewed; and Ctrl+D restored the shell. Evidence is retained at `E:\ReProject\StageG-Runtime-Acceptance`.
- verified regression: visual review found a shadow diagnostic containing an absolute user path. The message is now path-free, covered by focused tests, rebuilt, rerun, and visually confirmed.
- verified: the documentation-synchronized canonical gate passed on Go 1.26.4 at exactly 3184/3184 statements (100.0%), including docs, links, architecture, formatting, all packages, build, CLI smoke, acceptance-scenario validation, and all 21 skill fixture sets.
- verified: exact implementation commit `6aaeb81ae01f6b0290c63af94791001fff021960` passed the Devstral Windows Terminal scenario; its executable SHA-256 is `3067cdc85457da674eec098507f5614170274c1956ab7d165da083132e0c851b`, with reviewed evidence at `E:\ReProject\StageG-Runtime-Acceptance`.
- verified: GitHub Actions run `29603832999` passed independently on `windows-latest` and `ubuntu-latest` against the implementation commit.
- closure: the stop condition is met. Every Stage G row is verified, no Stage H behavior entered production, and Stage H remains inactive pending a separate work packet.
