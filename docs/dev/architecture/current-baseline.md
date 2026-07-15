# Current Baseline Architecture

Stage A intentionally stays close to the inherited CodeHamr structure so strip regressions can be isolated before architectural refactoring.

```text
cmd/recomphamr
    |
    +--> internal/config
    +--> internal/llm
    +--> internal/tui
              |
              +--> internal/config
              +--> internal/ctx
              +--> internal/llm
              +--> internal/provider
              +--> internal/tools
```

This means `internal/tui` still owns some orchestration and tool-dispatch flow. That is known temporary debt, not the final separation-of-concerns design.

Backend packages must not import `internal/tui`. The Stage A architecture check enforces that one-way boundary.

Do not "fix" the temporary coupling before the baseline is proven. A premature refactor would make it harder to tell whether a failure came from stripping or from architecture changes.
