# Current Transitional Architecture

Stage A is accepted. Stage C slice 1 moved process startup and composition into `internal/app` without changing the concrete TUI contract.

```text
cmd/recomphamr
    |
    v
internal/app
    +--> internal/config
    +--> internal/llm
    +--> internal/tui
              |
              +--> internal/agent
              +--> internal/config
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

Stage C slice 2 is in progress. `internal/agent` now owns request packing/tool definitions, the mutable turn root, stable turn/round identity, model-round startup and opaque reading, event reduction/accounting, sequential local-tool execution, result pairing, cancellation identity, all loop-policy latch state, stream-close decisions, and provider-specific turn/probe diagnostic classification. `internal/app` constructs the typed agent runtime with its model and tool dependencies and injects it into the concrete frontend. The TUI currently unpacks that runtime into its value-model adapter, schedules opaque agent stream/tool work, and applies typed display/log/finish effects; production TUI code no longer imports `internal/tools`, opens model requests, reads raw transport channels, executes tool calls, decides loop policy, or classifies provider errors. Runtime-state encapsulation, logging ownership, and the final architecture/deletion boundary remain open.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
