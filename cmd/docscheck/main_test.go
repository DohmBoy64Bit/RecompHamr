package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "AGENTS.md")
		if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := run([]string{path}, os.Stat); got != 0 {
			t.Fatalf("run() = %d, want 0", got)
		}
	})

	t.Run("missing", func(t *testing.T) {
		stat := func(string) (os.FileInfo, error) { return nil, errors.New("missing") }
		if got := run([]string{"missing"}, stat); got != 1 {
			t.Fatalf("run() = %d, want 1", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty.md")
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if got := run([]string{path}, os.Stat); got != 1 {
			t.Fatalf("run() = %d, want 1", got)
		}
	})

	t.Run("directory", func(t *testing.T) {
		if got := run([]string{t.TempDir()}, os.Stat); got != 1 {
			t.Fatalf("run() = %d, want 1", got)
		}
	})
}

func TestCheckContent(t *testing.T) {
	read := func(path string) ([]byte, error) {
		if path == "error" {
			return nil, errors.New("read")
		}
		return []byte("documented term"), nil
	}

	if got := checkContent(map[string][]string{"ok": {"term"}}, read); got != 0 {
		t.Fatalf("checkContent present = %d, want 0", got)
	}
	if got := checkContent(map[string][]string{"ok": {"missing"}}, read); got != 1 {
		t.Fatalf("checkContent missing term = %d, want 1", got)
	}
	if got := checkContent(map[string][]string{"error": {"term"}}, read); got != 1 {
		t.Fatalf("checkContent read error = %d, want 1", got)
	}
}

func TestCheckContractFile(t *testing.T) {
	stat := func(string) (os.FileInfo, error) {
		path := filepath.Join(t.TempDir(), "doc.md")
		if err := os.WriteFile(path, []byte("term"), 0o644); err != nil {
			t.Fatal(err)
		}
		return os.Stat(path)
	}

	t.Run("valid", func(t *testing.T) {
		read := func(path string) ([]byte, error) {
			if path == "contract.json" {
				return []byte(`{"required":["doc.md"],"required_content":{"doc.md":["term"]}}`), nil
			}
			return []byte("term"), nil
		}
		if got := checkContractFile("contract.json", read, stat); got != 0 {
			t.Fatalf("checkContractFile() = %d, want 0", got)
		}
	})

	t.Run("read error", func(t *testing.T) {
		read := func(string) ([]byte, error) { return nil, errors.New("read") }
		if got := checkContractFile("contract.json", read, stat); got != 1 {
			t.Fatalf("checkContractFile() = %d, want 1", got)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		read := func(string) ([]byte, error) { return []byte("{"), nil }
		if got := checkContractFile("contract.json", read, stat); got != 1 {
			t.Fatalf("checkContractFile() = %d, want 1", got)
		}
	})

	t.Run("empty required list", func(t *testing.T) {
		read := func(string) ([]byte, error) { return []byte(`{"required":[]}`), nil }
		if got := checkContractFile("contract.json", read, stat); got != 1 {
			t.Fatalf("checkContractFile() = %d, want 1", got)
		}
	})

	t.Run("required file failure stops content check", func(t *testing.T) {
		readCalls := 0
		read := func(string) ([]byte, error) {
			readCalls++
			return []byte(`{"required":["missing.md"],"required_content":{"missing.md":["term"]}}`), nil
		}
		missingStat := func(string) (os.FileInfo, error) { return nil, errors.New("missing") }
		if got := checkContractFile("contract.json", read, missingStat); got != 1 {
			t.Fatalf("checkContractFile() = %d, want 1", got)
		}
		if readCalls != 1 {
			t.Fatalf("read calls = %d, want 1", readCalls)
		}
	})
}

func TestMainFunction(t *testing.T) {
	oldExit, oldRead, oldStat := exitProcess, readDocument, statPath
	t.Cleanup(func() {
		exitProcess, readDocument, statPath = oldExit, oldRead, oldStat
	})

	readDocument = func(path string) ([]byte, error) {
		if path == contractPath {
			return []byte(`{"required":["doc.md"],"required_content":{"doc.md":["content"]}}`), nil
		}
		return []byte("content"), nil
	}
	statPath = func(string) (os.FileInfo, error) {
		path := filepath.Join(t.TempDir(), "doc.md")
		if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
		return os.Stat(path)
	}

	called := false
	exitProcess = func(code int) {
		called = true
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
	}
	main()
	if !called {
		t.Fatal("main did not request exit")
	}
}
