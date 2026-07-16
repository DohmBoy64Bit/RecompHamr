//go:build !windows

package workspace

import (
	"os"
	"testing"
)

func TestLoadStateTightensPOSIXPermissions(t *testing.T) {
	workspace := makeWorkspace(t, []byte("state"))
	if err := os.Chmod(workspace.statePath, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.LoadState(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(workspace.statePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("state mode = %o, want 600", info.Mode().Perm())
	}
}
