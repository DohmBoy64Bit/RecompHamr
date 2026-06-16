# gc-decomp

GameCube static recompilation using sp00nznet/gcrecomp. PPC 750 → C
recompiler, GX → D3D11 with TEV shader generation, Dolphin OS HLE.

## What it teaches

- 6 phases: Setup → Recompile → OS HLE → GX Graphics → Audio+Input → Polish
- PPC 750 (Gekko) → C with Paired Singles SIMD, DOL/REL parsing
- GX graphics: TEV → HLSL, 16-stage combiners, D3D11 draw pipeline
- Dolphin OS HLE: timing, heap, DVD, VI, CARD, threads, interrupts
- DSP ADPCM audio: 4-bit samples, 64-voice mixer, XAudio2
- Asset pipeline: Yaz0 decompression, RARC archives, disc mounting
- Runtime validation via `dolphin_dump.py` + dolphin-memory-engine

## What it references

- `/skill ghidra-mcp` — static PPC analysis of DOL/REL
- `REPHAMR_STATE.md` — persistent project memory

## When to use

GameCube static recompilation — DOL/REL parsing, PPC→C lifting, GX/D3D11
graphics bringup, Dolphin OS HLE implementation, or TEV shader debugging.
