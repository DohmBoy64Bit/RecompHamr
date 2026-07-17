package skills

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type fakeBundledTemp struct {
	name     string
	writeErr error
	syncErr  error
	closeErr error
}

func (f fakeBundledTemp) Name() string { return f.name }
func (f fakeBundledTemp) Write(data []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(data), nil
}
func (f fakeBundledTemp) Sync() error  { return f.syncErr }
func (f fakeBundledTemp) Close() error { return f.closeErr }

func preserveBundledIO(t *testing.T) {
	t.Helper()
	manifest, write, parents, read, lstat, mkdir, temp, rename, restrict := bundledManifestFn, bundledWriteFn, bundledParentsFn, bundledReadFile, bundledLstat, bundledMkdir, bundledTemp, bundledRename, bundledRestrict
	t.Cleanup(func() {
		bundledManifestFn, bundledWriteFn, bundledParentsFn, bundledReadFile, bundledLstat, bundledMkdir, bundledTemp, bundledRename, bundledRestrict = manifest, write, parents, read, lstat, mkdir, temp, rename, restrict
	})
}

func TestBundledManifestFailuresAndOrdering(t *testing.T) {
	source := fstest.MapFS{
		"builtin/b/SKILL.md": {Data: []byte("b")},
		"builtin/a/SKILL.md": {Data: []byte("a")},
	}
	files, digest, err := bundledManifestFrom(source, fs.ReadFile)
	if err != nil || len(files) != 2 || files[0] != "a/SKILL.md" || len(digest) != 16 {
		t.Fatalf("manifest = %#v %q %v", files, digest, err)
	}
	nonregular := fstest.MapFS{"builtin/link": {Mode: fs.ModeSymlink}}
	if _, _, err := bundledManifestFrom(nonregular, fs.ReadFile); err == nil {
		t.Fatal("non-regular manifest accepted")
	}
	readErr := errors.New("read")
	if _, _, err := bundledManifestFrom(source, func(fs.FS, string) ([]byte, error) { return nil, readErr }); !errors.Is(err, readErr) {
		t.Fatalf("manifest read error = %v", err)
	}
	missing := fstest.MapFS{}
	if _, _, err := bundledManifestFrom(missing, fs.ReadFile); err == nil {
		t.Fatal("missing manifest root accepted")
	}
}

func TestInstallBundledCreatesDeterministicPrivateCatalog(t *testing.T) {
	privateRoot := t.TempDir()
	root, err := InstallBundled(privateRoot)
	if err != nil {
		t.Fatal(err)
	}
	again, err := InstallBundled(privateRoot)
	if err != nil || again != root || !strings.HasPrefix(root, filepath.Join(privateRoot, "bundled-skills")) {
		t.Fatalf("repeat root = %q, %v", again, err)
	}
	catalog := Discover([]Root{{Path: root, Scope: ScopeBundled, Trusted: true}})
	entries := catalog.Entries()
	want := []string{"build-fix-loop", "cdb-debug", "core-re", "evidence-mode", "file-format-reversing", "function-discovery", "gb-recomp", "gc-decomp", "gen-decomp", "imhex", "n64-decomp", "objdiff", "pcrecomp", "project-handoff", "ps2recomp", "ps3recomp", "snesrecomp", "vb-decomp", "windows-game-decomp", "xbox360-decomp", "xboxrecomp"}
	if len(entries) != len(want) {
		t.Fatalf("bundled entries = %#v diagnostics=%#v", entries, catalog.Diagnostics())
	}
	for i, name := range want {
		if entries[i].Name != name || entries[i].Scope != ScopeBundled {
			t.Fatalf("bundled entry %d = %#v", i, entries[i])
		}
	}
	activation, err := catalog.Activate("core-re")
	if err != nil || !strings.Contains(activation.Instructions, "Core reverse-engineering workflow") {
		t.Fatalf("activation = %#v, %v", activation, err)
	}
	if err := os.WriteFile(filepath.Join(root, "core-re", "SKILL.md"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallBundled(privateRoot); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "core-re", "SKILL.md"))
	if err != nil || strings.Contains(string(data), "tampered") {
		t.Fatalf("bundled restoration = %q, %v", data, err)
	}
}

func TestBundledSkillEvaluationContracts(t *testing.T) {
	type triggerCase struct {
		Prompt        string `json:"prompt"`
		ShouldTrigger bool   `json:"should_trigger"`
	}
	type outputCase struct {
		Prompt          string `json:"prompt"`
		ExpectedOutcome string `json:"expected_outcome"`
	}
	type evaluation struct {
		Skill        string        `json:"skill"`
		TriggerCases []triggerCase `json:"trigger_cases"`
		OutputCases  []outputCase  `json:"output_cases"`
	}
	entries := discoverFromBundledForTest(t)
	for _, entry := range entries {
		data, err := bundledReadFile(bundledFiles, "builtin/"+entry.Name+"/evals/evals.json")
		if err != nil {
			t.Fatalf("%s eval read: %v", entry.Name, err)
		}
		var eval evaluation
		if err := json.Unmarshal(data, &eval); err != nil {
			t.Fatalf("%s eval JSON: %v", entry.Name, err)
		}
		positive, negative := 0, 0
		for _, test := range eval.TriggerCases {
			if strings.TrimSpace(test.Prompt) == "" {
				t.Fatalf("%s has empty trigger prompt", entry.Name)
			}
			if test.ShouldTrigger {
				positive++
			} else {
				negative++
			}
		}
		if eval.Skill != entry.Name || positive < 8 || negative < 8 || len(eval.OutputCases) < 3 {
			t.Fatalf("%s eval shape = skill=%q positive=%d negative=%d output=%d", entry.Name, eval.Skill, positive, negative, len(eval.OutputCases))
		}
		for _, test := range eval.OutputCases {
			if strings.TrimSpace(test.Prompt) == "" || strings.TrimSpace(test.ExpectedOutcome) == "" {
				t.Fatalf("%s has incomplete output case", entry.Name)
			}
		}
	}
}

func discoverFromBundledForTest(t *testing.T) []Entry {
	t.Helper()
	root, err := InstallBundled(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	catalog := Discover([]Root{{Path: root, Scope: ScopeBundled, Trusted: true}})
	if diagnostics := catalog.Diagnostics(); len(diagnostics) != 0 {
		t.Fatalf("bundled diagnostics = %#v", diagnostics)
	}
	return catalog.Entries()
}

func TestBundledPathAndDirectoryFailures(t *testing.T) {
	preserveBundledIO(t)
	boom := errors.New("boom")
	bundledLstat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	bundledMkdir = func(string, os.FileMode) error { return boom }
	if err := ensureBundledDirectory("x"); !errors.Is(err, boom) {
		t.Fatalf("mkdir error = %v", err)
	}
	bundledMkdir = os.Mkdir
	if err := ensureBundledParents("root", filepath.Join("root", "..", "escape")); err == nil {
		t.Fatal("escaping parent accepted")
	}
	if err := ensureBundledParents("root", "root"); err != nil {
		t.Fatalf("same parent = %v", err)
	}
	bundledLstat = func(string) (os.FileInfo, error) { return testInfo{mode: fs.ModeSymlink}, nil }
	if err := ensureBundledDirectory("x"); err == nil {
		t.Fatal("unsafe directory accepted")
	}
	if err := writeBundledFile("x", nil); err == nil {
		t.Fatal("unsafe destination accepted")
	}
	bundledLstat = func(string) (os.FileInfo, error) { return nil, boom }
	if err := writeBundledFile("x", nil); !errors.Is(err, boom) {
		t.Fatalf("destination inspect error = %v", err)
	}
	if err := ensureBundledParents("root", filepath.Join("root", "child")); err == nil {
		t.Fatal("unsafe child directory accepted")
	}
}

func TestInstallBundledAndWriteFailures(t *testing.T) {
	preserveBundledIO(t)
	boom := errors.New("boom")
	bundledManifestFn = func() ([]string, string, error) { return nil, "", boom }
	if _, err := InstallBundled(t.TempDir()); !errors.Is(err, boom) {
		t.Fatalf("manifest error = %v", err)
	}
	bundledManifestFn = func() ([]string, string, error) { return []string{"core-re/SKILL.md"}, "digest", nil }
	bundledLstat = os.Lstat
	bundledMkdir = os.Mkdir
	bundledRestrict = func(path string, directory bool) error {
		if strings.HasSuffix(path, "bundled-skills") {
			return boom
		}
		return nil
	}
	if _, err := InstallBundled(t.TempDir()); !errors.Is(err, boom) {
		t.Fatalf("directory error = %v", err)
	}
	bundledRestrict = func(string, bool) error { return nil }
	root := t.TempDir()
	bundledReadFile = func(fs.FS, string) ([]byte, error) { return nil, boom }
	if _, err := InstallBundled(root); !errors.Is(err, boom) {
		t.Fatalf("read error = %v", err)
	}
	bundledReadFile = fs.ReadFile
	bundledManifestFn = func() ([]string, string, error) { return []string{"core-re/SKILL.md"}, "digest-two", nil }
	bundledParentsFn = func(string, string) error { return boom }
	if _, err := InstallBundled(t.TempDir()); !errors.Is(err, boom) {
		t.Fatalf("install parent error = %v", err)
	}
	bundledParentsFn = ensureBundledParents
	bundledManifestFn = func() ([]string, string, error) { return []string{"core-re/SKILL.md"}, "digest-three", nil }
	bundledWriteFn = func(string, []byte) error { return boom }
	if _, err := InstallBundled(t.TempDir()); !errors.Is(err, boom) {
		t.Fatalf("install write error = %v", err)
	}
	bundledWriteFn = writeBundledFile

	bundledLstat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	bundledTemp = func(string, string) (bundledTemporaryFile, error) { return nil, boom }
	if err := writeBundledFile("x", []byte("x")); !errors.Is(err, boom) {
		t.Fatalf("temp error = %v", err)
	}
	for name, file := range map[string]fakeBundledTemp{
		"write": {name: "temp", writeErr: boom},
		"sync":  {name: "temp", syncErr: boom},
		"close": {name: "temp", closeErr: boom},
	} {
		bundledTemp = func(string, string) (bundledTemporaryFile, error) { return file, nil }
		if err := writeBundledFile(name, []byte("x")); !errors.Is(err, boom) {
			t.Fatalf("%s error = %v", name, err)
		}
	}
	bundledTemp = func(string, string) (bundledTemporaryFile, error) { return fakeBundledTemp{name: "temp"}, nil }
	bundledRename = func(string, string) error { return boom }
	if err := writeBundledFile("rename", []byte("x")); !errors.Is(err, boom) {
		t.Fatalf("rename error = %v", err)
	}
	bundledRename = func(string, string) error { return nil }
	bundledRestrict = func(string, bool) error { return boom }
	if err := writeBundledFile("protect", []byte("x")); !errors.Is(err, boom) {
		t.Fatalf("protect error = %v", err)
	}
}
