# sega2asm

Sega Genesis ROM disassembler and splitter via hansbonini/sega2asm. 49
compression formats, 68000/Z80 disassembly, graphics/audio extraction.

## What it enables

- Full ROM split from YAML config (plan + execute)
- Compression detection across 49 formats, 297 games
- 68000 + Z80 disassembly with symbols, labels, VDP hints
- Graphics extraction (PNG), PCM audio (WAV), text decode
- Dry-run segment plan without executing full split

## Setup

1. `go install github.com/hansbonini/sega2asm@latest`
2. Ensure `sega2asm-mcp` on PATH (or set `RECOMPHAMR_MCP_SEGA2ASM_COMMAND`)

## When to use

Sega Genesis/Mega Drive ROM analysis — compression discovery,
ROM segment splitting, 68000/Z80 disassembly, asset extraction.
Pairs with `bizhawk` for runtime validation.
