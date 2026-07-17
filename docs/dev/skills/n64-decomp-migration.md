# `n64-decomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/n64-decomp.md` was read completely; its referenced tools and artifacts were project-dependent and no separate Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for N64 matching-decompilation and N64Recomp tracks: ROM/split identity, MIPS boundaries and delay slots, matching, overlays/relocations, generated code, indirect targets, guest/host address separation, and runtime bring-up. Exclude other platforms, ordinary emulation, generic MIPS work, and ROM acquisition.

**Improved.** The migration preserves the explicit dual-track model, layer diagnosis, metadata-owned generated output, boundary evidence, compiler/diff discipline, overlay and `jalr` reasoning, guest-pointer prohibition, legal boundary, and evidence handoff. It removes automatic dependency setup, `/doctor` and cross-skill calls, hard-coded commands/MCP tools, backend and file-layout assumptions, fixed success thresholds, and universal symptom fixes. It adds address/provenance identity, relocation/load-order and corruption alternatives, bounded emulator evidence, same-observable verification, and a repeated-failure stop rule. No script is bundled because project toolchains and runtime contracts vary; MCP debugging remains Stage H.

Final skill: `internal/skills/builtin/n64-decomp/SKILL.md`. Its eval set has eight positive, eight negative, and three boundary/evidence cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No N64 target/tool runtime is claimed because none was available.
