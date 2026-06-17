# imhex

ImHex hex editor skill. Guides the LLM to use ImHex Pattern Language for
binary format analysis, the Patterns database for format discovery, and the
LLM-queryable documentation for syntax reference.


## Kickoff

`/skill imhex` — then "Search ImHex Patterns for this file format" and paste hexdump of first 64 bytes.

## What it teaches

- Query ImHex docs on demand via LLM-queryable GitBook API
- Search the ImHex Patterns database for known file format definitions
- Guide the user through interactive Pattern Language analysis
- Draft Pattern Language structures for unknown binary formats
- Pair with `file-format-reversing` for methodology and `evidence-mode`
  for classification

## What it references

- `/skill file-format-reversing` — methodology for unknown formats
- `/skill evidence-mode` — classification of findings
- `repomixr` — clone ImHex-Patterns database
- `recomp_reference` — fetch ImHex documentation pages

## When to use

Binary file analysis — magic byte identification, format validation,
structure mapping, or comparing files against the ImHex Patterns database.
