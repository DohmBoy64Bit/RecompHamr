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

The built-in `SkillServers` mapping is:

| Skill | Gates server |
|---|---|
| `ghidra-mcp` | `ghidra` |
| `n64-debug-mcp` | `n64-debug-mcp` |
| `pcrecomp` | `pcrecomp` |
| `mcp-pine` | `mcp-pine` |
| `objdiff` | `objdiff` |
| `pcsx2` | `pcsx2` |
| `bizhawk` | `bizhawk` |
| `sega2asm` | `sega2asm` |

## Built-in skills

Twenty-eight skills are compiled into the binary:

| `/skill <name>` | Purpose | Details |
|---|---|---|
| `bizhawk` | Multi-system emulator debug (gates `bizhawk.*`) | [doc](skill-bizhawk.md) |
| `cdb-debug` | Windows CDB debugger for recompiled native EXEs | [doc](skill-cdb-debug.md) |
| `core-re` | General RE workflow | [doc](skill-core-re.md) |
| `evidence-mode` | Evidence-first methodology | [doc](skill-evidence-mode.md) |
| `build-fix-loop` | Iterate on build failures | [doc](skill-build-fix-loop.md) |
| `file-format-reversing` | Binary format analysis | [doc](skill-file-format-reversing.md) |
| `function-discovery` | Find and classify functions | [doc](skill-function-discovery.md) |
| `ghidra-mcp` | Ghidra integration (gates `ghidra.*`) | [doc](skill-ghidra-mcp.md) |
| `imhex` | Hex editor / pattern language for binary formats | [doc](skill-imhex.md) |
| `n64-debug-mcp` | N64 runtime debugging (gates `n64-debug-mcp.*`) | [doc](skill-n64-debug-mcp.md) |
| `n64-decomp` | N64 matching decomp + N64Recomp pipeline | [doc](skill-n64-decomp.md) |
| `objdiff` | Object file diffing (gates `objdiff.*` tools) | [doc](skill-objdiff.md) |
| `pcrecomp` | PC recomp pipeline (gates `pcrecomp.*`) | [doc](skill-pcrecomp.md) |
| `vb-decomp` | Virtual Boy static recomp (V810→C, VIP/VSU) | [doc](skill-vb-decomp.md) |
| `windows-game-decomp` | Windows game matching decomp + compiler-matrix | [doc](skill-windows-game-decomp.md) |
| `xbox360-decomp` | Xbox 360 static recompilation (4 tracks, ReXGlue+Xenon) | [doc](skill-xbox360-decomp.md) |
| `gb-recomp` | Game Boy static recompilation (trace-guided, PyBoy) | [doc](skill-gb-recomp.md) |
| `gc-decomp` | GameCube static recomp (PPC→C, GX/D3D11, OS HLE) | [doc](skill-gc-decomp.md) |
| `gen-decomp` | Sega Genesis decomp (sega2asm + bizhawk) | [doc](skill-gen-decomp.md) |
| `mcp-pine` | RPCS3 debug bridge (gates `mcp-pine.*` tools) | [doc](skill-mcp-pine.md) |
| `ps3recomp` | PS3 static recompilation (PPU/SPU lifting, HLE, RSX) | [doc](skill-ps3recomp.md) |
| `xboxrecomp` | OG Xbox static recompilation (XBE→C, kernel, D3D, ICALL) | [doc](skill-xboxrecomp.md) |
| `pcsx2` | PCSX2 debug bridge (gates `pcsx2.*` tools) | [doc](skill-pcsx2.md) |
| `ps2recomp` | PS2 static recomp (MIPS→C++, syscalls, PCSX2) | [doc](skill-ps2recomp.md) |
| `recomp-foundations` | Recomp theory router (recompclass module map) | [doc](skill-recomp-foundations.md) |
| `sega2asm` | Genesis ROM disassembler (gates `sega2asm.*`) | [doc](skill-sega2asm.md) |
| `snesrecomp` | SNES static recomp (65816→C, LakeSnes HW) | [doc](skill-snesrecomp.md) |
| `project-handoff` | Generate project docs | [doc](skill-project-handoff.md) |

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

Skills range 55–85 lines each (~1,500–6,000 tokens depending on the
model's tokenizer). Loading all twenty-eight adds substantial prompt
overhead — load only the skills you need. Loading none adds zero:
`buildSystem()` skips the `## Active RE Skills` block when `activeSkills`
is empty.

See **[tools-vs-skills.md](tools-vs-skills.md)** for a cost comparison
between tools and skills.
