# Default Devstral Profile Work Packet

## Outcome

Fresh RecompHamr projects seed LM Studio `mistralai/devstral-small-2-2512` as the active `local` profile and retain `google/gemma-4-12b-qat` as the selectable `gemma` profile. Existing user configuration remains untouched.

## Scope and evidence

The user selected Devstral on 2026-07-17 and explicitly required the Gemma profile to remain. This changes only fresh seeded configuration, focused assertions, documentation, Stage G evaluation defaults, and later Stage G runtime acceptance. Automatic model loading, existing-config migration, transport behavior, TUI layout, and prior accepted runtime records are out of scope.

Both profiles use the existing loopback `http://localhost:1234` endpoint, empty key, and conservative 16,177-token packing fallback. Users must set `context_size` to the window actually loaded by LM Studio. No credential or remote endpoint is introduced.

## Verification and stop condition

Focused configuration tests must prove Devstral is active, Gemma remains selectable, stable profile ordering is preserved, and existing configurations are not merged or overwritten. Stage G skill evaluations and exact-build runtime acceptance must use Devstral. Closure still requires the canonical 100% gate and dual-platform CI.
