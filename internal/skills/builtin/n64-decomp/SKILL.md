---
name: n64-decomp
description: Run Nintendo 64 matching decompilation or N64Recomp static-port work across ROM/splat metadata, MIPS overlays, guest VRAM/ROM/host addresses, RecompiledFuncs and symbol regeneration, jalr targets, and VI/DMA/RSP/RDP runtime failures. Use for N64 ROM projects, matching and split configuration, generated-code ownership, or guest-to-host bring-up; do not use for other platforms, ordinary emulation, or ROM acquisition.
compatibility: Requires a legally obtained target ROM and the selected project's documented split, compiler, matching, recompilation, and runtime toolchain.
---
# Nintendo 64 decompilation and static recompilation

Choose and record the track before changing code:

- **Matching decompilation** targets the project's defined binary/ROM identity with its pinned compiler, linker, split metadata, and objective diff tool.
- **Static recompilation** translates discovered guest code and links it to a host runtime; semantic/runtime parity is distinct from a byte match.

Inspect the project documentation, ROM hash/byte order, entry point, segment and overlay map, build configuration, symbol provenance, generated-output policy, tool help/version output, and current failure. Do not assume filenames, commands, TOML/YAML keys, compilers, renderers, memory sizes, save hardware, or runtime APIs from another project. Never acquire, distribute, or commit ROMs, SDK leaks, keys, or copyrighted extracted assets.

For matching work:

1. Establish a reproducible asm/data baseline from the exact ROM and split configuration.
2. Resolve segment, BSS, relocation, alignment, data/code, and overlay boundaries in metadata rather than permanently patching generated assembly.
3. Prove function boundaries from raw MIPS control flow, delay slots, relocations, references, and compiler behavior; decompiler output is a hypothesis.
4. Compile with the pinned identity and classify objective differences before changing C. Distinguish code, data, relocation, linker, and compiler mismatches.

For static recompilation:

1. Verify guest virtual/physical/ROM address spaces, relocatable sections, overlays, symbols, entry path, and indirect control-flow targets before code generation.
2. Fix configuration, discovery, translator, or runtime ownership rather than hand-editing generated recompiled functions.
3. Never cast a guest address to a host pointer. Use the runtime's verified translation and registration contracts.
4. At a `jalr` or unregistered-target failure, inspect caller, target provenance, loaded overlay identity, relocation/load order, and corruption evidence. Do not register an address solely to silence a crash.
5. Bring up the earliest missing host contract—memory/DMA, timing, threads, saves, RSP/RDP/VI, audio, input, or rendering—and validate the same observable after each change.

Reference-emulator, debugger, and static-tool observations are bounded evidence, not hardware truth. Record tool/core/version, ROM hash, address space, state/input, time/frame, and exact observed fields. Stage G assumes no MCP bridge.

If one emulator frame or observable matches, record only that versioned captured result. Do not call the whole static port accurate; separately retain execution coverage, memory and DMA, timing, threads, saves, RSP/RDP/VI and rendering, audio, input, cleanup, and remaining guest/runtime contracts as unverified until exercised.

After repeated identical failures, stop editing and gather a new class of evidence: raw control flow, overlay/relocation map, guest runtime snapshot, generated dispatch, linker/diff output, or host trace. Record track, build/tool identities, verified commands, before/after evidence, changed owning layer, remaining uncertainty, and the next falsifiable step. Never install dependencies, clone projects, or run scripts merely because this skill was activated.
