---
name: ps2recomp
description: Statically recompile PlayStation 2 EE/IOP executables through ELF/ISO identity, R5900 translation, generated code, configuration, syscall and hardware runtime bring-up, and reference comparison. Use for PS2 recompilation build, boot, indirect-call, or runtime divergences; do not use for other PlayStation generations, ordinary emulation, or copyrighted artifact acquisition.
compatibility: Requires lawfully obtained target artifacts and the selected project's documented PS2 recompiler, generated-build, runtime, and comparison contracts.
---
# PlayStation 2 static recompilation

Identify the exact title/build and executable before changing code: hashes, ELF/program headers, EE and IOP components/modules, entry points, relocations, load addresses, configuration, recompiler/runtime revisions, build graph, generated-output policy, and first reproducible failure. Never acquire, distribute, or commit ISOs, BIOS dumps, SDK material, executables, keys, or copyrighted assets.

Inspect the current project and tool documentation/help before issuing commands. Do not assume Legacy filenames, TOML keys, compiler, build directory, generated-file count, build duration, runner layout, or host backend.

Diagnose by owning layer:

1. **Input/loader:** ISO filesystem and boot metadata, ELF segments, modules, relocations, address spaces, and initialization order.
2. **Translation:** R5900/MIPS control flow, delay slots, 128-bit GPR/MMI behavior, COP/FPU/VU interactions, memory access, and indirect targets.
3. **Configuration/generated output:** durable discovery, stubs, patches, exclusions, and regeneration. Do not make generated C/C++ the primary permanent fix.
4. **Runtime:** observed syscalls, EE/IOP coordination, threads/synchronization, DMA/VIF/GIF/GS, CD/DVD, SPU2, input, timing, and errors. Implement the smallest evidenced contract; do not turn unresolved symbols into unconditional success stubs.
5. **Host build/execution:** environment, ABI, stack, exceptions, graphics/audio integration, and process lifecycle.

Preserve expensive build artifacts and avoid clean/full rebuilds unless current project evidence requires one and the user accepts the cost. Header changes are not universally forbidden; estimate the actual dependency fan-out and choose the correct durable interface rather than hiding necessary declarations.

At a crash in generated code, map the host frame back to guest address, function/config identity, input state, and owning layer. For an indirect target, verify loaded module, relocation, call-site provenance, and corruption before registering or overriding it.

Reference-emulator comparison is bounded evidence: record emulator/build, title hash, state/input, guest address space, stop condition, register widths, memory ranges, and corresponding recomp state. Compare the fields relevant to the contract; no arbitrary number of breakpoints proves the whole runtime. Stage G assumes no PCSX2 MCP bridge.

After repeated identical failures, stop and gather a different evidence class—raw guest disassembly, ELF/relocation/module map, coherent runtime snapshot, generated dispatch, host trace, or syscall caller—before editing again. Record verified commands, build/incremental cost, before/after observations, changed owning layer, unsupported surfaces, and next falsifiable step. Never install tools or run project scripts merely because this skill was activated.
