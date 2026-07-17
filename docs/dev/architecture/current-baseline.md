# Current Transitional Architecture

Stage A, all four Stage C slices, Stage D, Stage E, and Stage F are accepted. Stage G command registries and Agent Skills are implemented, runtime-verified, and canonical-gate clean but remain open pending dual-platform CI and accepted-commit evidence.

```text
cmd/recomphamr
    |
    v
internal/app/terminal --> internal/tui --> internal/frontend
    |
    v
internal/app
    +--> internal/config
    +--> internal/agent
    +--> internal/logging
    +--> internal/session
    +--> internal/frontend
    +--> internal/app/controller
    +--> internal/workspace
    +--> internal/skills

internal/agent
    +--> internal/ctx
    +--> internal/llm
    +--> internal/provider
    +--> internal/tools

internal/logging
    +--> internal/config
    +--> internal/ctx

internal/session
    +--> internal/config
    +--> internal/llm
    +--> internal/provider

internal/skills
    +--> strict Agent Skills parsing and validation
    +--> bounded user/project/bundled discovery and precedence
    +--> activation-time revalidation and exact resource reads
```

## Completed ownership

- `cmd/recomphamr` parses process-level help/version arguments, delegates terminal startup/help to `internal/app/terminal`, and converts startup errors into the process exit contract.
- Core `internal/app` owns working-directory discovery, configuration bootstrap, environment overrides, session/agent/controller composition, debug-log lifecycle, absolute project identity, system-prompt construction, and idempotent backend cleanup.
- `internal/app/terminal` owns concrete TUI construction, accepted inline-screen clearing, Bubble Tea focus/program creation and execution, terminal help formatting, and application-lifetime cleanup.
- The architecture check prevents the process entrypoint from bypassing `internal/app` to import runtime or presentation packages directly.

## Accepted backend ownership

Stage C slice 2 is accepted. `internal/agent` owns request packing/tool definitions, the mutable turn root, stable turn/round identity, model-round startup and opaque reading, stable-identity delivery validation, raw event reduction/accounting, sequential local-tool execution, result pairing, cancellation identity, all loop-policy latch state, stream-close decisions, provider-specific turn/probe diagnostic classification, immutable presentation snapshots, and causal runtime observation. `internal/app` constructs the typed agent runtime with its model/tool dependencies and private observer, and owns the protected debug-log lifecycle through `internal/logging`; the runtime allocates the single turn, stream, and loop state roots shared by Bubble Tea model copies. Turn contexts and cancel functions are private agent capabilities; presentation submits lifecycle intents and applies typed effects/snapshots. Production TUI code no longer imports `internal/tools` or `internal/logging`, reads mutable agent component fields, packs model history, opens model requests, reads raw transport channels, inspects raw delivered events, executes tool calls, stores cancellation capabilities, decides loop policy, classifies provider errors, emits orchestration records, or opens the private debug log. Exact-build Windows Terminal evidence with LM Studio Gemma proves ordered tools, cancellation cleanup, no stale result or cancelled-goal reissue, recovery, unchanged representative rendering, and restored shell control.

Stage C slice 3 is accepted. Checkpoint 3A moved the complete prompt-history filesystem implementation into `internal/session`. Checkpoint 3B added one app-composed `session.Runtime` that owns configuration reload, active-profile persistence, resolved credential use, concrete client replacement, and captured reachability/probe work. The TUI receives immutable non-secret snapshots and opaque bounded work, and production TUI code no longer imports config, LLM, or provider packages or performs history/config filesystem I/O. Focused/canonical verification and the complete user-confirmed Windows Terminal checklist preserve model switching, real response, persistence, history, representative rendering, clean exit, and shell restoration.

Stage C slice 4 is accepted. Backend-neutral `internal/frontend` contracts route every session and agent action through `internal/app/controller`, and `internal/app/terminal` isolates concrete terminal composition. The controller owns model reads, stream reduction, sequential tools, close policy, cancellation, stale work draining/rejection, and final accounting. Production TUI code imports only `internal/frontend` among project runtime packages. Core `internal/app` imports neither TUI nor Bubble Tea and exposes only a neutral controller plus idempotent cleanup; deleting the terminal adapter removes the runnable presentation target without removing buildable application/backend behavior. Canonical verification, the positive deletion graph, dual-platform CI, and exact-build Windows Terminal evidence close the separation boundary.

Backend packages do not import `internal/tui`. `internal/app/terminal` is the sole concrete TUI/Bubble Tea wiring edge.

## Accepted Stage D ownership

`internal/workspace` owns canonical absolute project identity and bounded optional reads of `.rehamr/REPHAMR_STATE.md`. It refuses links/reparse points, non-regular files, replacement races, invalid UTF-8, oversize state, and security/I/O failures; it creates no files or directories. Core `internal/app` is its only production consumer and supplies a refreshable system-prompt function to `internal/app/controller`. The controller refreshes before startup accounting and every model round. Missing, empty, or unsafe optional state leaves the embedded prompt and working-directory anchor intact and does not cross the frontend boundary.

Focused and canonical tests, architecture enforcement, dual-platform CI, and exact-build Windows Terminal evidence with LM Studio Gemma verify absent/present/changed/cleared state, frozen rendering, model and tool cancellation, stale-result rejection, recovery, and restored shell control.

## Accepted Stage E parity

Stage E audited the accepted `internal/agent`, `internal/ctx`, and `internal/llm` owners against every in-scope Legacy agent/runtime branch. Exact source/test traceability and runtime evidence established that Stage C already preserved the required loop policies and lifecycle while improving typed ownership, cancellation cleanup, context pairing, stream error handling, and deterministic testing. No genuine production gap existed, so Stage E made no redundant runtime change. The Legacy skills-template classifier remains outside this graph until Stage G.

## Accepted Stage F tools

Core `internal/app` constructs an immutable `internal/tools.Set` with the protected `.rehamr` root and platform-native path protection, then injects its executor into `internal/agent`. The agent exposes six stable schemas and retains sequential execution, cancellation, stale-result rejection, and bounded conversation output. `internal/tools` privately owns Git/process, public-network policy, deterministic packing/reduction, and protected atomic cache persistence. Frontend/TUI contracts are unchanged and receive only the existing one-line status and semantic tool events.

Stage F scope is exactly `repomixr` and `recomp_reference`. Unknown tools still fail closed. Complete focused/canonical verification, exact-build Gemma runtime evidence, Windows cache ACL inspection, and dual-platform CI accept the boundary and both improved Legacy parity contracts.

## Active Stage G ownership

`internal/frontend` owns one immutable ordered seven-command registry used by dispatch, CLI/TUI help, completion, and argument metadata. The TUI renders registry facts and translates input to typed intents; `internal/app/controller` dispatches command semantics and composes `internal/workspace`, `internal/session`, and `internal/skills` behavior below presentation.

`internal/skills` owns bounded immediate-child discovery from bundled, user, and explicitly trusted project `.agents/skills` roots; strict specification parsing; deterministic precedence and disable filtering; path-free display diagnostics; activation-time replacement-race checks; deduplication; and exact on-demand resource reads. Catalog disclosure is capped at 8 KiB, skill bodies load only on activation, and resources remain names-only until an activated skill requests one. Discovery and activation never execute scripts.

Core `internal/app` restores bundled skills into a protected content-addressed directory and composes the skills runtime with the controller/agent. `internal/agent` exposes constrained activation/resource tools and incorporates activated instructions into the system-context budget. Production TUI code imports no skills, filesystem, process, network, persistence, or trust owner.

Exact-build Windows Terminal acceptance with LM Studio Devstral verified the full registry, retained Gemma profile, explicit and implicit activation, exact resource loading, evidence-workspace commands, three representative sizes, and terminal restoration. The documentation-synchronized canonical gate passed at 3183/3183 statements; dual-platform CI and accepted-commit evidence remain required before Stage G acceptance.
