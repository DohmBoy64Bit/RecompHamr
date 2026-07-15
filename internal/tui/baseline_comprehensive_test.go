package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
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
	return New(cfg, llm.New(cfg.ActiveURL(), cfg.ActiveProfile().LLM, ""), t.TempDir(), "test")
}

func TestCommandBoundariesAndErrorHints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	msg := runToolCall(ctx, 7, chmctx.ToolCall{ID: "1", Name: tools.ReadFileName, Arguments: map[string]any{"path": path}})().(toolResultMsg)
	if msg.Msg.Content != "hello" || msg.turnID != 7 {
		t.Fatalf("tool result = %#v", msg)
	}
	m := baselineModel(t)
	if got := m.errorMessage(llm.Event{}); got != "" {
		t.Fatalf("nil error = %q", got)
	}
	if got := m.errorMessage(llm.Event{Err: provider.ErrUnauthorized}); !strings.Contains(got, "key rejected") {
		t.Fatalf("auth = %q", got)
	}
	un := provider.ErrUnreachable{Err: errors.New("offline")}
	if !isUnreachable(un) || !strings.Contains(m.errorMessage(llm.Event{Err: un}), "unreachable") {
		t.Fatal("unreachable mapping failed")
	}
	if got := m.errorMessage(llm.Event{Err: errors.New("bad")}); !strings.Contains(got, "bad") {
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

func TestDebugLogLifecycleAndPayloads(t *testing.T) {
	CloseDebugLog()
	OpenDebugLog("")
	if dbgEnabled() {
		t.Fatal("empty path enabled logging")
	}
	dir := t.TempDir()
	OpenDebugLog(dir)
	if !dbgEnabled() {
		t.Fatal("logging not enabled")
	}
	dbgWriteSession("v", "p", "m", "u", 100, 10, []string{"read_file"})
	dbgWriteRequest("m", 100, 80, 2, []chmctx.Message{{Role: chmctx.RoleSystem}, {Role: chmctx.RoleTool, Content: "x\n───── truncated:"}})
	dbgWriteMessage("assistant", chmctx.Message{Role: chmctx.RoleAssistant, Content: "answer", ToolCalls: []chmctx.ToolCall{{ID: "1", Name: "read_file", Arguments: map[string]any{"path": "x"}}}})
	dbgWriteMessage("tool", chmctx.Message{Role: chmctx.RoleTool, ToolName: "read_file", ToolCallID: "1"})
	CloseDebugLog()
	CloseDebugLog()
	raw, err := os.ReadFile(filepath.Join(dir, "log.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"session", "request", "TOOL_CALL", "tool=read_file"} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("log missing %q", want)
		}
	}
	dbgWritef("off", "ignored")
	dbgWriteRequest("m", 1, 1, 0, nil)
	dbgWriteMessage("off", chmctx.Message{})
	OpenDebugLog(filepath.Join(dir, "missing", "child"))
}

func TestToolOutcomeAndNudgeContracts(t *testing.T) {
	cases := []struct {
		name, result string
		failed       bool
	}{
		{tools.WriteFileName, "(write error: x)", true}, {tools.EditFileName, "edited x", false},
		{tools.ReadFileName, "(read error: x)", true}, {tools.ReadFileName, "(valid lisp)", false},
		{tools.PowerShellName, "x\n(exit: 1)", true}, {tools.PowerShellName, "(cancelled)", false},
		{"future", "(unknown tool: future)", true}, {"future", "ok", false},
	}
	for _, tc := range cases {
		if got := toolResultFailed(tc.name, tc.result); got != tc.failed {
			t.Fatalf("%s %q = %v", tc.name, tc.result, got)
		}
	}
	for _, call := range []chmctx.ToolCall{
		{Name: tools.ReadFileName, Arguments: map[string]any{"path": "x"}},
		{Name: tools.PowerShellName, Arguments: map[string]any{"script": " echo x \nnext"}},
		{Name: "future"},
	} {
		if toolTargetKey(call) == "" {
			t.Fatal("empty target key")
		}
	}
	m := baselineModel(t)
	m.lastToolKey = "read_file|x"
	m.recordToolOutcome(tools.ReadFileName, "ok")
	for range maxToolFailStreak {
		m.recordToolOutcome(tools.ReadFileName, "(read error: x)")
	}
	m.maybeFailureNudge()
	if m.failStreak != 0 || len(m.turn.History) == 0 {
		t.Fatal("failure nudge missing")
	}
	m.maybeFailureNudge()
	m.toolRounds = maxToolRounds
	m.maybeRunawayNudge()
	before := len(m.turn.History)
	m.maybeRunawayNudge()
	if !m.runawayNudged || len(m.turn.History) != before {
		t.Fatal("runaway latch failed")
	}
	m.turn.History = append(m.turn.History, chmctx.Message{Role: chmctx.RoleAssistant, Content: "done"})
	m.toolRounds = verifyNudgeMinRounds
	if !m.maybeVerifyNudge() || m.maybeVerifyNudge() {
		t.Fatal("verify latch failed")
	}
}
