# Baseline Configuration

On first run, RecompHamr creates `.rehamr/config.yaml` in the current project directory.

The baseline seeds one LM Studio OpenAI-compatible profile for the accepted
Gemma runtime. Start LM Studio's local server and load the named model before
the first real turn:

```yaml
active: local
models:
  local:
    llm: google/gemma-4-12b-qat
    url: http://localhost:1234
    key: ""
    context_size: 16177
```

The values are editable. Any endpoint used by the baseline must expose the OpenAI-compatible chat-completions interface expected by `internal/llm`.
Changing these source defaults affects newly created configuration files only;
RecompHamr never overwrites an existing `.rehamr/config.yaml`.

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
debug logs, and optional workspace state can contain credentials or sensitive prompts.

- On POSIX systems, `.rehamr` is enforced as `0700`; `config.yaml`, `history`,
  `log.txt`, and workspace state opened by RecompHamr are enforced as `0600`.
- On Windows, Unix mode bits are not claimed as a security boundary. RecompHamr
  installs a protected DACL granting full control only to the current process
  user. Directory inheritance applies that protection to child objects.
- Existing loose paths are tightened when opened. Configuration saves use a
  sibling temporary file, sync, close, atomic rename, and then reapply the
  platform-native owner-only protection.
- A symlinked `.rehamr`, `config.yaml`, or workspace-state file is refused rather than followed.

Do not commit a credential-bearing `.rehamr/config.yaml`. Prefer an environment
reference such as `key: ${LM_STUDIO_API_KEY}` when a server requires a key.
