# windows-game-decomp

Windows game matching decompilation and compiler-matrix research skill.
Covers native PE, .NET, Unity (Mono + IL2CPP), and Unreal (UE4/UE5) targets.


## Kickoff

`/skill windows-game-decomp` — then "I have game.exe at [PATH]. Detect runtime family and start Phase 0."

## What it teaches

- Five tracks: matching decomp (A), behavioral port (B), compat shims (C),
  modding SDK (D), Lua/native ABI modding (E)
- Four match levels: compiles (0) → asm match (1) → C match same compiler (2)
  → C match any compiler (3) → functional port with shims (4)
- Runtime family detection from workspace layout
- Operational phases: Recon → Inventory → First Match → Bulk Match →
  Runtime Fill → Validation
- Engine-specific workflows for Unity Mono, Unity IL2CPP, Unreal, Native PE
- Three-build strategy: vanilla_matching, vanilla_behavioral, modding_sdk

## What it references

- `/skill ghidra-mcp` — unlocks `ghidra.*` tools for static analysis
- `/skill core-re` — RE workflow discipline
- `/skill evidence-mode` — evidence classification
- `REPHAMR_STATE.md` — persistent project memory

## When to use

Windows game reverse engineering — matching decompilation, 1:1 rebuilds,
compiler-matrix research, or modding SDK design on DOS/Win16/Win9x/Win32
native PE, .NET, Unity, or Unreal targets.
