package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

func activeTurn(t *testing.T) TurnState {
	t.Helper()
	turn := NewTurnState(nil)
	turn.Begin(context.Background(), time.Now())
	return turn
}

func TestToolWorkAndSequentialPairing(t *testing.T) {
	turn := activeTurn(t)
	stream := NewStreamState()
	loop := LoopState{}
	if work, ok := loop.NextTool(&turn, &stream, NewToolExecutor(nil)); ok || work != nil {
		t.Fatal("empty queue")
	}
	calls := []chmctx.ToolCall{
		{ID: "1", Name: tools.ReadFileName, Arguments: map[string]any{"path": "a"}},
		{ID: "2", Name: tools.ReadFileName, Arguments: map[string]any{"path": "b"}},
	}
	stream.Pending = append(stream.Pending, calls...)
	var executed []string
	executor := NewToolExecutor(func(ctx context.Context, call chmctx.ToolCall) chmctx.Message {
		if ctx != turn.context {
			t.Fatal("wrong context")
		}
		executed = append(executed, call.ID)
		return chmctx.Message{Role: chmctx.RoleTool, ToolCallID: call.ID, ToolName: call.Name, Content: "ok"}
	})
	work, ok := loop.NextTool(&turn, &stream, executor)
	if !ok || !strings.Contains(work.Status(), "read_file") || stream.Phase != PhaseRunning || loop.ToolRounds != 1 {
		t.Fatal("first work")
	}
	first := work.Run()
	if first.TurnID != turn.ID || first.Message.ToolCallID != "1" {
		t.Fatal("first delivery")
	}
	effect := loop.ApplyToolResult(&turn, &stream, first)
	if !effect.Accepted || !effect.ContinueTools || effect.Failed || len(turn.History) != 1 {
		t.Fatal("first result")
	}
	work, _ = loop.NextTool(&turn, &stream, executor)
	effect = loop.ApplyToolResult(&turn, &stream, work.Run())
	if !effect.Accepted || effect.ContinueTools || stream.Phase != PhaseThinking || len(turn.History) != 2 || strings.Join(executed, ",") != "1,2" {
		t.Fatal("second result")
	}
	if effect = loop.ApplyToolResult(&turn, &stream, ToolDelivery{TurnID: turn.ID + 1}); effect.Accepted {
		t.Fatal("stale result")
	}
	production := LocalToolExecutor()
	stream.Pending = []chmctx.ToolCall{{ID: "3", Name: tools.ReadFileName, Arguments: map[string]any{"path": ""}}}
	work, _ = loop.NextTool(&turn, &stream, production)
	if delivery := work.Run(); delivery.Message.ToolCallID != "3" {
		t.Fatal("production executor")
	}
}

func TestToolCancellationAndStaleCleanup(t *testing.T) {
	turn := activeTurn(t)
	stream := NewStreamState()
	loop := LoopState{}
	stream.Pending = []chmctx.ToolCall{{ID: "slow", Name: tools.PowerShellName}}
	executor := NewToolExecutor(func(ctx context.Context, call chmctx.ToolCall) chmctx.Message {
		<-ctx.Done()
		return chmctx.Message{Role: chmctx.RoleTool, ToolCallID: call.ID, ToolName: call.Name, Content: "(cancelled)"}
	})
	work, _ := loop.NextTool(&turn, &stream, executor)
	result := make(chan ToolDelivery, 1)
	go func() { result <- work.Run() }()
	oldID := turn.ID
	turn.End()
	delivery := <-result
	if delivery.TurnID != oldID || delivery.Message.Content != "(cancelled)" {
		t.Fatal("cancel delivery")
	}
	before := len(turn.History)
	if effect := loop.ApplyToolResult(&turn, &stream, delivery); effect.Accepted || len(turn.History) != before {
		t.Fatal("cancelled result accepted after turn ended")
	}
	turn.Begin(context.Background(), time.Now())
	if effect := loop.ApplyToolResult(&turn, &stream, delivery); effect.Accepted || len(turn.History) != before {
		t.Fatal("stale cancelled result")
	}
}

func TestToolFailureAndRunawayPolicy(t *testing.T) {
	turn := activeTurn(t)
	stream := NewStreamState()
	loop := LoopState{}
	executor := NewToolExecutor(func(context.Context, chmctx.ToolCall) chmctx.Message {
		return chmctx.Message{Role: chmctx.RoleTool, ToolName: tools.ReadFileName, Content: "(read error: x)"}
	})
	for i := 1; i <= MaxToolFailStreak; i++ {
		stream.Pending = []chmctx.ToolCall{{Name: tools.ReadFileName, Arguments: map[string]any{"path": "x"}}}
		work, _ := loop.NextTool(&turn, &stream, executor)
		effect := loop.ApplyToolResult(&turn, &stream, work.Run())
		if !effect.Failed || effect.FailStreak != i {
			t.Fatalf("failure %d", i)
		}
		if i < MaxToolFailStreak && effect.FailureNudged {
			t.Fatal("early failure nudge")
		}
		if i == MaxToolFailStreak && (!effect.FailureNudged || effect.FailureCount != MaxToolFailStreak || loop.FailStreak != 0) {
			t.Fatal("missing failure nudge")
		}
	}
	loop.FailKey, loop.FailStreak = "x", 4
	stream.Pending = []chmctx.ToolCall{
		{Name: tools.ReadFileName, Arguments: map[string]any{"path": "x"}},
		{Name: tools.ReadFileName, Arguments: map[string]any{"path": "x"}},
	}
	work, _ := loop.NextTool(&turn, &stream, executor)
	effect := loop.ApplyToolResult(&turn, &stream, work.Run())
	if !effect.ContinueTools || effect.FailureNudged {
		t.Fatal("nudge before queue drain")
	}
	loop.ToolRounds = MaxToolRounds - 1
	stream.Pending = []chmctx.ToolCall{{Name: tools.ReadFileName, Arguments: map[string]any{"path": "y"}}}
	work, _ = loop.NextTool(&turn, &stream, executor)
	effect = loop.ApplyToolResult(&turn, &stream, work.Run())
	if !effect.RunawayNudged || effect.ToolRounds != MaxToolRounds {
		t.Fatal("runaway nudge")
	}
	before := len(turn.History)
	stream.Pending = []chmctx.ToolCall{{Name: tools.ReadFileName, Arguments: map[string]any{"path": "z"}}}
	work, _ = loop.NextTool(&turn, &stream, executor)
	loop.ApplyToolResult(&turn, &stream, work.Run())
	if len(turn.History) != before+1 {
		t.Fatal("runaway repeated")
	}
	loop.ResetUserGoal()
	loop.ResetTurn()
	if loop.FailKey != "" || loop.FailStreak != 0 || loop.ToolRounds != 0 || loop.RunawayNudged || loop.EmptyNudged || loop.VerifyNudged {
		t.Fatal("reset")
	}
}

func TestCloseDecisions(t *testing.T) {
	turn := activeTurn(t)
	stream := NewStreamState()
	stream.Phase = PhaseThinking
	loop := LoopState{}
	stream.Pending = []chmctx.ToolCall{{Name: "x"}}
	loop.EmptyNudged = true
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseRunTool || decision.Reason != CloseToolPending || loop.EmptyNudged {
		t.Fatal("pending")
	}
	stream.Pending = nil
	turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant}}
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseRestartModel || decision.Reason != CloseEmptyNudge || !loop.EmptyNudged || stream.Phase != PhaseThinking {
		t.Fatal("empty nudge")
	}
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseFinishStopped || decision.Reason != CloseEmptyStall {
		t.Fatal("empty stall")
	}
	turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "<tool_call>bad"}}
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseFinishStopped || decision.Reason != CloseToolLeak {
		t.Fatal("leak")
	}
	turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "done"}}
	loop.ToolRounds = VerifyNudgeMinRounds
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseRestartModel || decision.Reason != CloseVerifyNudge || decision.ToolRounds != VerifyNudgeMinRounds || !loop.VerifyNudged {
		t.Fatal("verify")
	}
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseFinishDone || decision.Reason != CloseClean {
		t.Fatal("clean after verify")
	}
	loop = LoopState{ToolRounds: VerifyNudgeMinRounds}
	turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "UNVERIFIED: runtime"}}
	if decision := loop.DecideClose(&turn, &stream); decision.Action != CloseFinishDone || loop.VerifyNudged {
		t.Fatal("unverified suppression")
	}
}
