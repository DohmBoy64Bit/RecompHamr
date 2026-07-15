package tui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

func TestPhaseOutcomeContextAndInitContracts(t *testing.T) {
	for p, want := range map[phase]string{phaseIdle: "", phaseThinking: "thinking", phaseStreaming: "generating", phaseRunning: "running"} {
		if p.Label() != want {
			t.Fatalf("phase %d = %q", p, p.Label())
		}
	}
	if outcomeNone.marker() != "" || outcomeDone.marker() != "✓" || outcomeStopped.marker() != "✗" {
		t.Fatal("outcome markers")
	}
	m := baselineModel(t)
	if m.Init() == nil {
		t.Fatal("init command missing")
	}
	m.cfg.ActiveProfile().Key = "secret"
	if m.Init() == nil {
		t.Fatal("keyed init command missing")
	}
	m.runtime.LiveContextSize[m.cfg.Active] = 1234
	if m.activeContextSize() != 1234 {
		t.Fatal("live context")
	}
	delete(m.runtime.LiveContextSize, m.cfg.Active)
	m.cfg.ActiveProfile().ContextSize = 0
	if m.activeContextSize() != defaultPackFallback {
		t.Fatal("fallback context")
	}
	m.installTurnContext()
	first := m.turn.CancelFunc
	m.installTurnContext()
	if first == nil || m.turn.CancelFunc == nil {
		t.Fatal("turn context")
	}
	m.status = queueSlashHint
	m.runtime.Retrying = true
	m.endTurn()
	if m.status != "" || m.runtime.Phase != phaseIdle {
		t.Fatal("end turn cleanup")
	}
}

func TestStreamEventStateMachine(t *testing.T) {
	m := baselineModel(t)
	m.runtime.Phase = phaseThinking
	m.turnStart = time.Now().Add(-time.Second)
	ch := make(chan llm.Event, 2)
	m.runtime.BeginStream(m.turn.ID, ch)
	next, _ := m.handleStream(llm.Event{Kind: llm.EventRetry, Content: "retry"})
	m = next.(Model)
	if !m.runtime.Retrying || m.status != "retry" {
		t.Fatal("retry")
	}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventReasoning, Content: "reasoning bytes"})
	m = next.(Model)
	if m.runtime.Retrying || m.runtime.StreamingEstimate == 0 {
		t.Fatal("reasoning")
	}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventToolArgs, Content: "arguments"})
	m = next.(Model)
	if m.runtime.Phase != phaseStreaming {
		t.Fatal("tool args phase")
	}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventContent, Content: "a\tb"})
	m = next.(Model)
	if !strings.Contains(m.streaming.String(), "a    b") {
		t.Fatal("content tabs")
	}
	call := chmctx.ToolCall{ID: "1", Name: tools.ReadFileName, Arguments: map[string]any{"path": "x"}}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventToolCall, ToolCall: &call})
	m = next.(Model)
	if len(m.runtime.Pending) != 1 {
		t.Fatal("tool call not queued")
	}
	final := chmctx.Message{Role: chmctx.RoleAssistant, Content: "done", ToolCalls: []chmctx.ToolCall{call}}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventDone, Final: &final, Tokens: 9, PromptTokens: 4000, ContextWindow: 4096})
	m = next.(Model)
	if m.runtime.TurnTokens != 9 || m.runtime.LiveContextSize[m.cfg.Active] != 4096 || len(m.turn.History) == 0 {
		t.Fatal("done accounting")
	}
	next, cmd := m.handleStreamClosed()
	m = next.(Model)
	if cmd == nil || m.runtime.Phase != phaseRunning || len(m.runtime.Pending) != 0 {
		t.Fatal("dispatch pending")
	}

	turnCtx, cancel := context.WithCancel(context.Background())
	cancel()
	m.turn.Context = turnCtx
	m.turn.ID = 1
	toolMsg := chmctx.Message{Role: chmctx.RoleTool, ToolName: tools.ReadFileName, Content: "ok"}
	next, _ = m.update(toolResultMsg{Msg: toolMsg, turnID: 1})
	m = next.(Model)
	if m.runtime.Phase != phaseThinking {
		t.Fatal("tool result did not resume")
	}
	before := len(m.turn.History)
	next, _ = m.update(toolResultMsg{Msg: toolMsg, turnID: 2})
	m = next.(Model)
	if len(m.turn.History) != before {
		t.Fatal("stale tool result")
	}

	m.runtime.Phase = phaseStreaming
	m.turnStart = time.Now().Add(-time.Second)
	m.streaming.WriteString("partial")
	next, _ = m.handleStream(llm.Event{Kind: llm.EventError, Err: provider.ErrUnreachable{Err: errors.New("down")}})
	m = next.(Model)
	if m.runtime.Phase != phaseIdle || m.runtime.Connected {
		t.Fatal("error unwind")
	}
	next, _ = m.handleStream(llm.Event{Kind: llm.EventContent, Content: "stale"})
	if next.(Model).runtime.Phase != phaseIdle {
		t.Fatal("inactive event")
	}
}

func TestUpdateTypedMessagesAndClosedOutcomes(t *testing.T) {
	m := baselineModel(t)
	for _, msg := range []tea.Msg{tea.FocusMsg{}, tea.BlurMsg{}, tea.KeyMsg{Type: tea.KeyRunes}, spinner.TickMsg{}} {
		next, _ := m.update(msg)
		m = next.(Model)
	}
	current := make(chan llm.Event)
	staleEvents := make(chan llm.Event, 1)
	staleEvents <- llm.Event{Kind: llm.EventContent}
	close(staleEvents)
	staleState := agent.NewStreamState()
	stale := staleState.BeginStream(m.turn.ID+1, staleEvents)
	m.runtime.BeginStream(m.turn.ID, current)
	m.runtime.Phase = phaseThinking
	if _, cmd := m.update(streamEventMsg{stream: stale, delivery: stale.Read()}); cmd == nil {
		t.Fatal("stale event not drained")
	}
	if _, cmd := m.update(streamClosedMsg{stream: stale, delivery: stale.Read()}); cmd != nil {
		t.Fatal("stale close acted")
	}
	m.cli.BaseURL = "http://current"
	next, _ := m.update(pingMsg{ok: false, baseURL: "http://stale"})
	m = next.(Model)
	if !m.runtime.Connected {
		t.Fatal("stale ping")
	}
	next, _ = m.update(pingMsg{ok: false, baseURL: "http://current"})
	m = next.(Model)
	if m.runtime.Connected {
		t.Fatal("live ping")
	}
	m.quitArmedAt = time.Now().Add(-time.Second)
	m.status = quitArmText
	next, _ = m.update(quitArmResetMsg{})
	m = next.(Model)
	if m.status != "" {
		t.Fatal("quit reset")
	}
	next, _ = m.update(struct{}{})
	_ = next.(Model)

	m = baselineModel(t)
	m.runtime.Phase = phaseThinking
	m.turnStart = time.Now().Add(-time.Second)
	closedCtx, closeCancel := context.WithCancel(context.Background())
	closeCancel()
	m.turn.Context = closedCtx
	m.turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant}}
	next, cmd := m.handleStreamClosed()
	m = next.(Model)
	if cmd == nil || !m.emptyNudged {
		t.Fatal("empty nudge")
	}
	m.runtime.Phase = phaseThinking
	m.runtime.Stream = nil
	next, _ = m.handleStreamClosed()
	m = next.(Model)
	if m.runtime.Phase != phaseIdle || !strings.Contains(m.scroll.String(), "ended its turn") {
		t.Fatal("empty stall")
	}
	m = baselineModel(t)
	m.runtime.Phase = phaseThinking
	m.turnStart = time.Now().Add(-time.Second)
	m.turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "<tool_call>bad"}}
	next, _ = m.handleStreamClosed()
	m = next.(Model)
	if m.lastOutcome != outcomeStopped {
		t.Fatal("leak outcome")
	}
}

func TestProbeSuccessFailureAndBackendCommands(t *testing.T) {
	m := baselineModel(t)
	next, _ := m.handleProbe(probeMsg{profile: m.cfg.Active, err: provider.ErrUnauthorized})
	m = next.(Model)
	if m.runtime.Connected {
		t.Fatal("failed probe connected")
	}
	next, _ = m.handleProbe(probeMsg{profile: m.cfg.Active, silent: true, err: errors.New("quiet")})
	m = next.(Model)
	next, _ = m.handleProbe(probeMsg{profile: "vanished", contextWindow: 10})
	m = next.(Model)
	next, _ = m.handleProbe(probeMsg{profile: m.cfg.Active, contextWindow: 8192})
	m = next.(Model)
	if !m.runtime.Connected || m.runtime.LiveContextSize[m.cfg.Active] != 8192 {
		t.Fatal("probe success")
	}
	if probeErrorMessage(provider.ErrUnauthorized) != "key rejected" || !strings.Contains(probeErrorMessage(provider.ErrUnreachable{Err: errors.New("down")}), "unreachable") || probeErrorMessage(errors.New("raw")) != "raw" {
		t.Fatal("probe hints")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v1/models") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"completion_tokens":1}}`))
	}))
	defer srv.Close()
	if got := pingBackend(srv.URL)().(pingMsg); !got.ok {
		t.Fatal("ping failed")
	}
	cli := llm.New(srv.URL, "model", "key")
	if got := probeBackend(cli, "p", false)().(probeMsg); got.err != nil {
		t.Fatalf("probe command = %v", got.err)
	}
}

func TestRemainingStateBranches(t *testing.T) {
	dir := t.TempDir()
	OpenDebugLog(dir)
	m := baselineModel(t)
	CloseDebugLog()
	if len(m.promptHistory) != 0 {
		t.Fatal("unexpected history")
	}
	next, _ := m.update(probeMsg{profile: m.cfg.Active, silent: true})
	m = next.(Model)
	m.runtime.Pending = []chmctx.ToolCall{{Name: tools.ReadFileName}, {Name: tools.ReadFileName}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m.turn.Context = ctx
	m.turn.ID = 1
	next, cmd := m.update(toolResultMsg{Msg: chmctx.Message{Role: chmctx.RoleTool}, turnID: 1})
	m = next.(Model)
	if cmd == nil || len(m.runtime.Pending) != 1 {
		t.Fatal("pending tool chain")
	}
	m.resizeGen = 2
	_, tick := m.handleWindowSize(tea.WindowSizeMsg{Width: 20, Height: 10})
	if tick != nil {
		_ = tick()
	}

	OpenDebugLog(dir)
	m.runtime.Phase = phaseThinking
	m.runtime.BeginStream(m.turn.ID, make(chan llm.Event))
	next, _ = m.handleStream(llm.Event{Kind: llm.EventReasoning, Content: "logged reasoning"})
	m = next.(Model)
	m.runtime.StreamingEstimate = 8
	m.reasoning.WriteString("reason")
	m.applyDone(llm.Event{Kind: llm.EventDone})
	CloseDebugLog()
	if m.runtime.TurnTokens == 0 {
		t.Fatal("estimated done tokens")
	}
	m.runtime.Phase = phaseIdle
	next, _ = m.handleStreamClosed()
	if next.(Model).runtime.Phase != phaseIdle {
		t.Fatal("idle close")
	}
	if _, ok := newestAssistant([]chmctx.Message{{Role: chmctx.RoleUser}}); ok {
		t.Fatal("found absent assistant")
	}
	if toolCallLeakWarning([]chmctx.Message{{Role: chmctx.RoleAssistant, Content: "<tool_call>", ToolCalls: []chmctx.ToolCall{{Name: "x"}}}}) != "" {
		t.Fatal("structured call warned")
	}
	m.turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "UNVERIFIED: runtime"}}
	m.toolRounds = verifyNudgeMinRounds
	m.verifyNudged = false
	if m.maybeVerifyNudge() {
		t.Fatal("honest unverified finish nudged")
	}
	if _, ok := quitArmReset(time.Now()).(quitArmResetMsg); !ok {
		t.Fatal("quit reset callback")
	}
	if got := resizeSettled(7)(time.Now()).(resizeSettleMsg); got.gen != 7 {
		t.Fatal("resize callback")
	}
}
