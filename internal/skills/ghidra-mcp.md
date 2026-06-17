# ghidra-mcp

Use this skill when Ghidra is needed for static binary analysis — decompiling
functions, tracing cross-references, renaming symbols, searching strings and
imports, and exporting evidence from a loaded binary.

> Ghidra output is evidence, but interpretation still needs classification.
> Decompiler output is not source truth — it is a tool-derived hypothesis
> unless confirmed by symbols, behavior, or matching code. The 20 most-used
> RE tools are enabled by default.

## What it enables

- Decompile functions with `ghidra.decompile_function` (hint only — never final proof)
- Trace callers and callees via `ghidra.get_xrefs_to`, `ghidra.get_function_callers`
- Rename functions and labels after evidence-backed classification
- Search strings, list imports/exports, analyze function completeness
- Batch decompile with `ghidra.decompile_all` and function statistics
- Export evidence to `.rehamr/evidence/` or `.rehamr/functions/`

## When to use

Use this tool for:
- Static analysis of any loaded binary (native PE, PPC, MIPS, ARM, x86)
- Tracing cross-references before renaming or reclassifying
- Batch function analysis for inventory building
- Exporting decompiler output, symbols, strings, and imports as evidence

Do not use it for:
- Runtime debugging (use `cdb-debug`, `pcsx2`, `mcp-pine`, `n64-debug-mcp`,
  or `bizhawk` per platform)
- Managed/.NET code where ILSpy/dnSpy is more appropriate
- Final boundary proof — raw disassembly and delay slots take priority over
  decompiler output

## Boot / Connection Check

1. Verify Ghidra is reachable:
   - `/mcp tools ghidra` — lists 20 enabled tools by default
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Verify the correct binary is loaded:
   - `ghidra.get_entry_points` — confirms the target program is active
   - Confirm program name matches the expected target
3. Verify MIPS N64 / PPC / x86 architecture is correctly detected if
   working with platform-specific binary (N64LoaderWV, XEXLoaderWV, etc.)
4. If unavailable:
   - Check `/mcp status ghidra` for connection state
   - Verify `ghidra-mcp` is on PATH or `RECOMPHAMR_MCP_GHIDRA_COMMAND` is set
   - Provide exact setup steps from `docs/mcp-ghidra.md`

## Setup

1. Install: `pip install ghidra-mcp` or via [REPlugins](https://github.com/DohmBoy64Bit/REPlugins)
2. Ghidra 12.1.2 with a project open and binary loaded
3. Ensure `ghidra-mcp` is on PATH (or set `RECOMPHAMR_MCP_GHIDRA_COMMAND`)
4. Start recomphamr — connect with `/mcp connect ghidra`
5. Load `/skill ghidra-mcp` — unlocks `ghidra.*` tools (20 by default)
6. Verify: `/mcp tools ghidra`
7. For all ~100+ tools: `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` or `/mcp enable ghidra *`

## Evidence Protocol

Every Ghidra output should be classified:
- Raw disassembly + xrefs → CONFIRMED (strongest static evidence)
- Decompiler output → HYPOTHESIS (tool-derived, not final proof)
- Symbol/import/string references → CONFIRMED (if verified against disassembly)
- Function boundaries from decompiler → HYPOTHESIS (confirm with raw disasm + delay slots)

Save evidence to:
- `.rehamr/evidence/` — decompiler output, xref summaries, string lists
- `.rehamr/functions/` — function inventory, classification notes
- `REPHAMR_STATE.md` — active Ghidra session, program name, architecture

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Decompile function | `ghidra.decompile_function` | Pseudocode | Hint only — never final boundary proof |
| Trace callers | `ghidra.get_xrefs_to` | Address list | Use before renaming |
| Trace callees | `ghidra.get_function_callers` | Address list | Understand call graph |
| Full function dump | `ghidra.analyze_function_complete` | xrefs, callees, callers, vars | Before classifying |
| Rename function | `ghidra.rename_function_by_address` | Confirmation | Only after evidence packet complete |
| Search strings | `ghidra.search_strings` | String matches | Find OS/engine version strings |
| List imports | `ghidra.list_imports` | Import table | External dependencies |
| Read raw bytes | `ghidra.read_memory` | Hex dump | Verify at crash or boundary address |
| Batch decompile | `ghidra.decompile_all` | Full program pseudocode | For inventory building |
| Function statistics | `ghidra.function_stats` | Counts + categories | Phase 0 recon summary |

## Guardrails

1. Decompiler output is not source truth — raw MIPS/PPC/x86 + delay slots +
   jump-table proof takes priority.
2. Never ask the user to look at Ghidra for you if GhidraMCP is connected —
   gather narrow evidence yourself via MCP tools.
3. Ghidra output is evidence, but interpretation still needs classification
   (use `/skill evidence-mode`).
4. Never rename functions or symbols without an evidence packet: disassembly,
   xref, string reference, or symbol table entry.
5. Confirm the correct architecture is loaded (MIPS N64 via N64LoaderWV,
   PPC Xenon via XEXLoaderWV, etc.) before trusting decompiler output.
6. Prefer reproducible exports: functions, symbols, decompiler output, cross
   references, data references, strings, imports, and entry points.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `ghidra-mcp` not on PATH | Server binary missing | Install via pip or REPLugins; set `RECOMPHAMR_MCP_GHIDRA_COMMAND` |
| Server Disconnected | Ghidra not running or project not open | Start Ghidra, open project, load binary, reconnect |
| Decompiler timeout | Function too large or complex | Use `disassemble_function` for raw asm; break into smaller ranges |
| Wrong architecture detected | Incorrect loader used | Verify Ghidra project uses correct loader (N64LoaderWV, XEXLoaderWV, etc.) |
| `decompile_function` empty output | Function not recognized or at wrong address | Verify address; use `read_memory` to confirm bytes at that location |

## Session Close

1. Save evidence exports to `.rehamr/evidence/` or `.rehamr/functions/`.
2. Update `REPHAMR_STATE.md` with active Ghidra program name, architecture, and session summary.
3. Report: functions decompiled, xrefs traced, symbols renamed, evidence files created, remaining addresses to investigate.
