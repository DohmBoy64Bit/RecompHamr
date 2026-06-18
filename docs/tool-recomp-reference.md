# recomp_reference

Built-in tool that fetches and caches web pages locally for offline reading.
Always available alongside `bash`, `read_file`, `write_file`, `edit_file`, and
`repomixr`.

## What it does

1. Fetches a URL over HTTP(S) with a 15-second timeout
2. If the response is HTML, extracts readable text (strips scripts, styles,
   nav, footer, header, noscript)
3. Saves a plain-text copy to `.rehamr/reference/<host>-<path>.txt`
4. Returns the cached file path for the LLM to `read_file`

Subsequent requests to the same URL return the cached copy for 24 hours
without re-fetching.

## Output

```
Fetched https://example.com/doc
  → .rehamr/reference/example.com-doc.txt · 12 KB

Use read_file to inspect. Cached for 24 hours.
```

Or for cached content:

```
Cached (from 2026-06-17 14:30)
  → .rehamr/reference/example.com-doc.txt

Use read_file to inspect.
```

## Cache location

```
.rehamr/reference/
├── en.wikipedia.org-wiki_N64
├── github.com-user_repo_blob_main_doc
└── ...
```

The cache expires after 24 hours per entry. Delete files manually to force
a fresh fetch.

## Usage

The LLM calls it when the `recomp-foundations` skill points to a reference
URL. Example parameter:

```json
{
  "url": "https://en.wikipedia.org/wiki/N64"
}
```
