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
	if !strings.Contains(result, "RecompHAMR doctor") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "OS/arch:") {
		t.Error("missing OS/arch")
	}
	if !strings.Contains(result, "Go runtime:") {
		t.Error("missing Go runtime")
	}
	if !strings.Contains(result, "Project:") {
		t.Error("missing project dir")
	}
	if !strings.Contains(result, "Active profile:") {
		t.Error("missing active profile")
	}
	if !strings.Contains(result, "Model:") {
		t.Error("missing model")
	}
	if !strings.Contains(result, "Memory/GPU hints") {
		t.Error("missing memory section")
	}
	if !strings.Contains(result, "Toolchain hints") {
		t.Error("missing toolchain section")
	}
	if !strings.Contains(result, "Endpoint check") {
		t.Error("missing endpoint section")
	}
}

func TestRunShowsWorkspaceStatus(t *testing.T) {
	cfg := *config.Default()
	cfg.Dir = t.TempDir()

	result := Run(t.TempDir(), cfg, "")
	if !strings.Contains(result, "not initialized") {
		t.Errorf("expected 'not initialized' when workspace missing, got: %s", result)
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
