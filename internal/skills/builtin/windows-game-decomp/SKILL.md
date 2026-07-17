---
name: windows-game-decomp
description: Classify and decompile Windows games across native PE, managed .NET/Mono, Unity IL2CPP, Unreal, DOS/Win16, matching, behavioral-port, compatibility-shim, and modding-SDK tracks. Use when reconstructing a Windows game's code, toolchain, ABI, engine metadata, or clean vanilla/mod boundary; do not use for console targets, generic PE inspection, or PC static recompilation already owned by a recompiler pipeline.
compatibility: Requires lawfully obtained target artifacts and runtime-family-specific current tools; exact matching additionally requires the project's pinned compiler/linker/build and objective comparison contract.
---
# Windows game decompilation

Classify both runtime family and intended result before choosing tools:

- native PE or DOS/Win16 code reconstruction and optional exact matching;
- managed .NET or Unity Mono IL and metadata recovery;
- Unity IL2CPP native code plus matching metadata and registration evidence;
- Unreal native code/reflection/assets with versioned engine evidence;
- behavioral port or compatibility shim where byte identity is not the goal;
- modding SDK/plugin ABI kept separate from a clean vanilla baseline.

Record hashes, PE/container and architecture, CLR/native/engine evidence, modules/resources, symbols, protection/packing status, compiler/linker/runtime hints, target track, accepted match/parity definition, and first reproducible question. Never acquire, distribute, bypass protection on, or commit retail binaries, keys, proprietary assets, symbols, or leaked SDK material.

Use the representation that owns the behavior. For managed assemblies, inspect IL/metadata before native stubs. For IL2CPP, pair the exact native binary with compatible metadata/registration evidence; generated dummy assemblies are aids, not original source. For Unreal, reflection/SDK dumps are structural aids, not complete source. For native code, prove function/data boundaries and ABI from raw instructions, relocations, xrefs, unwind/symbol/import evidence, and runtime observations.

For exact matching, establish compiler, linker, CRT/SDK, flags, headers, libraries, sections, relocations, and target provenance from evidence. A unit comparison proves only that unit under the recorded normalization; percentages must expose missing, excluded, failed, and unverified surfaces. Scratch compilers and decompilers generate hypotheses, not final proof.

For behavioral ports and compatibility shims, define API/ABI, state, threading, timing, filesystem/registry, graphics, audio, input, networking, failure, and cleanup contracts with focused conformance and scenario tests. Do not call functional parity an exact match.

Keep vanilla reconstruction, behavioral adaptation, and mod/plugin hooks in explicit independently buildable boundaries. Do not contaminate the artifact used for matching with instrumentation or enhancements; when a hook is necessary for observation, document and remove or isolate it from acceptance evidence.

After repeated matching or classification failure, gather a new raw-disassembly, compiler-output, metadata/registration, xref, trace, or ABI evidence class. Record verified tools/commands and versions, function/runtime provenance, match or parity scope, before/after result, remaining unknowns, and next falsifiable step. Stage G assumes no MCP integrations and never installs or runs tools merely because this skill was activated.
