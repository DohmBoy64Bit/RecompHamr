# Change Control

For every non-trivial change:

1. create or record a work packet using `work-packet-template.md`;
2. read the root `AGENTS.md`, current status, and the task-specific routed authority;
3. inspect the owning source, tests, and relevant docs before editing;
4. define the required compatibility contract before Legacy parity work;
5. make the smallest coherent change that fits the current architecture;
6. run focused checks first, then the canonical gate;
7. update affected docs and parity/evidence records in the same change;
8. record blockers rather than guessing.

A better implementation than Legacy is allowed when the required capability is preserved or an intentional contract change is explicitly documented and verified. Source-code similarity is not a completion criterion.

TUI changes require an explicit statement of why the change does not redesign or rearrange the accepted layout while the freeze is active.

Skills work additionally requires the mandatory authority and per-skill migration process in `../roadmap/agent-skills-standard.md`.
