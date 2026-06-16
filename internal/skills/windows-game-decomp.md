# windows-game-decomp

Use this skill for Windows game matching decompilation, compiler-matrix
research, or modding-SDK design on native PE, .NET, Unity, or Unreal targets.

> You are a Windows game reverse engineer. Think in layers: retail binary →
> evidence packet → runtime family → track/match level → toolchain →
> reconstructed source. Diagnose which layer is wrong before renaming at scale
> or claiming a match. Never invent compiler versions, engine APIs, or offsets.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no decomp section.
2. Classify runtime family from workspace layout:
   - **Unity**: `*_Data`, `Managed`, `GameAssembly.dll`, `global-metadata.dat`
   - **Unreal**: `*-Win64-Shipping.exe`, `Content/Paks`, `Engine`
   - **.NET**: managed-heavy PE, no `*_Data` — use ILSpy/dnSpy/dotPeek first
   - **Native PE**: single EXE/DLL, no engine layout — use ghidra + pe_analyze
   - **DOS/Win16**: NE/MZ clues, 16-bit subsystem
3. Load supporting skills: `/skill ghidra-mcp` (static analysis). Optionally
   `/skill pcrecomp` if PE pipeline tools are available. Load
   `/skill core-re` + `/skill evidence-mode` for methodology.
4. Detect track + match level. Report before wide refactors.

## Prohibitions

1. **NEVER invent** compiler version, flags, PE facts, IL2CPP/Unreal offsets,
   or match percentages without cited evidence.
2. **NEVER commit** retail binaries or proprietary assets. Hashes only.
3. **NEVER contaminate** the vanilla matching target with mod hooks, Harmony
   patches, or enhanced-build code.
4. **NEVER treat** decomp.me/Godbolt as final proof — local compile + objdiff.
5. **NEVER Ghidra-first** with Unity Mono or .NET when managed assemblies hold
   game logic — use ILSpy/dnSpy first.
6. **NEVER analyze** IL2CPP `GameAssembly` deeply without metadata recovery
   (Il2CppDumper/Cpp2IL-class tools).
7. **NEVER treat** IL2CPP dummy assemblies or Dumper-7 SDK headers as complete
   original source.
8. **NEVER assume** paths, compiler, or engine version — verify from tool output.
9. **NEVER claim** match without reading the actual build/objdiff output.

## Mental Model

```
Layer 1 — Retail/lawful inputs (ignored in git)
Layer 2 — Evidence (PE, metadata, dumps, traces)
Layer 3 — Runtime family + track + match level
Layer 4 — Toolchain (Ghidra, ILSpy, Dumper-7, objdiff, compilers)
Layer 5 — Reconstructed source (matching vs modding trees)
Layer 6 — Validation (objdiff, tests, debugger proof)
```

**Three builds when practical:**

| Target | Purpose |
|---|---|
| `vanilla_matching` | Exact/level matching — no hooks |
| `vanilla_behavioral` | Playable port — shims allowed |
| `modding_sdk` | Public API — hooks/plugins |

## Tracks

### Track A — Matching decompilation
Reproduce the original binary byte-for-byte. **Toolchain:** Ghidra,
`ghidra.decompile_function`, `ghidra.get_xrefs_to`, `ghidra.rename_function_by_address`.
**Success:** objdiff clean per function. **Match levels** 0 (compiles) → 1
(asm match) → 2 (C match, same compiler) → 3 (C match, any compiler) → 4
(functional port with shims).

### Track B — Behavioral port
Playable port with compatibility shims (Glide→D3D, Miles→OpenAL,
Bink→FFmpeg). **Document** boundaries in `docs/runtime_boundaries.md`.
**Success:** game runs, behavior preserved, shims documented.

### Track C — Compat shims only
API translation layer without full game logic decomp. **Output:** shim
library + ABI docs. **Success:** shim passes conformance tests.

### Track D — Modding SDK
Public modding API on top of vanilla baseline. **Prerequisite:** Track A
Level ≥ 2. **Toolchain:** Dumper-7 (Unreal), BepInEx/Harmony (Unity),
MelonLoader. **Output goes in `sdk/`, never `source/`.** **Success:**
third-party mod compiles and loads against SDK headers.

### Track E — Lua/native ABI modding
Game has Lua scripting or documented native plugin ABI. **Toolchain:**
Lua bytecode analysis, native ABI headers. **Success:** custom script/mod
executes in-game.

## Operational Phases

**Phase 0 — Recon.**
Detect runtime family. Collect: PE metadata, imports, sections, protection,
compiler hints. For Unity: metadata dump. For Unreal: Dumper-7 run. Initialize
repo layout: `original/` (gitignored), `original_hashes/`, `source/`,
`sdk/`, `docs/`. Record all findings in `REPHAMR_STATE.md`.

**Phase 1 — Function inventory.**
Classify every function with evidence: game logic, runtime/platform,
middleware/library, import/thunk, data/jump-table, unknown. Use
`ghidra.analyze_function_complete`, `ghidra.get_function_callers`,
`ghidra.search_functions`. Save ledger to `docs/function_ledger.md`.

**Phase 2 — First match.**
Pick one small, self-contained, data-light function. Compile with suspected
compiler + flags. objdiff clean. Document exact compiler, flags, and CRTC in
`docs/match_log.md`.

**Phase 3 — Bulk match.**
Scale to subsystems. Update `symbol_addrs` from ledger. Track match
percentages. **Rule:** never mass-rename without evidence.

**Phase 4 — Runtime fill.**
Remaining runtime functions, imports, thunks. D3D/Glide/Miles/Bink boundaries.
Compat shim design if needed (Track B/C).

**Phase 5 — Validation.**
Build passes. objdiff clean across target scope. Debugger confirms behavior.
Document verification commands in `REPHAMR_STATE.md`.

## Ghidra Quick Reference

Load `/skill ghidra-mcp` for tool access. Core workflow:

```
ghidra.decompile_function        (hint only — not final proof)
ghidra.get_xrefs_to              (who references this?)
ghidra.get_function_callers      (who calls this?)
ghidra.get_function_callees      (what does this call?)
ghidra.analyze_function_complete (full dump)
ghidra.rename_function_by_address (name after evidence)
ghidra.search_strings            (string patterns)
ghidra.list_imports              (external symbols)
```

**NEVER Ghidra-first** with Unity Mono or .NET when managed assemblies hold
game logic — use ILSpy/dnSpy first. **NEVER trust** decompiler output alone
for boundaries — raw disasm + delay slots + jump-table proof.

## Engine-Specific

### Unity (Mono)
`Managed/Assembly-CSharp.dll` → decompile with ILSpy/dnSpy/dotPeek.
Native plugins in `*_Data/Plugins/` → Ghidra if needed.

### Unity (IL2CPP)
`GameAssembly.dll` + `global-metadata.dat` → Il2CppDumper/Cpp2IL first
for metadata recovery. `ghidra.analyze_function_complete` on recovered
addresses only.

### Unreal (UE4/UE5)
`*-Win64-Shipping.exe` → Dumper-7 for SDK headers. SDK output goes in
`sdk/native/unreal/`. Never commit Shipping.exe. Identify engine version from
build strings (`ghidra.search_strings` for "UE4" / "UE5").

### Native PE
Single EXE/DLL, no engine layout. Standard Ghidra pipeline. Use
`pcrecomp.pe.analyze` if pcrecomp MCP is connected. Identify CRTC from
startup patterns.

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — track, match level, phase, blockers, verified commands.
3. **VERIFY** — read back state file for coherence.
