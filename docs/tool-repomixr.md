# repomixr

Built-in tool that clones a GitHub repository and packs all source files into
a single XML file for LLM consumption. Always available alongside `bash`,
`read_file`, `write_file`, and `edit_file`.

## What it does

1. Runs `git clone --depth 1` into `.rehamr/repos/<owner>-<repo>/repo/`
2. Walks all text files (UTF-8 validated, <512KB, binary extensions filtered)
3. Packs into XML at `.rehamr/repos/<owner>-<repo>/packed.xml`
4. Returns the path + file count + size for the LLM to `read_file`

The clone is kept on disk — subsequent calls can `git pull` instead of
re-cloning. Delete `.rehamr/repos/<name>/` manually when done.

## Output format

```xml
<?xml version="1.0" encoding="UTF-8"?>
<repository url="https://github.com/user/repo" branch="main" files="3">
  <directory_structure>
    include/
      header.h
    src/
      main.c
    README.md
  </directory_structure>
  <file path="README.md" lines="1">
    <![CDATA[readme content]]>
  </file>
  <file path="include/header.h" lines="5">
    <![CDATA[header content]]>
  </file>
  <file path="src/main.c" lines="42">
    <![CDATA[source content]]>
  </file>
  <instruction>
    <![CDATA[Focus on the rendering pipeline.]]>
  </instruction>
</repository>
```

### Directory structure

A tree view is built from sorted file paths and placed before the file
contents. Directories are marked with `/`. This gives the LLM an instant
navigation map before reading files.

### Instruction block

If `.rehamr/repomix-instruction.md` exists, its content is injected into
the `<instruction>` block. Use this to provide project-specific guidance
("focus on the renderer", "ignore audio code", etc.). If the file doesn't
exist, the block is omitted.

## Flags

| Parameter | Default | Description |
|---|---|---|
| `repo_url` | *(required)* | GitHub repo URL |
| `branch` | `main` | Branch or tag to checkout |
| `remove_comments` | `false` | Strip `//` and `#` comments from source files |
| `remove_empty_lines` | `false` | Remove blank lines from output |
| `show_line_numbers` | `false` | Prefix each line with `    N |` |
| `compress` | `false` | Collapse multiple whitespace to single space |

## File filtering

- **Skipped by extension:** `.exe`, `.dll`, `.so`, `.o`, `.png`, `.jpg`,
  `.zip`, `.tar`, `.gz`, `.pdf`, `.pyc`, `.class`, `.jar`, and ~20 more
- **Skipped by size:** files > 512 KB
- **Skipped by encoding:** non-UTF-8 files
- **Skipped by path:** `.git/` directory
- **Included:** everything else (`.c`, `.h`, `.cpp`, `.py`, `.go`, `.rs`,
  `.md`, `.txt`, `.yaml`, `.toml`, `.json`, `.xml`, `.js`, `.ts`, `.html`,
  `.css`, `.sh`, `.bat`, `.ps1`, etc.)

## Usage

The LLM calls it when it needs a reference codebase. Example parameters:

```json
{
  "repo_url": "https://github.com/KuromeSan/RecompHamr",
  "branch": "main",
  "remove_comments": false,
  "remove_empty_lines": false,
  "show_line_numbers": false,
  "compress": false
}
```

After the call, the LLM receives the output path and uses `read_file` to ingest
the packed XML. For deep-diving into individual files, the LLM can also
`read_file` specific paths from the clone directory.

## Cache location

```
.rehamr/repos/
├── owner-repo/
│   ├── repo/           # git clone
│   └── packed.xml      # packed output
└── another-repo/
    ├── repo/
    └── packed.xml
```

The cache is never automatically cleaned. Delete directories manually when
no longer needed.
