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

Stage C slice 2 is in progress. `internal/agent` now owns request packing/tool definitions, pure loop-policy contracts, the mutable turn root, stable turn/round identity, model-round startup, opaque stream reading, event reduction, pending-call collection, phase/retry/connectivity state, live context hints, and turn/session token accounting. The TUI schedules opaque agent stream reads and applies typed display/log effects; it no longer opens model requests, reads raw transport channels, or defines stream accounting transitions. Tool execution, stream-close loop decisions, policy latches, and final diagnostic styling remain transitional TUI ownership.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
