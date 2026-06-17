# function-discovery

Use this skill when building a function inventory for a binary — identifying
entry points, classifying functions, discovering jump tables and vtables, and
preparing data for `symbol_addrs`, recompiler metadata, or matching work.

## Goal

Build a high-confidence function ledger with evidence-backed classification —
separate game/project logic from runtime, platform, middleware, import thunks,
data, and unknown code. Feed the ledger into symbol lists, recompiler configs,
and function matching workflows.

## Rules

1. **Start from the strongest evidence.** PDB/MAP/export symbols, ELF symbols,
   PE `.pdata`, XEX metadata, ROM/splat config, loader metadata — any
   structured symbol data takes priority over disassembly-only guesses.
2. **Record all entry points first.** Process entry, TLS callbacks, static
   constructors/destructors, init arrays, reset/boot entry, thread starts.
3. **Trace direct calls before indirect.** Direct call targets establish
   function boundaries. Indirect calls (jump tables, vtables, switch
   dispatchers) are classified afterward with additional evidence.
4. **Classify with evidence, not instinct.** Every classification needs a
   source: xref pattern, string reference, import match, SDK fingerprint,
   or behavioral analysis.
5. **Keep unknown as unknown.** UNKNOWN is a valid classification. Do not
   promote to a stronger label without new evidence.
6. **Populate the CSV before bulk symbol imports.** The function ledger is
   the gate for `symbol_addrs` or recompiler metadata — no bulk naming
   without a classified inventory.

## Workflow

1. **Extract strongest evidence.** Parse PDB, MAP, ELF symbols, PE metadata,
   ROM config, or XEX metadata for known symbols and boundaries.
2. **Record entry points.** Process entry, TLS callbacks, constructors,
   destructors, init arrays, reset/boot entry, thread starts.
3. **Trace direct calls and cross-references.** Use `ghidra.get_xrefs_to`,
   `ghidra.get_function_callers`, `ghidra.get_function_callees` for each
   identified function.
4. **Identify jump tables, switch dispatchers, vtables.** Record indirect
   call patterns — classify only after confirming dispatch structure.
5. **Classify every function.** Use the classification taxonomy below with
   evidence sources recorded per entry.
6. **Populate the CSV ledger.** Write to `.rehamr/functions/inventory.csv`
   before using the data for symbol lists or recompiler metadata.

## Classification Taxonomy

| Classification | Description | Typical Evidence |
|---|---|---|
| `game_logic` | Game/project-specific logic | String refs to game data, unique behavior |
| `runtime_platform` | OS, kernel, HAL, hardware access | Syscall patterns, hardware register access |
| `middleware_library` | Third-party engine/SDK code | SDK fingerprint, license strings, known patterns |
| `import_thunk` | Wrapper around external import | Single jump to import table entry |
| `data_jump_table` | Jump table or switch dispatcher | Sequential address table, indirect jump pattern |
| `data_vtable` | Virtual method table | Vtable-like dispatch, object pattern |
| `data_other` | Non-code data misidentified as code | No valid prologue, repetitive bytes, string-like |
| `unknown` | Insufficient evidence to classify | — |

## Required Output

```csv
address_or_symbol,name,status,classification,evidence_source,confidence,notes
0x00401000,EntryPoint,CONFIRMED,game_logic,PE entry point + disasm,high,main process entry
0x00401100,sub_401100,CONFIRMED,runtime_platform,PDB symbol + xref to kernel,high,thread init
0x00412000,sub_412000,HYPOTHESIS,middleware_library,string ref "FMOD",medium,needs SDK fingerprint confirmation
```

**CSV columns:** `address_or_symbol`, `name`, `status` (CONFIRMED/HYPOTHESIS/TODO),
`classification` (see taxonomy), `evidence_source`, `confidence` (high/medium/low),
`notes`

## Evidence / Artifact Targets

Required:
- `.rehamr/functions/inventory.csv` — classified function ledger (primary output)
- `.rehamr/functions/unknown.md` — functions awaiting classification evidence

Optional:
- `.rehamr/functions/game_logic.md` — game/project logic function notes
- `.rehamr/functions/runtime_platform.md` — runtime/platform function notes
- `.rehamr/evidence/function_discovery.log` — discovery session log
- `REPHAMR_STATE.md` — function count, classification progress, active unknowns

## Stop Conditions

Stop and gather better evidence when:
- You are about to classify a function without a cited evidence source
- You are about to promote an UNKNOWN to a stronger label without new evidence
- A function boundary is contested (different tools disagree) — document the
  disagreement with offsets from each source
- You are about to do bulk symbol renaming without a populated CSV ledger

## Session Close

1. Update `.rehamr/functions/inventory.csv` with new classifications.
2. Update `REPHAMR_STATE.md` with function count, classification progress,
   and confidence distribution.
3. Report: new classifications, promoted/demoted entries, remaining unknowns,
   next 3 functions to investigate with expected evidence.
