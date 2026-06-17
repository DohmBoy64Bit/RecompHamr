# n64-decomp

N64 matching decompilation and N64Recomp static PC port methodology skill.
Teaches the LLM the full pipeline from ROM recon through build verification.

## What it teaches

- Two-track system: Track A (matching decomp) + Track B (N64Recomp static port)
- 9 operational phases with exit criteria for each track
- Boot sequence: detect workspace, load supporting skills, report before acting
- Build Gate: inspect → verify → execute → read before claiming success
- Guardrails: four fix tools, circuit breaker, degradation canary
- Debug format: Phase, Structural Cause, Evidence, Address Mapping, Fix, etc.

## What it references

This skill instructs the LLM to load supporting skills:
- `/skill ghidra-mcp` — unlocks `ghidra.*` tools for static analysis
- `/skill n64-debug-mcp` — unlocks `n64-debug-mcp.*` tools for guest runtime debug
- `/skill core-re` — RE workflow discipline
- `/skill evidence-mode` — evidence classification

State is maintained in `REPHAMR_STATE.md` (persistent memory).

## Kickoff

`/skill n64-decomp` — then "I have baserom.z64 at [PATH]. Start Track A, Phase 0."

## When to use

Any N64 decompilation or recompilation project — from fresh ROM recon through
matching C functions to native PC port bringup. Pair with `/skill ghidra-mcp`
and `/skill n64-debug-mcp` for full tool access.
