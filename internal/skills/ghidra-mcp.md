# ghidra-mcp

Use this skill only when a Ghidra MCP bridge or local Ghidra automation is relevant.

Checklist:
1. Check whether Ghidra is installed and reachable.
2. Check whether Java is available.
3. Check whether the current project already has Ghidra scripts, exports, labels, notes, or prior MCP setup.
4. If MCP is not configured, provide exact setup steps rather than pretending it is available.
5. Prefer reproducible exports: functions, symbols, decompiler output, cross references, data references, strings, imports, and entry points.
6. Save derived evidence into `.rehamr/evidence/` or `.rehamr/functions/`.

Guardrails:
- Ghidra output is evidence, but interpretation still needs classification.
- Decompiler output is not source truth; it is a tool-derived hypothesis unless confirmed by symbols, behavior, or matching code.

