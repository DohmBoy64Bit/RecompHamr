---
name: xbox360-decomp
description: Analyze, matching-decompile, or statically recompile Xbox 360 XEX and lawfully obtained STFS/ISO content across Xenon PowerPC discovery, generated code, imports, indirect control flow, runtime stubs, and Xenos bring-up. Use for ReXGlue-, XenonRecomp-, matching-, or extraction-led Xbox 360 projects; do not use for original Xbox, generic PowerPC work, or protected-content/key acquisition.
compatibility: Requires lawfully obtained target artifacts and the selected project's documented extraction, XEX loader, recompiler, generated-build, runtime, and validation contracts.
---
# Xbox 360 decompilation and static recompilation

Choose and record the actual track from the workspace and user goal: matching PPC reconstruction, ReXGlue-family recompilation, XenonRecomp-family recompilation, or lawful package extraction feeding one of those tracks. Do not default by repository name or Legacy preference.

Establish artifact identity and provenance: package/container type, title/build hashes, XEX headers/security state, image base and sections from the actual loader, imports/exports, relocations, TLS, modules/overlays, entry points, configuration, tool/runtime revisions, generated-output policy, and first reproducible failure. Never acquire, redistribute, decrypt with unauthorized material, or commit retail packages/XEX files, keys, SDK leaks, or copyrighted assets.

Inspect current tool documentation, help/version output, schemas, runtime source, and build graph. Do not assume `rexglue` commands, TOML fields, extraction scripts, generated directories, SDK patches, compiler/backend, or image base.

Diagnose by owning layer:

1. **Container/XEX loader:** STFS/ISO filesystem, compression/encryption state, sections, relocations, imports, module identity, and guest address translation.
2. **Xenon translation:** big-endian PPC control flow, ABI/TOC, VMX128, atomics/reservations, function boundaries, branch tables, indirect calls, and data/code separation.
3. **Configuration/generated code:** correct discovery/configuration/translator ownership and regenerate; generated C/C++ is not the primary permanent patch.
4. **Runtime:** implement evidenced kernel/XAM/XAudio/file/thread/time/memory/network services with signatures, side effects, failures, callbacks, synchronization, and cleanup. Never cast guest virtual addresses to host pointers.
5. **Xenos/host:** command/state/shaders, tiling/endian formats, resolve/MSAA/ROV behavior, presentation/timing, input, audio, saves, and host process lifecycle.

At an unregistered address or switch-table failure, verify caller, target provenance, section/module load state, table bounds/encoding, relocations, and corruption before adding a function or table entry. A half-speed or graphics symptom does not authorize a remembered timing/SDK patch; reproduce, identify the owning clock/presentation/state contract, inspect the exact patch and version, and test side effects.

For matching, record compiler/linker identity and objective unit scope. For native runtime validation, pair guest-address evidence with host traces and same-scenario observations. A build or first boot proves neither semantic parity nor stable graphics/audio.

After repeated identical failures, gather a new raw PPC, XEX/relocation/import, generated dispatch, guest-state, host-trace, or graphics-capture evidence class. Record verified commands, track/tool/build identities, before/after results, changed layer, remaining unsupported surfaces, and next falsifiable step. Stage G assumes no MCP and never installs or runs external tools merely because this skill was activated.
