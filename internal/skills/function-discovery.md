# function-discovery

Goal: build a high-confidence inventory of functions and classify which are game/project logic versus runtime/platform/middleware code.

Methodology:
1. Start from the strongest function-boundary evidence available: PDB/map/export symbols, ELF symbols, PE `.pdata`, XEX metadata, ROM/splat config, loader metadata.
2. Record entry points: process entry, TLS callbacks, static constructors/destructors, init arrays, reset/boot entry, thread starts.
3. Record direct call targets and cross references.
4. Identify jump tables, switch dispatchers, and indirect-call tables.
5. Identify vtables or object-like dispatch tables when evidence supports them.
6. Classify import thunks, wrappers, platform calls, middleware/library code, game/project logic, data, and unknown.
7. Keep unknown as unknown until stronger evidence appears.

Preferred output files:
- `.rehamr/functions/inventory.csv`
- `.rehamr/functions/game_logic.md`
- `.rehamr/functions/runtime_platform.md`
- `.rehamr/functions/unknown.md`
- `.rehamr/evidence/function_discovery.log`

CSV columns:
`address_or_symbol,name,status,classification,evidence_source,confidence,notes`

