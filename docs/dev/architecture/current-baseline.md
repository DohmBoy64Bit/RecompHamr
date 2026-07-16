# Current Transitional Architecture

Stage A, all four Stage C slices, and Stage D are accepted. The accepted Stage D implementation commit is `449a83cb379e79fd84b817b0e95f63de7472578a`.

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
