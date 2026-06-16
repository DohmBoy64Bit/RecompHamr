package doctor

import (
	"strings"
	"testing"

	"github.com/DohmBoy64Bit/recomphamr/internal/config"
)

func TestRunContainsExpectedSections(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	sections := []string{
		"RecompHAMR doctor",
		"OS/arch:",
		"Go runtime:",
		"Project:",
		"Active profile:",
		"Model:",
		"Profiles",
		"Memory/GPU hints",
		"Toolchain hints",
		"MCP servers",
		"Endpoint check",
		"Workspace (.rehamr/)",
	}
	for _, s := range sections {
		if !strings.Contains(result, s) {
			t.Errorf("missing section %q", s)
		}
	}
}

func TestRunShowsWorkspaceFiles(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	// All workspace files should show as missing when workspace isn't initialized
	for _, label := range []string{
		"PROJECT.md", "REPHAMR_STATE.md", "EVIDENCE.md",
		"BLOCKERS.md", "CHANGELOG.md", "repomix-instruction.md",
	} {
		if !strings.Contains(result, label) {
			t.Errorf("workspace should mention %s", label)
		}
	}
}

func TestRunShowsProfileList(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	profiles := []string{"lmstudio-amd", "lmstudio-fast", "ollama-amd", "llama-vulkan"}
	for _, p := range profiles {
		if !strings.Contains(result, p) {
			t.Errorf("should list profile %q", p)
		}
	}
}

func TestRunShowsMCPEnvVars(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	envVars := []string{
		"RECOMPHAMR_MCP_GHIDRA_COMMAND",
		"RECOMPHAMR_MCP_N64_COMMAND",
		"RECOMPHAMR_MCP_GHIDRA_TOOLS",
		"RECOMPHAMR_MCP_AUTOSTART",
	}
	for _, ev := range envVars {
		if !strings.Contains(result, ev) {
			t.Errorf("should show env var %s", ev)
		}
	}
}

func TestRunShowsMCPServerTools(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	if !strings.Contains(result, "ghidra-mcp") {
		t.Error("should mention ghidra-mcp")
	}
	if !strings.Contains(result, "n64-debug-mcp") {
		t.Error("should mention n64-debug-mcp")
	}
}

func TestWhichFound(t *testing.T) {
	result := which("go")
	if !strings.Contains(result, "go") {
		t.Errorf("which(go) should contain 'go', got: %s", result)
	}
}

func TestWhichMissing(t *testing.T) {
	result := which("nonexistent-tool-xyz-123")
	if !strings.Contains(result, "missing") {
		t.Errorf("which should report missing, got: %s", result)
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}
	for _, tt := range tests {
		got := humanSize(tt.n)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
