# file-format-reversing

Use this skill for archives, assets, maps, scripts, configs, model formats, texture formats, audio banks, save files, and custom binary/text formats.

Evidence rules:
- Every claimed field needs an offset, sample, observed value, code reference, tool output, or repeated pattern.
- Unknown bytes stay unknown.
- Do not name fields by vibes. Use `unknown_XX`, `tentative_*`, or HYPOTHESIS labels.
- Parser tests must use real sample files where possible.

Preferred output structure:
- `.rehamr/formats/inventory.md`
- `.rehamr/formats/hypotheses.md`
- `.rehamr/formats/samples.md`
- `.rehamr/formats/parsers/`
- `.rehamr/formats/tests/`

Workflow:
1. Inventory sample files and hashes.
2. Identify magic, endian clues, sizes, offsets, counts, tables, compression, and references.
3. Write a tiny parser/dumper before a full editor.
4. Validate parser output against multiple samples.
5. Keep failures and unknown ranges documented.

