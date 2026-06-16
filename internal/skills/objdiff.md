# objdiff

Object file diffing and match validation for decompilation projects. Load
this skill to unlock `objdiff.*` tools for comparing compiled object files
against original assembly targets.

## What it enables

- List all units from `objdiff.json` with match status
- Diff a specific object: compare target (original) vs base (recompiled)
- Generate progress reports with match percentages
- Validate match claims with automated object comparison

Used by Build Gates in `n64-decomp`, `windows-game-decomp`, and `xboxrecomp`
skills to verify function-level matching before claiming success.

## Setup

1. Install objdiff: `cargo install --locked --git https://github.com/encounter/objdiff.git objdiff-cli`
2. Ensure `objdiff-mcp` is on PATH (or set `RECOMPHAMR_MCP_OBJDIFF_COMMAND`)
3. Project must have `objdiff.json` config
4. Start recomphamr — auto-connects at launch
5. Load `/skill objdiff` — unlocks `objdiff.*` tools

## When to use

Any matching decompilation project — verifying that recompiled C produces
identical assembly to the original binary. Core validation for n64-decomp,
windows-game-decomp, and xboxrecomp tracks.
