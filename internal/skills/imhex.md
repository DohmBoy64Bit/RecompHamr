# imhex

Use this skill when working with hex data, binary file formats, or the ImHex
Pattern Language. ImHex is a hex editor for reverse engineers with a custom
pattern language for parsing and highlighting binary structures.

> ImHex is a GUI tool the user runs. You cannot drive it directly — you
> guide the user through interactive analysis using ImHex's Pattern Language,
> then process the exported results.

## Knowledge base

ImHex documentation is LLM-queryable via GitBook:

```
GET https://docs.werwolv.net/imhex/readme.md?ask=<question>
```

Use `bash curl -sL 'https://docs.werwolv.net/imhex/readme.md?ask=<question>'`
to query specific topics on demand. Never load the full docs — query only
what you need for the current task.

The Pattern Language database is at:
[https://github.com/WerWolv/ImHex-Patterns](https://github.com/WerWolv/ImHex-Patterns) —
clone via `repomixr` to search known file format definitions.

Complete LLM-friendly docs index:
[https://docs.werwolv.net/imhex/llms.txt](https://docs.werwolv.net/imhex/llms.txt)

## How to use

1. When analyzing an unknown binary format, search ImHex Patterns for known
   format definitions matching the magic bytes or structure
2. Guide the user to open the binary in ImHex and apply the matching pattern
3. The user can export parsed data (JSON/CSV) for you to `read_file`
4. For unknown formats, use the Pattern Language syntax to draft a pattern
   based on your file-format-reversing methodology — the user tests it in ImHex
5. Pair with `/skill file-format-reversing` for the methodology;
   `/skill evidence-mode` for classification of findings

## Pattern language quick reference

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

The Pattern Language is a full data-parsing DSL. For syntax details, query
the docs: `bash curl -sL 'https://docs.werwolv.net/imhex/readme.md?ask=<syntax question>'`

## When to use

Any task involving binary file inspection — magic byte identification,
structure mapping, format validation, or comparing a file against known
format definitions from the ImHex Patterns database.
