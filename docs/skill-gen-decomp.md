# gen-decomp

Sega Genesis decompilation skill. ROM splitting via sega2asm + runtime
validation via bizhawk. 68000/Z80 disassembly, 49 compression formats,
VDP register hints.


## Kickoff

`/skill gen-decomp` — then "I have a Genesis ROM at [PATH]. Start Phase 0 recon."

## What it teaches

- 6 operational phases: Recon → Config → Split → Symbols → Runtime → Matching
- sega2asm for static analysis: plan, split, compression detection
- bizhawk for dynamic proof: memory r/w, buttons, save states, A/B comparison
- Hardware: 68000 CPU, Z80 sound, YM7101 VDP, 49 Genesis compression formats
- Prohibitions: never guess compression, segment boundaries, or claim match without bizhawk

## What it references

- `/skill sega2asm` — unlocks sega2asm.* tools for ROM splitting
- `/skill bizhawk` — unlocks bizhawk.* tools for runtime validation
- `REPHAMR_STATE.md` — persistent project memory

## When to use

Sega Genesis / Mega Drive decompilation — ROM analysis, segment discovery,
compression cataloging, 68000/Z80 disassembly, or function-level matching
with runtime comparison.
