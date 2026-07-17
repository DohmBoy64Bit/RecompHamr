# `build-fix-loop` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/build-fix-loop.md` was read completely and has no linked resources. The mandatory current Agent Skills creator, evaluation, script, specification, and Codex guidance set had been read before this individual edit.

The trigger is a currently reproducible build, test, generation, or recompilation failure. Passing-build refactors, performance work, dependency upgrades, and new build-system design are excluded.

**Improved.** The skill preserves exact-command reproduction, earliest-cause diagnosis, single-change iteration, generated-output ownership, repeated-failure strategy change, explicit blockers, and evidence-backed closure. It removes unsafe blanket-revert language, corrects evidence paths, distinguishes build success from runtime/binary acceptance, and adds data minimization. No script or resource is needed.

Final skill: `internal/skills/builtin/build-fix-loop/SKILL.md`. Its eval set contains eight positive, eight negative, and three boundary/output cases. Specification/client validation is automated. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts; the two positive fixtures that lacked the skill's required exact failing command were corrected before the final review.
