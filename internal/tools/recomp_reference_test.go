package tools

import "testing"

func TestRecompRefSchemaShape(t *testing.T) {
	s := RecompRefSchema()
	fn, ok := s["function"].(map[string]any)
	if !ok {
		t.Fatal("missing function key")
	}
	if fn["name"] != RecompRefName {
		t.Errorf("expected name %q, got %q", RecompRefName, fn["name"])
	}
	params, ok := fn["parameters"].(map[string]any)
	if !ok {
		t.Fatal("missing parameters")
	}
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing properties")
	}
	if _, ok := props["url"]; !ok {
		t.Errorf("missing url property")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"example.com", "example.com"},
		{"example.com/path", "example.com_path"},
		{"github.com/user/repo", "github.com_user_repo"},
		{"a?b=c&d=e", "a_b_c_d_e"},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.in)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRecompRefNoDir(t *testing.T) {
	old := RecompRefDir
	RecompRefDir = ""
	defer func() { RecompRefDir = old }()

	result := RecompRef("https://example.com/doc")
	t.Logf("result: %s", result)
}
