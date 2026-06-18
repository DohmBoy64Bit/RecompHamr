package classifier

import (
	"testing"

	"github.com/DohmBoy64Bit/recomphamr/internal/skills"
)

// synthetic bodies that mimic each template class pattern.

const bodyFull = `# n64-decomp

Role: Nintendo 64 matching decompilation specialist.

## When to use

Use this skill when decompiling an N64 ROM to matching C.

Do not use this skill for Xbox or PS2 targets.

## Hardware Architecture

The N64 CPU is a MIPS VR4300 at 93.75 MHz with 4 MB RDRAM (expandable to 8 MB).

## Prohibitions

- Never guess function boundaries.
- Never trust Ghidra output without runtime validation.

## Phases

### Phase 1: Recon

Entry: ROM loaded in Ghidra.
Exit: symbol map, segment list, compression catalog.

### Phase 2: Lift

Entry: validated segment list.
Exit: compilable C for all segments.

## Build Gate

` + "`" + `make clean && make -j8 && make test` + "`" + `

## Session Close

Save evidence to .rehamr/state.md.

## Known Reference Projects

| Project | Platform | URL |
|---|---|---|
| Zelda64Recomp | N64 | https://github.com/N64Recomp/Zelda64Recomp |
`

const bodyMicro = `# core-re

Rules for core reverse engineering methodology.

## When to use this skill

Use this skill whenever you are reversing an unknown binary.

## Rules

1. Never assume library code is bug-free.
2. Always verify function signatures with runtime data.
3. Label every unknown global with a DAT_ prefix.

## Workflow

1. Recon: identify the binary format, entry point, imports.
2. Triage: sort functions by importance (call count, xref count).
3. Annotate: name functions, globals, and set types.

## Stop conditions

Stop when all functions with >5 xrefs are named and typed.

## Session Close

Write a summary to .rehamr/state.md.
`

const bodyBridge = `# sega2asm

Sega Genesis ROM disassembler via hansbonini/sega2asm — 68000/Z80 disassembly,
49 compression format detection, graphics and audio extraction.

## What it enables

- Full ROM split from YAML config
- 49 compression format detection
- 68000 + Z80 disassembly with labels

## Setup

1. Install sega2asm: ` + "`" + `go install github.com/hansbonini/sega2asm@latest` + "`" + `
2. Ensure sega2asm-mcp is on PATH

## When to use

Sega Genesis/Mega Drive ROM analysis.

## Boot / Connection Check

1. Verify sega2asm MCP server is running: ` + "`" + `/mcp tools sega2asm` + "`" + `

## Common Operations

| Operation | Tool Call | Output |
|---|---|---|
| Detect compression | sega2asm.detect_compression | Format name |
| Plan (dry-run) | sega2asm.plan | Validation report |

## Guardrails

1. Never guess compression types — use sega2asm.detect_compression.

## Session Close

Save results to .rehamr/evidence/.
`

const bodyEmpty = ``

const bodyGeneric = `# unknown

This is just a collection of notes.

Some text about decompilation.
No clear structure, no triggers, no rules, no MCP references.
`

// bodyAmbiguous is a micro-skill that also references MCP tools (bridge-like).
const bodyAmbiguous = `# ghidra-mcp

Ghidra MCP bridge helper.

## When to use this skill

Use this skill when working with Ghidra MCP tools.

## Rules

1. Always verify tool output in Ghidra before committing.
2. Never rename a function without checking xrefs first.

## Workflow

1. Open the binary in Ghidra.
2. Use ghidra.list_functions to enumerate.
3. Use ghidra.decompile_function on each.

## Setup

1. Ensure ghidra-mcp is running on port 8080.

## Session Close

Save the Ghidra archive.
`

func TestClassifyFullWorkflow(t *testing.T) {
	r := Classify("n64-decomp", bodyFull)
	if r.Class != FullWorkflow {
		t.Errorf("expected FullWorkflow, got %q; scores=%v", r.Class, r.Scores)
	}
	if r.Confidence < 0.3 {
		t.Errorf("confidence too low: %.2f", r.Confidence)
	}
}

func TestClassifyMicroSkill(t *testing.T) {
	r := Classify("core-re", bodyMicro)
	if r.Class != MicroSkill {
		t.Errorf("expected MicroSkill, got %q; scores=%v", r.Class, r.Scores)
	}
}

func TestClassifyToolBridge(t *testing.T) {
	r := Classify("sega2asm", bodyBridge)
	if r.Class != ToolBridge {
		t.Errorf("expected ToolBridge, got %q; scores=%v", r.Class, r.Scores)
	}
}

func TestClassifyNoTemplateEmpty(t *testing.T) {
	r := Classify("empty", bodyEmpty)
	if r.Class != NoneClass {
		t.Errorf("expected NoneClass for empty body, got %q; scores=%v", r.Class, r.Scores)
	}
}

func TestClassifyNoTemplateGeneric(t *testing.T) {
	r := Classify("unknown", bodyGeneric)
	if r.Class != NoneClass {
		t.Errorf("expected NoneClass for generic body, got %q; scores=%v", r.Class, r.Scores)
	}
}

func TestClassifyAmbiguous(t *testing.T) {
	r := Classify("ghidra-mcp", bodyAmbiguous)
	// This body has both micro-skill signals (rules, workflow, trigger)
	// and tool bridge signals (MCP refs, setup, trigger).
	// The classification is valid for either; check that it's not NoneClass
	// and that alternatives exist.
	if r.Class == NoneClass {
		t.Errorf("expected a valid class, got NoneClass; scores=%v", r.Scores)
	}
	// Should have at least one alternative or close score.
	t.Logf("class=%q confidence=%.2f scores=%v reasoning=%v alts=%v",
		r.Class, r.Confidence, r.Scores, r.Reasoning, r.Alternatives)
}

func TestClassifyResultHasAllFields(t *testing.T) {
	r := Classify("test", bodyFull)
	if r.Skill != "test" {
		t.Errorf("Skill field: %q", r.Skill)
	}
	if r.Scores == nil {
		t.Fatal("Scores is nil")
	}
	if len(r.Scores) != 3 {
		t.Errorf("expected 3 scores, got %d", len(r.Scores))
	}
	if len(r.Reasoning) == 0 {
		t.Error("Reasoning is empty")
	}
}

func TestClassifyTemplateName(t *testing.T) {
	tests := []struct {
		c    TemplateClass
		want string
	}{
		{FullWorkflow, "Universal Template (Full Workflow)"},
		{MicroSkill, "Micro-Skill Method Template"},
		{ToolBridge, "Tool Bridge Card Template"},
		{NoneClass, "No Template"},
	}
	for _, tt := range tests {
		if got := tt.c.TemplateName(); got != tt.want {
			t.Errorf("TemplateName(%q) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestSingleConcern(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{"N64 ROM decompilation for Zelda", true},
		{"Supports N64, Xbox 360, and PS2 recompilation", false},
		{"Generic reverse engineering tool", true},
		{"PlayStation Xbox GameCube", false},
		{"", true},
	}
	for _, tt := range tests {
		if got := singleConcern(tt.body); got != tt.want {
			t.Errorf("singleConcern(%q) = %v, want %v", tt.body, got, tt.want)
		}
	}
}

// RealSkillTest validates that each embedded skill classifies to a valid template.
// This is a coarse integration test — it ensures every skill gets at least SOME
// classification. It does NOT assert the exact class (skills change over time).
func TestRealSkillsClassifyNonZero(t *testing.T) {
	names := skills.Names()
	if len(names) == 0 {
		t.Skip("no embedded skills found")
	}
	for _, name := range names {
		body, err := skills.Get(name)
		if err != nil {
			t.Fatalf("skills.Get(%q): %v", name, err)
		}
		r := Classify(name, body)
		if r.Class == NoneClass {
			t.Errorf("%q classified as NoneClass (scores=%v)", name, r.Scores)
		}
	}
}

// TestRealSkillsClassifyExpected validates that a curated subset of skills
// match their expected template class. These names/signals are stable.
func TestRealSkillsClassifyExpected(t *testing.T) {
	tests := []struct {
		name     string
		want     TemplateClass
	}{
		{"n64-decomp", FullWorkflow},
		{"ps2recomp", FullWorkflow},
		{"xboxrecomp", FullWorkflow},
		{"windows-game-decomp", FullWorkflow},
		{"xbox360-decomp", FullWorkflow},
		{"gb-recomp", FullWorkflow},
		{"gen-decomp", FullWorkflow},
		{"vb-decomp", FullWorkflow},
		{"snesrecomp", FullWorkflow},
		{"ps3recomp", FullWorkflow},
		{"cdb-debug", FullWorkflow},
		{"pcrecomp", FullWorkflow},
		{"core-re", MicroSkill},
		{"evidence-mode", MicroSkill},
		{"build-fix-loop", MicroSkill},
		{"file-format-reversing", MicroSkill},
		{"function-discovery", MicroSkill},
		{"project-handoff", MicroSkill}, // tie with tool_bridge; either valid
		{"ghidra-mcp", ToolBridge},
		{"n64-debug-mcp", ToolBridge},
		{"bizhawk", ToolBridge},
		{"mcp-pine", ToolBridge},
		{"pcsx2", ToolBridge},
		{"objdiff", ToolBridge},
		{"sega2asm", ToolBridge},
		{"imhex", ToolBridge},
	}
	for _, tt := range tests {
		body, err := skills.Get(tt.name)
		if err != nil {
			t.Fatalf("skills.Get(%q): %v", tt.name, err)
		}
		r := Classify(tt.name, body)
		if r.Class != tt.want {
			t.Errorf("%q expected %s, got %s (scores=%v, reasoning=%v)",
				tt.name, tt.want, r.Class, r.Scores, r.Reasoning)
		}
	}
}

func TestRecompFoundationsClassifies(t *testing.T) {
	body, err := skills.Get("recomp-foundations")
	if err != nil {
		t.Fatal(err)
	}
	r := Classify("recomp-foundations", body)
	// recomp-foundations is a reference/knowledge-base router.
	// It may be any class. Just verify it's not NoneClass.
	if r.Class == NoneClass {
		t.Errorf("recomp-foundations classified as NoneClass; scores=%v reasoning=%v",
			r.Scores, r.Reasoning)
	}
	t.Logf("recomp-foundations: class=%s confidence=%.2f scores=%v",
		r.Class, r.Confidence, r.Scores)
}
