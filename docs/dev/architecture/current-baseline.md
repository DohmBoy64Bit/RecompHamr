# Current Transitional Architecture

Stage A is accepted. Stage C slice 1 moved process startup and composition into `internal/app` without changing the concrete TUI contract.

```text
cmd/recomphamr
    |
    v
internal/app
    +--> internal/config
    +--> internal/llm
    +--> internal/agent
    +--> internal/logging
    +--> internal/tui
              |
              +--> internal/agent
              +--> internal/config
              +--> internal/ctx
              +--> internal/llm
              +--> internal/provider

internal/agent
    +--> internal/ctx
    +--> internal/llm
    +--> internal/provider
    +--> internal/tools

internal/logging
    +--> internal/config
    +--> internal/ctx
```

## Completed ownership

- `cmd/recomphamr` parses process-level help/version arguments, delegates application startup, and converts startup errors into the process exit contract.
- `internal/app` owns working-directory discovery, configuration bootstrap, environment overrides, LLM-client construction, debug-log lifecycle, absolute project identity, concrete frontend construction, accepted inline-screen clearing, Bubble Tea creation, and program execution.
- The architecture check prevents the process entrypoint from bypassing `internal/app` to import runtime or presentation packages directly.

## Remaining temporary ownership

Stage C slice 2 is accepted. `internal/agent` owns request packing/tool definitions, the mutable turn root, stable turn/round identity, model-round startup and opaque reading, stable-identity delivery validation, raw event reduction/accounting, sequential local-tool execution, result pairing, cancellation identity, all loop-policy latch state, stream-close decisions, provider-specific turn/probe diagnostic classification, immutable presentation snapshots, and causal runtime observation. `internal/app` constructs the typed agent runtime with its model/tool dependencies and private observer, and owns the protected debug-log lifecycle through `internal/logging`; the runtime allocates the single turn, stream, and loop state roots shared by Bubble Tea model copies. Turn contexts and cancel functions are private agent capabilities; presentation submits lifecycle intents and applies typed effects/snapshots. Production TUI code no longer imports `internal/tools` or `internal/logging`, reads mutable agent component fields, packs model history, opens model requests, reads raw transport channels, inspects raw delivered events, executes tool calls, stores cancellation capabilities, decides loop policy, classifies provider errors, emits orchestration records, or opens the private debug log. Exact-build Windows Terminal evidence with LM Studio Gemma proves ordered tools, cancellation cleanup, no stale result or cancelled-goal reissue, recovery, unchanged representative rendering, and restored shell control. Broader configuration/client/persistence ownership remains later Stage C work.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
