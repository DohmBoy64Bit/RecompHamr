# imhex

Use this skill when working with hex data, binary file formats, or the ImHex
Pattern Language for file-format analysis and interactive binary inspection.

> ImHex is a GUI tool the user runs. You cannot drive it directly — you
> guide the user through interactive analysis using ImHex's Pattern Language,
> then process the exported results.

## What it enables

- Query ImHex documentation on demand via LLM-queryable GitBook API
- Search the [ImHex Patterns database](https://github.com/WerWolv/ImHex-Patterns)
  for known file format definitions matching magic bytes or structure
- Guide the user through Pattern Language analysis of unknown binary formats
- Process exported JSON/CSV parse results via `read_file`
- Draft Pattern Language structures based on file-format-reversing methodology

## When to use

Use this tool for:
- Binary file inspection — magic byte identification, structure mapping
- Discovering known format definitions from the ImHex Patterns database
- Drafting Pattern Language structures for unknown formats
- Validating format hypotheses against real sample files

Do not use it for:
- Automated binary parsing without user interaction — ImHex is GUI-driven
- File-format methodology without ImHex (use `file-format-reversing`)
- Binary analysis via Ghidra or programmatic tools (use `ghidra-mcp` or `bash`)

## Boot / Connection Check

1. Verify ImHex documentation is reachable:
   ```bash
   curl -sL 'https://docs.werwolv.net/imhex/readme.md?ask=What%20is%20ImHex%3F' | head -20
   ```
2. Verify the Patterns database is accessible:
   `bash curl -sI https://github.com/WerWolv/ImHex-Patterns`
3. Clone the Patterns database via `repomixr` if offline access is needed:
   `repo_url: https://github.com/WerWolv/ImHex-Patterns`
4. If documentation or database is unreachable, use the LLM-friendly index:
   `https://docs.werwolv.net/imhex/llms.txt`

## Setup

1. User installs ImHex from [imhex.werwolv.net](https://imhex.werwolv.net)
2. User opens the target binary file in ImHex
3. User loads patterns from the ImHex-Patterns database or your drafted pattern
4. User runs the pattern and exports results (JSON/CSV)
5. You process the exported results via `read_file`

## Evidence Protocol

Every analysis should record:
- target file name and hash (if available)
- pattern used (name + source: database, drafted, or modified)
- exported result path (JSON/CSV)
- interpretation status: CONFIRMED / HYPOTHESIS / TODO / BLOCKED

Save evidence to:
- `.rehamr/formats/` — inventory, hypotheses, parser output
- `.rehamr/evidence/imhex_analysis.md` — pattern results and format findings
- `REPHAMR_STATE.md` — active format investigations, confirmed structures

## Knowledge Base

Query ImHex documentation on demand — never load the full docs:

```bash
curl -sL 'https://docs.werwolv.net/imhex/readme.md?ask=<question>'
```

Additional references:
- Pattern Language database: [ImHex-Patterns](https://github.com/WerWolv/ImHex-Patterns)
- LLM-friendly docs index: [llms.txt](https://docs.werwolv.net/imhex/llms.txt)

## Common Operations

| Operation | Command / Action | Output | Notes |
|---|---|---|---|
| Query ImHex docs | `bash curl -sL 'https://docs.werwolv.net/imhex/readme.md?ask=<q>'` | Doc excerpt | Ask specific, self-contained questions |
| Search known patterns | User browses ImHex-Patterns clone or GitHub | Pattern file list | Match by magic bytes or file extension |
| Draft pattern | Write Pattern Language code; user pastes into ImHex | Parse errors or structured output | Test iteratively on the target file |
| Export results | User: File → Export → JSON/CSV | Structured data | Read via `read_file` for classification |
| Pattern language syntax | Query docs: `?ask=Pattern+Language+syntax` | Syntax reference | On-demand, not pre-loaded |

## Pattern Language Quick Reference

```rust
// Magic bytes check
u8 magic[4] @ 0x00;
if (magic != "FILE") { return; }

// Structured parsing
u32 file_count @ 0x04;
struct Entry {
    u32 offset;
    u32 size;
    char name[32];
};
Entry entries[file_count] @ 0x08;
```

## Guardrails

1. ImHex output is evidence, but interpretation needs the file-format-reversing
   methodology — structure alone doesn't explain purpose.
2. A single-sample pattern match is a hypothesis — confirm against ≥2 samples.
3. Drafted patterns are untested until the user runs them — mark as HYPOTHESIS
   until the user confirms results.
4. ImHex cannot automate batch processing — for bulk format analysis, use
   programmatic tools via `bash`.
5. Never assume the user has ImHex installed or running — verify before
   giving pattern instructions.

## Failure Handling

| Failure | Meaning | Next Step |
|---|---|---|
| Pattern doesn't parse | Syntax error in pattern code | Query docs for syntax; fix pattern; user retests |
| Pattern parses but wrong output | Logic error — offset/type mismatch | Compare hex dump vs pattern structure; adjust offsets |
| No matching pattern in database | Unknown format | Draft pattern from scratch using `file-format-reversing` methodology |
| GitBook API unreachable | Network issue or docs offline | Use `llms.txt` index or cloned Patterns database |
| User doesn't have ImHex | Not installed | Recommend install from imhex.werwolv.net; use `bash` hexdump in meantime |

## Session Close

1. Update `.rehamr/formats/` with confirmed structure discoveries.
2. Update `REPHAMR_STATE.md` with active pattern investigations and results.
3. Report: patterns tested, confirmed structures, failed patterns, next 3 hex regions or formats to investigate.
