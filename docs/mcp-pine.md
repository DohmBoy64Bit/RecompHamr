# mcp-pine Setup

RPCS3 emulator debug bridge. Connects to a running RPCS3 instance via the
PINE IPC interface for A/B comparison, dynamic memory probing, and savestate
management.

## Dependencies

- **RPCS3** — PS3 emulator with PINE IPC enabled
- **Python 3.10+** — the bridge runtime
- **mcp-pine** — the bridge itself

## Install

1. Enable PINE IPC in RPCS3: **Settings → Emulator → Enable PINE IPC**
2. Start RPCS3 with the target game loaded
3. Install mcp-pine: `pip install mcp-pine` or from source
4. Ensure `mcp-pine` is on PATH (or set `RECOMPHAMR_MCP_PINE_COMMAND`)

## Enable

1. Start recomphamr — connect with `/mcp connect mcp-pine`
2. Or use HTTP: set `RECOMPHAMR_MCP_PINE_URL` to a streamable-http endpoint
3. Run `/skill mcp-pine` — unlocks `mcp-pine.*` tools
4. Verify: `/mcp tools mcp-pine`

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
