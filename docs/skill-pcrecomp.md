# pcrecomp

PC static recompilation pipeline skill. Loads PCRECOMP-Next methodology AND
unlocks `pcrecomp.*` tools for PE analysis, disassembly, and code lifting.


## Kickoff

`/skill pcrecomp` — then "Run pe_analyze on mystery.exe and classify the compiler."

## What it teaches

- Follow the pipeline: Analyze → Disassemble → Classify → Graph → Lift →
  Build → Iterate
- Check that PCRECOMP-Next is installed and the target binary is accessible
- Prefer targeted lifts over batch — one subsystem at a time
- Save evidence to `.rehamr/recomp/` after each phase
- Never edit lifted code as the primary fix — fix metadata, reclassify, re-lift
- Lifted code is scaffold, not finished product

## What it unlocks

8 pipeline tools by default: `pcrecomp.pe.analyze`, `pcrecomp.pe.extract_imports`,
`pcrecomp.disasm32.run`, `pcrecomp.disasm32.callgraph`, `pcrecomp.lift32.run`,
`pcrecomp.classify.run`, `pcrecomp.ghidra.decompile_all`,
`pcrecomp.ghidra.function_stats`. Set `RECOMPHAMR_MCP_PCRECOMP_TOOLS=*` for
all 30+ including 16-bit DOS, NE/Win16, DRM, and asset extractors.

## When to use

When working on a PC static recompilation project — turning an old Windows/DOS
executable into modern C code that compiles and runs natively.

## Setup

See **[mcp-pcrecomp.md](mcp-pcrecomp.md)** for install and enable instructions.
