# recomp-foundations

Use this skill when you need foundational theory about static recompilation —
binary formats, control-flow recovery, instruction lifting, indirect calls,
CPU architecture reference, or GPU pipeline translation. It does NOT contain
the knowledge itself — it tells you where to find it.

> recompclass is a learning reference. Always verify claims against tool
> output, Ghidra behavior, runtime evidence, and upstream source code.
> recompclass explains theory; your project proves it.

## Setup (once per project)

Clone recompclass via repomixr: `repo_url: https://github.com/sp00nznet/recompclass`

Then use `read_file` to read specific modules from
`.rehamr/repos/sp00nznet-recompclass/repo/units/` and
`.rehamr/repos/sp00nznet-recompclass/repo/docs/`.

## Module map

| Topic | Module | Path |
|---|---|---|
| Binary formats (ELF, PE, ROM headers) | Module 2 | `units/unit-1-foundations/module-02-binary-formats/` |
| CPU architectures overview | Module 3 | `units/unit-1-foundations/module-03-cpu-architectures/` |
| Reading assembly (x86, MIPS, ARM, Z80, PPC) | Module 4 | `units/unit-1-foundations/module-04-reading-assembly/` |
| Tooling (Ghidra, Capstone) | Module 5 | `units/unit-1-foundations/module-05-tooling-ghidra-capstone/` |
| Control-flow recovery | Module 6 | `units/unit-2-core-techniques/module-06-control-flow-recovery/` |
| Lifting fundamentals | Module 7 | `units/unit-2-core-techniques/module-07-lifting-fundamentals/` |
| First lift (Z80 → C) | Module 8 | `units/unit-2-core-techniques/module-08-first-lift-z80/` |
| Game Boy recompilation | Module 9 | `units/unit-3-first-targets/module-09-game-boy/` |
| NES/6502 | Module 10 | `units/unit-3-first-targets/module-10-nes-6502/` |
| SNES/65816 | Module 11 | `units/unit-3-first-targets/module-11-snes/` |
| GBA/ARM7 | Module 12 | `units/unit-3-first-targets/module-12-gba-arm7/` |
| DOS/x86 real-mode | Module 13 | `units/unit-3-first-targets/module-13-dos/` |
| Indirect calls + jump tables | Module 14 | `units/unit-4-pipeline-essentials/module-14-indirect-calls/` |
| Hardware shims + SDL2 | Module 15 | `units/unit-4-pipeline-essentials/module-15-hardware-shims/` |
| Build systems + CMake | Module 17 | `units/unit-5-pipeline-mastery/module-17-build-systems/` |
| Testing + validation | Module 18 | `units/unit-5-pipeline-mastery/module-18-testing-validation/` |
| Optimization | Module 19 | `units/unit-5-pipeline-mastery/module-19-optimization/` |
| N64 / MIPS | Module 20 | `units/unit-6-console-architectures/module-20-n64-mips/` |
| N64 RSP/RDP deep dive | Module 21 | `units/unit-6-console-architectures/module-21-n64-rsp-rdp/` |
| GameCube / PowerPC | Module 22 | `units/unit-6-console-architectures/module-22-gamecube-ppc/` |
| Wii / Broadway | Module 23 | `units/unit-6-console-architectures/module-23-wii-broadway/` |
| Dreamcast / SH-4 | Module 24 | `units/unit-6-console-architectures/module-24-dreamcast-sh4/` |
| PS2 / Emotion Engine | Module 25 | `units/unit-6-console-architectures/module-25-ps2-ee/` |
| Saturn / Dual SH-2 | Module 26 | `units/unit-7-advanced-targets/module-26-saturn-sh2/` |
| Xbox / Win32 | Module 27 | `units/unit-7-advanced-targets/module-27-xbox-win32/` |
| Xbox 360 / Xenon PPC | Module 28 | `units/unit-7-advanced-targets/module-28-xbox360-xenon/` |
| GPU pipeline translation | Module 29 | `units/unit-7-advanced-targets/module-29-gpu-translation/` |
| PS3 / Cell | Module 30 | `units/unit-8-extreme-targets/module-30-ps3-cell/` |
| Multi-threaded recomp | Module 31 | `units/unit-8-extreme-targets/module-31-multithreaded-recomp/` |

### Quick references

| Topic | Path |
|---|---|
| Glossary of terms | `docs/glossary.md` |
| Tool setup guide | `docs/tool-setup.md` |
| CPU ISA references (12 architectures) | `docs/architecture-reference/` |
| Tool cheat sheets | `docs/cheat-sheets/` |
| Recommended reading | `docs/recommended-reading.md` |

## How to use

1. Clone recompclass once via `repomixr` at project start
2. When you encounter an unknown concept, find the module in this table
3. Use `read_file` to read the relevant module (first 200 lines, then more if needed)
4. Apply the theory to your project, verifying against real tool output
5. Never cite recompclass as evidence — it explains concepts, not your binary
