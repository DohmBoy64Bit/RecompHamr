# Workspace State

RecompHamr identifies the current working directory as the project root. Stage D supports one optional Legacy-compatible project-state file:

```text
.rehamr/REPHAMR_STATE.md
```

When the file exists, is non-empty, valid UTF-8, and no larger than 64 KiB, its text is added to the model system message under `## Persistent Memory`. RecompHamr labels it as **untrusted project-maintained context**: it can summarize project facts, paths, decisions, and blockers, but it does not outrank the embedded application rules and its claims should be verified against evidence.

The state is refreshed before every model round. It therefore survives conversation clearing and process restart, and edits made between model rounds are visible without restarting. Missing or empty state preserves the normal baseline prompt. Unsafe or unreadable optional state is omitted silently so its contents and filesystem details do not enter the TUI or private debug log.

## Security and bounds

- The canonical state path is fixed beneath the current project’s private `.rehamr` directory; callers cannot supply another relative path.
- Symlinks, Windows reparse points represented as links, directories, devices, and other non-regular files are refused.
- RecompHamr verifies that the opened file is the same regular file inspected before opening, rejecting replacement races.
- The file is tightened with the same POSIX owner-only mode or Windows current-user-only protected DACL used for other private state.
- The 64 KiB limit is checked before and during reading, preventing misleading metadata or concurrent growth from creating unbounded prompt input.

Stage D does not automatically create this file and does not add `/init-re`, `/status-re`, or `/doctor`. New slash commands remain Stage G work; reverse-engineering templates, tools, skills, and MCP configuration remain in their separately authorized stages. You may create or edit `REPHAMR_STATE.md` manually if you want persistent project context now.
