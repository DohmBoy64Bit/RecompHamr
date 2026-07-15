# RecompHamr-Legacy Parity and Modernization Policy

## Purpose

RecompHamr-Legacy is a behavior and feature reference. It is not the reconstruction base, not an architecture specification, and not a requirement to preserve historical implementation choices.

The goal is **capability parity or an explicitly documented better replacement**, not source-code similarity.

## What must be preserved

For each feature selected for integration, identify the compatibility contract that actually matters, including as applicable:

- user-visible capability;
- command names and arguments that users or automation depend on;
- configuration/data formats that require migration or backward compatibility;
- tool schemas and result/error behavior relied on by callers;
- persistent state behavior;
- security and trust expectations;
- observable lifecycle behavior.

Only verified requirements belong in the parity contract.

## What may be improved

The implementation may be redesigned when the new design is better supported by the current architecture and evidence. Improvements may include:

- stronger separation of concerns;
- clearer typed interfaces;
- safer configuration and secret handling;
- better cancellation and lifecycle management;
- simpler package ownership;
- fewer hidden fallbacks;
- deterministic registries instead of duplicated command lists;
- stronger tests and diagnostics;
- standards-based replacement of proprietary formats or loaders;
- removal of obsolete defects or technical debt;
- better performance or context efficiency when measured honestly.

Do not preserve a bug or poor internal design merely for internal similarity. Preserve bug-compatible behavior only when a real external compatibility requirement is documented.

## Per-feature parity workflow

1. **Inventory Legacy evidence**
   - read the exact Legacy source, tests, docs, configuration, and user-facing behavior relevant to the feature;
   - record unknowns rather than inferring them.
2. **Define the required contract**
   - separate required behavior from incidental internals and defects.
3. **Select the current owner**
   - place behavior in the current architecture, not automatically in the old package path.
4. **Compare implementation options**
   - direct port;
   - adapted port;
   - clean rewrite behind the same contract;
   - standards-based replacement.
5. **Choose the best evidence-backed option**
   - document why it is safer, simpler, more maintainable, or more compatible.
6. **Inventory the complete behavioral surface**
   - include retained Legacy behavior, compatibility behavior, replacements, intentional changes, and every new or improved surface introduced by the modern implementation;
   - no old surface is grandfathered and no new surface is exempt.
7. **Test below presentation first**
   - prove backend behavior before TUI exposure;
   - reach 100% behavioral surface coverage with every applicable contract, error, boundary, lifecycle, platform, persistence, migration, and security case represented.
8. **Document the complete contract**
   - maintain 100% meaningful documentation coverage for the required capability, approved changes, configuration, migration, errors, trust boundaries, and extension behavior.
9. **Connect presentation last**
   - preserve the accepted TUI contract unless a later approved UI phase explicitly changes it.
10. **Close the parity record**
   - implementation, statement coverage, behavioral-surface coverage, documentation coverage, runtime evidence, and disposition must agree.

## Dispositions

Every tracked behavior ends in one of these states:

- `equivalent` — required Legacy capability is preserved;
- `improved` — required capability is preserved and the new contract has a documented upgrade;
- `intentionally changed` — an approved incompatibility or replacement is explicit;
- `not applicable` — obsolete or outside the current product, with evidence;
- `blocked` — named external input or environment prevents completion;
- `unverified` — evidence is incomplete.

A feature is not closed merely because code was copied or a similarly named command exists.

## Standards take precedence over historical custom formats when approved

When a mature standard better fits the capability, prefer the standard over restoring a proprietary Legacy design, provided required compatibility and migration behavior are handled explicitly.

The planned skills migration is the first explicit example: future RecompHamr skills support must implement the Agent Skills standard and convert Legacy skill knowledge into compliant skills rather than restoring the old skill format one-for-one.
