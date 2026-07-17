---
name: function-discovery
description: Build an evidence-backed function-boundary and classification inventory for an executable, ROM, firmware image, or static-recompilation target. Use for entry points, direct and indirect call targets, jump tables, vtables, thunks, library code, data mistaken for code, and symbol/recompiler metadata preparation; do not use for source-only call graphs or naming functions without binary evidence.
compatibility: Requires target-format metadata or static/runtime analysis output; no particular disassembler is required.
---
# Function discovery and classification

Identify the target binary and its hash or equivalent immutable identity first. Prefer structured boundary evidence over heuristic disassembly: debug/map/export symbols, unwind or exception metadata, loader records, relocation tables, overlay/segment configuration, and platform executable metadata.

Build the inventory in this order:

1. Record process/boot/reset entries, callbacks, constructors, destructors, init arrays, interrupt handlers, thread starts, and exported entries.
2. Add direct call targets and reconcile overlapping or contested boundaries.
3. Analyze indirect targets only with supporting dispatch evidence: switch tables, function-pointer tables, vtables, callbacks, or computed branches.
4. Separate executable functions from thunks, padding, literal pools, jump-table data, vtables, and other non-code.
5. Classify each row and cite its strongest evidence. Keep insufficiently supported rows `unknown`.

Use these classifications when applicable: `project_logic`, `runtime_platform`, `middleware_library`, `import_thunk`, `data_jump_table`, `data_vtable`, `data_other`, and `unknown`. A semantic name is separate from a boundary: a confirmed start address does not prove behavior. Record tool disagreements rather than choosing one silently.

When initialized, `.rehamr/functions/inventory.csv` is the canonical ledger with columns:

```text
address_or_symbol,name,status,classification,evidence_source,confidence,notes
```

Use `CONFIRMED`, `HYPOTHESIS`, or `TODO` for status; use bounded, precise evidence references. Preserve unknowns and avoid bulk symbol or recompiler-metadata imports until the ledger is reviewable. CSV values containing commas, quotes, or newlines must be encoded correctly.

Finish with target identity, counts by classification/confidence, new or changed boundaries, contested rows, remaining unknowns, and the next evidence-producing queries.
