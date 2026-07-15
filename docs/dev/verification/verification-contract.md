# Verification Contract

## Automated gate

`scripts/verify.ps1` runs:

1. baseline policy checks;
2. the required documentation contract;
3. documentation-link checks;
4. Stage A architecture checks;
5. `gofmt` verification;
6. `go test ./...` with an atomic coverage profile;
7. the strict 100% statement coverage gate;
8. `go build ./cmd/recomphamr`;
9. a CLI help smoke test.

## Documentation contract

`cmd/docscheck` requires durable project documents to exist, be non-empty, and retain mandatory project terms. `scripts/check-docs.ps1` runs that contract first and then independently checks all relative Markdown links.

This is a structural/documentation-coverage guard. A passing keyword contract does not, by itself, prove that every sentence is semantically correct; code-versus-documentation review remains required.

## Statement coverage contract

`scripts/check-coverage.ps1` runs:

```powershell
go test ./... -covermode=atomic -coverprofile=<temporary-profile>
```

and then validates that profile with `cmd/coveragecheck`. Every instrumented Go statement for the active build platform must be covered. Anything below 100.0% fails the canonical gate.

The coverage profile is temporary and is removed after the check.

## Behavioral surface coverage contract

The repository also requires **100% behavioral surface coverage**. This is separate from `coverage.out` and cannot be inferred from line or statement execution alone.

Every retained upstream behavior, modified behavior, replacement, Legacy parity behavior, optimization with observable consequences, and newly added behavior must appear in the active behavioral-surface inventory. Each row must map the surface to implementation ownership, applicable tests, documentation, verification evidence, and status. Applicable success, failure, malformed-input, boundary, cancellation, timeout, cleanup, platform, Unicode, persistence, migration, compatibility, concurrency, and security cases must be covered. A category may be marked not applicable only with a recorded reason.

No old behavior is grandfathered and no new behavior is exempt. Phase or task closure requires all in-scope surface rows to be complete. See [`behavioral-surface-coverage.md`](behavioral-surface-coverage.md).

## Meaningful documentation coverage contract

The repository requires **100% meaningful documentation coverage** for all retained, modified, replaced, parity, and new surfaces. Every Go package and exported symbol must have appropriate Go documentation, and every relied-on user, integration, configuration, persistence, security, lifecycle, and extension contract must be documented. Trivial private locals do not require artificial comments.

The documentation contract and link checker are structural gates; reviewers and task evidence must also verify semantic agreement between implementation and documentation.

## What automation does not prove

Automation alone does not prove:

- visual parity;
- real terminal behavior;
- interactive key handling on the target terminal;
- Windows PowerShell child-process behavior;
- end-to-end model compatibility with every OpenAI-compatible endpoint.

Those require runtime evidence.
