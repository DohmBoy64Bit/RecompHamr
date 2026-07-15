//go:build windows

package config

import "golang.org/x/sys/windows"

var (
	openProcessToken     = windows.OpenCurrentProcessToken
	getTokenUser         = func(token windows.Token) (*windows.Tokenuser, error) { return token.GetTokenUser() }
	aclFromEntries       = windows.ACLFromEntries
	setNamedSecurityInfo = windows.SetNamedSecurityInfo
)

func restrictPrivatePath(path string, directory bool) error {
	token, err := openProcessToken()
	if err != nil {
		return err
	}
	defer token.Close()
	user, err := getTokenUser(token)
	if err != nil {
		return err
	}
	inheritance := uint32(windows.NO_INHERITANCE)
	if directory {
		inheritance = uint32(windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT)
	}
	acl, err := aclFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       inheritance,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
	if err != nil {
		return err
	}
	return setNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
}
