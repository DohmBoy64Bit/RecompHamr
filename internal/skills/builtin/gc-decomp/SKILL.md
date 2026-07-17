---
name: gc-decomp
description: Analyze, recompile, and bring up Nintendo GameCube DOL or REL code across Gekko PowerPC lifting, generated builds, SDK/runtime replacements, GX rendering, audio, input, and disc assets. Use for GameCube static recompilation and evidence-based emulator comparison; do not use for Wii/other consoles, ordinary Dolphin setup, or obtaining copyrighted game data.
compatibility: Requires legally obtained target artifacts and the selected GameCube recompiler/runtime project's documented toolchain; host graphics and audio backends are project-specific.
---
# GameCube static recompilation

Work in layers: input identity and address map; DOL/REL discovery; Gekko PowerPC translation; generated build; OS/runtime services; GX, audio, input, and storage; then game behavior. Inspect the project's current documentation, configuration, tool help/version output, generated-code contract, symbol sources, and supported host backends before issuing commands.

Establish a baseline with hashes and revisions, entry point and sections, REL loading rules, symbol-map provenance, build results, first observable failure, logs, and a reference-emulator reproduction. Never acquire, distribute, or commit disc images, executables, keys, SDK material, or extracted assets.

For translation failures, verify address space, section mapping, endianness, branch targets, relocations, paired-single semantics, and indirect-call evidence. Improve the symbol/configuration or translator that owns generated output; do not make generated C/C++ the durable patch.

For runtime bring-up, classify the earliest missing or divergent contract before implementing it:

- OS/service behavior such as allocation, threads, timing, VI, DVD, or card storage;
- GX command/state/texture/shader behavior;
- DSP/audio decoding and mixing;
- controller structure and mapping;
- disc filesystem, compression, or archive handling.

Do not assume the Legacy implementation's D3D11, XAudio2, XInput, memory sizes, timing constants, voice counts, or supported formats describe the current project. Derive each relied-on contract from target code, authoritative hardware documentation, lawful clean-room observations, and the selected runtime source/tests.

Use a reference emulator only as an observation oracle. Record emulator version, configuration, input sequence, frame/timing point, and the exact bounded state or image compared. A screenshot proves presentation at one point; a state capture proves only the fields and moment captured. Resolve contradictions before claiming parity.

Treat third-party emulator and decompilation source as license-constrained reference material. Record provenance and license; do not copy incompatible implementation code. Prefer independently derived interfaces, tests, and minimal behavior supported by target evidence.

After each change, rebuild the owning translator/runtime and regenerated project, rerun the same reproduction, and compare the same observable. Record source/target address mapping, evidence classification, changed layer, before/after result, remaining uncertainty, and the next falsifiable step. Never clone repositories, install packages, start an emulator, or run capture scripts merely because this skill was activated.
