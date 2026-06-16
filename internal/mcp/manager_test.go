package mcp

import (
	"os"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestRegister(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test", Command: "echo"})
	names := m.AllStatus()
	found := false
	for _, s := range names {
		if s.Name == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Register did not add server")
	}
}

func TestRegisterIdempotent(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test", Command: "echo"})
	m.Register(ServerConfig{Name: "test", Command: "other"})
	if len(m.AllStatus()) != 1 {
		t.Fatalf("expected 1 server, got %d", len(m.AllStatus()))
	}
}

func TestStatusDisconnected(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	s := m.Status("test")
	if s.State != StateDisconnected {
		t.Errorf("expected Disconnected, got %s", s.State)
	}
	if s.Tools != 0 {
		t.Errorf("expected 0 tools, got %d", s.Tools)
	}
}

func TestStatusUnknownServer(t *testing.T) {
	m := NewManager()
	s := m.Status("unknown")
	if s.State != StateDisconnected {
		t.Errorf("expected Disconnected for unknown, got %s", s.State)
	}
}

func TestDisconnectUnknownServer(t *testing.T) {
	m := NewManager()
	m.Disconnect("unknown")
}

func TestFormatStatus(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	out := m.FormatStatus()
	if !strings.Contains(out, "MCP Servers:") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "test") {
		t.Errorf("missing server name: %s", out)
	}
	if !strings.Contains(out, "Disconnected") {
		t.Errorf("missing Disconnected state: %s", out)
	}
}

func TestAllToolsEmptyWhenDisconnected(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	tools := m.AllTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools when disconnected, got %d", len(tools))
	}
}

func TestToolsForSkillsEmpty(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "ghidra", RequireSkill: true})
	tools := m.ToolsForSkills(nil)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools without skills, got %d", len(tools))
	}
}

func TestToolsForSkillsNonRequireSkill(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "always-on", RequireSkill: false})
	// Still 0 because server is not connected
	tools := m.ToolsForSkills(nil)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools for disconnected server, got %d", len(tools))
	}
}

func TestSkillServersMapping(t *testing.T) {
	if s, ok := SkillServers["ghidra-mcp"]; !ok || s != "ghidra" {
		t.Errorf("ghidra-mcp should map to ghidra, got %q", s)
	}
	if s, ok := SkillServers["n64-debug-mcp"]; !ok || s != "n64-debug-mcp" {
		t.Errorf("n64-debug-mcp should map to n64-debug-mcp, got %q", s)
	}
}

func TestSetToolEnabled(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	// Enabling on disconnected server should be fine
	m.SetToolEnabled("test", "my_tool", true)

	m.SetAllToolsEnabled("test", false)
	m.SetToolEnabled("test", "tool_a", true)
	m.SetToolEnabled("test", "tool_a", false)
}

func TestSetToolEnabledUnknownServer(t *testing.T) {
	m := NewManager()
	err := m.SetToolEnabled("unknown", "tool", true)
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention unknown: %s", err)
	}
}

func TestSetAllToolsEnabled(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	m.SetAllToolsEnabled("test", false)
	m.SetAllToolsEnabled("test", true)
}

func TestSetAllToolsEnabledUnknown(t *testing.T) {
	m := NewManager()
	err := m.SetAllToolsEnabled("unknown", true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectedNamesEmpty(t *testing.T) {
	m := NewManager()
	names := m.ConnectedNames()
	if len(names) != 0 {
		t.Errorf("expected empty, got %v", names)
	}
}

func TestAllStatusSorted(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "zzz"})
	m.Register(ServerConfig{Name: "aaa"})
	status := m.AllStatus()
	if len(status) < 2 {
		t.Fatal("expected at least 2 servers")
	}
	if status[0].Name > status[1].Name {
		t.Errorf("expected sorted names, got %v", status)
	}
}

func TestServerConfigDefaults(t *testing.T) {
	cfg := ServerConfig{Name: "test", Command: "cmd"}
	if cfg.RequireSkill {
		t.Error("RequireSkill should default to false")
	}
	if cfg.AllowedTools != nil {
		t.Error("AllowedTools should default to nil")
	}
}

func TestFormatToolsDisconnected(t *testing.T) {
	m := NewManager()
	out := m.FormatTools("unknown")
	if !strings.Contains(out, "unknown") {
		t.Errorf("expected unknown server message, got: %s", out)
	}
}

func TestFormatToolsConnectedNoClient(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test"})
	out := m.FormatTools("test")
	if !strings.Contains(out, "not connected") {
		t.Errorf("expected not connected: %s", out)
	}
}

func TestCallToolUnknownServer(t *testing.T) {
	m := NewManager()
	_, err := m.CallTool(nil, "bad.server_tool", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCallToolInvalidFormat(t *testing.T) {
	m := NewManager()
	_, err := m.CallTool(nil, "no_dot", nil)
	if err == nil {
		t.Fatal("expected error for missing dot")
	}
}

func TestCallToolDisconnectedServer(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "ghost"})
	_, err := m.CallTool(nil, "ghost.tool", nil)
	if err == nil {
		t.Fatal("expected error for disconnected server")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("expected 'not connected' error: %s", err)
	}
}

func TestBuiltinServersGhidraHasAllowedTools(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_GHIDRA_TOOLS")
	os.Unsetenv("RECOMPHAMR_MCP_GHIDRA_TOOLS")
	defer func() { os.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "ghidra" {
			if len(s.AllowedTools) != 20 {
				t.Errorf("ghidra should have 20 default tools, got %d", len(s.AllowedTools))
			}
			if !s.RequireSkill {
				t.Error("built-in ghidra should require skill")
			}
		}
		if s.Name == "n64-debug-mcp" {
			if s.AllowedTools != nil {
				t.Error("n64 should have nil AllowedTools (all tools)")
			}
			if !s.RequireSkill {
				t.Error("built-in n64 should require skill")
			}
		}
		if s.Name == "pcrecomp" {
			if len(s.AllowedTools) != 8 {
				t.Errorf("pcrecomp should have 8 default tools, got %d", len(s.AllowedTools))
			}
			if !s.RequireSkill {
				t.Error("built-in pcrecomp should require skill")
			}
			if s.Command != "pcrecomp-mcp" {
				t.Errorf("pcrecomp command should be pcrecomp-mcp, got %q", s.Command)
			}
		}
	}
}

func TestBuiltinServersGhidraToolsEnvWildcard(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_GHIDRA_TOOLS")
	os.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", "*")
	defer func() { os.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "ghidra" {
			if s.AllowedTools != nil {
				t.Errorf("ghidra AllowedTools should be nil when wildcard, got %v", s.AllowedTools)
			}
		}
	}
}

func TestBuiltinServersGhidraToolsEnvCustomList(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_GHIDRA_TOOLS")
	os.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", "decompile_function,get_xrefs_to")
	defer func() { os.Setenv("RECOMPHAMR_MCP_GHIDRA_TOOLS", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "ghidra" {
			if len(s.AllowedTools) != 2 {
				t.Errorf("ghidra should have 2 tools, got %d: %v", len(s.AllowedTools), s.AllowedTools)
			}
		}
	}
}

func TestParseToolsEnv(t *testing.T) {
	tests := []struct {
		val  string
		want int
	}{
		{"", 0},
		{"*", 0},
		{"a", 1},
		{"a , b , c", 3},
	}
	for _, tt := range tests {
		result := parseToolsEnv(tt.val)
		if len(result) != tt.want {
			t.Errorf("parseToolsEnv(%q) = %d, want %d", tt.val, len(result), tt.want)
		}
	}
}

func TestFirstSentence(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"This is the first sentence. And more.", "This is the first sentence."},
		{"Single line no period", "Single line no period"},
		{"", ""},
		{"Short. Long second sentence. More stuff.", "Short."},
	}
	for _, tt := range tests {
		got := firstSentence(tt.in)
		if got != tt.want {
			t.Errorf("firstSentence(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestToolsForSkillsAutoMatch(t *testing.T) {
	// Test that a skill whose name matches a server name unlocks its tools
	m := NewManager()
	m.Register(ServerConfig{
		Name:         "my-mcp",
		Command:      "echo",
		RequireSkill: true,
	})

	// Server is disconnected, so no tools
	tools := m.ToolsForSkills([]string{"my-mcp"})
	if len(tools) != 0 {
		t.Errorf("expected 0 tools for disconnected server, got %d", len(tools))
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state ServerState
		want  string
	}{
		{StateDisconnected, "Disconnected"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateError, "Error"},
	}
	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestServerStatusFields(t *testing.T) {
	m := NewManager()
	m.Register(ServerConfig{Name: "test", Command: "echo"})

	s := m.Status("test")
	if s.Name != "test" {
		t.Errorf("expected name 'test', got %q", s.Name)
	}
	if s.State != StateDisconnected {
		t.Errorf("expected Disconnected, got %s", s.State)
	}
	if s.Tools != 0 {
		t.Errorf("expected 0 tools, got %d", s.Tools)
	}
	if s.Version != "" {
		t.Errorf("expected empty version, got %q", s.Version)
	}
}

func TestSkillServersHasPcrecomp(t *testing.T) {
	if s, ok := SkillServers["pcrecomp"]; !ok || s != "pcrecomp" {
		t.Errorf("pcrecomp should map to pcrecomp, got %q", s)
	}
}

func TestBuiltinServersHasThreeServers(t *testing.T) {
	servers := BuiltinServers()
	if len(servers) != 3 {
		t.Errorf("expected 3 built-in servers, got %d", len(servers))
	}
	names := map[string]bool{}
	for _, s := range servers {
		names[s.Name] = true
	}
	for _, want := range []string{"ghidra", "n64-debug-mcp", "pcrecomp"} {
		if !names[want] {
			t.Errorf("missing built-in server %q", want)
		}
	}
}

func TestPcrecompToolsEnvWildcard(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS")
	os.Setenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS", "*")
	defer func() { os.Setenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "pcrecomp" {
			if s.AllowedTools != nil {
				t.Errorf("pcrecomp AllowedTools should be nil when wildcard, got %v", s.AllowedTools)
			}
		}
	}
}

func TestPcrecompToolsEnvCustomList(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS")
	os.Setenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS", "pe.analyze,lift32.run")
	defer func() { os.Setenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "pcrecomp" {
			if len(s.AllowedTools) != 2 {
				t.Errorf("pcrecomp should have 2 tools, got %d: %v", len(s.AllowedTools), s.AllowedTools)
			}
		}
	}
}

func TestPcrecompToolsEnvCommand(t *testing.T) {
	old := os.Getenv("RECOMPHAMR_MCP_PCRECOMP_COMMAND")
	os.Setenv("RECOMPHAMR_MCP_PCRECOMP_COMMAND", "/custom/path/pcrecomp-mcp")
	defer func() { os.Setenv("RECOMPHAMR_MCP_PCRECOMP_COMMAND", old) }()

	servers := BuiltinServers()
	for _, s := range servers {
		if s.Name == "pcrecomp" {
			if s.Command != "/custom/path/pcrecomp-mcp" {
				t.Errorf("pcrecomp should use custom command, got %q", s.Command)
			}
		}
	}
}
