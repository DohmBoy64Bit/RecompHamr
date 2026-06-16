package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
