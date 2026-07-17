# `ps2recomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/ps2recomp.md` was read completely; its project layout, tool calls, runtime files, and build claims were project-dependent and no separate Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for PS2 EE/IOP static recompilation: ISO/ELF identity, R5900/MMI translation, generated configuration, syscalls and hardware runtime, expensive incremental builds, indirect targets, boot, and bounded emulator comparison. Exclude other generations, ordinary emulation, generic MIPS work, and artifact acquisition.

**Improved.** The migration preserves layered diagnosis, generated-code ownership, syscall/runtime bring-up, guest/host mapping, incremental-build protection, indirect-target evidence, legal boundaries, and A/B comparison. It removes automatic setup, MCP/cross-skill calls, fixed paths/TOML commands, runner counts/build durations, universal header bans, “95% environment,” exact hardware/backend assumptions, and arbitrary breakpoint success criteria. It adds ELF/module/relocation identity, syscall failure behavior, EE/IOP and vector semantics, actual dependency-cost assessment, bounded state alignment, and repeated-failure falsification. No script is bundled because project toolchains and runtimes vary; PCSX2 integration remains Stage H.

Final skill: `internal/skills/builtin/ps2recomp/SKILL.md`. Its eval set has eight positive, eight negative, and three syscall/generated/evidence cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No PS2 target/tool runtime is claimed because none was available.
