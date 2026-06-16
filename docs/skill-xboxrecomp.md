# xboxrecomp

OG Xbox static recompilation skill using sp00nznet/xboxrecomp. XBE extraction,
x86→C lifting, kernel/D3D/audio shim bringup, ICALL crash debugging.

## What it teaches

- Pipeline: parse XBE → disasm → classify → lift x86→C with global registers
- 6 operational phases: Setup → Lift → Game Project → ICALL Bringup → Runtime → Polish
- ICALL workflow: MAP file → classify garbage vs valid vs kernel VA
- Runtime libraries: kernel (147+ imports), D3D8→D3D11, audio, NV2A, input
- Build gaps: apu_xaudio2, CMAKE_SOURCE_DIR, subdirectory includes
- RenderWare handling: vtable/ICALL overrides in recomp_manual.c

## What it references

- `REPHAMR_STATE.md` — persistent project memory

## When to use

OG Xbox static recompilation — XBE extraction, x86 lifting, kernel/D3D
shim implementation, ICALL crash triage, or any xboxrecomp pipeline task.
