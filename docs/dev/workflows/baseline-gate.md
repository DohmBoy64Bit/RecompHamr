# Baseline Gate

The baseline gate prevents feature integration until the stripped upstream application is proven usable.

## Automated checks

Run:

```powershell
pwsh -NoProfile -File ./scripts/verify.ps1
```

All automated checks must pass on Go 1.26+, including the required documentation contract, documentation-link validation, and strict 100% statement coverage. Baseline closure additionally requires 100% behavioral surface coverage for every retained barebones surface and 100% meaningful documentation coverage; statement coverage alone is not sufficient.

## Runtime checklist

On the target Windows environment:

1. launch `recomphamr`;
2. verify the startup composition is not broken or unexpectedly wrapped;
3. type and edit a multi-line prompt;
4. complete a real model turn;
5. execute each retained tool through the agent;
6. switch models with `/models`;
7. verify history recall after restart;
8. cancel an active stream/tool run;
9. resize through wide and constrained terminal sizes;
10. exit cleanly and verify the terminal is restored.

## Visual acceptance

Capture screenshots from the actual target terminal at representative sizes. Compare interaction and composition against the inherited upstream baseline, allowing only intended RecompHamr branding and removal of hosted-service-only elements.

## Close the gate only when

- automated verification passes;
- every retained baseline surface is present in the behavioral-surface inventory with complete applicable test coverage;
- 100% meaningful documentation coverage is confirmed for the retained baseline contract;
- runtime behavior is demonstrated;
- visual acceptance is explicit;
- `docs/dev/verification/baseline-status.md` is updated with exact evidence.

Until then, feature integration remains blocked.
