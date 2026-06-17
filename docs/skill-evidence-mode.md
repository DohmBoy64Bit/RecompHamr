# evidence-mode

Strict evidence classification skill. Forces the LLM to separate facts from
guesses.


## Kickoff

`/skill evidence-mode` — then "Classify all findings in the current function ledger."

## What it teaches

- Classify every finding as CONFIRMED, HYPOTHESIS, TODO, or BLOCKED
- Never promote claims to confirmed without direct evidence
- Never rename functions, structs, or symbols based on guessing
- Preserve existing evidence unless a stronger source disproves it
- Include exact paths, commands, offsets, hashes, or log snippets as citation

## When to use

When accuracy matters over speed — function identification, symbol naming,
file format analysis, or any task where a wrong classification cascades into
wasted work.
