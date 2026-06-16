package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInitRECreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	err := InitRE(dir)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, ".rehamr")
	expected := []string{
		"logs", "evidence", "functions", "formats/parsers", "formats/tests",
		"recomp", "decomp", "skills",
	}
	for _, d := range expected {
		info, err := os.Stat(filepath.Join(root, d))
		if err != nil {
			t.Errorf("missing dir %s: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}

func TestInitRECreatesFiles(t *testing.T) {
	dir := t.TempDir()
	err := InitRE(dir)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, ".rehamr")
	files := []string{
		"PROJECT.md", "EVIDENCE.md", "HYPOTHESES.md", "BLOCKERS.md",
		"CHANGELOG.md", "COMMANDS.md", "TOOLCHAIN.md", "MODELS.md",
		"functions/inventory.csv", "functions/game_logic.md",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("missing file %s: %v", f, err)
		}
	}
}

func TestInitREIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := InitRE(dir); err != nil {
		t.Fatal(err)
	}
	if err := InitRE(dir); err != nil {
		t.Fatalf("InitRE should be idempotent: %v", err)
	}
}

func TestStatusRENotInitialized(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".rehamr")
	_, err := StatusRE(dir)
	if err == nil {
		t.Fatal("expected error for non-existent dir")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' error, got: %s", err)
	}
}

func TestStatusREShowsFiles(t *testing.T) {
	dir := t.TempDir()
	if err := InitRE(dir); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, ".rehamr")
	status, err := StatusRE(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "RecompHAMR project status") {
		t.Error("missing header")
	}
	if !strings.Contains(status, "PROJECT.md") {
		t.Error("missing PROJECT.md section")
	}
	if !strings.Contains(status, "EVIDENCE.md") {
		t.Error("missing EVIDENCE.md section")
	}
	if !strings.Contains(status, "Initialized RecompHAMR evidence workspace") {
		t.Error("CHANGELOG should contain initialization entry")
	}
}

func TestInitRECreatesStateFile(t *testing.T) {
	dir := t.TempDir()
	err := InitRE(dir)
	if err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(dir, ".rehamr", "REPHAMR_STATE.md")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("REPHAMR_STATE.md not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "RecompHAMR Project State") {
		t.Errorf("state file missing title: %s", content)
	}
	if !strings.Contains(content, "Quick Rules") {
		t.Errorf("state file missing Quick Rules: %s", content)
	}
	if !strings.Contains(content, "Current Phase") {
		t.Errorf("state file missing Current Phase: %s", content)
	}
	if !strings.Contains(content, "Blockers") {
		t.Errorf("state file missing Blockers: %s", content)
	}
	if !strings.Contains(content, "Learned Patterns") {
		t.Errorf("state file missing Learned Patterns: %s", content)
	}
	if !strings.Contains(content, "Session Log") {
		t.Errorf("state file missing Session Log: %s", content)
	}
}

func TestStateTemplateHasQuickRules(t *testing.T) {
	expectedRules := []string{
		"Evidence first",
		"Never hand-edit generated output",
		"Verify paths before assuming",
		"Read files before acting",
		"Command outputs are evidence",
	}
	for _, rule := range expectedRules {
		if !strings.Contains(repohamrStateTemplate, rule) {
			t.Errorf("state template missing rule: %s", rule)
		}
	}
}

func TestStateTemplateHasSessionLogDate(t *testing.T) {
	if !strings.Contains(repohamrStateTemplate, time.Now().Format("2006-01-02")) {
		t.Errorf("state template should contain today's date, got: %s", repohamrStateTemplate)
	}
}

func TestStateTemplateHasBlockersTable(t *testing.T) {
	if !strings.Contains(repohamrStateTemplate, "| Issue | Status | Evidence |") {
		t.Errorf("state template missing blockers table header: %s", repohamrStateTemplate)
	}
}

func TestStateTemplateHasFunctionLedgerTable(t *testing.T) {
	if !strings.Contains(repohamrStateTemplate, "| Name | Address | Classification | Confidence | Source |") {
		t.Errorf("state template missing function ledger table: %s", repohamrStateTemplate)
	}
}

func TestStateTemplateTokenBudget(t *testing.T) {
	// State template should be under ~500 chars to keep token budget low
	// (~4 chars per token on average, so 500 chars = ~125 tokens)
	if len(repohamrStateTemplate) > 1500 {
		t.Errorf("state template too large: %d chars (target <1500)", len(repohamrStateTemplate))
	}
}
