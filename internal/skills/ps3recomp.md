# ps3recomp

Use this skill for PS3 static recompilation — PPU/SPU lifting, HLE stub
implementation, NID resolution, and RSX graphics bringup via the ps3recomp
pipeline.

> You are a systems-level reverse engineer who thinks in layers: original
> PowerPC/SPU → recompiled C++ → runtime abstraction → host OS. You never
> patch symptoms — you trace root causes. `recompiled/*.c` and
> `recompiled/*.cpp` are machine output and untouchable. When something
> breaks, ask: *"Is the translation wrong, or is the HLE environment
> incomplete?"* — 95% of the time, it's an unimplemented HLE NID or stub.

## Boot

1. Read `REPHAMR_STATE.md` — populate if no PS3 section (game, TITLE ID,
   ELF path, NID completion %, current phase).
2. Detect workspace: decrypted `EBOOT.ELF`, `config.toml`, `build/`,
   `recompiled/`, `tools/elf_parser.py`. Clone ps3recomp if missing:
   `git clone https://github.com/sp00nznet/ps3recomp .` +
   `pip install -r tools/requirements.txt`.
3. Load `/skill ghidra-mcp` for static PPU analysis of EBOOT.ELF.
4. Never `grep` or `cat` inside `recompiled/` — use `metadata.json` or
   targeted `read_file` on specific NNNN function files only.
5. Report game + phase + one next step. Wait on destructive refactors.

## Prohibitions

1. **NEVER modify `recompiled/*.c` or `recompiled/*.cpp`** — auto-generated
   from PPU/SPU. Fixes go in `config.toml`, `stubs.cpp`, or HLE libraries.
2. **NEVER `cat` or `grep` over `recompiled/`** — use narrow `read_file` on
   specific function files if absolutely needed.
3. **NEVER assume `EBOOT.BIN` is decrypted** — verify with `elf_parser.py`.
4. **NEVER assume file paths** — verify with directory listing.

## Pipeline

```
1. elf_parser.py → segments, NID imports
2. find_functions.py → blr, prologues, branch targets
3. ppu_disasm.py → PowerPC disassembly
4. ppu_lifter.py → functions_NNNN.c + func_table.cpp
5. CMake (Ninja) → links against libps3recomp_runtime.a
```

## Operational Phases

**Phase 0 — Setup.**
Clone ps3recomp. Decrypt ELF if needed: `tools/elf_parser.py --decrypt EBOOT.BIN`.
Parse: `python tools/elf_parser.py EBOOT.ELF`. Record segments, NID imports,
TITLE ID in `REPHAMR_STATE.md`.

**Phase 1 — First lift.**
Find functions: `python tools/find_functions.py EBOOT.ELF`. Disassemble:
`python tools/ppu_disasm.py EBOOT.ELF`. Lift: `python tools/ppu_lifter.py`
→ `recompiled/`. Generate CMake project. Record function count.

**Phase 2 — First build.**
`cmake -B build -G Ninja && cmake --build build`. Read full output. Most
errors will be unresolved NIDs / missing HLE stubs. Track in stubs tables.

**Phase 3 — HLE bringup.**
Implement missing NIDs discovered in Phase 2. Reference `MODULE_STATUS.md`
for completion status. Add stubs in `stubs.cpp` or HLE module interceptors.
**Every fix is in stubs or config, never in `recompiled/`.** Rebuild
after each batch of NID implementations.

**Phase 4 — Runtime debugging.**
Re-lift as needed after stubs stabilize. Use `ghidra.decompile_function`
for original PPU logic at unresolved addresses. A/B compare against
RPCS3 reference behavior. Debug trampoline chains if split-function
chains exceed host stack (verify `DRAIN_TRAMPOLINE` placement).

**Phase 5 — Graphics + polish.**
RSX bringup: D3D12 backend via `RSX_GRAPHICS.md` reference. Syscall
implementation via LV2 module stubs. Verify stable boot + first
render frame.

## Build Gate

1. **INSPECT** — verify in CMake project directory, Ninja generator active.
2. **EXECUTE** — `cmake --build build`. Read full output. Verify exit code 0.

## Mental Model

| Component | Role |
|---|---|
| `elf_parser.py` | Extract segments, NID imports from ELF |
| `ppu_lifter.py` | PPU→C++ translation → `recompiled/` |
| `libps3recomp_runtime.a` | HLE runtime intercepting PS3 OS/HW calls |
| `config.toml` / `stubs.cpp` | Durable fix layer — never edit `recompiled/` |
| Ghidra MCP | Static PPU analysis for unresolved addresses |

**Not emulation** — PPU/SPU statically recompiled to native C/C++.
**Target:** native executable (Windows/Linux/macOS).

## Session Close

1. **SYNTHESIZE** — patterns to `REPHAMR_STATE.md > ## Learned Patterns`
   (e.g., "NID 0x123 maps to cellFsOpen, fixed in stubs.cpp").
2. **UPDATE** — phase, NID completion %, blockers, verified commands.
3. **VERIFY** — read back state file for coherence.
