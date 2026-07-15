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

## Local configuration security

RecompHamr treats `.rehamr` as private because `config.yaml`, prompt history,
and debug logs can contain credentials or sensitive prompts.

- On POSIX systems, `.rehamr` is enforced as `0700`; `config.yaml`, `history`,
  and `log.txt` are enforced as `0600`.
- On Windows, Unix mode bits are not claimed as a security boundary. RecompHamr
  installs a protected DACL granting full control only to the current process
  user. Directory inheritance applies that protection to child objects.
- Existing loose paths are tightened when opened. Configuration saves use a
  sibling temporary file, sync, close, atomic rename, and then reapply the
  platform-native owner-only protection.
- A symlinked `.rehamr` or `config.yaml` is refused rather than followed.

Do not commit a credential-bearing `.rehamr/config.yaml`. Prefer an environment
reference such as `key: ${LM_STUDIO_API_KEY}` when a server requires a key.
