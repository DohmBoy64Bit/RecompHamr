package workspace

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type failingEvidenceFile struct {
	writeErr error
	writeN   int
	syncErr  error
	closeErr error
}

func (f failingEvidenceFile) Write(data []byte) (int, error) {
	if f.writeN < 0 {
		return len(data), f.writeErr
	}
	return f.writeN, f.writeErr
}
func (f failingEvidenceFile) Sync() error  { return f.syncErr }
func (f failingEvidenceFile) Close() error { return f.closeErr }

type errorReadCloser struct{ err error }

func (r errorReadCloser) Read([]byte) (int, error) { return 0, r.err }
func (errorReadCloser) Close() error               { return nil }

func preserveEvidenceIO(t *testing.T) {
	t.Helper()
	now, mkdir, open, remove, status := evidenceNow, mkdirEvidence, openEvidenceFile, removeEvidence, openStatusFile
	lstat, restrict := lstatPath, restrictPath
	t.Cleanup(func() {
		evidenceNow, mkdirEvidence, openEvidenceFile, removeEvidence, openStatusFile = now, mkdir, open, remove, status
		lstatPath, restrictPath = lstat, restrict
	})
}

func TestEvidenceWorkspaceInitializationStatusAndPreservation(t *testing.T) {
	root := t.TempDir()
	workspace, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	evidenceNow = func() time.Time { return time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { evidenceNow = time.Now })
	if err := workspace.InitializeEvidence(); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, ".rehamr", "PROJECT.md")
	if err := os.WriteFile(project, []byte("# Project\n\nkept"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workspace.InitializeEvidence(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(project)
	if err != nil || !strings.Contains(string(data), "kept") {
		t.Fatalf("preserved project = %q, %v", data, err)
	}
	status, err := workspace.EvidenceStatus()
	if err != nil || !strings.Contains(status, "RecompHamr project status") || !strings.Contains(status, "kept") || !strings.Contains(status, "Initialized RecompHamr") {
		t.Fatalf("status = %q, %v", status, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".rehamr", StateFileName)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".rehamr", "mcp.json")); !os.IsNotExist(err) {
		t.Fatalf("Stage H file created: %v", err)
	}
}

func TestEvidenceStatusMissingFilesAndTruncation(t *testing.T) {
	root := t.TempDir()
	workspace, _ := Open(root)
	if _, err := workspace.EvidenceStatus(); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("uninitialized error = %v", err)
	}
	if err := workspace.InitializeEvidence(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".rehamr", "EVIDENCE.md")); err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("é", 1000)
	if err := os.WriteFile(filepath.Join(root, ".rehamr", "PROJECT.md"), []byte(long), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := workspace.EvidenceStatus()
	if err != nil || !strings.Contains(status, "missing or unavailable") || !strings.Contains(status, "...truncated...") || !strings.Contains(status, "é") {
		t.Fatalf("status = %q, %v", status, err)
	}
}

func TestEvidenceWorkspaceRefusesUnsafePaths(t *testing.T) {
	root := t.TempDir()
	workspace, _ := Open(root)
	if err := os.Mkdir(filepath.Join(root, ".rehamr"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".rehamr", "evidence"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workspace.InitializeEvidence(); err == nil {
		t.Fatal("unsafe directory accepted")
	}
	if err := os.Remove(filepath.Join(root, ".rehamr", "evidence")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".rehamr", "evidence"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".rehamr", "PROJECT.md"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := workspace.InitializeEvidence(); err == nil {
		t.Fatal("unsafe file accepted")
	}
}

func TestEvidenceWorkspaceInjectedFailureBoundaries(t *testing.T) {
	preserveEvidenceIO(t)
	boom := errors.New("boom")
	workspace := &Workspace{root: t.TempDir()}

	restrictPath = func(string, bool) error { return boom }
	if err := workspace.InitializeEvidence(); !errors.Is(err, boom) {
		t.Fatalf("root protection error = %v", err)
	}
	restrictPath = func(string, bool) error { return nil }
	lstatPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	mkdirEvidence = func(string, os.FileMode) error { return boom }
	if err := ensureEvidenceDirectory("x"); !errors.Is(err, boom) {
		t.Fatalf("mkdir error = %v", err)
	}

	lstatPath = func(string) (os.FileInfo, error) { return nil, boom }
	if err := ensureEvidenceFile("x", []byte("x")); !errors.Is(err, boom) {
		t.Fatalf("file inspect error = %v", err)
	}
	lstatPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	openEvidenceFile = func(string, int, os.FileMode) (writableEvidenceFile, error) { return nil, boom }
	if err := ensureEvidenceFile("x", []byte("x")); !errors.Is(err, boom) {
		t.Fatalf("file open error = %v", err)
	}
	removed := 0
	removeEvidence = func(string) error { removed++; return nil }
	for name, file := range map[string]failingEvidenceFile{
		"write": {writeErr: boom},
		"sync":  {writeN: -1, syncErr: boom},
		"close": {writeN: -1, closeErr: boom},
	} {
		openEvidenceFile = func(string, int, os.FileMode) (writableEvidenceFile, error) { return file, nil }
		if err := ensureEvidenceFile(name, []byte("x")); !errors.Is(err, boom) {
			t.Fatalf("%s error = %v", name, err)
		}
	}
	openEvidenceFile = func(string, int, os.FileMode) (writableEvidenceFile, error) { return failingEvidenceFile{}, nil }
	if err := ensureEvidenceFile("short", []byte("x")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error = %v", err)
	}
	openEvidenceFile = func(string, int, os.FileMode) (writableEvidenceFile, error) {
		return failingEvidenceFile{writeN: -1}, nil
	}
	restrictPath = func(string, bool) error { return boom }
	if err := ensureEvidenceFile("protect", []byte("x")); !errors.Is(err, boom) || removed != 5 {
		t.Fatalf("file protect error = %v removed=%d", err, removed)
	}
}

func TestReadStatusFileFailureBoundaries(t *testing.T) {
	preserveEvidenceIO(t)
	boom := errors.New("boom")
	path := filepath.Join(t.TempDir(), "status")
	if err := os.WriteFile(path, []byte("valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	openStatusFile = func(string) (io.ReadCloser, error) { return nil, boom }
	if _, err := readStatusFile(path); !errors.Is(err, boom) {
		t.Fatalf("open error = %v", err)
	}
	openStatusFile = func(string) (io.ReadCloser, error) { return errorReadCloser{err: boom}, nil }
	if _, err := readStatusFile(path); err == nil {
		t.Fatal("read error accepted")
	}
	openStatusFile = func(string) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("\xff")), nil }
	if _, err := readStatusFile(path); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	openStatusFile = func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("a" + strings.Repeat("é", 1000))), nil
	}
	text, err := readStatusFile(path)
	if err != nil || !strings.HasSuffix(text, "...truncated...") {
		t.Fatalf("rune truncation = %q, %v", text, err)
	}
}
