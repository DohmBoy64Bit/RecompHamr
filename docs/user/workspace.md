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

Stage G adds `/init-re`, which creates the evidence directories and canonical project, evidence, hypothesis, blocker, changelog, command, toolchain, function, format, recompilation, decompilation, and state files without overwriting existing content. Created and retained paths are tightened to the same POSIX owner-only modes or Windows current-user-only DACL as other private state; unsafe links and non-regular paths are refused. It deliberately does not create Legacy `mcp.json` or flat `.rehamr/skills` artifacts.

`/status-re` reads a fixed set of canonical evidence files, limits each displayed section to 1,800 UTF-8-safe bytes, and reports missing or unavailable files without exposing their contents through errors. `/doctor` remains unsupported in Stage G because its Legacy behavior mixed blocking environment probes and Stage H MCP details.
