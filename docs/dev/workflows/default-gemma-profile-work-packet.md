# Default Gemma Profile Work Packet

> Historical accepted packet. On 2026-07-17 the user selected Devstral as the
> fresh active default while requiring this Gemma profile to remain available.
> See [`default-devstral-profile-work-packet.md`](default-devstral-profile-work-packet.md).

## Outcome

Fresh RecompHamr projects seed the accepted LM Studio `google/gemma-4-12b-qat` profile at `http://localhost:1234` with a 16,177-token context, instead of the previous Qwen/Ollama values.

## In scope

Seeded configuration values, the matching pre-probe packing fallback, focused tests, and user configuration documentation.

## Out of scope

Existing user configuration migration, automatic model loading, LM Studio installation, transport changes, TUI changes, or Stage C ownership movement.

## Authorities and evidence

User direction and the accepted Stage A runtime record, which verified LM Studio, `google/gemma-4-12b-qat`, localhost port 1234, and a 16,177-token context.

## Verification

Focused config/TUI tests and `pwsh -NoProfile -File ./scripts/verify.ps1` at 100% statement coverage.

## Documentation impact

Update the fresh configuration example and explicitly state that existing `.rehamr/config.yaml` files are never overwritten.

## Security impact

The endpoint remains loopback-only and no credential is seeded. Existing private-path protections are unchanged.

## Stop condition

Fresh bootstrap tests prove the exact Gemma values, the fallback matches, documentation is synchronized, and the canonical gate passes.

## Completion evidence

- Changed: fresh `local` profiles now seed `google/gemma-4-12b-qat`, `http://localhost:1234`, and context size 16,177; the pre-probe fallback matches.
- Documented: configuration example, LM Studio prerequisite, and existing-config non-overwrite behavior.
- Verified: focused config and TUI packages pass; the serialized canonical gate passes every check.
- Coverage: repository statement coverage remains 100.0% (`1887/1887`), with the existing configuration behavioral surface covering bootstrap and persistence.
- Security: loopback endpoint and empty key remain the defaults; private-path enforcement is unchanged.
- Evidence: exact bootstrap assertions plus the previously accepted real Gemma/LM Studio runtime record.
- Known limits: RecompHamr does not install, load, or start LM Studio/models; existing configuration files intentionally retain their user-controlled values.
