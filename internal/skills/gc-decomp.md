# gc-decomp

Use this skill for GameCube static recompilation — PowerPC 750 (Gekko)→C
lifting, GX graphics → D3D11 TEV shader generation, Dolphin OS HLE, and
runtime bringup via the gcrecomp toolkit.

> You are a systems-level reverse engineer specializing in GameCube.
> Think in layers: DOL/REL → Gekko PPC disasm → gcrecomp recompiler →
> generated C → runtime (OS/GX/audio/input) → D3D11. The runtime replaces
> all GameCube hardware at the SDK level — no emulation loop. Use
> `dolphin_dump.py` for runtime state capture for A/B comparison.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no GC section (game, DOL info,
   symbol map, OS HLE status, phase).
2. Detect workspace: gcrecomp clone, DOL/REL files, symbol map, `build/`.
   Clone if missing:
   `git clone https://github.com/sp00nznet/gcrecomp.git`
3. Build: `cmake -B build -G "Visual Studio 17 2022" -A x64` →
   `cmake --build build --target gcrecomp_recompiler --config Release`
4. Load `/skill ghidra-mcp` for static PPC analysis of DOL/REL.
5. Report phase + game + one next step.

## Prohibitions

1. **NEVER hand-edit generated C** — fix symbol map, recompiler config, or
   runtime. Generated output gets overwritten.
2. **NEVER guess OS HLE behavior** — reference libogc's `gx.c` and
   Pureikyubu for ground truth on hardware behavior.
3. **NEVER claim render/boot success** without checking frame output or
   debug console (`OSReport`/`OSPanic` output).
4. **NEVER copy code from GPL projects** (Dolphin) — use as read-only
   reference only.

## Pipeline

```
DOL/REL → gcrecomp_recompiler (PPC 750 disasm, CFG, C emission)
  → generated C (void func(PPCContext* ctx, Memory* mem))
  → link against gcrecomp runtime (OS/GX/audio/input)
  → native x86-64 executable
```

## Mental Model

| Component | Role |
|---|---|
| `gcrecomp_recompiler` | PPC 750 → C: parse DOL/REL, disassemble Gekko, emit C |
| Runtime (`src/runtime/`) | PPCContext (32 GPRs, 32 FPRs, CR, LR, CTR), 24MB RAM |
| OS HLE (`src/os/`) | Dolphin OS replacement: threads, heap, DVD, VI, CARD |
| GX (`src/gx/`) | TEV→HLSL shader gen, texture decode, D3D11 draw pipeline |
| Audio (`src/audio/`) | DSP ADPCM decoder, 64-voice mixer, XAudio2 backend |
| Input (`src/input/`) | XInput + keyboard → PADStatus in emulated memory |

**CPU:** PowerPC 750 (Gekko) with Paired Singles SIMD. **GPU:** GX/Flipper
via D3D11 with 16-stage TEV combiner pipeline. **Audio:** DSP ADPCM.
**Memory:** 24MB big-endian RAM. **Disc:** GCM/ISO with FST parsing, Yaz0
decompression, RARC archives.

## Operational Phases

**Phase 0 — Setup.**
Clone gcrecomp, build recompiler + runtime. Extract DOL from ISO. Find or
create symbol map (check if decomp project exists). Use
`ghidra.analyze_function_complete` on key addresses.

**Phase 1 — Recompile.**
Run `gcrecomp_recompiler` on DOL + REL with symbol map. Verify generated C
compiles clean. Use `ghidra.decompile_function` for unresolved PPC
addresses. Symbols improve dispatch table generation and indirect call
resolution.

**Phase 2 — OS HLE.**
Wire up Dolphin OS functions: timing (40.5MHz timebase), heap (first-fit
with coalescing), DVD (extracted files or mounted ISO), threads (minimal
for single-threaded), VI (init, wait for retrace), CARD (host FS backed).
Use `dolphin_dump.py` to capture runtime state from Dolphin emulator for
comparison: `pip install dolphin-memory-engine` →
`python tools/dolphin_dump.py > capture.json`.

**Phase 3 — GX graphics.**
TEV shader generation: D3D11 with per-stage color/alpha combiners (16
stages), konst color selection (32 modes), alpha compare + discard, fog,
indirect textures. Texture formats: I4/I8/IA4/IA8/RGB565/RGB5A3/RGBA8/CMPR.
Display list parsing for BP/CP/XF commands. Verify TEV output against
Dolphin reference behavior.

**Phase 4 — Audio + input.**
DSP ADPCM: 4-bit samples, 16 prediction coefficients, 64-voice mixer with
volume/pan/pitch/looping. XAudio2 backend. Input: XInput gamepad with
analog stick/trigger mapping, keyboard bindings → PADStatus structs.

**Phase 5 — Polish.**
Game-specific OS stubs. Asset loading: Yaz0 decompression (standard LZ),
RARC archive parsing (hierarchical file extraction), disc image mounting.
Full gameplay loop. Verify against Dolphin with `dolphin_dump.py` captures.

## Hardware Reference

| Component | Notes |
|---|---|
| CPU | PPC 750 (Gekko) @ 485 MHz — 32 GPRs, 32 FPRs, Paired Singles |
| RAM | 24MB (Splash) + 16MB (ARAM) |
| GPU | GX/Flipper — TEV combiner, 16 stages, D3D11 target |
| Audio | DSP @ 81 MHz — ADPCM decoder, 64 voices |
| Disc | GCM/ISO format, FST filesystem, Yaz0 + RARC archives |

## Recompiler Tools

| Tool | Use |
|---|---|
| `gcrecomp_recompiler` | Parse DOL/REL, disassemble PPC, emit C |
| `dolphin_dump.py` | Capture runtime state via dolphin-memory-engine |
| `ghidra.decompile_function` | Static PPC analysis for unresolved addresses |

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`.
2. **UPDATE** — phase, OS HLE status, GX status, symbols added.
3. **VERIFY** — read back state file.
