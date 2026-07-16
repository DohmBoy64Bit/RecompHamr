# Definition of Done

A change or stage slice is complete only when:

- the intended behavior is implemented without placeholders or fake success paths;
- affected tests are updated or added;
- every reproducible bug fix has a regression test when practical;
- every retained, modified, replaced, and newly added behavioral surface is inventoried and reaches 100% behavioral surface coverage with all applicable contract, failure, boundary, lifecycle, platform, and security cases covered or explicitly justified as not applicable;
- 100% meaningful documentation coverage is maintained, including package documentation and appropriate Go doc comments for every exported symbol plus complete documentation of relied-on user, integration, configuration, persistence, security, lifecycle, and extension contracts;
- user and developer documentation agree with the code;
- `scripts/verify.ps1` passes in the supported toolchain, including the strict 100% statement coverage gate, or the exact blocker is recorded;
- no forbidden subsystem or TUI dependency drift is introduced;
- known limits are stated without pretending they are solved.

Baseline acceptance additionally requires real runtime and manual TUI evidence as defined by the baseline gate.

## Future Legacy feature work

A parity feature is complete only when:

- the required Legacy capability contract is documented from evidence;
- the chosen current architecture owner is explicit;
- the implementation is verified independently of the TUI where possible;
- the disposition is recorded as `equivalent`, `improved`, `intentionally changed`, `not applicable`, `blocked`, or `unverified`;
- compatibility or migration behavior is documented and tested when applicable;
- every Legacy-derived, replaced, improved, and newly introduced surface for the feature is included in the behavioral-surface inventory with complete applicable test and documentation coverage;
- source, tests, docs, runtime evidence, behavioral-surface evidence, and the parity record agree.

One-to-one source similarity is not required.

## Future Agent Skills work

A migrated skill is complete only when it satisfies `../roadmap/agent-skills-standard.md`, including standards-compliant structure, focused instructions, trigger-boundary evaluation, output-quality evaluation, supporting-resource justification, and per-skill evidence.
