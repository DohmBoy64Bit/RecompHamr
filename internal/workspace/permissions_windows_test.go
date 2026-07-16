//go:build windows

package workspace

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestLoadStateInstallsProtectedCurrentUserDACL(t *testing.T) {
	workspace := makeWorkspace(t, []byte("state"))
	if _, err := workspace.LoadState(); err != nil {
		t.Fatal(err)
	}
	sd, err := windows.GetNamedSecurityInfo(workspace.statePath, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if dacl == nil || dacl.AceCount != 1 {
		t.Fatalf("DACL ACE count = %v, want one current-user entry", dacl)
	}
	control, _, err := sd.Control()
	if err != nil {
		t.Fatal(err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatal("workspace-state DACL is not protected")
	}
}
