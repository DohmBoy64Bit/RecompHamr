# Stage A Runtime Acceptance

## Accepted target

- Date: 2026-07-15 (America/New_York).
- OS: Microsoft Windows 11 Pro 10.0.26200.
- Terminal: Windows Terminal 1.24.11321.0.
- Go: 1.26.4 windows/amd64.
- LM Studio CLI: 1.3.3, llama.cpp ROCm runtime 1.101.0.
- Model: `google/gemma-4-12b-qat`, 16,177-token context.
- Endpoint: loopback-only `http://localhost:1234` with no committed credential.
- Accepted implementation: `71f70e922ff4cf5c5e05fd2b8325610425365b02`.

The planned Devstral model could not initialize successfully on the available Vulkan runtime. The user explicitly selected Gemma 4 12B QAT as the replacement acceptance model. LM Studio loaded it through the ROCm runtime, `/v1/models` exposed it, and a direct completion returned the expected response before the TUI session.

## Observed session

The dedicated acceptance directory was outside the repository and contained no user data. The session demonstrated:

- startup and the frozen screen composition;
- composer/history behavior, multiline and large-paste behavior, and transcript scrolling, with automated interaction tests providing deterministic boundary coverage;
- a real streamed Gemma response with status/token progression;
- successful `powershell`, `write_file`, `edit_file`, and `read_file` calls against a disposable fixture, followed by exact readback of `beta`;
- `/models` switching from `gemma` to `gemma-copy`, persisted after restart;
- cancellation of a 30-second PowerShell command, followed by a clean next turn with no stale tool result;
- wide, 80x24, and constrained layouts;
- Ctrl+D exit with the Windows Terminal shell still usable, proving terminal restoration.

Sanitized event categories were inspected for request, assistant, tool call, tool result, cancellation, follow-up turn, and clean turn end. Raw prompts and debug logs were not copied into the repository.

## Security evidence

The acceptance `.rehamr` directory, `config.yaml`, history, and debug log were inspected with `icacls`; each was restricted to the current user. No key, prompt, tool argument, or unrelated user data appears in committed evidence. The fixture was isolated under the disposable acceptance directory.

## Automated evidence

- Local canonical gate: 100.0% statement coverage (`1887/1887` on Windows), build, and CLI smoke pass.
- GitHub Actions: run `29423037990`, with independent Windows and Ubuntu jobs.
- Runtime executable SHA-256: `F02986379487DC78B76448CE6107E456E2897760CEFD7E7E7421222039960700`.

## Screenshots

- [`evidence/stage-a-wide-120x36.png`](evidence/stage-a-wide-120x36.png)
- [`evidence/stage-a-80x24.png`](evidence/stage-a-80x24.png)
- [`evidence/stage-a-constrained-50x16.png`](evidence/stage-a-constrained-50x16.png)

The user visually confirmed the captured layouts on 2026-07-15.
