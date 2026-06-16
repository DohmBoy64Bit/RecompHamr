# objdiff

Object file diffing via encounter/objdiff. Validates match claims for
n64-decomp, windows-game-decomp, and xboxrecomp projects.

## What it enables

- List units from `objdiff.json` with match status
- Diff individual object files (target vs base)
- Generate progress reports with match percentages
- Automated match validation for Build Gates

## Setup

1. Install: `cargo install --locked --git https://github.com/encounter/objdiff.git objdiff-cli`
2. Ensure `objdiff-mcp` on PATH (or set `RECOMPHAMR_MCP_OBJDIFF_COMMAND`)
3. Project must have `objdiff.json` config

## When to use

Matching decomp validation — verifying recompiled C produces identical
assembly to the original binary. Used by n64-decomp, windows-game-decomp,
and xboxrecomp Build Gates.
