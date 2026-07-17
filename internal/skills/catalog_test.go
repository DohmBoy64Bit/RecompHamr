package skills

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testInfo struct {
	name string
	mode fs.FileMode
	size int64
}

func (i testInfo) Name() string       { return i.name }
func (i testInfo) Size() int64        { return i.size }
func (i testInfo) Mode() fs.FileMode  { return i.mode }
func (i testInfo) ModTime() time.Time { return time.Time{} }
func (i testInfo) IsDir() bool        { return i.mode.IsDir() }
func (i testInfo) Sys() any           { return nil }

type testEntry struct{ testInfo }

func (e testEntry) Type() fs.FileMode          { return e.mode.Type() }
func (e testEntry) Info() (fs.FileInfo, error) { return e.testInfo, nil }

func preserveIO(t *testing.T) {
	t.Helper()
	a, l, d, f, s, w, r := absPath, lstat, readDir, readFile, sameFile, walkDir, relPath
	t.Cleanup(func() { absPath, lstat, readDir, readFile, sameFile, walkDir, relPath = a, l, d, f, s, w, r })
}

func writeSkill(t *testing.T, root, name, description, body string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n" + body
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDiscoverPrecedenceCopiesAndActivation(t *testing.T) {
	user := t.TempDir()
	project := t.TempDir()
	writeSkill(t, user, "alpha", "User alpha. Use for alpha work.", "user body")
	dir := writeSkill(t, project, "alpha", "Project alpha. Use for alpha work.", "project body")
	writeSkill(t, user, "beta", "Beta helper. Use for beta work.", "beta body")
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "guide.md"), []byte("secretly lazy"), 0o600); err != nil {
		t.Fatal(err)
	}

	catalog := Discover([]Root{{Path: user, Scope: ScopeUser, Trusted: true}, {Path: project, Scope: ScopeProject, Trusted: true}})
	entries := catalog.Entries()
	if len(entries) != 2 || entries[0].Name != "alpha" || entries[0].Description != "Project alpha. Use for alpha work." || entries[1].Name != "beta" {
		t.Fatalf("entries = %#v", entries)
	}
	entries[0].Name = "mutated"
	if catalog.Entries()[0].Name != "alpha" {
		t.Fatal("entries shared storage")
	}
	diagnostics := catalog.Diagnostics()
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message, "shadowed") {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(diagnostics[0].Message, user) || strings.Contains(diagnostics[0].Message, project) || filepath.IsAbs(diagnostics[0].Message) {
		t.Fatalf("shadow diagnostic exposed a path: %#v", diagnostics[0])
	}
	diagnostics[0].Message = "mutated"
	if catalog.Diagnostics()[0].Message == "mutated" {
		t.Fatal("diagnostics shared storage")
	}
	activation, err := catalog.Activate("alpha")
	if err != nil || activation.Name != "alpha" || activation.Instructions != "project body" || len(activation.Resources) != 1 || activation.Resources[0] != "references/guide.md" || activation.Directory != dir {
		t.Fatalf("activation = %#v, %v", activation, err)
	}
	if _, err := catalog.Activate("missing"); err == nil {
		t.Fatal("unknown activation accepted")
	}
}

func TestCatalogWithoutFiltersKnownNames(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "Use alpha.", "body")
	writeSkill(t, root, "beta", "Use beta.", "body")
	filtered := Discover([]Root{{Path: root, Scope: ScopeUser}}).Without([]string{"alpha", "missing"})
	if entries := filtered.Entries(); len(entries) != 1 || entries[0].Name != "beta" {
		t.Fatalf("filtered entries = %#v", entries)
	}
	if diagnostics := filtered.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].Name != "alpha" || !strings.Contains(diagnostics[0].Message, "disabled") {
		t.Fatalf("filtered diagnostics = %#v", diagnostics)
	}
}

func TestDiscoverTrustRootsAndStrictValidation(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "valid", "Valid helper. Use for valid work.", "body")
	if got := Discover([]Root{{Path: root, Scope: ScopeProject}}); len(got.Entries()) != 0 || len(got.Diagnostics()) != 1 {
		t.Fatalf("untrusted = %#v %#v", got.Entries(), got.Diagnostics())
	}

	badRoot := t.TempDir()
	writeSkill(t, badRoot, "Wrong", "Bad name.", "body")
	writeSkill(t, badRoot, "empty", "", "body")
	unterminated := filepath.Join(badRoot, "unterminated")
	if err := os.Mkdir(unterminated, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unterminated, "SKILL.md"), []byte("---\nname: unterminated"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badRoot, "README.md"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := Discover([]Root{{Path: badRoot, Scope: ScopeUser, Trusted: true}, {Path: filepath.Join(badRoot, "missing"), Scope: ScopeUser}, {Path: badRoot, Scope: 99}})
	if len(got.Entries()) != 0 || len(got.Diagnostics()) != 4 {
		t.Fatalf("invalid = %#v %#v", got.Entries(), got.Diagnostics())
	}
}

func TestActivationDetectsChangesAndUnsafeResources(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "Alpha helper. Use for alpha work.", "body")
	catalog := Discover([]Root{{Path: root, Scope: ScopeUser, Trusted: true}})
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: alpha\ndescription: Changed.\n---\nbody"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Activate("alpha"); err == nil || !strings.Contains(err.Error(), "metadata changed") {
		t.Fatalf("changed error = %v", err)
	}

	dir = writeSkill(t, root, "beta", "Beta helper. Use for beta work.", "body")
	catalog = Discover([]Root{{Path: root, Scope: ScopeUser, Trusted: true}})
	if err := os.WriteFile(filepath.Join(dir, "references"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Activate("beta"); err == nil || !strings.Contains(err.Error(), "resource directory") {
		t.Fatalf("resource error = %v", err)
	}
}

func TestMetadataConstraintsAndDecode(t *testing.T) {
	cases := []metadata{
		{Name: strings.Repeat("a", 65), Description: "d"},
		{Name: "bad--name", Description: "d"},
		{Name: "good", Description: strings.Repeat("d", 1025)},
		{Name: "good", Description: "d", Compatibility: strings.Repeat("c", 501)},
	}
	for _, test := range cases {
		if err := validateMetadata(test, test.Name); err == nil {
			t.Fatalf("accepted %#v", test)
		}
	}
	if _, _, err := decodeSkill([]byte{0xff}); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if _, _, err := decodeSkill([]byte("no frontmatter")); err == nil {
		t.Fatal("missing frontmatter accepted")
	}
}

func TestInjectedDiscoveryFailuresAndLimits(t *testing.T) {
	preserveIO(t)
	boom := errors.New("boom")
	absPath = func(string) (string, error) { return "", boom }
	if got := Discover([]Root{{Path: "x", Scope: ScopeUser}}); len(got.Diagnostics()) != 1 {
		t.Fatalf("abs diagnostics = %#v", got.Diagnostics())
	}
	absPath = filepath.Abs
	lstat = func(string) (os.FileInfo, error) { return nil, boom }
	if got := Discover([]Root{{Path: "x", Scope: ScopeUser}}); len(got.Diagnostics()) != 1 {
		t.Fatalf("lstat diagnostics = %#v", got.Diagnostics())
	}
	lstat = func(string) (os.FileInfo, error) { return testInfo{mode: fs.ModeSymlink}, nil }
	if got := Discover([]Root{{Path: "x", Scope: ScopeUser}}); len(got.Diagnostics()) != 1 {
		t.Fatalf("link diagnostics = %#v", got.Diagnostics())
	}
	lstat = func(string) (os.FileInfo, error) { return testInfo{mode: fs.ModeDir}, nil }
	readDir = func(string) ([]os.DirEntry, error) { return nil, boom }
	if got := Discover([]Root{{Path: "x", Scope: ScopeUser}}); len(got.Diagnostics()) != 1 {
		t.Fatalf("read diagnostics = %#v", got.Diagnostics())
	}
	readDir = func(string) ([]os.DirEntry, error) { return make([]os.DirEntry, maxSkillsPerRoot+1), nil }
	if got := Discover([]Root{{Path: "x", Scope: ScopeUser}}); len(got.Diagnostics()) != 1 {
		t.Fatalf("limit diagnostics = %#v", got.Diagnostics())
	}
}

func TestInjectedMetadataAndActivationFailures(t *testing.T) {
	preserveIO(t)
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "Alpha helper. Use for alpha work.", "body")
	location := filepath.Join(dir, "SKILL.md")
	resource := filepath.Join(dir, "references", "guide.txt")
	if err := os.MkdirAll(filepath.Dir(resource), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resource, []byte("guide"), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog := Discover([]Root{{Path: root, Scope: ScopeUser}})
	activation, err := catalog.Activate("alpha")
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	readFile = func(string) ([]byte, error) { return nil, boom }
	if _, err := catalog.Activate("alpha"); err == nil {
		t.Fatal("read error accepted")
	}
	readFile = os.ReadFile
	lstat = func(path string) (os.FileInfo, error) {
		if path == location {
			return nil, boom
		}
		return os.Lstat(path)
	}
	if _, err := catalog.Activate("alpha"); err == nil {
		t.Fatal("activation lstat error accepted")
	}
	lstat = os.Lstat
	sameFile = func(os.FileInfo, os.FileInfo) bool { return false }
	if _, err := catalog.Activate("alpha"); err == nil || !strings.Contains(err.Error(), "changed while reading") {
		t.Fatalf("same-file error = %v", err)
	}
	sameFile = os.SameFile
	readFile = func(path string) ([]byte, error) {
		if path == resource {
			return nil, boom
		}
		return os.ReadFile(path)
	}
	if _, err := catalog.ReadResource(activation, "references/guide.txt"); err == nil {
		t.Fatal("resource read error accepted")
	}
	readFile = os.ReadFile
	sameFile = func(os.FileInfo, os.FileInfo) bool { return false }
	if _, err := catalog.ReadResource(activation, "references/guide.txt"); err == nil || !strings.Contains(err.Error(), "changed while reading") {
		t.Fatalf("resource same-file error = %v", err)
	}
	sameFile = os.SameFile
	absPath = func(path string) (string, error) {
		if strings.HasSuffix(path, "SKILL.md") {
			return "", boom
		}
		return filepath.Abs(path)
	}
	if got := Discover([]Root{{Path: root, Scope: ScopeUser}}); len(got.Entries()) != 0 || len(got.Diagnostics()) != 1 {
		t.Fatalf("metadata abs = %#v %#v", got.Entries(), got.Diagnostics())
	}
}

func TestReadRegularAndResourceBranches(t *testing.T) {
	preserveIO(t)
	boom := errors.New("boom")
	lstat = func(string) (os.FileInfo, error) { return testInfo{mode: fs.ModeSymlink}, nil }
	if _, _, err := readRegular("x", 1); err == nil {
		t.Fatal("link accepted")
	}
	lstat = func(string) (os.FileInfo, error) { return testInfo{mode: 0, size: 2}, nil }
	if _, _, err := readRegular("x", 1); err == nil {
		t.Fatal("oversize accepted")
	}
	lstat = func(string) (os.FileInfo, error) { return testInfo{mode: fs.ModeDir}, nil }
	walkDir = func(_ string, fn fs.WalkDirFunc) error { return fn("x", nil, boom) }
	if _, err := listResources("x"); !errors.Is(err, boom) {
		t.Fatalf("walk error = %v", err)
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		return fn("x/link", testEntry{testInfo{name: "link", mode: fs.ModeSymlink}}, nil)
	}
	if _, err := listResources("x"); err == nil || !strings.Contains(err.Error(), "linked") {
		t.Fatalf("link resource = %v", err)
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		return fn("x/dev", testEntry{testInfo{name: "dev", mode: fs.ModeDevice}}, nil)
	}
	if _, err := listResources("x"); err == nil || !strings.Contains(err.Error(), "non-regular") {
		t.Fatalf("device resource = %v", err)
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error {
		if err := fn("x/dir", testEntry{testInfo{name: "dir", mode: fs.ModeDir}}, nil); err != nil {
			return err
		}
		for i := 0; i <= maxResources; i++ {
			if err := fn("x/file", testEntry{testInfo{name: "file"}}, nil); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := listResources("x"); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("resource limit = %v", err)
	}
	walkDir = func(_ string, fn fs.WalkDirFunc) error { return fn("x/file", testEntry{testInfo{name: "file"}}, nil) }
	relPath = func(string, string) (string, error) { return "", boom }
	if _, err := listResources("x"); !errors.Is(err, boom) {
		t.Fatalf("rel error = %v", err)
	}
}

func TestSameScopeCollisionChildFilteringAndParserEdges(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	writeSkill(t, first, "alpha", "First alpha. Use for alpha work.", "first")
	writeSkill(t, second, "alpha", "Second alpha. Use for alpha work.", "second")
	if err := os.Mkdir(filepath.Join(first, "empty"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(first, "alpha"), filepath.Join(first, "linked")); err == nil {
		// Symlink support varies by Windows policy; when available it must not be discovered.
	}
	catalog := Discover([]Root{{Path: first, Scope: ScopeUser}, {Path: second, Scope: ScopeUser}})
	if entries := catalog.Entries(); len(entries) != 1 || entries[0].Description != "First alpha. Use for alpha work." {
		t.Fatalf("same-scope entries = %#v", entries)
	}
	if diagnostics := catalog.Diagnostics(); len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message, "shadowed") || strings.Contains(diagnostics[0].Message, first) || strings.Contains(diagnostics[0].Message, second) {
		t.Fatalf("same-scope diagnostics = %#v", diagnostics)
	}

	unsafe := filepath.Join(first, "unsafe")
	if err := os.MkdirAll(filepath.Join(unsafe, "SKILL.md"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, diagnostics, ok := parseMetadata(filepath.Join(unsafe, "SKILL.md"), "unsafe", ScopeUser); ok || len(diagnostics) != 1 {
		t.Fatalf("unsafe metadata = %v %#v", ok, diagnostics)
	}
	if _, diagnostics, ok := parseMetadata(filepath.Join(first, "absent", "SKILL.md"), "absent", ScopeUser); ok || diagnostics != nil {
		t.Fatalf("absent metadata = %v %#v", ok, diagnostics)
	}
	bad := filepath.Join(first, "bad", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(bad), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("---\nname: bad\ndescription: bad\nunknown: true\n---\nbody"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, diagnostics, ok := parseMetadata(bad, "bad", ScopeUser); ok || len(diagnostics) != 1 {
		t.Fatalf("bad yaml = %v %#v", ok, diagnostics)
	}
	meta, body, err := decodeSkill([]byte("---\nname: crlf\ndescription: CRLF helper.\n---\r\nbody"))
	if err != nil || meta.Name != "crlf" || body != "body" {
		t.Fatalf("CRLF decode = %#v %q %v", meta, body, err)
	}
}
