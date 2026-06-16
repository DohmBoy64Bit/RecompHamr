# sega2asm

Sega Genesis ROM disassembler and splitter via hansbonini/sega2asm.
Load this skill to unlock `sega2asm.*` tools for 68000/Z80 disassembly,
graphics extraction, and compression detection.

## What it enables

- Full ROM split from YAML config (plan + execute)
- 49 compression format detection and decompression
- 68000 + Z80 disassembly with labels and symbols
- Graphics extraction (PNG sheets), PCM audio (WAV), text decode
- VDP register + command hint decoding
- Segment plan dry-run without executing full split

## Setup

1. Install sega2asm: `go install github.com/hansbonini/sega2asm@latest`
2. Ensure `sega2asm-mcp` is on PATH (or set `RECOMPHAMR_MCP_SEGA2ASM_COMMAND`)
3. Start recomphamr — auto-connects at launch
4. Load `/skill sega2asm` — unlocks `sega2asm.*` tools

## When to use

Sega Genesis/Mega Drive ROM analysis — discovering compression types,
splitting ROMs into segments, disassembling 68000/Z80 code, extracting
graphics and audio assets. Pairs with `bizhawk` for runtime validation.
