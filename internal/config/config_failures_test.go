package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func restoreConfigHooks(t *testing.T) {
	origLstat, origMkdir, origRead := lstatPath, mkdirAllPath, readConfigFile
	origMarshal, origCreate, origRename, origRestrict := marshalYAML, createTempFile, renameConfig, restrictPath
	t.Cleanup(func() {
		lstatPath, mkdirAllPath, readConfigFile = origLstat, origMkdir, origRead
		marshalYAML, createTempFile, renameConfig, restrictPath = origMarshal, origCreate, origRename, origRestrict
	})
}

func TestBootstrapReportsFilesystemAndSecurityFailures(t *testing.T) {
	boom := errors.New("boom")
	for _, tc := range []struct {
		name string
		set  func()
	}{
		{"lstat", func() { lstatPath = func(string) (os.FileInfo, error) { return nil, boom } }},
		{"mkdir", func() {
			lstatPath = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
			mkdirAllPath = func(string, os.FileMode) error { return boom }
		}},
		{"secure directory", func() { restrictPath = func(string, bool) error { return boom } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreConfigHooks(t)
			tc.set()
			if _, _, err := Bootstrap(t.TempDir()); !errors.Is(err, boom) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestBootstrapInjectedSymlinksAndSeedFailure(t *testing.T) {
	originalLstat, originalCreate := lstatPath, createTempFile
	t.Cleanup(func() { lstatPath, createTempFile = originalLstat, originalCreate })
	symlink := specialConfigInfo{mode: os.ModeSymlink}
	lstatPath = func(string) (os.FileInfo, error) { return symlink, nil }
	if _, _, err := Bootstrap(t.TempDir()); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("dir error = %v", err)
	}

	lstatPath = originalLstat
	root := t.TempDir()
	dir := filepath.Join(root, DirName)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	lstatPath = func(path string) (os.FileInfo, error) {
		if filepath.Base(path) == "config.yaml" {
			return symlink, nil
		}
		return originalLstat(path)
	}
	if _, _, err := Bootstrap(root); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("file error = %v", err)
	}

	lstatPath = originalLstat
	createTempFile = func(string, string) (temporaryFile, error) { return nil, errors.New("seed failed") }
	if _, _, err := Bootstrap(t.TempDir()); err == nil || !strings.Contains(err.Error(), "seed failed") {
		t.Fatalf("seed error = %v", err)
	}
}

type specialConfigInfo struct{ mode os.FileMode }

func (specialConfigInfo) Name() string        { return "link" }
func (specialConfigInfo) Size() int64         { return 0 }
func (s specialConfigInfo) Mode() os.FileMode { return s.mode }
func (specialConfigInfo) ModTime() time.Time  { return time.Time{} }
func (specialConfigInfo) IsDir() bool         { return false }
func (specialConfigInfo) Sys() any            { return nil }

func TestBootstrapReportsExistingConfigFailures(t *testing.T) {
	boom := errors.New("boom")
	root := t.TempDir()
	dir := filepath.Join(root, DirName)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("active: x\nmodels: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("secure file", func(t *testing.T) {
		restoreConfigHooks(t)
		restrictPath = func(_ string, directory bool) error {
			if !directory {
				return boom
			}
			return nil
		}
		if _, _, err := Bootstrap(root); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("read file", func(t *testing.T) {
		restoreConfigHooks(t)
		readConfigFile = func(string) ([]byte, error) { return nil, boom }
		if _, _, err := Bootstrap(root); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
}

type failingTemp struct {
	name                        string
	writeErr, syncErr, closeErr error
}

func (f *failingTemp) Name() string              { return f.name }
func (f *failingTemp) Write([]byte) (int, error) { return 0, f.writeErr }
func (f *failingTemp) Sync() error               { return f.syncErr }
func (f *failingTemp) Close() error              { return f.closeErr }

func TestWriteYAMLFailureStages(t *testing.T) {
	boom := errors.New("boom")
	restoreConfigHooks(t)
	marshalYAML = func(any) ([]byte, error) { return nil, boom }
	if err := writeYAML("unused", Default()); !errors.Is(err, boom) {
		t.Fatalf("marshal error = %v", err)
	}
	marshalYAML = yaml.Marshal
	for _, tc := range []struct {
		name                              string
		temp                              *failingTemp
		createErr, renameErr, restrictErr error
	}{
		{name: "create", createErr: boom},
		{name: "write", temp: &failingTemp{name: filepath.Join(t.TempDir(), "tmp"), writeErr: boom}},
		{name: "sync", temp: &failingTemp{name: filepath.Join(t.TempDir(), "tmp"), syncErr: boom}},
		{name: "close", temp: &failingTemp{name: filepath.Join(t.TempDir(), "tmp"), closeErr: boom}},
		{name: "rename", temp: &failingTemp{name: filepath.Join(t.TempDir(), "tmp")}, renameErr: boom},
		{name: "restrict", temp: &failingTemp{name: filepath.Join(t.TempDir(), "tmp")}, restrictErr: boom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreConfigHooks(t)
			createTempFile = func(string, string) (temporaryFile, error) { return tc.temp, tc.createErr }
			renameConfig = func(string, string) error { return tc.renameErr }
			restrictPath = func(string, bool) error { return tc.restrictErr }
			if err := writeYAML(filepath.Join(t.TempDir(), "config.yaml"), Default()); !errors.Is(err, boom) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestConfigBoundaryHelpers(t *testing.T) {
	for _, bad := range []string{"", "1BAD", "BAD-NAME"} {
		if isEnvName(bad) {
			t.Fatalf("accepted %q", bad)
		}
	}
	if !isEnvName("A_9") {
		t.Fatal("rejected valid env name")
	}
	p := &Profile{Key: "${1BAD}"}
	if p.ResolvedKey() != "${1BAD}" {
		t.Fatal("expanded invalid env reference")
	}
	c := &Config{Active: "x", Models: map[string]*Profile{"x": {URL: "stored"}}}
	if got := c.ActiveURL(); got != "stored" {
		t.Fatalf("URL = %q", got)
	}
	if err := (&Config{}).Save(); err == nil || !strings.Contains(err.Error(), "Dir not set") {
		t.Fatalf("Save error = %v", err)
	}
}
