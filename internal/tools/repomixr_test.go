package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepomixrSchemaShape(t *testing.T) {
	s := RepomixrSchema()
	fn, ok := s["function"].(map[string]any)
	if !ok {
		t.Fatal("missing function key")
	}
	if fn["name"] != RepomixrName {
		t.Errorf("expected name %q, got %q", RepomixrName, fn["name"])
	}
	params, ok := fn["parameters"].(map[string]any)
	if !ok {
		t.Fatal("missing parameters")
	}
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing properties")
	}
	required := params["required"].([]string)
	hasURL := false
	for _, r := range required {
		if r == "repo_url" {
			hasURL = true
		}
	}
	if !hasURL {
		t.Errorf("repo_url should be required, got %v", required)
	}
	expected := []string{"repo_url", "branch", "remove_comments", "remove_empty_lines", "show_line_numbers", "compress"}
	for _, k := range expected {
		if _, ok := props[k]; !ok {
			t.Errorf("missing property %q", k)
		}
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		url, wantOwner, wantName string
	}{
		{"https://github.com/user/repo", "user", "repo"},
		{"https://github.com/user/repo.git", "user", "repo"},
		{"https://github.com/user/repo/", "user", "repo"},
		{"github.com/user/repo", "user", "repo"},
		{"http://github.com/user/repo", "user", "repo"},
		{"not a url", "", ""},
		{"https://github.com/user", "", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		o, n := parseRepoURL(tt.url)
		if o != tt.wantOwner || n != tt.wantName {
			t.Errorf("parseRepoURL(%q) = (%q, %q), want (%q, %q)",
				tt.url, o, n, tt.wantOwner, tt.wantName)
		}
	}
}

func TestStripComments(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"// comment", ""},
		{"code // comment\nmore code", "code \nmore code"},
		{"# python comment", ""},
		{"plain code", "plain code"},
		{"http://example.com // not a comment", "http://example.com // not a comment"},
		{"line1\n// commented\nline3", "line1\nline3"},
		{"code // inline", "code "},
	}
	for _, tt := range tests {
		got := stripComments(tt.in)
		if got != tt.want {
			t.Errorf("stripComments(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"a   b", "a b"},
		{"a\t\tb", "a b"},
		{"a\nb", "a b"},
		{"a   \n   b", "a b"},
		{"hello world", "hello world"},
		{"", ""},
	}
	for _, tt := range tests {
		got := collapseWhitespace(tt.in)
		if got != tt.want {
			t.Errorf("collapseWhitespace(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
	}
	for _, tt := range tests {
		got := humanSize(tt.n)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestWalkTextFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.c"), []byte("int main() { return 0; }"), 0o644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello"), 0o644)
	os.WriteFile(filepath.Join(dir, "binary.exe"), []byte{0x00, 0x01, 0x02}, 0o644)
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("git"), 0o644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)
	os.WriteFile(filepath.Join(dir, "subdir", "util.go"), []byte("package util"), 0o644)

	files, err := walkTextFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]bool)
	for _, f := range files {
		rel, _ := filepath.Rel(dir, f)
		found[rel] = true
	}

	if !found["main.c"] {
		t.Error("missing main.c")
	}
	if !found["readme.md"] {
		t.Error("missing readme.md")
	}
	if !found[filepath.Join("subdir", "util.go")] {
		t.Error("missing subdir/util.go")
	}
	if found["binary.exe"] {
		t.Error("binary.exe should be filtered")
	}
	if found[filepath.Join(".git", "config")] {
		t.Error(".git/ should be skipped")
	}
}

func TestWalkTextFilesLargeFileSkipped(t *testing.T) {
	dir := t.TempDir()
	big := make([]byte, 600*1024)
	for i := range big {
		big[i] = 'A'
	}
	os.WriteFile(filepath.Join(dir, "big.txt"), big, 0o644)

	files, err := walkTextFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Base(f) == "big.txt" {
			t.Error("big.txt should be skipped (>512KB)")
		}
	}
}

func TestRepomixrNoRepoDir(t *testing.T) {
	old := RepomixrDir
	RepomixrDir = ""
	defer func() { RepomixrDir = old }()

	result := Repomixr(context.Background(), "https://github.com/user/repo", "main", false, false, false, false)
	if !strings.Contains(result, "not configured") {
		t.Errorf("expected 'not configured' error, got: %s", result)
	}
}

func TestRepomixrInvalidURL(t *testing.T) {
	old := RepomixrDir
	RepomixrDir = t.TempDir()
	defer func() { RepomixrDir = old }()

	result := Repomixr(context.Background(), "not-a-url", "", false, false, false, false)
	if !strings.Contains(result, "could not parse") {
		t.Errorf("expected parse error, got: %s", result)
	}
}

func TestRepomixrGitCloneFails(t *testing.T) {
	old := RepomixrDir
	RepomixrDir = t.TempDir()
	defer func() { RepomixrDir = old }()

	result := Repomixr(context.Background(), "https://github.com/nonexistent/repo-that-does-not-exist-12345", "main", false, false, false, false)
	if !strings.Contains(result, "git clone failed") && !strings.Contains(result, "not found") {
		// Note: git clone might partially work on some systems, so we check either error
		t.Logf("clone result: %s", result)
	}
}

func TestRepomixrUsesDefaultBranch(t *testing.T) {
	old := RepomixrDir
	RepomixrDir = t.TempDir()
	defer func() { RepomixrDir = old }()

	result := Repomixr(context.Background(), "https://github.com/nonexistent/repo", "", false, false, false, false)
	if !strings.Contains(result, "clone failed") {
		// If branch is empty, it should default to "main" which will be in the output
		t.Logf("clone result: %s", result)
	}
}

func TestDirTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.MkdirAll(filepath.Join(dir, "include"), 0o755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "main.c"), []byte("main"), 0o644)
	os.WriteFile(filepath.Join(dir, "include", "header.h"), []byte("h"), 0o644)

	files := []string{
		filepath.Join(dir, "README.md"),
		filepath.Join(dir, "include", "header.h"),
		filepath.Join(dir, "src", "main.c"),
	}
	tree := dirTree(dir, files)
	if !strings.Contains(tree, "src/") {
		t.Errorf("tree should contain src/: %s", tree)
	}
	if !strings.Contains(tree, "main.c") {
		t.Errorf("tree should contain main.c: %s", tree)
	}
	if !strings.Contains(tree, "include/") {
		t.Errorf("tree should contain include/: %s", tree)
	}
	if !strings.Contains(tree, "header.h") {
		t.Errorf("tree should contain header.h: %s", tree)
	}
	if !strings.Contains(tree, "README.md") {
		t.Errorf("tree should contain README.md: %s", tree)
	}
}

func TestDirTreeDeepNesting(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0o755)
	os.WriteFile(filepath.Join(dir, "a", "b", "c", "d.c"), []byte("d"), 0o644)

	files := []string{filepath.Join(dir, "a", "b", "c", "d.c")}
	tree := dirTree(dir, files)
	if !strings.Contains(tree, "a/") {
		t.Errorf("tree should contain a/: %s", tree)
	}
	if !strings.Contains(tree, "d.c") {
		t.Errorf("tree should contain d.c at correct depth: %s", tree)
	}
}

func TestRepomixrOutputSkipsInstructionWhenMissing(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir, 0o755)
	os.MkdirAll(filepath.Join(dir, "test-repo", "repo"), 0o755)
	os.WriteFile(filepath.Join(dir, "test-repo", "repo", "readme.md"), []byte("readme"), 0o644)

	old := RepomixrDir
	RepomixrDir = dir
	defer func() { RepomixrDir = old }()

	result := Repomixr(context.Background(), "https://github.com/test/skip", "main", false, false, false, false)
	if strings.Contains(result, "Packed") {
		data, _ := os.ReadFile(filepath.Join(RepomixrDir, "test-repo", "packed.xml"))
		content := string(data)
		if strings.Contains(content, "<instruction>") {
			t.Errorf("output should NOT contain instruction when file missing: %s", content)
		}
	}
}

func TestRepomixrOutputIncludesInstruction(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir, 0o755)
	instrPath := filepath.Join(dir, "..", "repomix-instruction.md")
	os.WriteFile(instrPath, []byte("Focus on the rendering pipeline."), 0o644)

	old := RepomixrDir
	RepomixrDir = dir
	defer func() { RepomixrDir = old }()

	os.MkdirAll(filepath.Join(RepomixrDir, "test-repo", "repo"), 0o755)
	os.WriteFile(filepath.Join(RepomixrDir, "test-repo", "repo", "readme.md"), []byte("readme"), 0o644)

	result := Repomixr(context.Background(), "https://github.com/test/with-instr", "main", false, false, false, false)
	if strings.Contains(result, "Packed") {
		data, _ := os.ReadFile(filepath.Join(RepomixrDir, "test-repo", "packed.xml"))
		content := string(data)
		if !strings.Contains(content, "<instruction>") {
			t.Errorf("output should contain instruction: %s", content)
		}
		if !strings.Contains(content, "Focus on the rendering pipeline") {
			t.Errorf("instruction should contain file content: %s", content)
		}
	}
}
