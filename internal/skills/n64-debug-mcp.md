# n64-debug-mcp

Use this skill only when N64 runtime debugging via Mupen64Plus MCP is relevant.

Checklist:
1. Check whether the Mupen64MCP daemon is running and reachable.
2. Check whether a ROM is loaded and emulation state is valid.
3. If the daemon is not running, provide exact setup steps rather than pretending.
4. Prefer reproducible captures: register dumps, memory reads, breakpoint traces,
   display-list decodes, PI DMA logs, RSP task headers, and frame captures.
5. Save derived evidence into `.rehamr/evidence/` or `.rehamr/traces/`.

Guardrails:
- Runtime state is evidence, but interpretation needs static confirmation.
- Emulator output is not hardware truth — validate against ROM data and known
  hardware behavior.
