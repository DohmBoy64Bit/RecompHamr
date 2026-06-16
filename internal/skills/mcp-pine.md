# mcp-pine

RPCS3 emulator debug bridge. Connects to a running RPCS3 instance via the
PINE IPC interface for A/B comparison, dynamic memory probing, and savestate
management. Load this skill to unlock `mcp-pine.*` tools.

## What it enables

- Read/write PS3 guest memory at runtime
- Save and load savestates for reproducible A/B comparison
- Compare recompiled native behavior against original RPCS3 execution
- Dynamic probing of unresolved addresses and NIDs

The PINE IPC interface must be enabled in RPCS3 settings. The mcp-pine
server connects over localhost — no remote exposure.

## Setup

1. Enable PINE IPC in RPCS3: Settings → Emulator → Enable PINE IPC
2. Install mcp-pine: `pip install mcp-pine` or from source
3. Start RPCS3 with the target game loaded
4. Ensure `mcp-pine` is on PATH (or set `RECOMPHAMR_MCP_PINE_COMMAND`)
5. Start recomphamr — auto-connects at launch
6. Load `/skill mcp-pine` — unlocks `mcp-pine.*` tools

## When to use

PS3 static recompilation A/B comparison — verifying recompiled C++ output
matches original PPU/SPU behavior at specific guest addresses. Used by the
`ps3recomp` skill during Phase 4 (runtime debugging).
