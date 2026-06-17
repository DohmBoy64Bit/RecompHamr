# sega2asm Setup

Sega Genesis ROM disassembler and splitter via hansbonini/sega2asm. 68000/Z80
disassembly, 49 compression formats, graphics extraction, VDP register hints.

## Dependencies

- **sega2asm** — installed via Go or from source
- **Go 1.21+** — for building the binary

## Install

1. `go install github.com/hansbonini/sega2asm@latest`
   or: `git clone ... && cd sega2asm && go build -o sega2asm .`
2. Ensure `sega2asm-mcp` is on PATH (or set `RECOMPHAMR_MCP_SEGA2ASM_COMMAND`)

## Enable

1. Start recomphamr — connect with `/mcp connect sega2asm`
2. Run `/skill sega2asm` — unlocks `sega2asm.*` tools
3. Verify: `/mcp tools sega2asm`

Refer to [common.md](mcp-common.md) for shared env vars and management commands.
