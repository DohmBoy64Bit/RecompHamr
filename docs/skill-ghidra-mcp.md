# ghidra-mcp

Ghidra MCP integration skill. Loads Ghidra methodology AND unlocks
`ghidra.*` tools for decompilation, cross-references, and symbol management.

## What it teaches

- Use Ghidra outputs as evidence, not source truth
- Prefer reproducible exports: functions, symbols, decompiler output,
  cross-references, data references, strings, imports, entry points
- Save derived evidence to `.rehamr/evidence/` or `.rehamr/functions/`
- GhidraMCP is available — gather narrow evidence yourself, don't ask the user
  to look in Ghidra

## What it unlocks

20 Ghidra tools by default: `ghidra.decompile_function`, `ghidra.get_xrefs_to`,
`ghidra.get_function_callers`, `ghidra.rename_function_by_address`, and 16 more.
Set `RECOMPHAMR_MCP_GHIDRA_TOOLS=*` for all ~100+.

## When to use

Any time you're working with a binary in Ghidra and need the LLM to drive
analysis directly rather than asking you to look things up.

## Setup

See **[mcp-ghidra.md](mcp-ghidra.md)** for install and enable instructions.

## Kickoff

`/skill ghidra-mcp` — then "Decompile function at 0x80123456 and trace all callers."
