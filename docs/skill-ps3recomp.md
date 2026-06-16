# ps3recomp

PS3 static recompilation skill using sp00nznet/ps3recomp. PPU/SPU lifting,
HLE stub implementation, NID resolution, RSX graphics bringup.

## What it teaches

- Pipeline: ELF parse → find functions → disassemble PPU → lift to C++ → CMake
- 6 operational phases: Setup → First Lift → First Build → HLE Bringup →
  Runtime Debug → Graphics + Polish
- HLE discipline: never edit `recompiled/`, fixes in stubs/config/TOML
- Build Gate: Ninja generator, full output reading, exit code 0 verification
- NID resolution, trampoline system debugging, RSX/D3D12 backend

## What it references

- `/skill ghidra-mcp` — static PPU analysis of EBOOT.ELF
- `REPHAMR_STATE.md` — persistent project memory

## When to use

PlayStation 3 static recompilation — PPU/SPU lifter execution, HLE stub
implementation, NID resolution, RSX graphics debugging, or any ps3recomp
pipeline task.
