---
name: gb-recomp
description: Improve or debug static recompilation of Game Boy and Game Boy Color ROMs using recompiler coverage, interpreter-fallback evidence, metadata, and emulator traces. Use for LR35902 code generation, unresolved indirect control flow, trace seeding, runtime fallback, or measured native performance; do not use for GBA/other consoles, ordinary emulation, or ROM acquisition.
compatibility: Requires a legally obtained GB/GBC ROM and the target project's documented recompiler, generated build, runtime, and optional trace-capable emulator.
---
# Game Boy static recompilation

Treat the ROM, recompiler, generated project, runtime, and reference emulator as distinct evidence layers. First inspect the project README, tool help/version output, build files, metadata format, and existing evidence. Do not assume a particular `gbrecomp` fork, flag, trace schema, directory, coverage target, or benchmark command.

Before changing anything:

1. Identify the exact ROM hash without copying, distributing, or committing the ROM, BIOS, or extracted assets.
2. Record the recompiler/runtime revisions, baseline build result, coverage or fallback metric, reproduction path, and active blocker.
3. Determine whether the gap is static discovery, unresolved indirect control flow, executable RAM, mapper/hardware emulation, generated-build failure, or measurement error. `JP HL` is a common indirect site, not a universal diagnosis.

Use this evidence loop:

1. Reproduce with the project's documented commands and retain bounded logs plus `metadata.json` or its current equivalent.
2. Map fallback addresses to ROM/RAM regions and candidate control-flow sites. Separate observed facts from hypotheses.
3. When static discovery lacks targets, capture a legal local emulator trace for the missing path and validate its address space and format before using it as an analysis seed.
4. Regenerate from the recompiler, then build both the recompiler and generated project. Never treat a hand edit to generated C/C++ as the durable fix.
5. Repeat the same scenario and compare the same metrics against the baseline. A trace proves only the exercised path; retain interpreter fallback for unobserved cases.
6. Fix the owning layer: discovery in analysis, mapper/PPU/APU behavior in the runtime, or bad measurement in the harness. Do not mask unknown targets with fabricated dispatch entries.

Make performance claims only from the project's uncapped, repeatable benchmark mode, with build identity, workload, configuration, and multiple measurements recorded. Windowed emulator or native runs may be paced by display or audio and are not sufficient by themselves.

Stop repeated reseeding when the same site does not improve. Recheck trace coverage, bank mapping, executable-RAM behavior, address translation, tool version, and classification before another attempt. Record commands actually verified, before/after metrics, trace provenance, changed source layer, remaining fallbacks, and the next falsifiable step in the evidence workspace.

Never install dependencies, clone repositories, fetch ROMs, or execute project scripts merely because this skill was activated. Use the existing bounded tools only after inspecting the project contract and preserve cancellation, output limits, and source artifacts.
