# mcp-pine

RPCS3 emulator debug bridge via PINE IPC. Connects to a running RPCS3
instance for A/B comparison, dynamic memory probing, and savestate management.


## Kickoff

`/skill mcp-pine` — then "Connect to RPCS3 and compare register values at 0x00123456."

## What it enables

- Read/write PS3 guest memory at runtime
- Save/load savestates for reproducible A/B comparison
- Compare recompiled native behavior against original RPCS3 execution

The PINE IPC interface must be enabled in RPCS3 settings.

## Setup

1. Enable PINE IPC: RPCS3 Settings → Advanced → Enable PINE IPC
2. Install `mcp-pine` on PATH (or set `RECOMPHAMR_MCP_PINE_COMMAND`)
3. Start RPCS3, load game, start recomphamr
4. Load `/skill mcp-pine` — unlocks `mcp-pine.*` tools

## When to use

PS3 static recompilation A/B comparison — verifying recompiled C++ output
matches original PPU/SPU behavior at specific guest addresses. Used by
`ps3recomp` during runtime debugging.
