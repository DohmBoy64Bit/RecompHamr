# windows-game-decomp

Use this skill for Windows game matching decompilation, compiler-matrix
research, or modding-SDK design on native PE, .NET, Unity (Mono/IL2CPP),
Unreal (UE4/UE5), or DOS/Win16 targets.

> You are a Windows game reverse engineer. Think in layers: retail binary →
> evidence packet → runtime family → track/match level → toolchain →
> reconstructed source. Diagnose which layer is wrong before renaming at scale
> or claiming a match. Never invent compiler versions, engine APIs, or offsets.

## When to use

Use this skill when:
- Starting a Windows game decompilation project from a PE binary
- Identifying compiler, CRTC, and build toolchain for matching
- Working with Unity (Mono or IL2CPP), Unreal (UE4/UE5), .NET, or native PE targets
- Designing a modding SDK on top of a vanilla matching baseline

Do not use this skill when:
- The target is OG Xbox (use `xboxrecomp`), Xbox 360 (use `xbox360-decomp`),
  or console ROM targets (use platform-specific skill)
- The task is general binary analysis without matching intent (use `ghidra-mcp`)
- You need PC static recompilation via pcrecomp pipeline (use `pcrecomp`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the decomp section with:
   - target binary name, PE hash (if available), runtime family
   - current track (A-E), match level (0-4), active blocker
   - detected compiler/CRTC if known
2. Detect workspace layout. Do not assume paths. Look for:
   - Target binary (EXE/DLL), any PDB or debug symbols
   - `original/` directory (gitignored), `original_hashes/`
   - `source/` (matching), `sdk/` (modding), `docs/` (ledgers, logs)
3. Classify runtime family from workspace evidence (do not guess):
   - **Unity**: `*_Data`, `Managed/`, `GameAssembly.dll`, `global-metadata.dat`
   - **Unreal**: `*-Win64-Shipping.exe`, `Content/Paks/`, `Engine/`
   - **.NET**: managed-heavy PE, no `*_Data` — use ILSpy/dnSpy/dotPeek first
   - **Native PE**: single EXE/DLL, no engine layout — use Ghidra + pe_analyze
   - **DOS/Win16**: NE/MZ clues, 16-bit subsystem
4. Verify required tools:
   - Ghidra (native/IL2CPP targets)
   - ILSpy/dnSpy/dotPeek (Unity Mono / .NET targets)
   - Il2CppDumper/Cpp2IL (Unity IL2CPP targets)
   - Dumper-7 (Unreal targets)
   - objdiff + suspected compiler (matching tracks)
5. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill ghidra-mcp` — static analysis for native/IL2CPP
   - `/skill objdiff` — object file match validation
   - `/skill core-re` + `/skill evidence-mode` — methodology
   - `/skill pcrecomp` — optional, if PCRECOMP MCP is connected
6. Detect track + match level from workspace evidence. Report: runtime family +
   track + one next step. Wait for go-ahead on destructive refactors.

## Prohibitions

1. **NEVER invent** compiler version, flags, PE facts, IL2CPP/Unreal offsets,
   or match percentages without cited evidence from tool output.
2. **NEVER commit** retail binaries, proprietary assets, or crack/key material.
   Hashes only in `original_hashes/`.
3. **NEVER contaminate** the vanilla matching target with mod hooks, Harmony
   patches, or enhanced-build code in `source/`. Modding output goes in `sdk/`.
4. **NEVER treat** decomp.me/Godbolt as final proof — local compile + objdiff
   for claims.
5. **NEVER Ghidra-first** with Unity Mono or .NET when managed assemblies hold
   game logic — use ILSpy/dnSpy/dotPeek first.
6. **NEVER analyze** IL2CPP `GameAssembly` deeply without metadata recovery
   (Il2CppDumper/Cpp2IL-class tools).
7. **NEVER treat** IL2CPP dummy assemblies or Dumper-7 SDK headers as complete
   original source.
8. **NEVER assume** paths, compiler, engine version, or PE layout — verify from
   tool output and local file inspection.
9. **NEVER claim** match without reading the actual build/objdiff output.
10. **NEVER mass-rename** functions without evidence-backed classification.
11. After 3 failed match attempts on the same function, STOP — update state file,
    gather fresh evidence (raw disasm, xrefs, compiler output) before retrying.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| Retail binary | Original artifact — hash verified, gitignored | `original/` (never committed) |
| Evidence | PE metadata, IL2CPP dumps, Dumper-7 output, traces | `.rehamr/evidence/` |
| Runtime family | Unity / Unreal / .NET / Native PE / DOS-Win16 | `REPHAMR_STATE.md` |
| Toolchain | Ghidra, ILSpy, Dumper-7, objdiff, compilers | Installed tools |
| Reconstructed source | Matching C/C++ or modding code | `source/` (matching), `sdk/` (modding) |
| Validation | objdiff, tests, debugger proof | `docs/match_log.md` |

### Three Builds

| Target | Purpose | Location |
|---|---|---|
| `vanilla_matching` | Exact/level matching — no hooks | `source/` |
| `vanilla_behavioral` | Playable port — shims allowed | project root |
| `modding_sdk` | Public API — hooks/plugins | `sdk/` |

### Evidence Ladder

1. Build output / objdiff clean (strongest)
2. Raw disassembly + PE metadata
3. `ghidra.get_xrefs_to` / `ghidra.analyze_function_complete` (static)
4. ILSpy/dnSpy decompiled C# (managed — higher trust than Ghidra for IL)
5. `ghidra.decompile_function` (hint only — never final proof for native)
6. decomp.me / Godbolt (scratch — never final proof)

## Tracks

| Track | Use When | Pipeline | Success Criteria |
|---|---|---|---|
| A — Matching decomp | Goal is byte-identical binary | PE → Ghidra → function inventory → match → objdiff | objdiff clean per function |
| B — Behavioral port | Playable with compat shims | PE → analyze → shim (Glide→D3D, Miles→OpenAL) → build | Game runs, behavior preserved |
| C — Compat shims only | API translation, no full game logic | PE → API analysis → shim library + ABI docs | Shim passes conformance tests |
| D — Modding SDK | Public modding API on Track A baseline | Track A L2+ → Dumper-7/BepInEx/Harmony → SDK headers | Third-party mod loads against SDK |
| E — Lua/native ABI | Lua scripting or native plugin ABI | Bytecode analysis → ABI headers → custom script | Custom script executes in-game |

**Match levels (Track A):** 0 (compiles) → 1 (asm match) → 2 (C match, same compiler) → 3 (C match, any compiler) → 4 (functional port with shims).

## Pipeline

```text
Retail binary → PE analysis → runtime family detection → function inventory
              → first match (objdiff) → bulk match → runtime fill → validation
```

## Operational Phases

**Phase 0 — Recon.**
Goal: detect runtime family and collect evidence baseline.
- Collect: PE metadata, imports, sections, protection, compiler hints
- Unity: metadata dump. Unreal: Dumper-7 run
- Initialize: `original/` (gitignored), `original_hashes/`, `source/`, `sdk/`, `docs/`
- Record: runtime family, compiler hints, PE hash in `REPHAMR_STATE.md`
- Exit: runtime family confirmed; repo layout initialized

**Phase 1 — Function inventory.**
Goal: classify every function with evidence.
- Classify: game logic, runtime/platform, middleware/library, import/thunk,
  data/jump-table, unknown
- Use `ghidra.analyze_function_complete`, `ghidra.get_function_callers`,
  `ghidra.search_functions`
- Save: `docs/function_ledger.md`
- Exit: ledger populated; no functions remain unclassified

**Phase 2 — First match.**
Goal: prove one small function matches byte-for-byte.
- Pick: small, self-contained, data-light function
- Compile with suspected compiler + flags
- Verify: objdiff clean
- Document: exact compiler, flags, CRTC in `docs/match_log.md`
- Exit: one function with confirmed compiler + flag set

**Phase 3 — Bulk match.**
Goal: scale to subsystems with tracked match percentages.
- Update `symbol_addrs` from function ledger
- Track match percentages in `REPHAMR_STATE.md`
- Never mass-rename without evidence
- Exit: target match percentage reached; objdiff clean across scope

**Phase 4 — Runtime fill.**
Goal: resolve remaining runtime functions, imports, thunks.
- Identify: D3D/Glide/Miles/Bink boundaries
- Design compat shims if needed (Track B/C)
- Document: boundaries in `docs/runtime_boundaries.md`
- Exit: no unresolved imports; shim strategy documented

**Phase 5 — Validation.**
Goal: prove match or behavioral parity.
- Build passes; objdiff clean across target scope
- Debugger confirms behavior at key addresses
- Document: verification commands in `REPHAMR_STATE.md`
- Exit: all validation gates met

## Build Gate / Validation Gate

Before claiming match:
1. **INSPECT** — build command matches documented compiler + flags from Phase 2.
2. **VERIFY** — objdiff clean for the function or scope being claimed.
3. **DOCUMENT** — match log entry in `docs/match_log.md` with compiler, flags, CRTC.

Success may only be claimed when:
- objdiff reports clean for the target function
- compiler + flags are documented
- match log is updated

## Tool Quick Reference

```bash
# Verified tool references. Do not invent flags.
# Unity Mono: ILSpy/dnSpy/dotPeek for Managed/Assembly-CSharp.dll
# Unity IL2CPP: Il2CppDumper/Cpp2IL for global-metadata.dat
# Unreal: Dumper-7 for Shipping.exe → sdk/native/unreal/
```

MCP tools (when connected):

```text
ghidra.decompile_function        — hint only, never final proof
ghidra.get_xrefs_to              — who references this address?
ghidra.get_function_callers      — who calls this function?
ghidra.get_function_callees      — what does this function call?
ghidra.analyze_function_complete — full dump: xrefs, callees, callers, vars
ghidra.rename_function_by_address — name after evidence
ghidra.search_strings            — string patterns (UE4/UE5 version strings)
ghidra.list_imports              — external symbols
pcrecomp.pe.analyze              — PE metadata, sections, compiler hints
```

## Engine-Specific Guidance

### Unity (Mono)
- **First:** `Managed/Assembly-CSharp.dll` → decompile with ILSpy/dnSpy/dotPeek
- Native plugins in `*_Data/Plugins/` → Ghidra if needed
- Never Ghidra-first for managed code

### Unity (IL2CPP)
- **First:** Il2CppDumper/Cpp2IL for metadata recovery from `global-metadata.dat`
- Then: `ghidra.analyze_function_complete` on recovered addresses only
- Never analyze `GameAssembly.dll` without metadata

### Unreal (UE4/UE5)
- **First:** Dumper-7 for SDK headers from `*-Win64-Shipping.exe`
- SDK output → `sdk/native/unreal/` (never commit Shipping.exe)
- Identify engine version: `ghidra.search_strings` for "UE4"/"UE5"
- Never treat Dumper-7 SDK headers as complete original source

### Native PE
- Standard Ghidra pipeline for single EXE/DLL
- Use `pcrecomp.pe.analyze` if pcrecomp MCP is connected
- Identify CRTC from startup patterns

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| objdiff not clean | Compiler/flags | objdiff output, function asm | Adjust compiler version/flags; document in match log |
| Unknown function boundary | Function inventory | Raw disasm, xrefs, delay slots | Classify with evidence; update ledger |
| IL2CPP addresses unresolved | Metadata | Il2CppDumper/Cpp2IL output | Recover metadata before analyzing GameAssembly |
| Managed code treated as native | Runtime family detection | ILSpy/dnSpy output | Use managed decompiler first for Unity/.NET |
| Mod hooks in `source/` | Track discipline | Directory listing | Move to `sdk/`; keep `source/` vanilla |
| Dumper-7 SDK incomplete | Source of truth | Original Shipping.exe analysis | SDK is scaffold — never treat as complete source |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — runtime family, track, match level, matched percentage, blocker
- `docs/function_ledger.md` — function inventory with classification and evidence
- `docs/match_log.md` — compiler, flags, CRTC per matched function

Optional:
- `.rehamr/evidence/windows_recon.md` — PE metadata, compiler hints, engine version
- `docs/runtime_boundaries.md` — D3D/Glide/Miles/Bink boundary documentation (Track B/C)
- `docs/il2cpp_metadata.md` — IL2CPP metadata dump (Unity IL2CPP targets)
- `original_hashes/` — verified hashes of original binaries (gitignored)

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — track, match level, matched percentage, blockers, verified commands.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified matches, remaining blockers, next 3 concrete steps.
