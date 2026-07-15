# Work Packet Template

Use this template for a non-trivial task or phase.

## Outcome

Describe the observable result that must be true.

## In scope

List behavior, packages, files, data, and user surfaces intentionally changed.

## Out of scope

List nearby work that must not be mixed into this task.

## Authorities

List the repository instructions, current external documentation, source references, and standards that govern the work.

## Evidence before editing

Record source, tests, runtime observations, baselines, screenshots, hashes, or reference behavior inspected before implementation.

## Implementation approach

Describe ownership boundaries and the smallest coherent approach. For parity work, state whether the target is direct compatibility, adapted compatibility, or an improved replacement.

## Behavioral surface inventory

List every retained, modified, replaced, parity, and newly added behavioral surface in scope. For each surface, identify implementation ownership, applicable test categories, documentation, verification evidence, and status. Mark a test category not applicable only with a concrete reason.

The task cannot close until all in-scope rows reach 100% behavioral surface coverage.

## Verification

List focused checks and the canonical gate. Include runtime/manual checks that automation cannot prove. State how 100% statement coverage and 100% behavioral surface coverage will both be demonstrated.

## Documentation impact

List documents/help/examples that must change, or explain why the behavior has no documentation impact. Confirm how 100% meaningful documentation coverage will be preserved, including package/exported-symbol Go documentation and all relied-on behavioral contracts.

## Security impact

Describe filesystem, process, network, permissions, secrets, trust, cancellation, or lifecycle implications.

## Stop condition

State the exact condition for closure.

## Completion evidence

- Changed:
- Documented:
- Verified:
- Coverage: statement coverage, behavioral-surface coverage, and meaningful documentation coverage
- Security:
- Evidence:
- Known limits:
