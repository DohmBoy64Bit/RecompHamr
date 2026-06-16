# n64-debug-mcp

N64 runtime debugging skill. Loads Mupen64Plus MCP methodology AND unlocks
`n64-debug-mcp.*` tools for live emulator debugging.

## What it teaches

- Check whether the Mupen64MCP daemon is running and a ROM is loaded
- Prefer reproducible captures: register dumps, memory reads, breakpoint
  traces, display-list decodes, PI DMA logs, RSP task headers, frame captures
- Save derived evidence to `.rehamr/evidence/` or `.rehamr/traces/`
- Runtime state is evidence, but interpretation needs static confirmation
  against ROM data and known hardware behavior

## What it unlocks

All N64 debug tools: `n64-debug-mcp.n64_read_memory`,
`n64-debug-mcp.n64_set_breakpoint`, `n64-debug-mcp.n64_get_registers`,
`n64-debug-mcp.n64_decode_display_list`, and 40+ more.

## When to use

When debugging N64 ROMs with Mupen64Plus and you need the LLM to set
breakpoints, trace execution, and capture runtime evidence directly.

## Setup

See **[mcp-n64.md](mcp-n64.md)** for install and enable instructions.
