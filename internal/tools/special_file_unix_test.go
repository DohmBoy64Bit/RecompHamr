//go:build !windows

package tools

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestFileToolsRefuseFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blocked.fifo")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}

	if got := ReadFile(path); !strings.Contains(got, "not a regular file") {
		t.Fatalf("ReadFile FIFO result = %q", got)
	}
	if got := WriteFile(path, "data"); !strings.Contains(got, "not a regular file") {
		t.Fatalf("WriteFile FIFO result = %q", got)
	}
	if got := EditFile(path, "old", "new"); !strings.Contains(got, "not a regular file") {
		t.Fatalf("EditFile FIFO result = %q", got)
	}

	if info, err := os.Stat(path); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatalf("FIFO should remain intact: info=%v err=%v", info, err)
	}
}
