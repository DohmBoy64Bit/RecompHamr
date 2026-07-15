# Agent Skills Standard and Migration Rules

## Status

This document governs future **Stage G** work. It does not authorize skills implementation during Stage A.

## Objective

Replace the historical RecompHamr-Legacy skills design with a standards-based, progressively disclosed skills system that follows the current Agent Skills specification and fits RecompHamr's separated architecture.

The migration target is the skill's useful expertise and workflow, not the old file layout or loader implementation.

## Mandatory external authority set

### Before implementing skills-client support

Read the complete current guide:

- <https://agentskills.io/client-implementation/adding-skills-support>

Also read the complete current specification before finalizing parser, validation, discovery, precedence, activation, or diagnostics behavior:

- <https://agentskills.io/specification>

### Before converting Legacy skills

Before editing **each individual Legacy skill**, confirm that the complete current versions of all of the following have been read in the current Codex task/session. If the skill is migrated in a separate task/session, reread the complete set before editing that skill:

- <https://agentskills.io/skill-creation/quickstart>
- <https://agentskills.io/skill-creation/best-practices>
- <https://agentskills.io/skill-creation/optimizing-descriptions>
- <https://agentskills.io/skill-creation/evaluating-skills>
- <https://agentskills.io/skill-creation/using-scripts>
- <https://agentskills.io/specification>

These documents are mandatory authorities for **every individual migrated skill**. Each per-skill migration packet must record that the complete set was read for that skill's task/session and must explicitly apply the requirements. If the authority set changes during the task, reread the affected current documents before closure.

For Codex-specific interoperability and authoring behavior, also consult the current official Codex skills documentation:

- <https://developers.openai.com/codex/build-skills>

## Client implementation requirements

The future RecompHamr skills subsystem must be owned below the TUI and expose typed application contracts.

At minimum, the design must cover:

- skill discovery;
- scope and deterministic precedence;
- trust policy for project-provided instructions;
- `SKILL.md` parsing;
- strict validation for RecompHamr-authored skills;
- clearly documented compatibility handling for third-party malformed-but-recoverable skills if supported;
- diagnostics for skipped, invalid, or shadowed skills;
- catalog generation;
- activation;
- relative resource resolution;
- cancellation and error handling where scripts or tools run;
- configuration and enable/disable behavior;
- testability without the TUI.

Do not hide discovery, parsing, tool execution, filesystem access, or lifecycle side effects inside TUI widgets.

## Progressive disclosure is mandatory

The implementation must preserve the three-level model:

1. **Catalog** — expose only discovery metadata needed to choose a skill, primarily `name` and `description`.
2. **Instructions** — load the full `SKILL.md` only when the skill is activated.
3. **Resources** — load scripts, references, assets, or other supporting files only when needed.

Do not eagerly inject the complete contents of every installed skill into the base system prompt.

## Discovery and locations

The Stage G design must explicitly evaluate both RecompHamr-native and cross-client locations. The preferred interoperability target is to support the `.agents/skills/` convention in addition to any documented RecompHamr-native location.

At minimum, decide and document:

- project-level locations;
- user-level locations;
- whether organization/bundled scopes exist;
- traversal bounds and excluded directories;
- deterministic collision/precedence behavior;
- shadowing diagnostics;
- trust requirements for project-level skills.

Do not copy Legacy precedence rules without re-evaluating them against the current Agent Skills client-implementation guidance.

## `SKILL.md` compliance for migrated RecompHamr skills

Each migrated skill must be a directory containing a file named exactly `SKILL.md`.

Required frontmatter:

- `name`;
- `description`.

For RecompHamr-authored skills, the `name` must follow the current specification, including matching the parent directory and using the allowed lowercase/hyphen form.

The `description` must:

- explain what the skill helps accomplish;
- explain when the skill should be used;
- use concrete user-intent and domain terms;
- include useful boundaries so adjacent tasks do not trigger incorrectly;
- remain within the specification limit.

Prefer concise, focused `SKILL.md` instructions. Move large detail into `references/` and load it on demand. Use `scripts/` only when deterministic or repeated executable behavior adds real value. Use `assets/` only for genuine static resources.

## No mechanical one-to-one conversion

For each Legacy skill:

1. read the entire Legacy skill and every directly referenced resource;
2. inspect related implementation, tests, docs, and real failure/correction history when available;
3. identify the actual reusable expertise and workflow;
4. remove stale product-specific assumptions, duplicated global rules, obsolete tools, and unsupported claims;
5. split unrelated concerns into separate skills when that improves triggering and coherence;
6. merge duplicate Legacy skills only when their user intent and activation boundary are genuinely the same;
7. produce a standards-compliant skill using progressive disclosure;
8. evaluate the converted skill before declaring it complete.

The conversion may improve wording, structure, scripts, references, safety checks, or workflow. Do not preserve old formatting or section names merely for visual similarity.

## Skill description optimization

Each skill must have trigger tests covering both positive and negative cases.

Use realistic user prompts. Include:

- explicit domain wording;
- implicit intent without the exact skill name;
- terse prompts;
- detailed prompts with realistic paths or context;
- likely typos or abbreviations where relevant;
- adjacent tasks that should **not** trigger the skill.

For substantial skills, target roughly 8-10 should-trigger and 8-10 should-not-trigger queries as a starting point. Run repeated trials when the client allows observation of activation behavior. Use a fixed train/validation split when iterating on descriptions so improvements do not merely overfit the test prompts.

Front-load the essential trigger intent and scope because clients may shorten descriptions when many skills are installed.

## Per-skill evaluations

Each migrated skill must contain or be linked to a reproducible evaluation set. Prefer:

```text
<skill>/
  SKILL.md
  evals/
    evals.json
```

Each evaluation case should contain:

- a realistic prompt;
- a human-readable expected outcome;
- optional input files when needed;
- at least one boundary or malformed-input case across the set.

Evaluate the new skill against a meaningful baseline such as:

- the same task without the skill;
- the previous skill version;
- the Legacy workflow when it is runnable and relevant.

Do not keep assertions that always pass without the skill merely to inflate success rates. Investigate tests that always fail in both configurations. Review actual outputs, not only aggregate scores.

A skill is complete only when its evaluation evidence shows that it adds reliable value for its intended scope without unacceptable false triggering.

## Script rules

Prefer instructions over scripts unless the task requires deterministic behavior, repeated non-trivial logic, or external tooling.

A bundled script must:

- have a narrow, documented purpose;
- be callable non-interactively;
- document dependencies and supported platforms;
- use pinned or bounded dependency versions where practical;
- provide useful `--help` or equivalent usage behavior when appropriate;
- produce helpful errors;
- use structured output when another agent step consumes the result;
- handle malformed input and cleanup;
- respect cancellation and output bounds where applicable;
- avoid exposing secrets.

Because RecompHamr is Windows-first, do not make a migrated built-in skill depend exclusively on Bash when a reasonable PowerShell or cross-platform implementation is available. Cross-platform scripts are allowed when their dependencies and invocation are explicit.

## Per-skill migration packet

Each Legacy skill conversion records:

- Legacy source paths and references read;
- current authority pages read;
- intended user intent and trigger boundary;
- what was preserved;
- what was removed as stale or duplicated;
- what was improved and why;
- final directory and `SKILL.md` name;
- scripts/references/assets added and why;
- trigger evaluation results;
- output-quality evaluation results;
- manual review notes;
- final disposition: `equivalent`, `improved`, `intentionally changed`, `blocked`, or `unverified`.

No skill is complete merely because it parses or appears in a picker.
