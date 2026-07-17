# `pcsx2` Skill Stage H Disposition

Legacy `RecompHamr-Legacy-main/internal/skills/pcsx2.md` was read completely. The mandatory current Agent Skills authority set and official Codex guidance had been read before this individual disposition.

**Intentionally deferred to Stage H; no Stage G skill is bundled.** The Legacy file is an operational contract for a PCSX2 DebugServer/PINE MCP bridge: connection negotiation, `/mcp` lifecycle, remote register and memory mutation, break/watchpoints, stepping, blocking waits, save states, and emulator-specific setup. Stage G cannot advertise those unavailable transport and tool contracts.

Reusable principles—record game/build/connection and guest-address identity, pause or stop at a coherent event, checkpoint before mutation, compare identical states, and treat emulator observations as bounded reference evidence—may inform `ps2recomp`. Stage H must verify current server/emulator source and schemas, local binding and trust, capability negotiation, 128-bit/register encoding, pause consistency, address spaces, conditional breakpoint safety, wait cancellation, cleanup, output limits, mutation authorization, and sanitized evidence before creating and evaluating a replacement skill. Legacy ports, versions, install commands, tool counts/names, and failure mappings are not accepted as current facts.

No parser, trigger, output, runtime, or human evaluation is claimed for a `pcsx2` skill because no such Stage G skill exists.
