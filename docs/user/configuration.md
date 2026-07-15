# Baseline Configuration

On first run, RecompHamr creates `.rehamr/config.yaml` in the current project directory.

The baseline seeds one local OpenAI-compatible profile:

```yaml
active: local
models:
  local:
    llm: qwen3.6:27b
    url: http://localhost:11434
    key: ""
    context_size: 32768
```

The values are editable. Any endpoint used by the baseline must expose the OpenAI-compatible chat-completions interface expected by `internal/llm`.

## Switching profiles

Add additional entries under `models:` and use:

```text
/models
/models <name>
```

The active selection is persisted to `.rehamr/config.yaml`.

## Runtime overrides

- `RECOMPHAMR_URL` overrides the active profile URL for the current process.
- `RECOMPHAMR_IDLE_TIMEOUT` overrides the inter-frame stream idle timeout using a Go duration such as `90m` or `1h`.

The baseline does not ship a hosted provider profile or bundled credentials.
