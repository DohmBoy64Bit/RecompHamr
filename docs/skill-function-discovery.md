# function-discovery

Function identification and classification skill. Load when mapping unknown
binaries — finding entry points, classifying code, and building a function
ledger.

## What it teaches

- Classify code as: game logic, runtime/platform, middleware/library,
  import/thunk, data/jump-table, or unknown
- Classification must be evidence-based — Ghidra decompiler output is a hint,
  not final proof
- Build a function ledger before bulk symbol imports
- Use cross-references and call graphs to confirm boundaries

## When to use

After initial binary analysis, before bulk symbol naming or recompilation
metadata work. Pairs with `ghidra-mcp` for tool access.
