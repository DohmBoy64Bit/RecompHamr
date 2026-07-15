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

`internal/tui` still owns configuration reload/model persistence, context packing, chat streaming, tool dispatch, cancellation, and agent-loop policy. This is the remaining inherited coupling for later Stage C slices, not the target design.

Backend packages must not import `internal/tui`. `internal/app` is the sole current exception because a composition root must select the concrete frontend. Later slices replace that concrete runtime coupling with typed frontend contracts while preserving the accepted layout and behavior.
