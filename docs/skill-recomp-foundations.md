# recomp-foundations

Static recompilation theory router. Points the LLM to specific recompclass
modules for foundational knowledge — binary formats, CFG recovery, lifting,
indirect calls, CPU architecture references, GPU pipeline translation.


## Kickoff

`/skill recomp-foundations` — then "I need to understand indirect calls. Find the relevant recompclass module."

## What it teaches

Nothing directly — it's a map. When the LLM encounters an unknown concept,
this skill routes it to the right recompclass module. 32 modules mapped
across 8 units covering 12 CPU architectures.

## How it works

1. Clone recompclass once via repomixr
2. Skill provides a table: topic → module file path
3. LLM uses `read_file` to read the relevant module
4. Always verify module content against tool output and real evidence

## When to use

Any decompilation/recompilation project where the LLM needs foundational
theory it doesn't already have. Extensible with additional knowledge base
sources in the future.
