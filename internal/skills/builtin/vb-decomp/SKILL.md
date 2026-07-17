---
name: vb-decomp
description: Statically recompile Nintendo Virtual Boy V810 ROMs through function discovery, generated code, hints/interception, cartridge/runtime integration, VIP/VSU/interrupt bring-up, and corpus regression analysis. Use for Virtual Boy recompilation build, boot, rendering, audio, interrupt, or cross-title regressions; do not use for other consoles, ordinary emulation, or ROM acquisition.
compatibility: Requires lawfully obtained ROMs and the selected project's documented V810 recompiler, generated-build, runtime, corpus, and comparison contracts.
---
# Virtual Boy static recompilation

Establish exact ROM and tool identity: hash, header/mapping and vectors, entry state, recompiler/runtime revisions, hints/configuration, generated-output policy, corpus manifest, build graph, and first reproducible failure. Never acquire, distribute, or commit ROMs, BIOS material, or copyrighted assets.

Inspect the current project's documentation, schemas, helper APIs, runtime, and tool help/version output. Do not assume Legacy `v810recomp` commands, file layout, corpus size, SDL/GUI backend, hardware addresses, channel counts, or known-bug fixes apply.

Work by owning layer:

1. **Discovery/translation:** prove V810 functions and data from entry/vectors, direct and indirect control flow, references, alignment, instruction semantics, flags/PSW, exceptions, and mirror/address mapping.
2. **Hints/interception:** use names, overrides, or HLE only when the selected tool documents them and target evidence proves the function identity, signature, side effects, and replacement boundary.
3. **Generated output:** fix discovery, configuration, translator, or runtime ownership and regenerate; do not make generated C the primary permanent patch.
4. **Runtime:** validate CPU state, memory/cart mapping, VIP display/timing, VSU audio, timers, input, interrupts/priorities, saves, frame pacing, and cleanup as independent contracts.
5. **Corpus:** pin the lawful manifest and tool/build configuration, isolate each title, bound time/output, distinguish compile/boot/render/gameplay states, and retain regressions rather than reporting only aggregate improvement.

A build, screenshot, or a few matching registers proves only its observed surface. Align a versioned reference emulator or hardware observation by ROM, state/input, frame/time, address space, and fields; Stage G assumes no emulator MCP bridge. Cross-check function tables against raw V810 evidence and treat analyzer/decompiler output as hypotheses.

When a proposed fix appears to help many titles, add focused regression evidence for its instruction/control-flow contract plus the corpus sweep. Do not encode a symptom-specific rule such as mirror normalization, phase cycling, interrupt comparison, or missing-handler fallback without reproducing and falsifying the underlying cause.

After repeated identical failure, gather a new raw-disassembly, vector/mapping, trace, generated-dispatch, coherent runtime, or corpus-differential evidence class before editing again. Record verified commands, corpus/config identity, per-title statuses, before/after observations, changed layer, remaining unsupported features, and next falsifiable step. Never install or run tools merely because this skill was activated.
