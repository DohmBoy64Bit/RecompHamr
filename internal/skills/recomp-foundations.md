# recomp-foundations

Knowledge base router — maps topics to reference documentation across 9
platforms. Does NOT contain the knowledge itself; it tells you where to find
it. All sources are URLs — use `read_file` to inspect cached local copies
after cloning via `repomixr` or `bash curl`.

> **All references are external learning aids.** Always verify claims against
> tool output, Ghidra behavior, runtime evidence, and upstream source code.
> These sources explain theory and community knowledge; your project proves it.

## How to use

1. Find your platform + topic in the tables below
2. Use `recomp_reference` to fetch and cache the URL locally
3. Read the cached page with `read_file` (use **offset=0, limit=200** for
   large pages; read more chunks if needed). Never read entire documents at once
4. Apply to your project — verify every claim against real tool output
5. Never cite references as evidence — they explain concepts, not your binary

## Reference sources by platform

### N64 (Nintendo 64)

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware/wiki | n64brew | community-wiki | `https://n64brew.dev/wiki/Main_Page` |
| Curated index | n64.dev | community-index | `https://n64.dev/` |
| Hardware docs | n64docs (readthedocs) | emulator-docs | `https://n64.readthedocs.io/` |
| Official docs archive | N64Docs (DerekTurtleRoe) | archive | `https://github.com/DerekTurtleRoe/N64Docs` |
| Official manuals archive | ultra64.ca | archive | `https://ultra64.ca/resources/documentation/` |
| Static recompiler | N64Recomp | recompiler-source | `https://github.com/N64Recomp/N64Recomp` |
| Recomp runtime | N64ModernRuntime | runtime-source | `https://github.com/N64Recomp/N64ModernRuntime` |
| Recomp frontend | RecompFrontend | runtime-source | `https://github.com/N64Recomp/RecompFrontend` |
| Homebrew SDK | libdragon | open-sdk | `https://github.com/DragonMinded/libdragon` |
| Homebrew SDK docs | libdragon wiki | sdk-doc | `https://n64brew.dev/wiki/Libdragon` |
| Microcode/ hacking | Hack64 wiki | community-wiki | `https://hack64.net/wiki/doku.php?id=nintendo_64` |
| CPU reference | EN64 wiki | community-wiki | `https://en64.shoutwiki.com/wiki/N64_CPU` |
| RCP reference | EN64 wiki | community-wiki | `https://en64.shoutwiki.com/wiki/Reality_CoProcessor` |
| RE intro | RetroReversing | reverse-engineering-guide | `https://www.retroreversing.com/N64Reversing` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/nintendo-64/` |

### Xbox 360

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware wiki | Free60 | community-wiki | `https://free60.org/` |
| Hardware wiki | XenonLibrary | community-wiki | `https://xenonlibrary.com/wiki/Main_Page` |
| Error codes | XenonLibrary | community-wiki | `https://xenonlibrary.com/wiki/Errors` |
| POST codes | XenonLibrary | community-wiki | `https://xenonlibrary.com/wiki/Post_Codes` |
| Homebrew SDK docs | libXenon (Free60) | sdk-doc | `https://free60.org/Development/LibXenon/` |
| Homebrew SDK source | libXenon | open-sdk | `https://github.com/Free60Project/libxenon` |
| Reversing notes | x360-research (Invoxi) | reverse-engineering-guide | `https://github.com/InvoxiPlayGames/x360-research` |
| Reversing misc | Xbox_360_Research (Byrom90) | reverse-engineering-guide | `https://github.com/Byrom90/Xbox_360_Research` |
| Static recompiler | XenonRecomp | recompiler-source | `https://github.com/hedge-dev/XenonRecomp` |
| Shader recompiler | XenosRecomp | recompiler-source | `https://github.com/hedge-dev/XenosRecomp` |
| Runtime SDK | ReXGlue | runtime-source | `https://github.com/rexglue/rexglue-sdk` |
| Packaged tooling | 360tools | recompiler-source | `https://github.com/sp00nznet/360tools` |
| Emulator | Xenia (site) | emulator-source | `https://xenia.jp/` |
| Emulator source | Xenia | emulator-source | `https://github.com/xenia-project/xenia` |
| Filesystem | FATX (Free60) | filesystem | `https://free60.org/System-Software/Systems/FATX/` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/xbox-360/` |

### PlayStation 1 (PSX)

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware spec (modern) | psx-spx | hardware-spec | `https://psx-spx.consoledev.net/` |
| Hardware spec (original) | psx-spx (problemkaputt) | hardware-spec | `https://problemkaputt.de/psx-spx.htm` |
| Emulator/debugger | no$psx | emulator-source | `https://problemkaputt.de/psx.htm` |
| Tools/research SDK | pcsx-redux | emulator-source | `https://github.com/grumpycoders/pcsx-redux` |
| Open SDK | PSn00bSDK | open-sdk | `https://github.com/Lameguy64/PSn00bSDK/` |
| Open SDK docs | PSn00bSDK docs | sdk-doc | `https://www.breck-mckye.com/psnoobsdk-docs/setup.html` |
| Dev portal | psx.dev | community-index | `https://www.psx.dev/` |
| Dev resources | psx.dev resources | community-index | `https://www.psx.dev/resources` |
| Getting started | psx.arthus | tutorial | `https://psx.arthus.net/starting.html` |
| Ghidra RE guide | Tokimeki Memorial (Tetracorp) | reverse-engineering-guide | `https://tetracorp.github.io/tokimeki-memorial/methods/decompiling-psx-games.html` |
| Executable format | PSX EXE (RetroReversing) | executable-format | `https://www.retroreversing.com/psx-exe-format` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/playstation/` |

### PlayStation 2

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware spec (primary) | ps2tek | hardware-spec | `https://psi-rockin.github.io/ps2tek/` |
| Developer wiki | ps2devwiki | community-wiki | `https://www.psdevwiki.com/ps2/Main_Page` |
| Open SDK | ps2sdk | open-sdk | `https://github.com/ps2dev/ps2sdk` |
| Dev org | ps2dev | community-index | `https://github.com/ps2dev` |
| Ghidra loader | ghidra-emotionengine-reloaded | tooling | `https://github.com/chaoticgd/ghidra-emotionengine-reloaded` |
| Programming notes | Michael Kohn | tutorial | `https://www.mikekohn.net/software/playstation2.php` |
| RE intro | RetroReversing | reverse-engineering-guide | `https://www.retroreversing.com/playstation2` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/playstation-2/` |

### PlayStation 3

| Topic | Source | Quality | URL |
|---|---|---|---|
| Developer wiki | ps3devwiki | community-wiki | `https://www.psdevwiki.com/ps3/` |
| Executable format | SELF/SPRX | executable-format | `https://www.psdevwiki.com/ps3/SELF_-_SPRX` |
| Emulator source | RPCS3 | emulator-source | `https://github.com/RPCS3/rpcs3` |
| Emulator docs | RPCS3 wiki | emulator-docs | `https://wiki.rpcs3.net/` |
| Emulator dev info | RPCS3 dev | emulator-docs | `https://github.com/rpcs3/rpcs3/wiki/developer-information` |
| GPU/RSX docs | Nucleus (rsx) | hardware-spec | `https://github.com/AlexAltea/nucleus/blob/master/docs/technical/ps3/gpu.md` |
| System/LV1/LV2 | Nucleus (system) | hardware-spec | `https://github.com/AlexAltea/nucleus/blob/master/docs/technical/ps3/system.md` |
| Cell SDK docs (3.1) | IBM Cell SDK | archive | `https://www.ps3linux.net/ps3-filez/cellsdk-docs/3.1/` |
| Cell SDK docs (3.0) | IBM Cell SDK | archive | `https://arcb.csc.ncsu.edu/~mueller/cluster/ps3/SDK3.0/docs/` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/playstation-3/` |

### GameCube

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware spec (primary) | YAGCD | hardware-spec | `https://www.gc-forever.com/yagcd/` |
| Hardware spec (original) | YAGCD (hitmen) | hardware-spec | `https://hitmen.c02.at/files/yagcd/index.html` |
| Decomp tooling | decomp-toolkit | tooling | `https://github.com/encounter/decomp-toolkit` |
| Emulator source | Dolphin | emulator-source | `https://github.com/dolphin-emu/dolphin` |
| Emulator docs | Dolphin FAQ | emulator-docs | `https://dolphin-emu.org/docs/faq/` |
| Executable format | DOL (WiiBrew) | executable-format | `https://wiibrew.org/wiki/DOL` |
| Module format | REL | executable-format | `https://www.metroid2002.com/retromodding/wiki/REL_%28File_Format%29` |
| Homebrew SDK source | libogc | open-sdk | `https://github.com/devkitPro/libogc` |
| Homebrew SDK docs | libogc docs | sdk-doc | `https://libogc.devkitpro.org/` |
| Graphics API docs | libogc GX | sdk-doc | `https://devkitpro.org/wiki/libogc/GX` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/gamecube/` |

### Original Xbox

| Topic | Source | Quality | URL |
|---|---|---|---|
| Developer wiki | XboxDevWiki | community-wiki | `https://www.xboxdevwiki.net/` |
| Curated index | xbox1.dev | community-index | `https://xbox1.dev/` |
| Executable format | XBE | executable-format | `https://xboxdevwiki.net/Xbe` |
| Open SDK | nxdk | open-sdk | `https://github.com/XboxDev/nxdk` |
| Emulator source | Cxbx-Reloaded | emulator-source | `https://github.com/cxbx-reloaded/cxbx-reloaded` |
| Emulator source | XQEMU | emulator-source | `https://github.com/xqemu/xqemu` |
| Ghidra loader | ghidra-xbe | tooling | `https://github.com/XboxDev/ghidra-xbe` |
| Filesystem | FATX (Free60) | filesystem | `https://free60.org/System-Software/Systems/FATX/` |
| Modding doc | ConsoleMods | community-wiki | `https://consolemods.org/wiki/Xbox:Original_Xbox_Mods_Wiki` |
| Patching | XBE Patching | community-wiki | `https://consolemods.org/wiki/Xbox:Patching_XBEs` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/xbox/` |

### Game Boy / Color (DMG/MGB/SGB/GBC)

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware spec (primary) | Pan Docs | hardware-spec | `https://gbdev.io/pandocs/` |
| Hardware spec source | Pan Docs (GitHub) | hardware-spec | `https://github.com/gbdev/pandocs` |
| Curated index | gbdev.io | community-index | `https://gbdev.io/` |
| Full technical ref | GBCTR (gekkio) | hardware-spec | `https://gekkio.fi/files/gb-docs/gbctr.pdf` |
| Technical ref source | gb-ctr (GitHub) | hardware-spec | `https://github.com/Gekkio/gb-ctr` |
| Open doc project | gbdoc (mGBA) | hardware-spec | `https://mgba-emu.github.io/gbdoc/` |
| Emulator spec mirror | BGB Pan Docs | hardware-spec | `https://bgb.bircd.org/pandocs.htm` |

### Game Boy Advance (GBA/SP/Micro)

| Topic | Source | Quality | URL |
|---|---|---|---|
| Hardware spec (primary) | GBATEK | hardware-spec | `https://problemkaputt.de/gbatek.htm` |
| Hardware spec (mGBA) | GBATEK (mGBA markdown) | hardware-spec | `https://mgba-emu.github.io/gbatek/` |
| Hardware spec source | GBATEK (GitHub) | hardware-spec | `https://github.com/mgba-emu/gbatek` |
| Hardware spec (GBA only) | GBATEK (rust-console) | hardware-spec | `https://rust-console.github.io/gbatek-gbaonly/` |
| Programming guide | TONC | tutorial | `https://www.coranac.com/projects/tonc/` |
| Programming guide (modern) | TONC (gbadev) | tutorial | `https://gbadev.net/tonc/` |
| Dev portal | gbadev.net | community-index | `https://gbadev.net/` |
| Docs index | gbadev (org) | community-index | `https://www.gbadev.org/docs.php` |
| Open template | libtonc-template | open-sdk | `https://github.com/gbadev-org/libtonc-template` |
| Architecture overview | Copetti | hardware-spec | `https://www.copetti.org/writings/consoles/game-boy-advance/` |

## Theory + Tooling (cross-platform)

| Topic | Source | Quality | URL |
|---|---|---|---|
| M68K disassembler/splitter | sega2asm | tooling | `https://github.com/hansbonini/sega2asm` |
| Object diffing | objdiff | tooling | `https://github.com/encounter/objdiff` |
| Ghidra | Ghidra (NSA) | tooling | `https://github.com/NationalSecurityAgency/ghidra` |
| Decomp.me | decomp.me | tooling | `https://decomp.me/` |
| RetroReversing | retroreversing.com | community-index | `https://www.retroreversing.com/` |
| Copetti (all platforms) | Copetti.org | hardware-spec | `https://www.copetti.org/writings/consoles/` |

## recompclass curriculum

Structured course teaching static recompilation from theory to practice.
Clone once via repomixr: `repo_url: https://github.com/sp00nznet/recompclass`
Then read modules from `.rehamr/repos/sp00nznet-recompclass/repo/`.

| # | Topic | Quality | Path |
|---|---|---|---|
| 1 | What is static recomp? | tutorial | `units/unit-1-foundations/module-01-what-is-static-recomp/` |
| 2 | Binary formats (ELF, PE, ROM) | tutorial | `units/unit-1-foundations/module-02-binary-formats/` |
| 3 | CPU architectures overview | tutorial | `units/unit-1-foundations/module-03-cpu-architectures/` |
| 4 | Reading assembly (x86, MIPS, ARM, Z80, PPC) | tutorial | `units/unit-1-foundations/module-04-reading-assembly/` |
| 5 | Tooling (Ghidra, Capstone) | tutorial | `units/unit-1-foundations/module-05-tooling-ghidra-capstone/` |
| 6 | Control-flow recovery | tutorial | `units/unit-2-core-techniques/module-06-control-flow-recovery/` |
| 7 | Lifting fundamentals | tutorial | `units/unit-2-core-techniques/module-07-lifting-fundamentals/` |
| 8 | First lift (Z80 → C) | tutorial | `units/unit-2-core-techniques/module-08-first-lift-z80/` |
| 9 | Game Boy recomp | tutorial | `units/unit-3-first-targets/module-09-game-boy/` |
| 10 | NES / 6502 | tutorial | `units/unit-3-first-targets/module-10-nes-6502/` |
| 11 | SNES / 65816 | tutorial | `units/unit-3-first-targets/module-11-snes/` |
| 12 | GBA / ARM7 | tutorial | `units/unit-3-first-targets/module-12-gba-arm7/` |
| 13 | DOS / x86 real-mode | tutorial | `units/unit-3-first-targets/module-13-dos/` |
| 14 | Indirect calls + jump tables | tutorial | `units/unit-4-pipeline-essentials/module-14-indirect-calls/` |
| 15 | Hardware shims + SDL2 | tutorial | `units/unit-4-pipeline-essentials/module-15-hardware-shims/` |
| 17 | Build systems + CMake | tutorial | `units/unit-5-pipeline-mastery/module-17-build-systems/` |
| 18 | Testing + validation | tutorial | `units/unit-5-pipeline-mastery/module-18-testing-validation/` |
| 19 | Optimization | tutorial | `units/unit-5-pipeline-mastery/module-19-optimization/` |
| 20 | N64 / MIPS | tutorial | `units/unit-6-console-architectures/module-20-n64-mips/` |
| 21 | N64 RSP/RDP deep dive | tutorial | `units/unit-6-console-architectures/module-21-n64-rsp-rdp/` |
| 22 | GameCube / PowerPC | tutorial | `units/unit-6-console-architectures/module-22-gamecube-ppc/` |
| 23 | Wii / Broadway | tutorial | `units/unit-6-console-architectures/module-23-wii-broadway/` |
| 24 | Dreamcast / SH-4 | tutorial | `units/unit-6-console-architectures/module-24-dreamcast-sh4/` |
| 25 | PS2 / Emotion Engine | tutorial | `units/unit-6-console-architectures/module-25-ps2-ee/` |
| 26 | Saturn / Dual SH-2 | tutorial | `units/unit-7-advanced-targets/module-26-saturn-sh2/` |
| 27 | Xbox / Win32 | tutorial | `units/unit-7-advanced-targets/module-27-xbox-win32/` |
| 28 | Xbox 360 / Xenon PPC | tutorial | `units/unit-7-advanced-targets/module-28-xbox360-xenon/` |
| 29 | GPU pipeline translation | tutorial | `units/unit-7-advanced-targets/module-29-gpu-translation/` |
| 30 | PS3 / Cell | tutorial | `units/unit-8-extreme-targets/module-30-ps3-cell/` |
| 31 | Multi-threaded recomp | tutorial | `units/unit-8-extreme-targets/module-31-multithreaded-recomp/` |

Also available in the clone: `docs/glossary.md`, `docs/tool-setup.md`,
`docs/architecture-reference/` (12 CPU quickrefs), `docs/cheat-sheets/`,
`docs/recommended-reading.md`, and `labs/` (50 hands-on exercises).

## Known recomp projects (reference layouts)

### N64

| Project | URL |
|---|---|
| Zelda64Recomp | `https://github.com/N64Recomp/Zelda64Recomp` |
| BanjoRecomp (Banjo-Kazooie) | `https://github.com/BanjoRecomp/BanjoRecomp` |
| Kirby64Recomp | `https://github.com/Kirby64Ret/Kirby64Recomp` |
| MarioKart64Recomp | `https://github.com/sonicdcer/MarioKart64Recomp` |
| BM64Recomp (Bomberman 64) | `https://github.com/RevoSucks/BM64Recomp` |
| BMHeroRecomp (Bomberman Hero) | `https://github.com/RevoSucks/BMHeroRecomp` |
| dino-recomp (Dinosaur Planet) | `https://github.com/DinosaurPlanetRecomp/dino-recomp` |
| drmario64_recomp_plus | `https://github.com/theboy181/drmario64_recomp_plus` |
| DNZHRecomp (Densha de Go!) | `https://github.com/sonicdcer/DNZHRecomp` |
| Goemon64Recomp | `https://github.com/klorfmorf/Goemon64Recomp` |
| Quest64-Recomp | `https://github.com/Rainchus/Quest64-Recomp` |
| Starship (Star Fox 64) | `https://github.com/HarbourMasters/Starship` |

### Other platforms

| Project | Platform | URL |
|---|---|---|
| MK (Super Mario Kart) | SNES | `https://github.com/sp00nznet/mk` |
| Mario Paint | SNES | `https://github.com/sp00nznet/mariopaint` |
| UnleashedRecomp | Xbox 360 | `https://github.com/hedge-dev/UnleashedRecomp` |
| Daytona-XBLA-Recomp | Xbox 360 | `https://github.com/Subarasheese/daytona-xbla-recomp` |
| TiP-Recomp (Time is Priceless) | Xbox 360 | `https://github.com/SolarCookies/TiP-Recomp` |
| KameoRePowered | Xbox 360 | `https://github.com/birabittoh/KameoRePowered` |
| reNut | Xbox 360 | `https://github.com/masterspike52/reNut` |
| NaughtyBear_ReStuff | Xbox 360 | `https://github.com/MaxDeadBear/NaughtyBear_ReStuff` |
| Re-Cherry | Xbox 360 | `https://github.com/MaxDeadBear/Re-Cherry` |
| reDAHM | Xbox 360 | `https://github.com/masterspike52/reDAHM` |
| reblue | Xbox 360 | `https://github.com/zolaware/reblue` |

## Source quality key

| Label | Meaning |
|---|---|
| `hardware-spec` | Primary hardware reference — registers, memory maps, CPU docs |
| `community-wiki` | Community-maintained documentation — cross-verify |
| `emulator-source` | Emulator implementation — correct by behavior, not documentation |
| `emulator-docs` | Emulator developer documentation — generally accurate |
| `open-sdk` | Open-source SDK — authoritative for API behavior |
| `sdk-doc` | SDK documentation — accurate for API contracts |
| `recompiler-source` | Recompiler implementation — correct by output, not docs |
| `runtime-source` | Runtime library — correct by behavior |
| `reverse-engineering-guide` | Tutorial/technique — helpful patterns, not canonical |
| `executable-format` | File format docs — verify against headers |
| `filesystem` | Filesystem layout docs — verify on actual images |
| `tutorial` | Educational material — principles, not specifications |
| `archive` | Historical/docs archive — may be outdated |
| `tooling` | Development tool — verify CLI flags against `--help` |
| `community-index` | Curated link collection — follow to primary sources |
