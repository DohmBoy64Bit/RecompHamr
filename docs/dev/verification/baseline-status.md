# Barebones Baseline Status

## Gate

**OPEN — NOT YET MANUALLY ACCEPTED**

## Source lineage

Verified in the construction pass:

- the template was rebuilt from a fresh official CodeHamr source archive rather than RecompHamr-Legacy;
- the supplied upstream snapshot was identified as `40c7cca098b8d53d97d1b4fe4b343a68f063be1e`;
- the source tree was reconstructed to the pinned upstream revision `85409d167b97bec64ee330d51872d358d3ce2d57` before stripping;
- a full pre-strip SHA-256 tree manifest is included.

## Source-shaping status

Completed:

- hosted CodeHamr service behavior removed;
- `internal/cloud` removed;
- updater/re-exec subsystem removed;
- installers, demo/promotional material, and upstream release automation removed;
- active product naming moved to RecompHamr, `.rehamr`, and `RECOMPHAMR_*`;
- `bash` replaced by the four-tool baseline centered on native `powershell`;
- the inherited Bubble Tea v1 dependency stack remains frozen;
- no RecompHamr-Legacy feature family was integrated.

## Construction-environment probes

Completed without weakening the repository's Go 1.26 requirement:

- the pre-strip upstream SHA-256 manifest re-verified all 63 reconstructed upstream files;
- all Markdown relative links resolve;
- all local Go imports resolve to repository packages;
- active Go code contains no forbidden hosted-service/update/MCP/skills/legacy-shell references;
- `gofmt` completed cleanly;
- in an isolated temporary Go 1.23 probe containing only standard-library-dependent packages, `internal/provider`, `internal/tools`, `internal/llm`, and the config-independent `internal/ctx` tests passed;
- the isolated `internal/tools` test package cross-compiled for Windows/amd64.

These probes are supporting evidence only. They do not replace the required full Go 1.26+ application gate.

## Automated evidence required for closure

The canonical command is:

```powershell
pwsh -NoProfile -File ./scripts/verify.ps1
```

It must pass on the supported Go 1.26+ environment, including the required documentation contract and the strict 100% statement coverage gate. Closure also requires a complete behavioral-surface inventory proving 100% coverage of retained baseline behavior and confirmation of 100% meaningful documentation coverage.

## Manual evidence still required

The gate remains open until the real target environment proves:

- startup;
- composer editing;
- streaming output;
- transcript scrolling;
- model selection;
- history persistence;
- cancellation;
- resize handling;
- clean exit;
- native PowerShell execution;
- terminal restoration after exit;
- visual acceptance in the user-selected terminal.

## Current construction-environment limitation

The construction environment does not provide the required Go 1.26 toolchain and cannot resolve external module hosts. Full `go test ./...`, the strict 100% statement coverage gate, and the complete application build therefore remain unverified here. Static checks and isolated standard-library package probes do not replace that required gate.

## Next allowed work

1. run the canonical verification gate on Go 1.26+;
2. run the manual runtime checklist on Windows;
3. accept the frozen TUI baseline;
4. begin separation-of-concerns extraction with behavior and layout frozen;
5. only then begin RecompHamr feature integration.
