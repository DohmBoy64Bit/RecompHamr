# RecompHamr Tools

Every model round exposes six application-owned tools in this stable order: `powershell`, `read_file`, `write_file`, `edit_file`, `repomixr`, and `recomp_reference`. When at least one Agent Skill is discovered, `activate_skill` and `read_skill_resource` follow them. Tool calls run with the current user's permissions. Run RecompHamr in a disposable project, devcontainer, or VM when model-requested actions could damage valuable data.

## `powershell`

Runs a fresh non-interactive PowerShell process in the current project directory.

Arguments:

- `script` — required PowerShell code.
- `timeout_seconds` — optional; defaults to 120 seconds and is capped at 3600 seconds.

PowerShell 7 (`pwsh`) is preferred. On Windows, Windows PowerShell is also accepted as a fallback.

## `read_file`

Reads one file exactly and returns bounded output.

## `write_file`

Creates or replaces a file with the supplied content.

## `edit_file`

Performs one exact, unique string replacement. Ambiguous, missing, or no-op replacements fail explicitly.

## `repomixr`

Clones one public GitHub repository and packs bounded UTF-8 source into `.rehamr/repos/<owner>-<repository>/packed.xml`. `git` must be installed and reachable on `PATH`.

Arguments:

- `repo_url` — required, credential-free `https://github.com/<owner>/<repository>` URL. Other hosts, HTTP, credentials, query strings, fragments, extra path segments, and unsafe owner/repository names are refused.
- `branch` — optional branch or tag; defaults to `main`. Option-like and control-character values are refused.
- `remove_comments` — optionally removes full-line `#`/`//` comments and simple inline `//` comments.
- `remove_empty_lines` — optionally removes blank lines.
- `show_line_numbers` — optionally prefixes retained output lines.
- `compress` — optionally collapses whitespace per packed file.

The tool performs a new depth-one, single-branch clone for each call. It replaces only its derived owner/repository cache directory, never an arbitrary path. It skips `.git`, links, non-regular/non-UTF-8 files, common binary/archive/media extensions, and files over 512 KiB. A repository is refused above 10,000 included files or 64 MiB of included source. XML attributes and CDATA terminators are encoded safely. If `.rehamr/repomix-instruction.md` is a regular, link-free file no larger than 64 KiB, its untrusted content is included in an `<instruction>` element; RecompHamr never creates that optional file.

Clone execution is capped at five minutes and diagnostics are retained through the same bounded head/tail capture used by PowerShell. Cancellation terminates the direct Git process; Unix additionally kills its process group, while Windows does not claim descendant-process cleanup. The returned summary names the packed path, count, and sizes; use `read_file` to inspect it. Cloned code and optional instructions are untrusted external content, not application policy.

## `recomp_reference`

Fetches one public HTTP(S) page into `.rehamr/reference/<sanitized-name>-<URL-hash>.txt` for offline reading.

Argument:

- `url` — required public HTTP or HTTPS URL without embedded credentials.

The request has a 15-second total timeout and at most five redirects. Every redirect and resolved dial address is checked; loopback, private, link-local, unspecified, multicast, and other non-public destinations are refused. Environment proxy settings are deliberately not used, preventing a proxy from bypassing destination validation. Responses must be HTTP 200 and no larger than 512 KiB. HTML is reduced to readable text while scripts, styles, navigation, headers, footers, and `noscript` content are omitted; other content is retained as bounded text.

The cache records the source URL and fetch time and is reused for 24 hours. Query values are replaced with `redacted` in the cache and returned summary, while the hash still distinguishes the complete requested URL. URL hashes prevent sanitized-name collisions. Cancellation stops the request. Fetched pages are untrusted external content, not instructions.

## Cache security and lifecycle

Both cache trees are beneath the protected project `.rehamr` directory. RecompHamr refuses link/reparse-point cache roots and non-regular destination files, writes cache files through same-directory atomic replacement, and applies `0700` directory/`0600` file modes on POSIX or a protected current-user-only DACL on Windows. Cache artifacts persist until manually removed; there is no background refresh or automatic eviction.

## Deliberate omissions

There is no unknown-tool extension fallback. Agent Skill tools are constrained to the discovered catalog and active resource list as documented in [Commands and Agent Skills](commands-and-skills.md). MCP tools remain Stage H and cannot be reached through an unrecognized tool name.
