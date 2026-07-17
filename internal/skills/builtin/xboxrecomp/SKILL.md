---
name: xboxrecomp
description: Statically recompile original Xbox XBE executables through x86 discovery/lifting, generated code, kernel imports, indirect-call and vtable analysis, and D3D8/NV2A/APU/runtime bring-up. Use for OG Xbox XBE recompilation build, boot, ICALL, kernel, graphics, audio, or game-runtime divergences; do not use for Xbox 360, generic Win32 binaries, emulator setup, or protected-content acquisition.
compatibility: Requires lawfully obtained target artifacts and the selected project's documented XBE loader, x86 recompiler, generated-build, runtime, and validation contracts.
---
# Original Xbox static recompilation

Establish artifact and tool identity first: XBE hash, certificate/title/build, entry encoding as resolved by the verified loader, sections and virtual ranges, TLS, imports/ordinals and libraries, relocations, configuration, recompiler/runtime revisions, generated-output policy, build graph, and first reproducible failure. Never acquire, distribute, or commit disc images, XBEs, keys, SDK leaks, or copyrighted assets.

Inspect current tool documentation, help/version output, schemas, runtime source, and project build files. Do not assume Legacy Python modules/flags, generated layout, compilers, map format, kernel counts, address ranges, host graphics/audio backends, or known upstream gaps.

Diagnose by owning layer:

1. **XBE loader:** certificate, entry/thunk decoding, sections, imports/ordinals, TLS, memory protection, resources, and guest virtual-address mapping.
2. **x86 discovery/translation:** code/data/function boundaries, calling conventions, stack/flags/x87/SSE, exceptions, self-modification, jump tables, callbacks, thunks, vtables, and indirect calls.
3. **Generated output/configuration:** correct loader/discovery/configuration/translator ownership and regenerate; do not make generated C the primary permanent patch.
4. **Runtime services:** implement evidenced kernel, filesystem, threads/synchronization, time, memory, D3D8/NV2A, DirectSound/APU, input, networking, saves, failures, callbacks, and cleanup.
5. **Host execution:** ABI, guest-memory translation, stack/exception handling, process lifecycle, presentation, audio pacing, and diagnostics.

Never cast a guest address to a host pointer. At an ICALL/vtable failure, retain the bounded caller/target trace and build/map identity, then classify the target as verified code, import/thunk, data/vtable, unloaded or relocated content, or corruption. Validate object initialization and call-site provenance before extending dispatch or overriding a function; do not use address ranges alone as proof.

For lifted middleware such as RenderWare, determine whether behavior is translated game code, imported library code, or runtime replacement from target evidence. Overrides require proven identity, signature, state/lifetime, and an explicit durable registration boundary.

A build, linked stub set, or clean ICALL trace proves only that surface. Compare the same scenario against versioned emulator/hardware evidence and native host traces, then validate kernel failures, graphics/state/shaders, audio/input, timing, saves, and cleanup separately.

After repeated identical failures, gather a new XBE/import, raw x86/xref, object/vtable initialization, generated dispatch, guest-state, native trace, or GPU/APU capture before editing again. Record verified commands, tool/build identities, before/after results, changed layer, unsupported services, and next falsifiable step. Stage G assumes no MCP and never installs or runs tools merely because this skill was activated.
