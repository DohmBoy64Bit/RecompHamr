# pcrecomp

Use this skill only when PC static recompilation is relevant — turning an old
Windows/DOS .exe into modern C code with the PCRECOMP-Next toolkit.

## Pipeline

```
Phase 1 — Analyze:  pcrecomp.pe.analyze target.exe
Phase 2 — Disassemble: pcrecomp.disasm32.run target.exe
Phase 3 — Classify:  pcrecomp.classify.run target.exe functions.json
Phase 4 — Graph:     pcrecomp.disasm32.callgraph target.exe
Phase 5 — Lift:      pcrecomp.lift32.run functions.json → src/recomp/
Phase 6 — Build:     bash cmake -B build && cmake --build build
Phase 7 — Iterate:   fix build errors, re-lift, re-test
```

## Checklist
1. Check whether PCRECOMP-Next is installed (tool should report path).
2. Check whether the target binary is accessible and readable.
3. Check whether Python 3.10+ and capstone/pefile are available.
4. Prefer targeted lifts over batch — lift one subsystem at a time.
5. Save evidence to `.rehamr/recomp/` after each phase.
6. Never edit lifted code as the primary fix — fix metadata, reclassify, re-lift.

## Guardrails
- PE analysis output is evidence, but compiler ID is a hypothesis until confirmed.
- Disassembler output is not source truth — verify against known data patterns.
- Lifted code is scaffold, not finished product — expect build errors and iterate.
- Runtime shims (recomp32/recomp16) are template C source — copy them, don't call them through MCP.
- Ghidra tools (decompile_all, function_stats) complement static analysis but don't replace it.
