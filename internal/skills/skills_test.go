package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNamesReturnsEmbeddedSkills(t *testing.T) {
	names := Names()
	if len(names) < 1 {
		t.Fatal("Names() returned empty slice")
	}
	want := []string{"build-fix-loop", "core-re", "evidence-mode", "file-format-reversing",
		"function-discovery", "ghidra-mcp", "n64-debug-mcp", "project-handoff"}
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Names() missing %q, got %v", w, names)
		}
	}
}

func TestNamesIsSorted(t *testing.T) {
	names := Names()
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Errorf("Names() not sorted: %q >= %q", names[i-1], names[i])
		}
	}
}

func TestResolveExactMatch(t *testing.T) {
	name, err := Resolve("core-re")
	if err != nil {
		t.Fatal(err)
	}
	if name != "core-re" {
		t.Errorf("expected core-re, got %q", name)
	}
}

func TestResolveCaseInsensitive(t *testing.T) {
	name, err := Resolve("Core-RE")
	if err != nil {
		t.Fatal(err)
	}
	if name != "core-re" {
		t.Errorf("expected core-re, got %q", name)
	}
}

func TestResolveWithMDSuffix(t *testing.T) {
	name, err := Resolve("core-re.md")
	if err != nil {
		t.Fatal(err)
	}
	if name != "core-re" {
		t.Errorf("expected core-re, got %q", name)
	}
}

func TestResolveUnknown(t *testing.T) {
	_, err := Resolve("nonexistent-skill")
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestGetEmbedded(t *testing.T) {
	body, err := Get("core-re")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "# core-re") {
		t.Errorf("body should contain skill title, got: %s", body)
	}
}

func TestGetUnknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMarkdownNoActive(t *testing.T) {
	out := ListMarkdown(nil)
	if !strings.Contains(out, "Built-in RE skills:") {
		t.Errorf("missing header: %s", out)
	}
	if strings.Contains(out, "* ") {
		t.Errorf("should have no active marker when none active: %s", out)
	}
}

func TestListMarkdownWithActive(t *testing.T) {
	out := ListMarkdown([]string{"core-re"})
	if !strings.Contains(out, "* core-re") {
		t.Errorf("core-re should be marked active: %s", out)
	}
	lines := strings.Split(out, "\n")
	activeCount := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "* ") {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Errorf("expected exactly 1 active skill, got %d: %s", activeCount, out)
	}
}

func TestCustomDirMergesNames(t *testing.T) {
	dir := t.TempDir()
	SetCustomDir(dir)
	defer SetCustomDir("")

	os.WriteFile(filepath.Join(dir, "my-skill.md"), []byte("# My Skill\n"), 0o644)

	names := Names()
	found := false
	for _, n := range names {
		if n == "my-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Names() missing custom skill, got %v", names)
	}
}

func TestCustomDirGetDiskFirst(t *testing.T) {
	dir := t.TempDir()
	SetCustomDir(dir)
	defer SetCustomDir("")

	os.WriteFile(filepath.Join(dir, "core-re.md"), []byte("# Custom Override\n"), 0o644)

	body, err := Get("core-re")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Custom Override") {
		t.Errorf("custom skill should override embedded, got: %s", body)
	}
}

func TestListMarkdownLabelsCustom(t *testing.T) {
	dir := t.TempDir()
	SetCustomDir(dir)
	defer SetCustomDir("")

	os.WriteFile(filepath.Join(dir, "my-skill.md"), []byte("# My Skill\n"), 0o644)

	out := ListMarkdown(nil)
	if !strings.Contains(out, "my-skill (custom)") {
		t.Errorf("custom skill should have (custom) label: %s", out)
	}
}

func TestSetCustomDirEmpty(t *testing.T) {
	SetCustomDir("")
	defer SetCustomDir("")
	names := Names()
	foundEmbedded := false
	for _, n := range names {
		if n == "core-re" {
			foundEmbedded = true
			break
		}
	}
	if !foundEmbedded {
		t.Error("embedded skills should still be available with empty custom dir")
	}
}

func TestResolveCustomSkill(t *testing.T) {
	dir := t.TempDir()
	SetCustomDir(dir)
	defer SetCustomDir("")

	os.WriteFile(filepath.Join(dir, "my-skill.md"), []byte("# My Skill\n"), 0o644)

	name, err := Resolve("my-skill")
	if err != nil {
		t.Fatal(err)
	}
	if name != "my-skill" {
		t.Errorf("expected my-skill, got %q", name)
	}
}

func TestResolveSucceedsWhenCustomDirIsMissing(t *testing.T) {
	SetCustomDir(filepath.Join(t.TempDir(), "nonexistent"))
	defer SetCustomDir("")

	name, err := Resolve("core-re")
	if err != nil {
		t.Fatal(err)
	}
	if name != "core-re" {
		t.Errorf("expected core-re, got %q", name)
	}
}

func TestIsEmbeddedTrue(t *testing.T) {
	if !IsEmbedded("core-re") {
		t.Error("core-re should be embedded")
	}
	if !IsEmbedded("n64-decomp") {
		t.Error("n64-decomp should be embedded")
	}
}

func TestIsEmbeddedFalseForCustom(t *testing.T) {
	dir := t.TempDir()
	SetCustomDir(dir)
	defer SetCustomDir("")

	os.WriteFile(filepath.Join(dir, "my-custom.md"), []byte("# custom\n"), 0o644)

	if IsEmbedded("my-custom") {
		t.Error("custom skill should not be reported as embedded")
	}
}

func TestIsEmbeddedFalseForUnknown(t *testing.T) {
	if IsEmbedded("nonexistent-skill") {
		t.Error("unknown skill should not be embedded")
	}
}

func TestConcurrentSetCustomDirAndNames(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			SetCustomDir(t.TempDir())
			Names()
		}
		close(done)
	}()
	SetCustomDir("")
	<-done
	SetCustomDir("")
}

func TestConcurrentGetAndSetCustomDir(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			SetCustomDir(t.TempDir())
			Get("core-re")
		}
		close(done)
	}()
	<-done
	SetCustomDir("")
}
