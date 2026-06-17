# snesrecomp

SNES static recompilation using sp00nznet/snesrecomp. 65816→C via
RECOMP_PATCH + LakeSnes hardware library for real PPU/APU/DMA.


## Kickoff

`/skill snesrecomp` — then "I have an SFC at [PATH]. Start Phase 0 setup."

## What it teaches

- Pipeline: disassemble 65816 → RECOMP_PATCH in C → link snesrecomp → native
- 5 phases: Setup → Disassembly → Recompilation → Integration → Iteration
- cpu_ops.h helpers + bus_read8/bus_write8 routing
- LakeSnes hardware: PPU Mode 0-7, SPC700 audio, 8 DMA channels, SRAM, DSP-1
- Auto function dispatch via hash table (JSR/JSL by SNES address)
- Modding: override functions by link order

## When to use

SNES static recompilation — 65816 to C conversion, hardware bringup with
snesrecomp library, or linking against LakeSnes for real PPU/APU/DMA.
