package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/logging"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

func baselineModel(t *testing.T) Model {
	t.Helper()
	cfg := config.Default()
	cfg.Dir = filepath.Join(t.TempDir(), config.DirName)
	if err := os.Mkdir(cfg.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	client := llm.New(cfg.ActiveURL(), cfg.ActiveProfile().LLM, "")
	return New(cfg, client, agent.NewRuntime(client, agent.LocalToolExecutor()).WithObserver(logging.NewObserver()), t.TempDir(), "test")
}

func TestCommandBoundariesAndErrorHints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	turn := agent.NewTurnState(nil)
	turn.Begin(ctx, time.Now())
	turn.ID = 7
	stream := agent.NewStreamState()
	stream.Pending = []chmctx.ToolCall{{ID: "1", Name: tools.ReadFileName, Arguments: map[string]any{"path": path}}}
	loop := agent.LoopState{}
	work, _ := loop.NextTool(&turn, &stream, agent.LocalToolExecutor())
	msg := runToolCall(work)().(toolResultMsg)
	if msg.delivery.Message.Content != "hello" || msg.delivery.TurnID != 7 {
		t.Fatalf("tool result = %#v", msg)
	}
	m := baselineModel(t)
	if got := m.errorMessage(nil); got != "" {
		t.Fatalf("nil error = %q", got)
	}
	if got := m.errorMessage(provider.ErrUnauthorized); !strings.Contains(got, "key rejected") {
		t.Fatalf("auth = %q", got)
	}
	un := provider.ErrUnreachable{Err: errors.New("offline")}
	if !agent.IsUnreachable(un) || !strings.Contains(m.errorMessage(un), "unreachable") {
		t.Fatal("unreachable mapping failed")
	}
	if got := m.errorMessage(errors.New("bad")); !strings.Contains(got, "bad") {
		t.Fatalf("generic = %q", got)
	}
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
	cfg := config.Default()
	if backendLabel(cfg, true) == backendLabel(cfg, false) {
		t.Fatal("backend states identical")
	}
	for n, want := range map[int]string{12: "12", 1234: "1,234", 123456: "123,456", 1234567: "1,234,567"} {
		if got := humanInt(n); got != want {
			t.Fatalf("int %d = %q", n, got)
		}
	}
}
