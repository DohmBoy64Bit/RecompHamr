# Skills

Skills inject focused context into the system prompt, giving the LLM
methodology and guardrails for specific tasks. MCP skills also gate which
server tools the LLM sees.

## How skills work

When loaded via `/skill <name>`, the skill's full markdown body is appended
to the system prompt under `## Active RE Skills`. This text travels with every
turn until recomphamr is restarted — skills survive `/clear` and `/models`
switches.

Skills also unlock MCP tools by convention: if a registered MCP server has
`RequireSkill: true`, loading a skill whose name matches the server (or maps
to it via the built-in `SkillServers` table) exposes that server's tools to
the LLM. For example, `/skill ghidra-mcp` maps to the `ghidra` server,
injecting `ghidra.*` tools.

## Built-in skills

Nine skills are compiled into the binary:

| `/skill <name>` | Purpose |
|---|---|
| `core-re` | General RE workflow |
| `evidence-mode` | Evidence-first methodology |
| `build-fix-loop` | Iterate on build failures |
| `file-format-reversing` | Binary format analysis |
| `function-discovery` | Find and classify functions |
| `ghidra-mcp` | Ghidra MCP integration (gates `ghidra.*` tools) |
| `n64-debug-mcp` | N64 runtime debugging via Mupen64Plus MCP (gates `n64-debug-mcp.*` tools) |
| `pcrecomp` | PC static recompilation pipeline (gates `pcrecomp.*` tools) |
| `project-handoff` | Generate project docs for handoff |

List them with `/skills`; active skills are marked `*`.

## Custom skills

Drop `.md` files into `.rehamr/skills/` and they appear in `/skills` with a
`(custom)` label. Custom skills take precedence over built-in skills with the
same name.

```
.rehamr/
├── config.yaml
├── skills/
│   ├── my-workflow.md       # /skill my-workflow
│   └── my-mcp.md            # gates my-mcp.* tools if server registered
└── history
```

To pair a custom skill with a custom MCP server, name the skill after the
server and register it with `RequireSkill: true`. See [docs/mcp.md](mcp.md)
for details.

## Token cost

Each skill `.md` file averages ~10-15 lines (~200-400 tokens). Loading all
nine adds ~2,800-3,600 tokens to the system prompt. Loading none adds zero —
`buildSystem()` skips the `## Active RE Skills` block entirely when
`activeSkills` is empty.
