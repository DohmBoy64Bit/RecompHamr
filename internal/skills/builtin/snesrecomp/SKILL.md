---
name: snesrecomp
description: Statically recompile Super Nintendo or Super Famicom 65C816 game code through ROM mapping, function discovery, generated dispatch, CPU-state translation, cartridge/runtime integration, and PPU/APU/DMA validation. Use for SNES recompilation build, boot, mapping, or hardware-runtime divergences; do not use for other consoles, ordinary emulation, or ROM acquisition.
compatibility: Requires a lawfully obtained ROM and the selected project's documented SNES recompiler, CPU helper, cartridge, hardware-runtime, build, and comparison contracts.
---
# SNES static recompilation

Establish target and tool identity first: ROM hash and copier-header status, mapping evidence, region, vectors, cartridge SRAM and enhancement chips, recompiler/runtime revisions, generated-output policy, build graph, and first reproducible failure. Never acquire, distribute, or commit ROMs, BIOS/firmware, or copyrighted assets.

Inspect the current project's documentation and code before using any helper, patch macro, function table, bus API, renderer, or command. Do not assume Legacy `RECOMP_PATCH`, `cpu_ops.h`, LakeSnes, SDL2, address signatures, or supported hardware describe the selected project.

Translate with explicit 65C816 state:

1. Preserve 24-bit program/data addresses and distinguish ROM offset, CPU bus address, bank registers, direct page, stack, and host storage.
2. Track emulation/native mode and M/X width changes across `REP`/`SEP`, interrupts, calls, and returns. Immediate and register widths are control-flow state, not local syntax.
3. Prove function boundaries from vectors, `JSR`/`JSL`/`RTS`/`RTL`/interrupt flow, bank context, references, traces, and embedded-data exclusions.
4. Route memory and I/O through the verified cartridge/bus contract. Never cast bus addresses to host pointers or bypass side effects for convenience.
5. Resolve dynamic/indirect calls and generated dispatch from observed targets and mapping/load state; do not register guessed addresses.

Diagnose failures by owning layer: ROM mapping/header, discovery/state propagation, translator/helper semantics, generated configuration, cartridge coprocessor, PPU, DMA/HDMA, APU/SPC, interrupts/timing, input, or host platform. Fix the owning source and regenerate rather than making generated output the primary permanent patch.

Visible output proves only a render path. Validate focused CPU/flag/addressing vectors, memory/register side effects, interrupts and DMA timing, audio/input, mapping variants, error paths, and the same scenario against a versioned reference emulator or hardware evidence when available. Stage G assumes no emulator MCP bridge.

After repeated identical failures, gather new raw disassembly/state-width, mapping, trace, generated-dispatch, bus-register, or host-runtime evidence before editing again. Record verified commands, function/mapping provenance, before/after observations, changed layer, unsupported chips/features, and next falsifiable step. Never install tools, clone projects, or run scripts merely because this skill was activated.
