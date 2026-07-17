---
name: gen-decomp
description: Split, analyze, and matching-decompile Sega Genesis or Mega Drive ROMs using evidence-backed M68000/Z80 segmentation, compression and asset classification, symbols, VDP interpretation, and emulator comparison. Use for Genesis disassembly/decompilation projects; do not use for other consoles, ordinary emulation, or copyrighted ROM acquisition.
compatibility: Requires a legally obtained target ROM and the selected project's documented splitter, assembler/compiler, comparison, and optional emulator tooling.
---
# Genesis and Mega Drive decompilation

Treat ROM identity, segment configuration, generated disassembly/assets, symbols, matching source, and runtime observations as separate layers. Inspect the project's current tool documentation and help/version output before choosing commands or configuration fields; do not assume Legacy `sega2asm` flags, formats, counts, or integrations still apply.

Start with the ROM hash, region/header evidence, size and mapping, vectors, existing configuration, symbol/charmap provenance, generated-output policy, baseline split/build result, and one reproducible discrepancy. Never acquire, distribute, or commit ROMs or copyrighted extracted assets.

Build the segment map from evidence rather than guessed boundaries:

1. Identify headers/vectors, M68000 and Z80 execution regions, tables, text, graphics, audio, padding, and unknown ranges.
2. Use the selected splitter's non-mutating plan or validation mode when it has one. Record exact diagnostics and byte ranges.
3. Classify compression from signatures, callers/decompressors, known test vectors, and verified tool output. Treat a detector result as a hypothesis until extraction and consumption agree.
4. Fix source configuration, symbols, charmaps, or the splitter—not regenerated assembly/assets—then rerun deterministically.

If asked to repair a bad label or boundary directly in generated assembly, refuse the generated-file edit. Trace the output back to its symbol file, segment configuration, charmap, or splitter logic; change the owning source, regenerate, verify deterministic output, and then rerun the applicable match or runtime observation.

For code, track CPU/address space explicitly. Verify M68000 branch/table targets, Z80 sound-driver boundaries, bank mappings, alignment, data embedded in code, and VDP command-table hypotheses. Never invent instruction behavior or VDP register meaning; corroborate with authoritative hardware material, target access patterns, and runtime observations.

Use a reference emulator as a bounded oracle: record emulator/core/version, ROM hash, state/input sequence, frame or instruction point, memory domain, address, width, and endianness. One or three matching addresses do not prove the whole disassembly. Preserve save states and traces only when lawful and sanitized; Stage G does not assume any MCP bridge.

If asked to treat a few matching emulator reads as proof of the entire disassembly, refuse and limit the claim to those versioned observations. Explicitly keep segment boundaries, control-flow and table classification, deterministic generated builds, objective matching, and broader runtime behavior unverified until each is checked.

For matching work, compile with the project's pinned compiler/flags and compare code/data using its objective matcher. Classify mismatches before editing source, preserve function and data ledgers, and separate semantic correctness from byte identity.

After repeated failure at one segment, stop and gather new boundary, compression, caller, or runtime evidence. Record verified commands, configuration changes, generated artifacts, before/after split or match results, runtime observations, remaining unknown ranges, and the next falsifiable step. Do not install tools, launch emulators, or execute project scripts merely because this skill was activated.
