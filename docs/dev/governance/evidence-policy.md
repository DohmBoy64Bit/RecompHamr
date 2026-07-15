# Evidence Policy

Repository claims must be supported by local source, tests, reproducible command output, or runtime observation.

Use these labels when evidence is incomplete:

- `verified:` reproduced by the stated check;
- `unverified:` not yet executed in the required environment;
- `blocked:` cannot currently be executed because a named dependency or environment is unavailable;
- `unsupported:` intentionally outside the current baseline.

Do not infer runtime success from static inspection. Do not infer visual parity from unit tests. Do not treat planning documents as stronger evidence than current source and tests.
