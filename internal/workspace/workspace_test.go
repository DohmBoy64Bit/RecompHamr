package workspace

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func makeWorkspace(t *testing.T, data []byte) *Workspace {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".rehamr")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if data != nil {
		if err := os.WriteFile(filepath.Join(dir, StateFileName), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	workspace, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	return workspace
}

func TestOpenAndLoadStateContracts(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("empty project path accepted")
	}
	workspace := makeWorkspace(t, nil)
	if !filepath.IsAbs(workspace.Root()) || filepath.Clean(workspace.Root()) != workspace.Root() {
		t.Fatalf("root = %q", workspace.Root())
	}
	if state, err := workspace.LoadState(); err != nil || state != "" {
		t.Fatalf("missing state = %q, %v", state, err)
	}

	workspace = makeWorkspace(t, []byte("# State\nUnicode 🐹\n"))
	state, err := workspace.LoadState()
	if err != nil || state != "# State\nUnicode 🐹\n" {
		t.Fatalf("state = %q, %v", state, err)
	}
	prompt, err := workspace.SystemPrompt("base")
	if err != nil || !strings.Contains(prompt, "base\n\nWorking directory: "+workspace.Root()) ||
		!strings.Contains(prompt, "## Persistent Memory") || !strings.Contains(prompt, "untrusted project-maintained context") ||
		!strings.HasSuffix(prompt, state) {
		t.Fatalf("prompt = %q, %v", prompt, err)
	}

	empty := makeWorkspace(t, []byte{})
	if state, err := empty.LoadState(); err != nil || state != "" {
		t.Fatalf("empty state = %q, %v", state, err)
	}
	prompt, err = empty.SystemPrompt("base")
	if err != nil || strings.Contains(prompt, "Persistent Memory") {
		t.Fatalf("empty prompt = %q, %v", prompt, err)
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got, readErr := workspace.LoadState(); readErr != nil || got != state {
				t.Errorf("concurrent state = %q, %v", got, readErr)
			}
		}()
	}
	wg.Wait()
}

func TestLoadStateRejectsUnsafeInputs(t *testing.T) {
	t.Run("oversize", func(t *testing.T) {
		workspace := makeWorkspace(t, []byte(strings.Repeat("x", MaxStateBytes+1)))
		if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "limit") {
			t.Fatalf("oversize error = %v", err)
		}
	})
	t.Run("invalid UTF-8", func(t *testing.T) {
		workspace := makeWorkspace(t, []byte{0xff})
		if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "UTF-8") {
			t.Fatalf("UTF-8 error = %v", err)
		}
	})
	t.Run("directory", func(t *testing.T) {
		workspace := makeWorkspace(t, nil)
		if err := os.Mkdir(workspace.statePath, 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "regular") {
			t.Fatalf("directory error = %v", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		workspace := makeWorkspace(t, nil)
		target := filepath.Join(t.TempDir(), "state")
		if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, workspace.statePath); err != nil {
			t.Skipf("symlink unsupported: %v", err)
		}
		if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "refused") {
			t.Fatalf("symlink error = %v", err)
		}
	})
}

type fakeInfo struct {
	mode os.FileMode
	size int64
}

func (f fakeInfo) Name() string       { return StateFileName }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() os.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeInfo) Sys() any           { return nil }

type fakeFile struct {
	reader   io.Reader
	info     os.FileInfo
	statErr  error
	closeErr error
}

func (f *fakeFile) Read(p []byte) (int, error) { return f.reader.Read(p) }
func (f *fakeFile) Stat() (os.FileInfo, error) { return f.info, f.statErr }
func (f *fakeFile) Close() error               { return f.closeErr }

func TestWorkspaceFailureBoundaries(t *testing.T) {
	origAbs, origLstat, origOpen, origRestrict, origSame := absPath, lstatPath, openPath, restrictPath, sameFile
	t.Cleanup(func() {
		absPath, lstatPath, openPath, restrictPath, sameFile = origAbs, origLstat, origOpen, origRestrict, origSame
	})
	boom := errors.New("boom")
	absPath = func(string) (string, error) { return "", boom }
	if _, err := Open("x"); !errors.Is(err, boom) {
		t.Fatalf("abs error = %v", err)
	}
	absPath = origAbs
	workspace := &Workspace{statePath: "state"}

	lstatPath = func(string) (os.FileInfo, error) { return nil, boom }
	if _, err := workspace.LoadState(); !errors.Is(err, boom) {
		t.Fatalf("lstat error = %v", err)
	}
	if prompt, err := workspace.SystemPrompt("base"); !errors.Is(err, boom) || prompt != "base\n\nWorking directory: " {
		t.Fatalf("prompt failure = %q, %v", prompt, err)
	}
	lstatPath = func(string) (os.FileInfo, error) { return fakeInfo{mode: os.ModeSymlink}, nil }
	if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("link error = %v", err)
	}
	regular := fakeInfo{mode: 0o600, size: 1}
	lstatPath = func(string) (os.FileInfo, error) { return regular, nil }
	sameFile = func(os.FileInfo, os.FileInfo) bool { return true }
	restrictPath = func(string, bool) error { return boom }
	if _, err := workspace.LoadState(); !errors.Is(err, boom) {
		t.Fatalf("restrict error = %v", err)
	}
	restrictPath = func(string, bool) error { return nil }
	openPath = func(string) (stateFile, error) { return nil, boom }
	if _, err := workspace.LoadState(); !errors.Is(err, boom) {
		t.Fatalf("open error = %v", err)
	}
	openPath = func(string) (stateFile, error) { return &fakeFile{reader: strings.NewReader("x"), statErr: boom}, nil }
	if _, err := workspace.LoadState(); !errors.Is(err, boom) {
		t.Fatalf("stat error = %v", err)
	}
	openPath = func(string) (stateFile, error) {
		return &fakeFile{reader: strings.NewReader("x"), info: fakeInfo{mode: os.ModeDir}}, nil
	}
	if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("replacement error = %v", err)
	}
	openPath = func(string) (stateFile, error) {
		return &fakeFile{reader: strings.NewReader("x"), info: fakeInfo{mode: 0o600, size: MaxStateBytes + 1}}, nil
	}
	if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("post-open size error = %v", err)
	}
	openPath = func(string) (stateFile, error) { return &fakeFile{reader: errorReader{err: boom}, info: regular}, nil }
	if _, err := workspace.LoadState(); !errors.Is(err, boom) {
		t.Fatalf("read error = %v", err)
	}
	openPath = func(string) (stateFile, error) {
		return &fakeFile{reader: strings.NewReader(strings.Repeat("x", MaxStateBytes+1)), info: regular}, nil
	}
	if _, err := workspace.LoadState(); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("streamed oversize error = %v", err)
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
