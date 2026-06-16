# xbox360-decomp

Xbox 360 static recompilation methodology skill. Covers 4 tracks from XEX
extraction through runtime bringup to full game loop.

## What it teaches

- Four tracks: A (ReXGlue), B (XenonRecomp), C (matching decomp), D (360toolsUpdated)
- 6 operational phases: Extract → Init → First Build → Runtime Bringup → Crash Triage → Polish
- Build Gate: no-clean enforcement, MSVC + clang-cl verification, regen discipline
- Four fix tools: TOML → stubs → SDK patches → regen
- Guest PPC evidence via Ghidra MCP at `0x82000000` base
- Track D pipeline: STFS/ISO extract → `rexglue init` → `rexglue codegen` → cmake
- VdSwap QPC, switch tables, ROV/MSAA, SDK patches 0001-0005

## What it references

- `/skill ghidra-mcp` — unlocks `ghidra.*` tools for PPC guest-VA analysis
- `REPHAMR_STATE.md` — persistent project memory

## When to use

Xbox 360 static recompilation projects — XBLA, LIVE, PIRS, CON, 360 ISO
targets using ReXGlue or XenonRecomp toolchains. Merger of
xbox360-decomp-skill (tracks A-D, operational phases) and 360tools-skill
(build gate, extraction tools, run script conventions).
