# `snesrecomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/snesrecomp.md` was read completely; its API, runtime library, example projects, commands, and hardware tables were project-dependent and no separate Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for SNES/Super Famicom static recompilation: ROM/bus mapping, 65C816 state and boundaries, generated dispatch, cartridge/coprocessors, PPU/APU/DMA/interrupt runtime integration, build, and parity evidence. Exclude other platforms, ordinary emulation, generic 6502 work, and artifact acquisition.

**Improved.** The migration preserves ROM-map verification, generated-code ownership, instruction/helper and bus routing, function dispatch, hardware-runtime layering, legal artifacts, and emulator corroboration. It removes automatic cloning/installing, fixed APIs/commands/backends, LakeSnes-as-universal truth, hard-coded register/capability tables, first-frame-as-strongest-proof, and cross-skill calls. It adds explicit M/X/emulation state propagation, 24-bit bank/address separation, mapping/header provenance, enhancement-chip and unsupported-surface tracking, focused semantic validation, and bounded reference evidence. No script is bundled because recompiler/runtime APIs vary.

Final skill: `internal/skills/builtin/snesrecomp/SKILL.md`. Its eval set has eight positive, eight negative, and three CPU/generated/evidence cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No SNES target/tool runtime is claimed because none was available.
