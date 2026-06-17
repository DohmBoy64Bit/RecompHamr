# sega2asm

Sega Genesis ROM disassembler and splitter via hansbonini/sega2asm —
68000/Z80 disassembly, 49 compression format detection, graphics and audio
extraction, VDP register/command decoding, and segment plan validation.

> Pairs with `bizhawk` for runtime validation — sega2asm provides static
> analysis (ROM split, disasm, compression), bizhawk provides dynamic proof
> (memory r/w, runtime comparison). The `gen-decomp` skill orchestrates both.

## What it enables

- Full ROM split from YAML config — plan (dry-run validate) then execute
- 49 compression format detection and decompression covering 297 games
- 68000 + Z80 disassembly with labels, symbols, and VDP register hints
- VDP register and command hint decoding (`vdp_regs`, `vdp_cmds`)
- Graphics extraction as PNG sheets, PCM audio as WAV, text with charmap decode
- Segment plan dry-run without executing full split (save time, validate first)
- Symbols file in multiple formats (name=addr, addr:name, colon-separated)

## When to use

Use this tool for:
- Sega Genesis/Mega Drive ROM analysis — initial reconnaissance
- Discovering compression types at specific ROM offsets
- Splitting ROMs into segments: header, m68k code, z80 sound, gfx, pcm, text
- Building and validating YAML segment configs via dry-run
- Extracting graphics and audio assets from ROMs
- Annotating disassembly with VDP register/command hints

Do not use it for:
- Runtime validation (use `bizhawk` for emulator verification)
- Manual hex editing or interactive binary inspection (use `imhex`)
- Non-Genesis platforms (use platform-specific tools)
- Full decompilation methodology without the skill (use `gen-decomp`)

## Boot / Connection Check

1. Verify sega2asm MCP server is running:
   - `/mcp tools sega2asm` — lists all available tools
   - Server shows `Connected` on the splash screen or `/mcp` output
2. Verify sega2asm binary is available:
   - `which sega2asm` or verify the binary exists on PATH
3. If unavailable:
   - Check `/mcp status sega2asm` for connection state
   - Install: `go install github.com/hansbonini/sega2asm@latest`
   - Ensure `sega2asm-mcp` is on PATH or `RECOMPHAMR_MCP_SEGA2ASM_COMMAND` is set

## Setup

1. Install sega2asm: `go install github.com/hansbonini/sega2asm@latest`
2. Ensure `sega2asm-mcp` is on PATH (or set `RECOMPHAMR_MCP_SEGA2ASM_COMMAND`)
3. Start recomphamr — connect with `/mcp connect sega2asm`
4. Load `/skill sega2asm` — unlocks `sega2asm.*` tools
5. Verify: `/mcp tools sega2asm`

## Evidence Protocol

Every ROM analysis should record:
- ROM SHA-1, region (NTSC/PAL/Japan), ROM size
- compression types detected at specific offsets (catalog for reuse)
- segment boundaries (start/end addresses per segment)
- symbol addresses verified via bizhawk runtime comparison
- interpretation status: CONFIRMED / HYPOTHESIS / TODO / BLOCKED

Save evidence to:
- `.rehamr/evidence/genesis_recon.md` — ROM info, compression catalog
- `.rehamr/evidence/genesis_split_log.md` — segment split results
- `config.yaml` — validated segment definitions (source of truth)
- `REPHAMR_STATE.md` — active ROM, phase, segment count

## Common Operations

| Operation | Tool Call | Output | Notes |
|---|---|---|---|
| Detect compression | `sega2asm.detect_compression` | Format name + confidence | Run at suspected asset offsets first |
| Plan (dry-run) | `sega2asm.plan` | Validation report | Validate config before full split |
| Full split | `sega2asm.run` | `asm/` + `assets/` output | Run after plan passes validation |
| List supported compression | `sega2asm.list_compressions` | 49 format names | Reference for unknown compression |
| Validate config | `sega2asm.validate_config` | Lint errors | Check YAML before plan or run |

CLI fallback (when MCP tools are unavailable):

```bash
sega2asm -c config.yaml -s symbols.txt -t charmap.tbl -v   # full split
sega2asm -c config.yaml --dry-run                           # validate config
```

## Guardrails

1. **Never guess compression types** — use `sega2asm.detect_compression` at
   unknown offsets. 49 formats exist across 297 games; guessing wastes time.
2. **Never assume segment boundaries** — use `sega2asm.plan` (dry-run) to
   validate config before full split.
3. **Never hand-edit disassembly output** — fix the YAML config or symbol
   file, re-run the split. The generated `.asm` is output, not source.
4. **Never trust disassembly without runtime validation** — use `bizhawk` to
   verify values at runtime addresses.
5. **Validate config before full split** — a full ROM split can take minutes;
   dry-run validation catches errors in seconds.
6. **Catalog compression types early** — detected compression at one offset
   often repeats at other offsets in the same ROM.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| `sega2asm-mcp` not on PATH | Server binary missing | `go install github.com/hansbonini/sega2asm@latest` |
| `detect_compression` returns unknown | Format not in the 49 supported | Document as UNKNOWN; compare raw bytes manually; file an upstream issue |
| `plan` returns validation errors | Segments overlap or invalid | Fix YAML config boundaries; re-run plan |
| `run` fails on a segment | Segment config error or ROM corruption | Check segment start/end; verify ROM hash; re-run plan on failing segment |
| Disassembly missing labels | Symbols file incomplete | Add symbols from bizhawk verification; re-run split |
| Server Disconnected | sega2asm-mcp not running | Reconnect; verify binary on PATH |

## Session Close

1. Save split results and compression catalog to `.rehamr/evidence/`.
2. Update `config.yaml` with validated segment definitions.
3. Update `REPHAMR_STATE.md` with ROM info, phase, segment count, compression types.
4. Report: segments split, compression types identified, symbols added, VDP hints applied, remaining segments or offsets to analyze.
