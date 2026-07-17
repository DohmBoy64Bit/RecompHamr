---
name: imhex
description: Design, debug, and validate ImHex Pattern Language definitions for interactive binary-format inspection, and interpret ImHex exports against raw bytes and multiple samples. Use when the user is actively working with ImHex patterns or exported results; do not use for generic format reversing, unattended batch parsers, or unrelated static analysis.
compatibility: Requires a user-controlled ImHex installation for interactive execution; Pattern Language syntax and export behavior must be verified against the installed version.
---
# ImHex pattern analysis

ImHex is a user-controlled interactive tool. Establish the installed version, target-file hash, pattern source/version/license, and the exact hypothesis before drafting or changing a pattern. Consult current official Pattern Language documentation for the installed version; do not invent syntax or rely on remembered examples.

Work from bytes to structure:

1. Preserve a bounded hex view around the candidate field and record offset, address base, byte order, alignment, and file-size constraints.
2. Identify observed constants, counts, offsets, lengths, tags, and repeated records. Separate observations from semantic interpretations.
3. Draft the smallest pattern that can falsify the hypothesis. Add bounds before variable arrays, pointer/offset dereferences, recursion, or computed seeks.
4. Have the user run it in ImHex and return the exact diagnostic or a sanitized structured export. A drafted pattern is unverified until the installed tool accepts it.
5. Compare displayed fields with the original bytes and downstream consumers. Validate across representative positive, boundary, truncated, malformed, and variant samples available to the project; one successful file proves only that file.
6. Promote a field meaning only when independent evidence—code, format documentation, cross-sample invariants, or controlled mutation—supports it.

Before reusing a community pattern, verify its target format/version, license, assumptions, byte order, bounds, and behavior on the project's samples. Do not fetch or install remote patterns merely because this skill was activated, and never treat an extension match as format identity.

For exports, retain the pattern identity and file hash with the JSON/CSV or other result. Check that numeric precision, byte arrays, duplicate names, nested records, and offsets survive the export before relying on it. Sanitize paths, personal data, keys, proprietary payloads, and copyrighted assets from reports.

ImHex patterns are analysis aids, not production parsers. When deterministic batch processing or security-sensitive parsing is required, translate confirmed contracts into a bounded programmatic parser with independent tests rather than automating the GUI.

If an untrusted sample supplies a count, length, offset, pointer, or allocation size, do not emit a pattern from an arbitrary hard-coded cap alone. Derive and enforce file-size and remaining-byte bounds, checked offset-plus-length arithmetic, count-times-entry-size overflow limits, and valid seek ranges; exercise zero, maximum valid, truncated, malformed, and overflow cases. State explicitly that the interactive pattern is not the production parser.

Record verified syntax/documentation version, tested samples, exact diagnostics, confirmed fields, rejected hypotheses, unresolved ranges, and the next falsifiable experiment. Do not claim that a pattern explains field purpose merely because it parses without error.
