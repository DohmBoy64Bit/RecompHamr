package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/logging"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

func baselineModel(t *testing.T) Model {
	t.Helper()
	cfg := config.Default()
	cfg.Dir = filepath.Join(t.TempDir(), config.DirName)
	if err := os.Mkdir(cfg.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	sessionRuntime := session.NewRuntime(cfg)
	return newModelWithRuntime(sessionRuntime, agent.NewRuntime(sessionRuntime, agent.LocalToolExecutor()).WithObserver(logging.NewObserver()), "test system", "test")
}

func TestFormatBoundaries(t *testing.T) {
	for n, want := range map[int]string{0: "0 tok", 999: "999 tok", 1000: "1.0k tok", 1_000_000: "1.0M tok"} {
		if got := humanTokens(n); got != want {
			t.Fatalf("tokens %d = %q", n, got)
		}
	}
	for d, want := range map[time.Duration]string{time.Second: "1s", 61 * time.Second: "1m 01s", 3660 * time.Second: "1h 01m"} {
		if got := liveElapsed(d); got != want {
			t.Fatalf("elapsed %s = %q", d, got)
		}
	}
	if humanRate(0, time.Second) != "" || humanRate(1, 0) != "" || humanRate(5, time.Second) != "5.0 tok/s" || humanRate(20, time.Second) != "20 tok/s" {
		t.Fatal("rate boundaries")
	}
	if backendLabel("local", true) == backendLabel("local", false) {
		t.Fatal("backend states identical")
	}
	for n, want := range map[int]string{12: "12", 1234: "1,234", 123456: "123,456", 1234567: "1,234,567"} {
		if got := humanInt(n); got != want {
			t.Fatalf("int %d = %q", n, got)
		}
	}
}
