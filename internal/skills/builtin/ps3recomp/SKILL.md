---
name: ps3recomp
description: Statically recompile PlayStation 3 PPU/SPU executables through lawful SELF/ELF identity, PowerPC/SPU discovery and lifting, NID/HLE runtime contracts, generated builds, RSX bring-up, and reference comparison. Use for PS3 recompilation build, boot, import, translation, or runtime divergences; do not use for other platforms, ordinary RPCS3 debugging, or decryption/acquisition without lawful user-provided material.
compatibility: Requires lawfully obtained target artifacts and the selected project's documented PS3 loader, recompiler, HLE runtime, generated-build, and comparison contracts.
---
# PlayStation 3 static recompilation

Establish lawful artifact and build identity first: title/build hashes, SELF versus ELF state, program/section headers, PPU and SPU images, relocations, entry points, imports/exports and NIDs, configuration, recompiler/runtime revisions, generated-output policy, and first reproducible failure. Never request, distribute, commit, or help obtain retail executables, keys, firmware, SDK leaks, or copyrighted assets. Do not claim a decryption command exists; use only a verified project-supported path with material the user is authorized to use.

Inspect the current tool documentation, help/version output, schemas, runtime modules, and build graph. Do not assume Legacy script names, TOML fields, compiler/generator, generated-file layout, backend, or module-status table.

Diagnose by owning layer:

1. **Container/loader:** SELF/ELF validation, segments, relocations, TLS, PRX/SPRX modules, import/export tables, and address-space mapping.
2. **Discovery/translation:** big-endian PPU PowerPC control flow, ABI and TOC, VMX/FP semantics, atomics/reservations, callbacks and indirect calls; separately identify SPU images, local-store/DMA/mailbox behavior, and unsupported dynamic code.
3. **Generated output/configuration:** make durable changes in discovery, configuration, or translator/runtime ownership and regenerate. Generated C/C++ is diagnostic output, not the primary permanent patch.
4. **HLE/runtime:** resolve each NID to an evidenced module, signature, ABI, side effects, return/error behavior, synchronization, filesystem/network/user-data boundary, and callback lifecycle. An unresolved NID is not evidence that a no-op or success stub is correct.
5. **Host/RSX:** host ABI, trampoline/stack lifecycle, threads, memory, graphics command/state/shaders, audio, input, timing, and cleanup. Backend choice is project-specific.

At a generated-code crash, map host frame to guest address/function/module and distinguish translation, bad signature/TOC, HLE state, unsupported SPU behavior, or corruption. For trampoline or indirect-call failures, verify ownership and lifetime before adding dispatch or stack-drain behavior.

Reference-emulator comparison is bounded: record RPCS3/build/configuration, title hash, state/input, guest address space, PPU/SPU context, stop condition, memory/register fields, and matching recomp point. No arbitrary count of matching addresses proves global parity, and Stage G assumes no PINE MCP bridge.

After repeated identical failures, gather a new evidence class—raw PPU/SPU disassembly, relocation/module/NID caller evidence, coherent runtime snapshot, generated dispatch, host trace, or RSX capture—before editing again. Record verified commands, implemented versus stubbed/unsupported NIDs, before/after observations, changed layer, remaining uncertainty, and next falsifiable step. Never install tools, clone projects, or execute scripts merely because this skill was activated.
