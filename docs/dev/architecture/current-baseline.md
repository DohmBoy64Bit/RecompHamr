# Current Transitional Architecture

Stage A is accepted. Stage C slices 1 through 3 are accepted.

```text
cmd/recomphamr
    |
    v
internal/app
    +--> internal/config
    +--> internal/agent
    +--> internal/logging
    +--> internal/session
    +--> internal/frontend
    +--> internal/app/controller
    +--> internal/tui
              |
              +--> internal/agent
              +--> internal/frontend

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

- `cmd/recomphamr` parses process-level help/version arguments, delegates application startup, and converts startup errors into the process exit contract.
- `internal/app` owns working-directory discovery, configuration bootstrap, environment overrides, session/agent runtime composition, debug-log lifecycle, absolute project identity and system-prompt construction, concrete frontend construction, accepted inline-screen clearing, Bubble Tea creation, and program execution.
- The architecture check prevents the process entrypoint from bypassing `internal/app` to import runtime or presentation packages directly.

## Remaining temporary ownership

Stage C slice 2 is accepted. `internal/agent` owns request packing/tool definitions, the mutable turn root, stable turn/round identity, model-round startup and opaque reading, stable-identity delivery validation, raw event reduction/accounting, sequential local-tool execution, result pairing, cancellation identity, all loop-policy latch state, stream-close decisions, provider-specific turn/probe diagnostic classification, immutable presentation snapshots, and causal runtime observation. `internal/app` constructs the typed agent runtime with its model/tool dependencies and private observer, and owns the protected debug-log lifecycle through `internal/logging`; the runtime allocates the single turn, stream, and loop state roots shared by Bubble Tea model copies. Turn contexts and cancel functions are private agent capabilities; presentation submits lifecycle intents and applies typed effects/snapshots. Production TUI code no longer imports `internal/tools` or `internal/logging`, reads mutable agent component fields, packs model history, opens model requests, reads raw transport channels, inspects raw delivered events, executes tool calls, stores cancellation capabilities, decides loop policy, classifies provider errors, emits orchestration records, or opens the private debug log. Exact-build Windows Terminal evidence with LM Studio Gemma proves ordered tools, cancellation cleanup, no stale result or cancelled-goal reissue, recovery, unchanged representative rendering, and restored shell control.

Stage C slice 3 is accepted. Checkpoint 3A moved the complete prompt-history filesystem implementation into `internal/session`. Checkpoint 3B added one app-composed `session.Runtime` that owns configuration reload, active-profile persistence, resolved credential use, concrete client replacement, and captured reachability/probe work. The TUI receives immutable non-secret snapshots and opaque bounded work, and production TUI code no longer imports config, LLM, or provider packages or performs history/config filesystem I/O. Focused/canonical verification and the complete user-confirmed Windows Terminal checklist preserve model switching, real response, persistence, history, representative rendering, clean exit, and shell restoration.

Stage C slice 4 is in progress. Checkpoints 4B and 4C add backend-neutral `internal/frontend` contracts and an independently tested controller under `internal/app/controller`, then route startup/history, reload, activation, reachability/probe work, and conversation-history clearing through that boundary. Production TUI code no longer imports or stores `internal/session` or `internal/ctx`; it runs opaque controller work and returns completions without inspecting backend payloads. The TUI still retains the temporary `internal/agent` façade and agent lifecycle aliases until checkpoint 4D.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
