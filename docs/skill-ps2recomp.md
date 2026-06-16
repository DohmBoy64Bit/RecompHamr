# ps2recomp

PS2 static recompilation skill. MIPS→C++ lifting, syscall stubbing, A/B
comparison with PCSX2, and runtime bringup.

## What it teaches

- 6 operational phases: Setup → First Build → Syscall Bringup → First Boot →
  A/B Comparison → Polish
- Four fix tools: TOML → Runtime C++ → Game Override → Recompiler
- Build discipline: no-clean, no runner/ edits, no .h edits, vcvars64 required
- PS2 hardware: RDRAM 32 MB, EE mask 0x1FFFFFFF, GS/VIF/GIF/Scratchpad
- PCSX2-MCP A/B comparison at Phase 4

## What it references

- `/skill pcsx2` — PCSX2-MCP for A/B comparison (Phase 4)
- `/skill ghidra-mcp` — static MIPS analysis
- `REPHAMR_STATE.md` — persistent project memory

## When to use

PS2 static recompilation — MIPS analysis, syscall implementation, C++
runtime debugging, or PCSX2 A/B comparison for the PS2Recomp pipeline.
