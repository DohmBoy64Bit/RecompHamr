//go:build windows

package config

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRestrictPrivatePathPropagatesWindowsAPIFailures(t *testing.T) {
	boom := errors.New("boom")
	origOpen, origUser, origACL, origSet := openProcessToken, getTokenUser, aclFromEntries, setNamedSecurityInfo
	t.Cleanup(func() {
		openProcessToken, getTokenUser, aclFromEntries, setNamedSecurityInfo = origOpen, origUser, origACL, origSet
	})

	openProcessToken = func() (windows.Token, error) { return 0, boom }
	if err := restrictPrivatePath("unused", false); !errors.Is(err, boom) {
		t.Fatalf("open error = %v", err)
	}

	openProcessToken = origOpen
	getTokenUser = func(windows.Token) (*windows.Tokenuser, error) { return nil, boom }
	if err := restrictPrivatePath("unused", false); !errors.Is(err, boom) {
		t.Fatalf("user error = %v", err)
	}

	getTokenUser = origUser
	aclFromEntries = func([]windows.EXPLICIT_ACCESS, *windows.ACL) (*windows.ACL, error) { return nil, boom }
	if err := restrictPrivatePath("unused", false); !errors.Is(err, boom) {
		t.Fatalf("ACL error = %v", err)
	}

	aclFromEntries = origACL
	setNamedSecurityInfo = func(string, windows.SE_OBJECT_TYPE, windows.SECURITY_INFORMATION, *windows.SID, *windows.SID, *windows.ACL, *windows.ACL) error {
		return boom
	}
	if err := restrictPrivatePath(t.TempDir(), true); !errors.Is(err, boom) {
		t.Fatalf("set error = %v", err)
	}
}
