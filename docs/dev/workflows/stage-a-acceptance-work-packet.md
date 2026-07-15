# Stage A Acceptance Work Packet

## Outcome

The stripped RecompHamr baseline passes the canonical Go 1.26 gate on Windows and Ubuntu, has complete statement, behavioral-surface, and meaningful documentation coverage, and is manually accepted in Windows Terminal against LM Studio using the user-selected replacement model `google/gemma-4-12b-qat`.

## In scope

- Verification-script and CI determinism.
- Correct POSIX permissions and Windows ACL protection for `.rehamr` state.
- Focused refactors required to test retained behavior deterministically.
- Complete tests and documentation for every retained Stage A surface.
- Windows Terminal runtime, resize, cancellation, tool, persistence, and restoration evidence.

## Out of scope

- Legacy feature integration, MCP, skills, reverse-engineering helpers, new commands, and Stage C extraction.
- Bubble Tea, Bubbles, Lip Gloss, or Glamour upgrades.
- TUI redesign, rearrangement, restyling, or migration to the alternate screen.

## Authorities

- Repository and subtree `AGENTS.md` files.
- `baseline-gate.md`, `verification-contract.md`, and `behavioral-surface-coverage.md`.
- Current source, focused tests, Go 1.26 behavior, and Bubble Tea v1 behavior.

## Evidence before editing

- Local Go 1.26.4 Windows gate reached tests but failed POSIX-mode assertions on Windows.
- Initial GitHub Actions run `29418319255` failed Windows formatting because checkout line endings were not pinned.
- Package statement coverage was below 100%, with the largest gaps in CLI startup, LLM transport, tools, and TUI orchestration.
- LM Studio 1.3.3 is installed; Devstral Small 2 is present but the server and model were not loaded at packet creation.

## Implementation approach

Preserve runtime behavior and expose only narrow deterministic seams for testing. Implement platform-native state protection, add contract-level tests, and close surface rows only with source, automation, documentation, and required runtime evidence.

## Behavioral surface inventory

The active inventory is `../verification/stage-a-behavioral-surface.md`. Every row must reach `complete`; statement execution alone cannot close a row.

## Verification

- Focused package tests and coverage while changing each owner.
- `pwsh -NoProfile -File ./scripts/verify.ps1` on Go 1.26.4.
- Passing Windows and Ubuntu GitHub Actions jobs.
- Recorded Windows Terminal acceptance at wide, 80x24, and constrained sizes.

## Documentation impact

Update user configuration/tool contracts, developer verification evidence, Go package/exported-symbol documentation, the documentation map, and the machine-readable contract. Do not use keyword-only edits without matching source evidence.

## Security impact

Configuration, history, and debug logs may contain credentials, prompts, or tool arguments. POSIX paths require owner-only modes; Windows paths require a current-user-only DACL. Tests and evidence must never print secret values.

## Stop condition

The exact accepted commit passes local and dual-platform CI gates, every behavioral row is complete, manual evidence is recorded without secrets, and `baseline-status.md` says accepted with exact evidence.

## Completion evidence

- Changed: verification portability, platform-native private-state protection, and deterministic test seams completed without changing the frozen layout or dependency stack.
- Documented: configuration/security guarantees, behavioral inventory, and runtime acceptance are durable and contract-checked.
- Verified: canonical local gate plus independent Windows and Ubuntu CI jobs.
- Coverage: 100% meaningful statement and behavioral-surface coverage.
- Security: POSIX modes and Windows current-user DACLs tested; committed evidence contains no secrets or raw prompts.
- Evidence: `../verification/stage-a-runtime-acceptance.md` and its three representative screenshots.
- Known limits: visual acceptance targets Windows Terminal only; the accepted real-model runtime is Gemma 4 12B QAT, replacing Devstral at the user's direction.
