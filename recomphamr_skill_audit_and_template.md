# RecompHAMR Skill Audit, Refactor Template, and Skill Quality Guide

## Scope

This audit covers the 25 uploaded Markdown skill files:

- `bizhawk.md`
- `build-fix-loop.md`
- `cdb-debug.md`
- `core-re.md`
- `evidence-mode.md`
- `file-format-reversing.md`
- `function-discovery.md`
- `gb-recomp.md`
- `gc-decomp.md`
- `gen-decomp.md`
- `ghidra-mcp.md`
- `imhex.md`
- `mcp-pine.md`
- `n64-debug-mcp.md`
- `n64-decomp.md`
- `objdiff.md`
- `pcrecomp.md`
- `pcsx2.md`
- `project-handoff.md`
- `ps2recomp.md`
- `snesrecomp.md`
- `vb-decomp.md`
- `windows-game-decomp.md`
- `xbox360-decomp.md`
- `xboxrecomp.md`

The two Go files, `skills.go` and `skills_test.go`, are not Markdown skills, but they matter for refactoring because the loader treats skill names as the Markdown filename without `.md`, resolves case-insensitively, and allows a custom directory to override embedded skills.

---

## Executive Summary

Most of the uploaded Markdown files are already skill-like, but they are not all the same kind of skill. They fall into three useful classes:

1. **Full workflow skills** — complete operating procedures for a domain, console, platform, or track. These are the strongest examples and should be the model for large skills.
2. **Micro-protocol skills** — compact rules or methods, such as evidence handling, function discovery, build fixing, or project handoff. These are valid skills, but they need a smaller standardized template.
3. **Tool bridge cards** — setup/capability notes for MCP or emulator bridges. These are useful, but most are more doc-like than skill-like because they explain what the tool enables without fully specifying boot checks, evidence outputs, failure handling, and session close behavior.

The best skills share these traits:

- They start with a clear **“Use this skill for…”** trigger.
- They define the agent’s **role and mental model**.
- They require **workspace detection instead of path assumptions**.
- They contain hard **prohibitions** that prevent expensive or destructive mistakes.
- They use **phased workflows** rather than vague advice.
- They include **build gates, validation gates, or proof requirements**.
- They require evidence and state updates before closing the session.
- They distinguish generated code, source-of-truth config, runtime glue, and evidence artifacts.

The weakest files are not bad; they are just incomplete as skills. Most of them need a boot contract, evidence artifact targets, failure modes, and session-close requirements.

---

## Classification Legend

| Classification | Meaning | Refactor Priority |
|---|---|---|
| **Full Skill** | Operational enough to guide an AI through a complete workflow with boot, guardrails, phases, and session close. | Low to medium. Standardize section order and fill minor gaps. |
| **Near-Full Skill** | Mostly operational, but missing one major section such as prohibitions, build gate, or validation proof. | Medium. Add missing hard gates. |
| **Micro-Skill** | Small but valid behavior module: a focused checklist, doctrine, or protocol. | Medium. Add compact trigger, boot, outputs, and stop rules. |
| **Tool Bridge Card** | Explains a tool or MCP bridge, but mostly setup/capability documentation. | High if intended to be loaded as an active skill. Add runtime behavior. |
| **Doc-Like Reference** | More explanation than action. Useful documentation, but not enough agent procedure. | High if it must behave as a skill. Convert to checklist/phase format. |

---

## Per-File Audit

| File | Classification | Why | Refactor Recommendation |
|---|---|---|---|
| `bizhawk.md` | **Tool Bridge Card** | Good capability/setup summary for BizHawk bridge, but it mostly says what the bridge enables and how to install it. | Add boot checks, connection verification, common commands, evidence outputs, save-state policy, and session close. |
| `build-fix-loop.md` | **Micro-Skill** | Strong compact loop and report format. Very skill-like despite being short. | Add trigger conditions, stop conditions, retry/circuit-breaker rule, and evidence file update behavior. |
| `cdb-debug.md` | **Full Skill** | Has prerequisites, boot, workflow, patterns, crash classification, diagnostic logging, evidence archiving, and session close. | Add “when not to use,” destructive-command rules, and explicit output files for traces. |
| `core-re.md` | **Micro-Skill** | Good general RE policy, but very short. | Expand into the base micro-template: trigger, boot, source-of-truth discovery, validation, output artifacts, session close. |
| `evidence-mode.md` | **Micro-Skill / Policy Skill** | Excellent evidence taxonomy and anti-hallucination policy. | Add enforcement checklist, required output format, promotion rules from hypothesis to confirmed, and examples. |
| `file-format-reversing.md` | **Micro-Skill / Method Skill** | Skill-like because it has evidence rules, preferred output structure, and workflow. | Add boot checklist, sample inventory command expectations, parser validation gate, and failure handling. |
| `function-discovery.md` | **Micro-Skill / Method Skill** | Good goal, methodology, and output CSV schema. | Add phased workflow, confidence rubric, tool references, stopping conditions, and validation rules for jump tables/vtables. |
| `gb-recomp.md` | **Full Skill** | Strong boot, prohibitions, pipeline, phases, commands, mental model, and close behavior. | Add explicit build gate and source-of-truth artifact list. |
| `gc-decomp.md` | **Full Skill** | Strong role, boot, prohibitions, pipeline, phases, hardware model, tools, and close behavior. | Add validation gate for generated C build, frame proof, and Dolphin comparison outputs. |
| `gen-decomp.md` | **Full Skill** | Good Genesis workflow, prohibitions, phases, hardware reference, and session close. | Add legal/asset boundary note and explicit generated-output policy. |
| `ghidra-mcp.md` | **Tool Micro-Skill** | Useful checklist and guardrails, but short and tool-centered. | Add boot verification, export commands/targets, evidence output files, and session close. |
| `imhex.md` | **Tool Guide / Micro-Skill** | Good explanation of ImHex and Pattern Language usage; more guide-like than operational. | Add trigger, boot, prohibited assumptions, pattern validation gate, output files, and session close. |
| `mcp-pine.md` | **Tool Bridge Card** | Good setup/capability summary for RPCS3 PINE bridge. | Add connection boot checks, runtime capture protocol, A/B comparison output, failure handling, and session close. |
| `n64-debug-mcp.md` | **Tool Micro-Skill** | Good checklist and guardrails for Mupen64Plus MCP, but very compact. | Add setup, common captures, exact output paths, reproducibility policy, and session close. |
| `n64-decomp.md` | **Full Skill / Exemplary** | One of the strongest: role, boot, prohibitions, mental model, constants, build gate, tracks, phases, session close. | Use this as a model. Minor improvement: add standardized “outputs” section and function-discovery links. |
| `objdiff.md` | **Tool Bridge Card** | Good capability/setup/when-to-use card. | Add boot check for `objdiff.json`, report format, validation gate, failed-diff handling, and session close. |
| `pcrecomp.md` | **Pipeline Micro-Skill** | Has pipeline, checklist, and guardrails. More operational than a tool card, less complete than a full skill. | Add boot section, prohibitions, phases, evidence outputs, and session close. |
| `pcsx2.md` | **Tool Bridge Card** | Good setup/capability summary for PCSX2 MCP. | Add connection checks, pause-before-read rule, A/B comparison workflow, output captures, and session close. |
| `project-handoff.md` | **Micro-Skill** | Clear and useful compact handoff protocol. | Add a fixed handoff template and rules for evidence/hypothesis separation. |
| `ps2recomp.md` | **Full Skill / Exemplary** | Very strong: boot, prohibitions, build gate, phases, fix tools, mental model, MCP reference, session close. | Use as a model. Minor improvement: add explicit source artifact paths and phase success criteria. |
| `snesrecomp.md` | **Near-Full Skill** | Has role, boot, pipeline, mental model, phases, hardware reference, and session close. | Add prohibitions, build gate, validation proof, emulator comparison, and generated-code policy. |
| `vb-decomp.md` | **Full Skill** | Strong boot, prohibitions, pipeline, mental model, phases, hardware reference, and close behavior. | Add explicit build/validation gate and corpus report output format. |
| `windows-game-decomp.md` | **Full Skill / Exemplary** | Strong layer model, runtime-family detection, prohibitions, tracks, phases, Ghidra references, engine-specific guidance, and session close. | Use as a model. Add standard output artifact list and exact report file names. |
| `xbox360-decomp.md` | **Full Skill / Exemplary** | Strong track matrix, prohibitions, phases, build gate, mental model, fix tools, Ghidra reference, and close behavior. | Use as a model. Add explicit validation success criteria for each track. |
| `xboxrecomp.md` | **Full Skill** | Strong boot, prohibitions, mental model, pipeline, phases, crash quick reference, runtime libraries, known build gaps, and session close. | Add explicit build gate section and evidence output paths for ICALL traces/map files. |

---

## Common Traits of the Strongest Skills

The best uploaded skills are not just notes. They behave like runtime instructions for an AI agent. Across the strongest examples, the common pattern is:

1. **Trigger statement**
   - The first line says exactly when to load the skill.
   - Good format: `Use this skill for <domain/task> — <specific workflows>.`

2. **Role framing**
   - The skill tells the AI what kind of specialist it is acting as.
   - Strong examples use a blockquote after the title to define the mental stance.

3. **Layered mental model**
   - Good skills separate source inputs, generated outputs, runtime glue, debugger evidence, and final validation.
   - This prevents the AI from patching the wrong layer.

4. **Boot contract**
   - Good skills define what must happen at the start of every session.
   - Typical boot steps: read `REPHAMR_STATE.md`, detect workspace layout, verify tools, load supporting skills, report phase and next step.

5. **No-assumption workspace detection**
   - Good skills never assume paths.
   - They inspect for ROMs, EXEs, configs, generated folders, build dirs, symbols, state files, and tool output.

6. **Prohibitions**
   - The best skills include hard “NEVER” rules.
   - This is especially important for generated code, destructive build commands, copyrighted assets, retail binaries, SDK leaks, massive directories, and long rebuild paths.

7. **Pipeline**
   - Strong skills show the whole transformation path.
   - Example pattern: `Input → analysis → generated code → runtime → native executable → validation`.

8. **Operational phases**
   - Strong skills are phase-based, not generic.
   - Each phase has a goal, expected artifacts, and a next action.

9. **Build or validation gates**
   - Good skills say what must be inspected before running a build or claiming success.
   - They also say what proof is required: command output, objdiff, trace hit, frame capture, emulator comparison, or runtime log.

10. **Tool quick reference**
    - Good skills include commands or MCP calls only when they are verified and scoped.
    - They avoid inventing APIs or flags.

11. **Failure-pattern table**
    - Strong debugging skills include symptom → likely cause → fix layer.
    - This turns recurring crashes into repeatable triage instead of guesswork.

12. **Evidence/state update requirement**
    - Good skills require `.rehamr/`, `REPHAMR_STATE.md`, `docs/`, or logs to be updated with confirmed facts.

13. **Session close protocol**
    - Strong skills end with: synthesize learned patterns, update state, verify the state file.

---

# Universal RecompHAMR Skill Template

Use this template for large workflow skills such as `n64-decomp`, `ps2recomp`, `windows-game-decomp`, `xbox360-decomp`, `xboxrecomp`, `gc-decomp`, `gb-recomp`, `gen-decomp`, `vb-decomp`, and `snesrecomp`.

```md
# <skill-name>

Use this skill for <specific domain/task> — <specific workflows this skill owns>.

> You are a <specialist role>. Think in layers: <input/original artifact> →
> <analysis/config> → <generated/reconstructed output> → <runtime/toolchain> →
> <validation>. Diagnose which layer is broken before changing code.

## When to use

Use this skill when:
- <trigger 1>
- <trigger 2>
- <trigger 3>

Do not use this skill when:
- <out-of-scope case 1>
- <out-of-scope case 2>

## Boot

1. Read `REPHAMR_STATE.md`. If missing or incomplete, populate the `<domain>` section with:
   - target/game/project name
   - original artifact path or hash, if lawful and available
   - toolchain/version assumptions
   - current track
   - current phase
   - active blocker
2. Detect workspace layout. Do not assume paths. Look for:
   - <original inputs>
   - <config files>
   - <generated directories>
   - <build directories>
   - <logs/evidence>
3. Verify required tools:
   - <tool 1>
   - <tool 2>
   - <tool 3>
4. Load supporting skills when relevant:
   - `/skill core-re`
   - `/skill evidence-mode`
   - `/skill build-fix-loop`
   - `/skill ghidra-mcp`
   - <domain-specific skill>
5. Report:
   - detected track
   - phase
   - current evidence
   - one next step

## Prohibitions

1. **NEVER <dangerous action>.** Reason: <why>.
2. **NEVER <generated-output edit>.** Durable fixes go in <config/runtime/source-of-truth>.
3. **NEVER claim success** without <proof command/output>.
4. **NEVER invent** tool flags, APIs, offsets, symbols, paths, compiler versions, or hardware behavior.
5. **NEVER run destructive commands** such as <examples> without explicit user approval.
6. **NEVER request, commit, or redistribute** copyrighted binaries, ROMs, retail assets, keys, or SDK leaks.
7. After <N> repeated failures with the same symptom, STOP, update state, and gather fresh evidence.

## Mental Model

| Layer | Role | Source of Truth | Durable Fix Location |
|---|---|---|---|
| Original artifact | <binary/ROM/XEX/etc.> | hash/tool output | never edited |
| Analysis metadata | <symbols/config/functions> | <tool output> | <config/docs> |
| Generated output | <generated dir> | generated from metadata | never hand-edit first |
| Runtime/glue | <runtime/stubs/shims> | source + tests | <runtime path> |
| Validation | <build/trace/objdiff/emulator> | command output | <evidence path> |

## Tracks

| Track | Use When | Pipeline | Success Criteria |
|---|---|---|---|
| A — <track name> | <condition> | <pipeline> | <proof> |
| B — <track name> | <condition> | <pipeline> | <proof> |

If there is only one track, replace this section with `## Pipeline`.

## Pipeline

```text
<input> → <analysis> → <configuration> → <codegen/reconstruction>
        → <build/runtime> → <validation/proof>
```

## Operational Phases

### Phase 0 — Recon / Setup

Goal: identify the target and create the evidence baseline.

Actions:
- <action>
- <action>

Artifacts:
- `REPHAMR_STATE.md`
- `.rehamr/evidence/<domain>_recon.md`
- <other files>

Exit criteria:
- <confirmed fact>
- <tool verified>

### Phase 1 — First Analysis / Split / Lift

Goal: <goal>.

Actions:
- <action>

Exit criteria:
- <proof>

### Phase 2 — First Build / First Boot

Goal: <goal>.

Actions:
- <action>

Exit criteria:
- <proof>

### Phase 3 — Debug / Runtime Bringup

Goal: <goal>.

Actions:
- <action>

Exit criteria:
- <proof>

### Phase 4 — Validation / A-B Comparison

Goal: prove behavior with the narrowest reliable check.

Actions:
- <trace, objdiff, emulator, debugger, frame capture, tests>

Exit criteria:
- <proof>

### Phase 5 — Polish / Documentation

Goal: stabilize and document.

Actions:
- <action>

Exit criteria:
- <proof>

## Build Gate / Validation Gate

Before building or claiming success:

1. Inspect command for destructive options.
2. Verify environment and build directory.
3. Run the narrowest useful command.
4. Read full output and exit code.
5. Record proof in `REPHAMR_STATE.md` or `.rehamr/evidence/`.

Success may only be claimed when:
- <exact command> succeeds with exit code 0
- <secondary validation> passes
- <evidence artifact> is updated

## Tool Quick Reference

```bash
# Verified commands only. Do not invent flags.
<command 1>
<command 2>
```

MCP/tool calls, when available:

```text
<tool.method> — <purpose>
<tool.method> — <purpose>
```

## Failure Patterns

| Symptom | Likely Layer | Evidence to Gather | Durable Fix |
|---|---|---|---|
| <symptom> | <layer> | <tool/log/trace> | <fix location> |

## Output Artifacts

Required:
- `REPHAMR_STATE.md`
- `.rehamr/evidence/<skill>.md`
- `.rehamr/CHANGELOG.md` if `/init-re` has run

Optional:
- `.rehamr/functions/inventory.csv`
- `.rehamr/formats/inventory.md`
- `docs/<domain>_notes.md`
- `logs/<tool>_<date>.txt`

## Session Close

1. **SYNTHESIZE** — write learned patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, blocker, commands, evidence paths, crash/build status.
3. **VERIFY** — read back the state file or evidence artifact for coherence.
4. **REPORT** — changed files, verified commands, remaining blockers, next 3 concrete steps.
```

---

# Tool Bridge Skill Template

Use this for skills such as `bizhawk`, `pcsx2`, `mcp-pine`, `objdiff`, `ghidra-mcp`, and `n64-debug-mcp`.

```md
# <tool-skill-name>

Use this skill when <tool> is needed for <specific evidence or action>.

## What it enables

- <capability>
- <capability>
- <capability>

## When to use

Use this tool for:
- <scenario>
- <scenario>

Do not use it for:
- <scenario better served by another skill/tool>

## Boot / Connection Check

1. Verify <tool> is installed or reachable:
   ```bash
   <safe check command>
   ```
2. Verify the target is loaded/running:
   - <condition>
3. Verify connection:
   - <expected response>
4. If unavailable, provide exact setup steps. Do not pretend the bridge is connected.

## Setup

1. <install step>
2. <launch step>
3. <bridge step>
4. <verification step>

## Evidence Protocol

Every capture should record:
- target name
- target hash or version, if available
- timestamp/frame/address
- command/tool used
- output path
- interpretation status: CONFIRMED / HYPOTHESIS / TODO / BLOCKED

Save evidence to:
- `.rehamr/evidence/<tool>.md`
- `.rehamr/traces/<tool>/`
- `logs/<tool>_trace.txt`

## Common Operations

| Operation | Command / Tool Call | Output | Notes |
|---|---|---|---|
| <operation> | `<command>` | <output> | <notes> |

## Guardrails

1. Runtime state is evidence, but interpretation needs static confirmation.
2. Screenshots/logs/traces do not prove root cause alone.
3. Do not mutate emulator/game state unless the task calls for it.
4. Prefer save states/checkpoints before writes or risky actions.
5. Do not expose remote debug bridges beyond localhost unless explicitly designed for it.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| <failure> | <meaning> | <step> |

## Session Close

1. Save captures/logs.
2. Update `REPHAMR_STATE.md` with last connection status and evidence paths.
3. Report what was captured, what it proves, what remains unknown.
```

---

# Micro-Skill Template

Use this for compact method or policy skills such as `evidence-mode`, `core-re`, `build-fix-loop`, `function-discovery`, `file-format-reversing`, and `project-handoff`.

```md
# <micro-skill-name>

Use this skill when <specific trigger>.

## Goal

<One sentence describing the behavior this skill enforces.>

## Rules

1. <rule>
2. <rule>
3. <rule>

## Workflow

1. <step>
2. <step>
3. <step>

## Required Output

Use this format:

```md
## CONFIRMED
- <evidence-backed facts>

## HYPOTHESIS
- <plausible but unproven ideas>

## TODO
- <next evidence-gathering or implementation steps>

## BLOCKED
- <missing tools/files/permissions>
```

## Evidence / Artifact Targets

- `<path>`
- `<path>`

## Stop Conditions

Stop and ask for/gather better evidence when:
- <condition>
- <condition>

## Session Close

1. Update the relevant evidence/state file.
2. Report changed files, verified facts, and next steps.
```

---

# What Makes a Good Skill?

A good skill is not a normal documentation page. It is an **operating contract** for how an AI should behave in a specific workflow.

A normal doc answers: “What is this?”

A good skill answers:

- When should I load this?
- What role am I taking?
- What files/tools must I inspect first?
- What am I forbidden to do?
- What evidence is acceptable?
- What layer owns the fix?
- What command proves success?
- What should I save before ending the session?

## A Skill Must Have

### 1. Clear trigger

The first paragraph should make the skill’s scope obvious.

Good:

```md
Use this skill for PS2 static recompilation — ISO/ELF extraction, MIPS R5900
analysis, TOML config, syscall stubbing, and C++ runtime debugging.
```

Weak:

```md
This document describes PS2 recompilation.
```

The good version tells the agent when to load it and what it owns.

### 2. Role and mental stance

A skill should tell the AI how to think.

Good:

```md
> You are a systems-level reverse engineer who thinks in layers: original
> MIPS → recompiled C++ → runtime abstraction → host OS.
```

This prevents shallow fixes and keeps the agent focused on layer diagnosis.

### 3. Boot checklist

Every major skill should start with a repeatable boot routine:

- read `REPHAMR_STATE.md`
- detect workspace layout
- verify toolchain
- identify current phase
- load supporting skills
- report one next step

Boot sections prevent the AI from assuming paths, rebuilding the wrong target, or ignoring previous evidence.

### 4. Prohibitions

The strongest skills have hard safety rails. These should be specific and practical.

Examples:

- never edit generated code as the primary fix
- never clean a long build without approval
- never invent tool flags or APIs
- never claim success without command output
- never commit retail binaries, ROMs, keys, or proprietary assets
- never rename functions based on vibes

Prohibitions are one of the most important parts of a skill because they prevent expensive mistakes.

### 5. Layer ownership

A good skill tells the AI where durable fixes belong.

Example:

| Problem | Bad Fix | Durable Fix |
|---|---|---|
| Generated code crash | edit generated file | fix TOML/config/runtime and regenerate |
| Unknown binary field | name it based on vibes | keep `unknown_XX` until evidence supports a name |
| Missing function call target | patch symptom | update function inventory / switch table / dispatch metadata |
| Runtime wait loop | delete loop | prove hardware/event condition and stub runtime state |

### 6. Evidence ladder

A good skill defines which sources are stronger than others.

Typical evidence ladder:

1. Build output / failing command
2. Hashes / metadata / symbols
3. Raw disassembly and xrefs
4. Emulator/debugger traces
5. Decompiler output as hypothesis
6. Human interpretation as hypothesis until proven

The exact ladder can vary by platform, but every skill needs one.

### 7. Phased workflow

Phases keep the AI from jumping straight to code edits.

Good phase design:

- Phase 0 — Recon
- Phase 1 — Split / disassemble / lift
- Phase 2 — first build
- Phase 3 — first boot
- Phase 4 — debug/A-B comparison
- Phase 5 — polish/documentation

Each phase should include exit criteria. Without exit criteria, the AI may claim progress too early.

### 8. Validation gate

A good skill says what counts as success.

Examples:

- `cmake --build build` exits 0
- `configure.py --diff` is clean
- objdiff reports a function match
- a CDB breakpoint is hit
- emulator and recomp state match at a given address/frame
- a headless screenshot exists and matches expected render state
- parser output validates against multiple real samples

“Looks fixed” is not enough.

### 9. Output artifacts

Skills should leave behind structured project memory.

Recommended locations:

- `REPHAMR_STATE.md` — current phase, blocker, learned patterns
- `.rehamr/evidence/` — command outputs, trace summaries, confirmed facts
- `.rehamr/functions/` — function inventory and classifications
- `.rehamr/formats/` — file-format inventories, samples, parser tests
- `.rehamr/CHANGELOG.md` — meaningful changes if `/init-re` has run
- `logs/` — raw tool logs and traces

A skill that does not update evidence creates repeated work in future sessions.

### 10. Session close

Every skill should have an ending routine:

1. synthesize learned patterns
2. update state/evidence
3. verify the state file
4. report changed files, verified commands, remaining blockers, and next steps

This is especially important for long reverse-engineering projects.

---

## What Makes a Skill Bad or Doc-Like?

A skill becomes doc-like when it mostly explains instead of directing action.

Common weak patterns:

- lots of background, few commands
- no boot checklist
- no “never do this” section
- no output artifacts
- no validation gate
- no stop condition
- no distinction between generated code and durable source
- no evidence standard
- no session close
- vague verbs such as “investigate,” “look into,” or “fix” without method
- tool setup but no runtime workflow

Tool cards are not bad, but if they are loaded as skills, they need operational behavior.

---

## Recommended Refactor Strategy

### 1. Keep three templates, not one

Use:

- **Full Workflow Template** for platform/domain skills
- **Tool Bridge Template** for MCP/emulator/debugger bridge skills
- **Micro-Skill Template** for compact policies and methods

Trying to force every file into one huge template would make short skills bloated.

### 2. Standardize section order

Recommended order for full skills:

1. Title
2. Use statement
3. Role blockquote
4. When to use / when not to use
5. Boot
6. Prohibitions
7. Mental model
8. Tracks or pipeline
9. Operational phases
10. Build/validation gate
11. Tool quick reference
12. Failure patterns
13. Output artifacts
14. Session close

### 3. Refactor tool bridge cards first

Highest-priority refactor targets:

- `bizhawk.md`
- `mcp-pine.md`
- `pcsx2.md`
- `objdiff.md`
- `ghidra-mcp.md`
- `n64-debug-mcp.md`

These are useful, but they should define active behavior: connection checks, evidence capture, failure handling, and session close.

### 4. Refactor micro-skills second

Targets:

- `core-re.md`
- `evidence-mode.md`
- `function-discovery.md`
- `file-format-reversing.md`
- `build-fix-loop.md`
- `project-handoff.md`
- `pcrecomp.md`

These mostly need standardized output and stop conditions.

### 5. Normalize full skills last

Targets:

- `n64-decomp.md`
- `ps2recomp.md`
- `windows-game-decomp.md`
- `xbox360-decomp.md`
- `xboxrecomp.md`
- `gc-decomp.md`
- `gb-recomp.md`
- `gen-decomp.md`
- `vb-decomp.md`
- `snesrecomp.md`

These already work. Do not over-edit them. Standardize headings and fill missing validation/output sections.

---

## Minimum Acceptance Checklist for a RecompHAMR Skill

A Markdown file is ready to be treated as a skill when it passes this checklist:

- [ ] Filename is lowercase, hyphenated, and matches the intended `/skill <name>` command.
- [ ] First paragraph begins with `Use this skill for...`.
- [ ] Scope is clear enough to know when not to use it.
- [ ] Boot checklist exists.
- [ ] Workspace detection avoids hardcoded path assumptions.
- [ ] Source-of-truth files are identified.
- [ ] Prohibitions cover destructive, generated, legal, and hallucination risks.
- [ ] Workflow is phased or checklist-driven.
- [ ] Validation gate defines what proves success.
- [ ] Output artifacts are named.
- [ ] Session close updates state/evidence.
- [ ] Unknowns stay unknown until evidence promotes them.
- [ ] Decompiler output is not treated as final truth.
- [ ] Tool commands are verified or clearly marked as examples.
- [ ] The skill tells the AI when to stop and gather better evidence.

---

## Recommended Naming and Loader Conventions

Because the loader resolves skills from Markdown filenames, the filename is the skill ID. Use:

```text
lowercase-hyphen-name.md
```

Good:

- `n64-decomp.md`
- `windows-game-decomp.md`
- `file-format-reversing.md`

Avoid:

- `N64 Decomp Guide.md`
- `xboxrecomp notes final.md`
- duplicated upload names like `n64-decomp.md` inside the embedded skill directory

Recommended cleanup before embedding:

- rename `*.md` files to their clean names
- keep one canonical file per skill
- let custom skills override embedded ones intentionally
- keep tests for custom override behavior

---

## Final Recommendation

Use the existing **N64, PS2, Windows, Xbox 360, and OG Xbox** skills as the “gold standard” for full skills. Use **evidence-mode, build-fix-loop, function-discovery, and file-format-reversing** as micro-skill modules, but give them standardized output and stop rules. Convert **BizHawk, PCSX2, RPCS3/PINE, objdiff, Ghidra MCP, and N64 debug MCP** from setup cards into active tool-bridge skills with boot checks and evidence protocols.

The guiding rule should be:

> A doc explains a workflow. A skill controls agent behavior inside the workflow.

If a file does not tell the agent what to inspect, what not to do, what proof is required, and what state to update, it is still a doc — not a finished skill.
