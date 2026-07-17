# `ps3recomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/ps3recomp.md` was read completely; its pipeline scripts, runtime files, build commands, and referenced status/graphics documents were project-dependent and no separate Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for PS3 SELF/ELF static recompilation: PPU/SPU discovery and translation, PRX/relocations, NID/HLE contracts, generated configuration/builds, trampoline and indirect failures, RSX/runtime bring-up, and bounded emulator comparison. Exclude other platforms, ordinary emulator debugging, generic PowerPC work, and unauthorized artifact/key acquisition.

**Improved.** The migration preserves layered translation/HLE/host diagnosis, generated-code ownership, NID evidence, runtime/graphics bring-up, legal boundary, and A/B comparison. It removes automatic cloning/installing/decryption, MCP calls, hard-coded scripts/TOML/build/backend assumptions, file-count warnings, “95% NID,” build-as-strongest-evidence, and arbitrary address success criteria. It adds SELF/ELF and PRX/TLS/relocation provenance, PPU TOC/ABI/atomics, distinct SPU lifecycle, NID signatures/errors/callbacks, unauthorized-key refusal, and bounded aligned-state evidence. No script is bundled because toolchains and runtimes vary; PINE integration remains Stage H.

Final skill: `internal/skills/builtin/ps3recomp/SKILL.md`. Its eval set has eight positive, eight negative, and three NID/security/evidence cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No PS3 target/tool runtime is claimed because none was available.
