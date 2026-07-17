# `pcrecomp` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/pcrecomp.md` was read completely; its PCRECOMP-Next paths, tools, templates, and artifacts were project-dependent and no separate Legacy resource file was supplied. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual edit.

Trigger for legacy DOS MZ/16-bit, Win16 NE, and native Win32 PE static-recompilation pipelines: loader identity, discovery/classification, x86 lifting, compatibility runtime, generated builds, and behavioral validation. Exclude managed/modern-engine projects, console targets, source ports, isolated PE education, and protection bypass.

**Improved.** The migration preserves layered analysis/discovery/lifting/runtime/build ownership, generated-code regeneration, targeted diagnosis, legal artifacts, and evidence handoff. It removes automatic setup, `/doctor` and MCP calls, hard-coded PCRECOMP-Next commands/paths/dependencies/templates, assumed output schemas, compile-as-strongest-evidence, and universal batch-size rules. It adds format/bitness/segmentation selection, flags/ABI/self-modification and failure semantics, explicit translator ownership, focused semantic vectors, unsupported dynamic behavior, and protection boundaries. No script is bundled because executable formats and project recompiler/runtime contracts vary.

Final skill: `internal/skills/builtin/pcrecomp/SKILL.md`. Its eval set has eight positive, eight negative, and three format/semantic cases. Automated parser/client checks apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts. No lawful PC target/recompiler fixture was available, so no external recompiler run is claimed.
