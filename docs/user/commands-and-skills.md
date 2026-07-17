# Commands and Agent Skills

RecompHamr uses one ordered command registry for terminal completion, `/help`, and CLI help. The active Stage G commands are:

- `/clear` — reset the conversation and prompt recall;
- `/models [name]` — list profiles or persistently activate one;
- `/skills` — list discovered Agent Skills, active markers, and display-safe diagnostics;
- `/skill <name>` — explicitly activate one discovered skill for the conversation;
- `/init-re` — idempotently initialize the private evidence workspace without overwriting existing files;
- `/status-re` — show a bounded summary of canonical evidence files;
- `/help` — render the same registry.

Unknown commands retain the quiet `unknown command - type / to see options` warning. Configuration is reloaded before slash dispatch, so valid profile and skill trust/disable edits take effect without restart.

## Skill locations, trust, and precedence

The standards-based client discovers exact `SKILL.md` files one directory below:

```text
~/.agents/skills/<name>/SKILL.md
<project>/.agents/skills/<name>/SKILL.md
```

User skills are trusted because the current user installed them. Project skills are repository-provided instructions and are skipped unless `skills.trust_project: true` is explicitly set. A trusted project skill outranks a user skill with the same name; within one scope the first configured root wins. Names are sorted for deterministic disclosure. Discovery does not recurse beyond immediate skill directories and is bounded to 2,000 entries per root.

RecompHamr-authored built-in skills ship inside the executable. At startup they are restored atomically into a content-addressed, current-user-protected `.rehamr/bundled-skills/<digest>/` directory; local edits there are overwritten from the binary. Built-ins have the lowest precedence, so a user skill or explicitly trusted project skill with the same name can replace one for that project/session. Old content-addressed versions are not automatically deleted.

RecompHamr-authored skills are validated strictly against the current Agent Skills specification: the directory and frontmatter `name` must match and use 1–64 lowercase letters, numbers, or single hyphens; `description` is required and limited to 1,024 characters; `compatibility` is limited to 500 characters; unknown frontmatter fields, invalid UTF-8, unsafe links, non-regular files, and `SKILL.md` over 512 KiB are rejected. Diagnostics never expose skill bodies or absolute paths in the TUI.

## Progressive disclosure and resources

Model rounds initially receive only a bounded metadata catalog. The catalog is capped at 8 KiB and reports how many later entries were omitted. When the model’s task matches a description it can call `activate_skill`; `/skill` provides the same explicit activation. Activation rereads and revalidates the selected file, refuses replacement races, and injects its instructions once per conversation.

Only names of files under `scripts/`, `references/`, and `assets/` are disclosed at activation. The model must use `read_skill_resource` to load an exact listed path on demand. The reader accepts only activated skills, caps each resource at 2 MiB, and refuses traversal, links, non-regular files, and replacement races. Discovery and activation never execute scripts. Any later script execution still uses the ordinary tool permissions, cancellation, timeout, and output bounds.

`/clear` removes conversation activations. A configuration reload rediscovers skills, retains only activations that still revalidate, removes disabled or missing skills, and refreshes the model-facing tool enums. Activated instructions larger than the original fixed system reservation reduce history packing so response headroom remains protected.

## Configuration

```yaml
skills:
  trust_project: false
  disabled:
    - example-skill
```

`trust_project` defaults to false. `disabled` removes exact discovered names from catalog disclosure and activation and emits a display-safe diagnostic. Do not use project trust for repositories whose instructions you have not reviewed.

## Deliberate Legacy changes

The Legacy flat `.rehamr/skills/*.md` loader is unsupported; use the interoperable `.agents/skills/<name>/SKILL.md` structure. Legacy `/skill-audit` and `/skill-new` are intentionally replaced by specification validation and per-skill trigger/output evaluations: the old heuristic classifier and remote fetch-and-wrap flow cannot prove a useful or safe skill. Legacy `/doctor` is not exposed in Stage G because it synchronously mixed local process/network probes and Stage H MCP configuration into presentation. `/mcp` remains Stage H.
