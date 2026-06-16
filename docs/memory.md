# Persistent Memory

recomphamr maintains a project-wide state file at `.rehamr/REPHAMR_STATE.md`
that survives sessions, restarts, and `/clear` — giving the LLM persistent
context about what it's working on.

## How it works

1. **Auto-injected** — `buildSystem()` reads `.rehamr/REPHAMR_STATE.md` and
   injects it into the system prompt under `## Persistent Memory` on every
   turn. The LLM sees it before any loaded skills.

2. **LLM-maintained** — the system prompt instructs the LLM to update the
   file with `edit_file` after major actions: phase changes, new evidence,
   blockers found, commands that work.

3. **Survives `/clear`** — `rebuildSystem()` re-reads the file from disk, so
   clearing the conversation doesn't lose project context.

4. **Human-editable** — edit `.rehamr/REPHAMR_STATE.md` directly to correct
   addresses, add notes, or trim stale sections. The LLM picks up changes on
   the next turn.

## State file structure

Created by `/init-re` from a template with these sections:

| Section | Purpose |
|---|---|
| **Quick Rules** | 5 mandatory rules the LLM sees on every read |
| **Current Phase** | Track, phase, and current goal |
| **Project Info** | Project name, goal, source of truth, toolchain |
| **Workspace Paths** | Project root, evidence dir, repos cache, key files |
| **Active Commands** | Verbatim commands that work (never reconstructed from memory) |
| **Blockers** | Table of issues with status and evidence |
| **Function / Symbol Ledger** | Names, addresses, classification, confidence, sources |
| **Learned Patterns** | Session close synthesis: "X causes Y, fix with Z" |
| **Session Log** | Date-stamped summaries of each session |

## Token budget

The state file template is ~1,200 characters (~300 tokens). Mid-project,
a populated file runs 500-800 tokens. The LLM is instructed to keep it
under 1,500 tokens — about 2-3% of a 32K context window.

## Lifecycle

```
Session start → buildSystem() reads state → injected into prompt
    ↓
LLM works, updates state via edit_file after major actions
    ↓
Session end → LLM synthesizes ## Learned Patterns
    ↓
State persists on disk (.rehamr/REPHAMR_STATE.md)
    ↓
Next session → auto-injected, resume instantly
```
