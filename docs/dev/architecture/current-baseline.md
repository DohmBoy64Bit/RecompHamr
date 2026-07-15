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
```

## Completed ownership

- `cmd/recomphamr` parses process-level help/version arguments, delegates application startup, and converts startup errors into the process exit contract.
- `internal/app` owns working-directory discovery, configuration bootstrap, environment overrides, LLM-client construction, debug-log lifecycle, absolute project identity, concrete frontend construction, accepted inline-screen clearing, Bubble Tea creation, and program execution.
- The architecture check prevents the process entrypoint from bypassing `internal/app` to import runtime or presentation packages directly.

## Remaining temporary ownership

Stage C slice 2 is in progress. `internal/agent` now owns request packing/tool definitions, the mutable turn root, stable turn/round identity, model-round startup and opaque reading, stable-identity delivery validation, raw event reduction/accounting, sequential local-tool execution, result pairing, cancellation identity, all loop-policy latch state, stream-close decisions, and provider-specific turn/probe diagnostic classification. `internal/app` constructs the typed agent runtime with its model and tool dependencies and injects it into the concrete frontend; the runtime allocates the single turn, stream, and loop state roots shared by Bubble Tea model copies. Turn contexts and cancel functions are private agent capabilities; presentation can only begin, query, or end a turn through lifecycle methods. The TUI schedules opaque agent stream/tool work and applies typed display/log/finish effects; production TUI code no longer imports `internal/tools`, packs model history, opens model requests, reads raw transport channels, inspects raw delivered events, executes tool calls, stores cancellation capabilities, decides loop policy, or classifies provider errors. Further snapshot encapsulation, logging ownership, and the final architecture/deletion boundary remain later Stage C work.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
