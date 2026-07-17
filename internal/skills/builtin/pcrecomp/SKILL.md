---
name: pcrecomp
description: Analyze and statically recompile legacy DOS, Win16/NE, or native Win32 x86 binaries through PE/NE/MZ loading, relocations, x86 EFLAGS/ABI lifting, generated builds, and compatibility shims. Use for PC binary-to-native pipelines, wrong lifted CPU behavior, bad PE address mapping, or generated-code failures; do not use for managed applications, console targets, ordinary source ports, or acquiring protected executables.
compatibility: Requires a lawfully obtained target and the selected project's documented analyzer, lifter, compiler, generated-code, and compatibility-runtime contracts.
---
# Legacy PC static recompilation

Identify the target format and execution model before choosing a pipeline: DOS MZ/16-bit, Win16 NE, native Win32 PE, or another format. Managed CLR, modern engine projects, packed/protected binaries, and console executables require different workflows. Record file hash, headers/sections/segments, entry points, imports/interrupts, relocations, bitness, memory model, tool revisions, and the first reproducible failure. Never acquire, distribute, bypass protection on, or commit proprietary executables or assets.

Inspect the selected recompiler's current documentation, help/version output, schemas, generated-output policy, runtime templates, and build system. Do not assume Legacy PCRECOMP-Next paths, flags, Python dependencies, file names, compiler hints, or supported formats.

Work in owned layers:

1. **Container/loader:** validate executable structure, sections or segments, relocations, overlays, imports, resources, entry and termination behavior.
2. **Discovery:** build conservative code/data/function/xref/indirect-control-flow evidence. Distinguish callable functions, thunks/imports, runtime/library code, tables, embedded data, and unknown ranges.
3. **Translation:** lift a bounded representative slice with explicit CPU, flags, stack, calling convention, memory, self-modification, and indirect-call semantics.
4. **Compatibility runtime:** implement only target-observed DOS/Win16/Win32 services and imports with documented ownership, side effects, error behavior, timing, filesystem, graphics, audio, and input contracts.
5. **Generated build:** regenerate from metadata/translator changes, build with the project toolchain, and verify runtime behavior. Generated code is not the durable first fix, but a translator bug may require changing the translator rather than metadata.

Do not infer correctness from compilation. Validate lifted instruction/flag behavior with focused vectors, compare call/return and memory effects, exercise imports and failures, and run the same target scenario against lawful reference evidence. Classify failures as loader, boundary/data, decoder, lifter, ABI, runtime shim, build, or unsupported dynamic behavior before editing.

If asked to declare a lifted executable correct merely because generated C compiles, refuse the overclaim. Report only the generated-build result; keep CPU and EFLAGS semantics, ABI and stack behavior, memory effects, imports and compatibility failures, indirect control flow, and same-scenario runtime behavior unverified until their focused evidence passes.

Expand coverage incrementally according to the project's risk and dependency graph; “never lift everything” is not a universal rule, but a whole-image attempt without discovery and diagnostics hides ownership. After repeated identical failures, gather new raw bytes/disassembly, relocation, call-site, trace, or ABI evidence before retrying.

Record target/tool/build identities, verified commands, discovered and unknown coverage, generated provenance, compatibility contracts, before/after observations, changed owning layer, and remaining unsupported behavior. Stage G assumes no MCP tools and never installs dependencies or executes external scripts merely because this skill was activated.
