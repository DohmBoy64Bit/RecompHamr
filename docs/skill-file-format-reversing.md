# file-format-reversing

Binary file format analysis skill. Load when reverse-engineering unknown file
formats, archives, asset containers, or data structures.


## Kickoff

`/skill file-format-reversing` — then "Analyze this unknown binary file at [PATH]."

## What it teaches

- Start from known offsets: magic bytes, headers, size fields
- Map structure before interpreting data
- Every claimed field needs an offset, sample value, observed value,
  code reference, or tool output as evidence
- Unknown bytes stay unknown until evidence supports a name

## When to use

When encountering undocumented file formats — game archives, proprietary
containers, custom serialization formats, or any binary blob that needs
structural analysis.
