# `file-format-reversing` Skill Migration Record

Legacy `RecompHamr-Legacy-main/internal/skills/file-format-reversing.md` was read completely and links no external resources. The full mandatory current Agent Skills authority set and official Codex guidance had been read before this edit.

Trigger for discovering unknown binary/container structures from samples and code/runtime evidence; exclude already documented formats, ordinary conversion, and new format design.

**Improved.** The skill preserves sample inventory, offsets and observations, unknown retention, tiny parser-first work, cross-sample validation, real-sample preference, failure recording, and non-completeness guards. It adds sample provenance, meaningful-variation criteria, malformed/boundary cases, bounded parser arithmetic/allocation/decompression, and data minimization; it removes the dependency on activating another skill and corrects state paths. No bundled script is justified because parser language and format are task-specific.

Final skill: `internal/skills/builtin/file-format-reversing/SKILL.md`. Its eval set includes eight positive, eight negative, and three safety/evidence output cases. Automated structure/client tests apply. Direct agent review on 2026-07-17 passed 8/8 positive, 8/8 negative, and 3/3 output contracts.
