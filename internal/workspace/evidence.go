package workspace

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const maxStatusFileBytes = 1800

var evidenceNow = time.Now

type writableEvidenceFile interface {
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

var (
	mkdirEvidence    = os.Mkdir
	openEvidenceFile = func(path string, flag int, mode os.FileMode) (writableEvidenceFile, error) {
		return os.OpenFile(path, flag, mode)
	}
	removeEvidence = os.Remove
	openStatusFile = func(path string) (io.ReadCloser, error) { return os.Open(path) }
)

var evidenceDirectories = []string{
	"logs", "evidence", "functions", "formats", "formats/parsers", "formats/tests", "recomp", "decomp",
}

func evidenceFiles(now time.Time) map[string]string {
	return map[string]string{
		"PROJECT.md":                    "# Project\n\n## Goal\nTODO\n\n## Source of truth\nTODO\n\n## Toolchain\nTODO\n",
		StateFileName:                   stateTemplate(now),
		"EVIDENCE.md":                   "# Evidence\n\nAdd confirmed facts only. Cite the source file, command output, log, symbol, offset, or other evidence.\n",
		"HYPOTHESES.md":                 "# Hypotheses\n\nUnconfirmed ideas live here until proven or disproven.\n",
		"BLOCKERS.md":                   "# Blockers\n\nTrack missing files, tools, build failures, runtime blockers, and required user decisions.\n",
		"CHANGELOG.md":                  fmt.Sprintf("# Changelog\n\n## %s\n- Initialized RecompHamr evidence workspace.\n", now.Format("2006-01-02")),
		"COMMANDS.md":                   "# Commands\n\nRecord useful commands and what they prove.\n",
		"TOOLCHAIN.md":                  "# Toolchain\n\nDocument compilers, SDKs, decompilers, reverse-engineering tools, scripts, and versions.\n",
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
	}
}

// InitializeEvidence creates the evidence workspace without overwriting
// existing files. Every created or retained path is tightened through the
// platform-native private-path policy, and links are refused.
func (w *Workspace) InitializeEvidence() error {
	privateRoot := filepath.Join(w.root, ".rehamr")
	if err := ensureEvidenceDirectory(privateRoot); err != nil {
		return err
	}
	for _, relative := range evidenceDirectories {
		if err := ensureEvidenceDirectory(filepath.Join(privateRoot, relative)); err != nil {
			return err
		}
	}
	for relative, content := range evidenceFiles(evidenceNow()) {
		path := filepath.Join(privateRoot, relative)
		if err := ensureEvidenceFile(path, []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

// EvidenceStatus returns a bounded, UTF-8 summary of canonical evidence files.
// Missing files are reported without failing; an unsafe workspace root fails.
func (w *Workspace) EvidenceStatus() (string, error) {
	root := filepath.Join(w.root, ".rehamr")
	info, err := lstatPath(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%s not initialized; run /init-re", root)
	}
	wanted := []string{"PROJECT.md", "EVIDENCE.md", "HYPOTHESES.md", "BLOCKERS.md", "CHANGELOG.md", "functions/inventory.csv", "formats/inventory.md", "recomp/runtime_gaps.md"}
	var out strings.Builder
	out.WriteString("RecompHamr project status\n=========================\n")
	for _, relative := range wanted {
		content, readErr := readStatusFile(filepath.Join(root, relative))
		fmt.Fprintf(&out, "\n## %s\n", relative)
		if readErr != nil {
			out.WriteString("missing or unavailable\n")
			continue
		}
		out.WriteString(content)
		out.WriteByte('\n')
	}
	return out.String(), nil
}

func ensureEvidenceDirectory(path string) error {
	info, err := lstatPath(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := mkdirEvidence(path, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("workspace initialize directory: %w", err)
		}
		info, err = lstatPath(path)
	}
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("workspace initialize: unsafe directory")
	}
	return restrictPath(path, true)
}

func ensureEvidenceFile(path string, content []byte) error {
	info, err := lstatPath(path)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("workspace initialize: unsafe existing file")
		}
		return restrictPath(path, false)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("workspace initialize file: %w", err)
	}
	file, err := openEvidenceFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("workspace initialize file: %w", err)
	}
	remove := true
	defer func() {
		if remove {
			_ = removeEvidence(path)
		}
	}()
	if written, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	} else if written != len(content) {
		_ = file.Close()
		return io.ErrShortWrite
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := restrictPath(path, false); err != nil {
		return err
	}
	remove = false
	return nil
}

func readStatusFile(path string) (string, error) {
	info, err := lstatPath(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > MaxStateBytes {
		return "", errors.New("unavailable")
	}
	file, err := openStatusFile(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaxStateBytes+1))
	if err != nil || len(data) > MaxStateBytes || !utf8.Valid(data) {
		return "", errors.New("unavailable")
	}
	text := strings.TrimSpace(string(data))
	if len(text) > maxStatusFileBytes {
		cut := maxStatusFileBytes
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		text = text[:cut] + "\n...truncated..."
	}
	return text, nil
}

func stateTemplate(now time.Time) string {
	return fmt.Sprintf("# RecompHamr Project State\n> Project-maintained context read at session start.\n\n## Current Phase\n- **Track**: (reversing | decompilation | static-recomp | analysis | general)\n- **Phase**: setup\n- **Current goal**:\n\n## Project Info\n- **Project / target**:\n- **Goal**:\n- **Source of truth**:\n- **Toolchain**:\n\n## Blockers\n| Issue | Status | Evidence |\n|---|---|---|\n\n## Learned Patterns\n-\n\n## Session Log\n| Date | Summary |\n|---|---|\n| %s | Initialized by /init-re |\n", now.Format("2006-01-02"))
}
