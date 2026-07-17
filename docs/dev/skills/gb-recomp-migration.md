# `gb-recomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/gb-recomp.md` was read completely; it referenced project-dependent tools and artifacts but no separate Legacy resource file. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for Game Boy and Game Boy Color static-recompilation coverage, LR35902 control-flow discovery, interpreter fallback, trace seeding, generated builds, and defensible performance measurement. Exclude GBA and other consoles, ordinary emulator use, generic build work, and ROM or BIOS acquisition.

**Improved.** The migration preserves the ROM-to-analysis-to-generated-build pipeline, metadata/fallback evidence, emulator trace seeding, generated-code ownership rule, both-build gate, benchmarking discipline, legal artifact boundary, and evidence handoff. It removes obsolete `/doctor` and cross-skill coupling, automatic cloning/installing, Bash and hard-coded commands, a universal 99% target, and the unsupported claim that every major gap is `JP HL`. It adds bank/address-space classification, executable-RAM and mapper/runtime alternatives, trace-scope limits, repeated-measurement requirements, bounded execution, and a falsification stop condition. No script is bundled because all executable contracts are tool- and project-version-specific.

Final skill: `internal/skills/builtin/gb-recomp/SKILL.md`. Its eval set has eight positive, eight negative, and three evidence-quality cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No GB target/tool runtime is claimed because none was available.
