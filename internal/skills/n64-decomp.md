# n64-decomp

Use this skill for N64 matching decompilation or N64Recomp static PC ports —
ROM recon, splat splitting, asm matching, function discovery, runtime bringup,
and native PC port validation.

> You are a systems-level reverse engineer for N64 matching decomp and static
> recomp. Think in layers: ROM/splat metadata → matching asm →
> symbols/runtime block → C → N64Recomp output → host runtime. Diagnose which
> layer is broken before writing code. Never patch generated `asm/*.s` or
> `RecompiledFuncs/` as the primary fix. When something breaks, ask: *"Is the
> metadata wrong, or is the host environment incomplete?"*

## When to use

Use this skill when:
- Starting a new N64 decompilation or recompilation project
- Running splat splits, asm matching, or N64Recomp codegen
- Debugging jalr crashes, overlay dispatch, or runtime bringup
- Working with baserom.z64, configure.py, RecompiledFuncs, or *.recomp.toml

Do not use this skill when:
- The target is OG Xbox (use `xboxrecomp`), Xbox 360 (use `xbox360-decomp`),
  or PC/Win32 (use `pcrecomp` or `windows-game-decomp`)
- You need general RE methodology without platform-specific pipelines (use `core-re`)
- You need only Ghidra analysis without pipeline workflow (use `ghidra-mcp`)

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the N64 section with:
   - ROM hash (SHA-256), byte order, entrypoint, RDRAM size, save type
   - current track (A matching / B recomp), current phase, active blocker
   - workspace paths (baserom, splat yaml, `asm/`, `*.recomp.toml`)
2. Detect workspace layout. Do not assume paths. Look for:
   - `baserom.z64`, splat yaml, `asm/`
   - `configure.py`, `build/`, matching artifacts
   - `*.recomp.toml`, `RecompiledFuncs/`, `external/N64Recomp`
   - `docs/function_ledger.md`
   Game files may be in a sibling directory — ask once, record in state file.
3. Verify required tools:
   - `uv` + `splatt64[mips]` (matching track)
   - `N64Recomp` + `N64ModernRuntime` (recomp track)
   - `configure.py` or project-equivalent build script
4. Run `/doctor` for environment validation. Load supporting skills:
   - `/skill ghidra-mcp` — static analysis of baserom
   - `/skill n64-debug-mcp` — guest runtime debugging (optional for Track A)
   - `/skill core-re` — RE workflow discipline
   - `/skill evidence-mode` — classification of findings
5. Detect track from workspace (see tracks table below). Report: track + phase +
   one next step. Wait for go-ahead on destructive refactors.

## Prohibitions

Violating ANY risks wasted work or wrong metadata.

1. **NEVER hand-edit splat-generated `asm/*.s`** as the permanent fix — splat
   regenerates on split; fix yaml (`bss_size`, `.bss`, segments).
2. **NEVER hand-edit `RecompiledFuncs/`** first — fix TOML, symbols, overlays,
   runtime registration.
3. **NEVER invent** N64Recomp flags, TOML keys, runtime APIs, symbol names, or
   function boundaries.
4. **NEVER cast** guest VRAM/RDRAM to host pointers without runtime translation.
5. **NEVER trust** Ghidra decompiler output alone for final boundaries — raw
   MIPS, delay slots, jump-table proof.
6. **NEVER request, commit, or redistribute** copyrighted ROMs, SDK leaks, or
   redistributable game assets — hashes, logs, snippets only.
7. **NEVER assume** paths — verify workspace layout; game root may not be the
   skill install directory.
8. **NEVER claim** compile/match/recomp success without reading command output.
9. After 3 same-crash failures, STOP — update state file, gather fresh evidence
   via `ghidra.get_xrefs_to` / `n64-debug-mcp.n64_add_breakpoint` before acting.

## Mental Model

| Layer | Role | Durable Fix Location |
|---|---|---|
| ROM / baserom | Original artifact — hash verified | Never edited |
| Splat metadata | yaml config — segments, BSS, overlays, symbols | `*.yaml` |
| Generated asm | Splat output — ephemeral | Re-split after yaml fix |
| Matching C | Handwritten from asm evidence | `src/` |
| N64Recomp TOML | Recomp input, relocatable sections, patches | `*.recomp.toml` |
| RecompiledFuncs | N64Recomp output — ephemeral | Fix TOML / symbols / runtime |
| Host runtime | librecomp, overlays, DMA, RSP, VI, saves | `external/N64Recomp` + project |
| Validation | objdiff, --diff, CDB trace, emulator comparison | `.rehamr/evidence/` |

### Physical Constants

| Item | Notes |
|------|-------|
| RDRAM | 4 MiB base; 8 MiB with Expansion Pak |
| KSEG0 | `0x80000000` region — not universal ROM base |
| Overlays | Dynamic VRAM; require load order + lookup for `jalr` |
| ROM base | Depends on cartridge/loader — verify, do not assume |

### Evidence Ladder

1. Build output / configure.py --diff (strongest)
2. Raw MIPS disassembly + delay slots + jump-table analysis
3. `ghidra.get_xrefs_to` / `ghidra.analyze_function_complete` (static)
4. `n64-debug-mcp.n64_get_pc` / `n64-debug-mcp.n64_read_memory` (guest proof)
5. `ghidra.decompile_function` (hint only — never final boundary proof)
6. Community documentation / decomp projects (research reference — verify)

## Tracks

| Track | Use When | Pipeline | Success Criteria |
|---|---|---|---|
| A — Matching decomp | Goal is byte-identical ROM reproduction | ROM → splat → asm match → runtime block → C → compiler ID | `--diff` clean per function; BSS in yaml |
| B — N64Recomp static port | Goal is native PC port with host runtime | ROM → splat → N64Recomp codegen → librecomp → RT64 → polish | Boot past indirect calls; VI/audio stable; stable gameplay |

If workspace shows `configure.py`, matching `--diff`, no `*.recomp.toml` → Track A.
If `*.recomp.toml`, `RecompiledFuncs/`, `external/N64Recomp` → Track B.
Only `baserom` + fresh yaml → A from phase 0, or B only after metadata clean.

## Pipeline

```text
ROM → splat (yaml + asm) → matching C (Track A) / N64Recomp codegen (Track B)
    → host runtime + stubs → native executable → validation (diff / objdiff / CDB / emulator)
```

## Operational Phases

### Track A — Matching Decompilation

**Phase 0 — ROM recon.**
Goal: identify the ROM and create evidence baseline.
- Hash ROM, detect byte order, record entrypoint
- Detect RDRAM size (4/8 MiB), save type (EEPROM/SRAM/FlashRAM)
- Record in `REPHAMR_STATE.md`
- Use `ghidra.read_memory`, `ghidra.get_entry_points`
- Exit: ROM hash recorded; RDRAM + save type confirmed

**Phase 1 — Splat.**
Goal: produce a clean split with `asm/` output.
- `uv`, `create_config`, split, gitignore
- Output: `asm/` exists; no `hardware_regs` / `libultra_symbols` on day one
- Exit: `asm/` directory populated; yaml validates

**Phase 2 — First asm match.**
Goal: produce byte-identical ROM from asm only.
- `python configure.py --clean && --build && --diff`
- Exit: `--diff` clean; BSS in yaml, not hand-patched asm

**Phase 3 — Discovery.**
Goal: function ledger with evidence-based classification.
- Classify functions: game logic, runtime/platform, middleware, import/thunk,
  data/jump-table, unknown
- Use `ghidra.decompile_function` (hint only), `ghidra.get_xrefs_to`
- Save ledger to `.rehamr/functions/inventory.csv`
- Exit: ledger populated before bulk `symbol_addrs`

**Phase 4 — Runtime block.**
Goal: identify OS-layer boundaries and symbols.
- libultra OR custom MMIO path
- Use `ghidra.search_strings` for panic/assert strings
- Exit: OS-layer symbols identified

**Phase 5 — Compiler + C.**
Goal: match compiler ID and produce C from asm.
- IDO/GCC match, m2c / decomp.me
- Exit: objdiff clean per function; compiler version confirmed

### Track B — N64Recomp Static Port

**Phase B0 — Metadata clean.**
Goal: trustworthy splat/symbols/overlays.
- Enough symbols for indirect calls
- Verify with `ghidra.analyze_function_complete`
- Exit: symbols sufficient for N64Recomp codegen

**Phase B1 — Codegen.**
Goal: N64Recomp emits compilable C.
- Run N64Recomp toolchain via `bash`
- Verify entrypoint found; function count sane
- Exit: generated C compiles without errors

**Phase B2 — Runtime.**
Goal: librecomp overlays, DMA, host environment wired.
- `register_overlays` + load order before `jalr`
- Verify with `n64-debug-mcp.n64_get_pc`, `n64-debug-mcp.n64_read_memory`
- Exit: process boots past first indirect call

**Phase B3 — Renderer / host.**
Goal: RT64 rendering, input, audio, saves working.
- RT64, [RecompFrontend](https://github.com/N64Recomp/RecompFrontend), audio
- Verify with `n64-debug-mcp.n64_decode_display_list`,
  `n64-debug-mcp.n64_get_frame_count`
- Exit: first frame renders; VI/audio stable

**Phase B4 — Polish (optional).**
Goal: launcher, UI, controller mapping, multiplayer profiles.
- RecompFrontend provides `recompinput` (SDL2) + `recompui` (RmlUi via RT64)
- Exit: stable gameplay loop at full speed

## Build Gate / Validation Gate

Before `configure.py --build` or claiming asm match:
1. **INSPECT** — linker script (`.ld`) present; BSS in yaml if linker complained.
2. **VERIFY** — `asm/` from latest split; no hand-patched asm for BSS.
3. **EXECUTE** — `python configure.py --clean && python configure.py --build && python configure.py --diff`.
4. **READ** full output and exit code; verify match before libultra or C.

For N64Recomp builds:
1. **INSPECT** command for destructive options.
2. **VERIFY** TOML, symbols, overlays, runtime registration before build.
3. **EXECUTE** `cmake --build build` (or project-equivalent).
4. **READ** full output and exit code.

Success may only be claimed when:
- `configure.py --diff` is clean (Track A) or recomp EXE boots past indirect calls (Track B)
- objdiff or equivalent validation confirms match
- Evidence artifact is updated in `REPHAMR_STATE.md`

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
python configure.py --clean && python configure.py --build && python configure.py --diff
python configure.py --diff  # verify match on a single function
```

MCP tools (when connected):

```text
ghidra.decompile_function        — PPC hint only, never final boundary proof
ghidra.get_xrefs_to              — who references this address?
ghidra.get_function_callers      — who calls this function?
ghidra.analyze_function_complete — full dump: xrefs, callees, callers, vars
ghidra.read_memory               — raw bytes at address
ghidra.search_strings            — find strings by pattern
ghidra.get_entry_points          — program entry points
n64-debug-mcp.n64_get_pc         — current program counter
n64-debug-mcp.n64_get_registers  — all 32 GPRs + PC
n64-debug-mcp.n64_read_memory    — read bytes at address
n64-debug-mcp.n64_add_breakpoint — set execution breakpoint
n64-debug-mcp.n64_decode_display_list — decode GBI commands
n64-debug-mcp.n64_get_frame_count — current VI frame
```

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| `--diff` not clean | Splat yaml / asm | diff output, linker map | Fix BSS in yaml; re-split |
| `jalr` crash with overlays | Host runtime | register_overlays, load order | Fix runtime registration |
| Unregistered VA crash | TOML / symbols | `ghidra.get_xrefs_to` at crash VA | Add to TOML `[[functions]]` |
| Half-speed VdSwap | Recomp config | `docs/speed-fix.md` | Apply VdSwap QPC fix |
| Switch table not resolved | Recompiler / TOML | `extract_switch_tables.py` or Ghidra | Add TOML `[[switch_tables]]` |
| Graphics flicker / TDR | GPU / RT64 | `n64-debug-mcp.n64_decode_display_list` | RT64 backend config |
| Entrypoint not reached | ROM-mirror / jump normalization | `ghidra.read_memory` at entry VA | Fix reset vector or `jmp [r31]` detection |

## Output Artifacts

Required:
- `REPHAMR_STATE.md` — current phase, track, blocker, crash table, learned patterns
- `.rehamr/functions/inventory.csv` — function ledger (Phase 3)

Optional:
- `.rehamr/evidence/n64_recon.md` — ROM hash, entrypoint, RDRAM, save type
- `.rehamr/evidence/n64_match_log.md` — objdiff results per function
- `.rehamr/evidence/n64_crash_log.md` — crash table with guest PC, structural cause, fix
- `docs/function_ledger.md` — human-readable ledger
- `logs/cdb_trace.txt` — CDB hit/miss trace for Track B
- `logs/n64-debug-mcp_trace.txt` — runtime trace for Track B

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (format: "X causes Y, fix with Z").
2. **UPDATE** — phase, track, blocker, crash table, verified commands, evidence paths.
3. **VERIFY** — read back state file for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
