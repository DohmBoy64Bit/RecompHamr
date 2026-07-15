# Target Separation of Concerns

This architecture becomes the next implementation target **after** the barebones baseline is accepted.

```text
cmd/recomphamr
        |
        v
   internal/app
    /    |     \
   v     v      v
agent  config   frontend contracts
  |                 |
  v                 v
 llm             internal/tui
  |
  v
tool contracts / execution
```

## Ownership

### `cmd/recomphamr`

Process entrypoint only. Parse process-level flags and delegate application composition.

### `internal/app`

Compose concrete services, own application lifecycle, translate frontend intents into exactly-once application actions, and provide immutable/runtime snapshots to presentation.

### `internal/tui`

Own rendering, local presentation state, focus, input translation, and visual behavior. It must not execute tools, mutate config files, own model-client lifecycle, or manage future MCP/process lifecycles.

### `internal/agent`

Own turn lifecycle, context packing coordination, model streaming, tool-call sequencing, cancellation, and agent-loop policy through interfaces.

### `internal/llm`

Own only OpenAI-compatible wire transport and transport-level errors/events.

### `internal/tools`

Own tool schemas and execution. Presentation receives tool state/results through application or agent contracts.

### `internal/config`

Own config parsing, persistence, validation, and profile resolution.

## Migration rule

The separation stage must preserve the accepted baseline behavior and layout. Move ownership in small slices with equivalence tests. Do not add RecompHamr-Legacy capabilities during the extraction.

## Closure condition

Deleting `internal/tui` from a separated build should remove presentation only, not tool execution, model orchestration, configuration persistence, or application lifecycle behavior.
