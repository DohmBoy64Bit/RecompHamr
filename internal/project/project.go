package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InitRE(projectDir string) error {
	root := filepath.Join(projectDir, ".rehamr")
	dirs := []string{
		"logs",
		"evidence",
		"functions",
		"formats/parsers",
		"formats/tests",
		"recomp",
		"decomp",
		"skills",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return err
		}
	}
	files := map[string]string{
		"PROJECT.md": `# Project

## Goal
TODO

## Source of truth
TODO

## Toolchain
TODO
`,
		"REPHAMR_STATE.md": rephamrStateTemplate,
		"EVIDENCE.md": `# Evidence

Add confirmed facts only. Each entry should include the source file, command output, log, symbol, offset, or other evidence.
`,
		"HYPOTHESES.md": `# Hypotheses

Unconfirmed ideas live here until proven or disproven.
`,
		"BLOCKERS.md": `# Blockers

Track missing files, missing tools, build failures, runtime blockers, and required user decisions.
`,
		"CHANGELOG.md": fmt.Sprintf("# Changelog\n\n## %s\n- Initialized RecompHAMR evidence workspace.\n", time.Now().Format("2006-01-02")),
		"COMMANDS.md": `# Commands

Record useful commands and what they prove.
`,
		"TOOLCHAIN.md": `# Toolchain

Document compilers, SDKs, decompilers, reverse-engineering tools, scripts, and versions.
`,
		"MODELS.md": `# Models

Record local model profiles, context settings, and known behavior.
`,
		"functions/inventory.csv":       "address_or_symbol,name,status,classification,evidence_source,confidence,notes\n",
		"functions/game_logic.md":       "# Game / Project Logic Functions\n\n",
		"functions/runtime_platform.md": "# Runtime / Platform / Middleware Functions\n\n",
		"functions/unknown.md":          "# Unknown Functions\n\n",
		"formats/inventory.md":          "# File Format Inventory\n\n",
		"formats/hypotheses.md":         "# File Format Hypotheses\n\n",
		"recomp/bridge_audit.md":        "# Bridge / Stub Audit\n\n",
		"recomp/runtime_gaps.md":        "# Runtime Gaps\n\n",
		"recomp/thread_trace.md":        "# Thread / Message Trace\n\n",
		"recomp/build_matrix.md":        "# Build Matrix\n\n",
		"decomp/compiler_detection.md":  "# Compiler Detection\n\n",
		"decomp/matching_status.md":     "# Matching Status\n\n",
		"decomp/symbols.md":             "# Symbols\n\n",
		"skills/active.md":              "# Active Skills\n\n",
	}
	for rel, body := range files {
		full := filepath.Join(root, rel)
		if _, err := os.Stat(full); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func StatusRE(root string) (string, error) {
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("%s not initialized; run /init-re", root)
	}
	wanted := []string{"PROJECT.md", "EVIDENCE.md", "HYPOTHESES.md", "BLOCKERS.md", "CHANGELOG.md", "functions/inventory.csv", "formats/inventory.md", "recomp/runtime_gaps.md"}
	var b strings.Builder
	b.WriteString("RecompHAMR project status\n")
	b.WriteString("=========================\n")
	for _, rel := range wanted {
		full := filepath.Join(root, rel)
		data, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(&b, "\n## %s\nmissing\n", rel)
			continue
		}
		text := strings.TrimSpace(string(data))
		if len(text) > 1800 {
			text = runeSafeTruncate(text, 1800) + "\n...truncated..."
		}
		fmt.Fprintf(&b, "\n## %s\n%s\n", rel, text)
	}
	return b.String(), nil
}

// runeSafeTruncate truncates s to at most maxLen bytes without splitting a
// multi-byte UTF-8 rune. If the cut point lands in the middle of a rune, the
// partial rune is dropped.
func runeSafeTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	for maxLen > 0 && maxLen < len(s) {
		if s[maxLen]&0xc0 != 0x80 {
			return s[:maxLen]
		}
		maxLen--
	}
	return s[:0]
}

var rephamrStateTemplate = `# RecompHAMR Project State
> Auto-maintained by agent. Read automatically at session start.

## Quick Rules
1. Evidence first — classify, cite, save to .rehamr/evidence/
2. Never hand-edit generated output as primary fix — fix metadata/config first
3. Verify paths before assuming; record them below
4. Read files before acting — never guess file contents
5. Command outputs are evidence — copy verbatim, don't summarize

## Current Phase
- **Track**: (reversing | decompilation | static-recomp | analysis | general)
- **Phase**: PHASE_SETUP
- **Current goal**:

## Project Info
- **Project / target**:
- **Goal**:
- **Source of truth**:
- **Toolchain**:

## Workspace Paths
- **Project root**:
- **Evidence dir**: .rehamr/evidence/
- **Repos cache**: .rehamr/repos/
- **Key reference files**:

## Active Commands
` + "``" + "`" + `
# Record verbatim commands that work — do not reconstruct from memory
` + "``" + "`" + `

## Blockers
| Issue | Status | Evidence |
|-------|--------|----------|

## Function / Symbol Ledger
| Name | Address | Classification | Confidence | Source |
|------|---------|----------------|------------|--------|

## Learned Patterns
> Session close: write patterns ("X causes Y, fix with Z"), not raw event logs.
- 

## Session Log
| Date | Summary |
|------|---------|
| ` + time.Now().Format("2006-01-02") + ` | Initialized by /init-re |`

