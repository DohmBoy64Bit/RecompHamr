package main

import (
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
)

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("RECOMPHAMR_URL", "http://127.0.0.1:9999")
	cfg := config.Default()
	applyEnvOverrides(cfg)
	if cfg.URLOverride != "http://127.0.0.1:9999" {
		t.Fatalf("URLOverride = %q", cfg.URLOverride)
	}
}
