---
name: file-format-reversing
description: Reverse engineer an unknown or partially known binary file/container format using real samples, offsets, parsers, and cross-sample evidence. Use for archives, assets, maps, saves, scripts, models, textures, audio banks, and custom containers; do not use for implementing a fully documented standard format or merely converting known data with an existing library.
compatibility: RecompHamr with representative sample files and a suitable parser language or inspection tool.
---
# Binary format reversing

Inventory samples before interpreting bytes. Record immutable hashes, sizes, provenance, and known producer/version facts without copying sensitive sample contents into logs.

Work from structure to meaning:

1. Identify magic/version bytes, endianness clues, alignment, sizes, offsets, counts, tables, checksums, compression signatures, and repeated records.
2. Track every claimed field by byte range, encoded type/width, observed values, samples, and evidence source.
3. Label unsupported ranges and meanings `unknown` or `hypothesis`; do not infer semantics from a single plausible value.
4. Build the smallest bounded parser/dumper that proves structural navigation before writing an editor or converter.
5. Validate every invariant across representative real samples, including truncated, malformed, empty, boundary-sized, and distinct-version cases when available.
6. Record parser failures and sample differences as format evidence. Never coerce a conflicting sample to preserve a theory.

Synthetic fixtures can test parser safety and known invariants, but cannot establish the real format by themselves. A second sample increases confidence only when it varies the relevant field or behavior. Use producer/consumer code or runtime traces to promote semantic hypotheses when static byte patterns are insufficient.

For parser safety, validate arithmetic before allocation or seeking; bound counts, offsets, sizes, recursion, decompression, and output; reject overlap or out-of-file ranges unless the format proves they are legal; preserve endianness explicitly.

When initialized, keep the catalog in `.rehamr/formats/inventory.md`, hypotheses and promotion criteria in `.rehamr/formats/hypotheses.md`, parsers under `.rehamr/formats/parsers/`, and focused tests under `.rehamr/formats/tests/`.

Finish with confirmed layout, hypotheses, unknown ranges, failing samples, parser limits, and the next discriminating evidence—not a claim of complete understanding while bytes or variants remain unexplained.
