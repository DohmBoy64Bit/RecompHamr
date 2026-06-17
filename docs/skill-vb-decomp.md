# vb-decomp

Virtual Boy static recompilation using sp00nznet/vbrecomp. V810→C recompiler
+ VIP video/VSU audio runtime. Corpus-driven — 76 ROMs, improvements benefit
the whole library.


## Kickoff

`/skill vb-decomp` — then "I have a .vb ROM at [PATH]. Start Phase 0 setup."

## What it teaches

- 6 phases: Setup → Recompile → Bringup → Render → Harden → Polish
- v810recomp pipeline: decode V810 → analyze CFG → emit annotated C
- Runtime: V810 CPU state, VIP video (384×224 LED), VSU 5-channel audio
- Corpus sweep: `sweep.ps1` runs against all 76 ROMs, `STATUS.md` auto-generated
- HLE hints: `rename` for function interception in custom drivers
- Runtime validation via `bizhawk` (BizHawk emulates Virtual Boy)

## What it references

- `/skill bizhawk` — bizhawk.* tools for runtime validation
- `REPHAMR_STATE.md` — persistent project memory

## When to use

Virtual Boy static recompilation — V810 lifting, VIP/VSU bringup, corpus
sweep hardening, or per-game driver development.
