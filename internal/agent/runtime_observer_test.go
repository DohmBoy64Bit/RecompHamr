package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
)

type observerRecord struct {
	category string
	body     string
}

type fakeObserver struct {
	enabled  bool
	records  []observerRecord
	messages []chmctx.Message
}

func (o *fakeObserver) Enabled() bool { return o.enabled }
func (o *fakeObserver) Writef(category, format string, args ...any) {
	o.records = append(o.records, observerRecord{category: category, body: fmt.Sprintf(format, args...)})
}
func (o *fakeObserver) WriteMessage(category string, message chmctx.Message) {
	o.records = append(o.records, observerRecord{category: category})
	o.messages = append(o.messages, message)
}

func observedRuntime(observer Observer) Runtime {
	client := &fakeChatClient{events: make(chan llm.Event)}
	return NewRuntime(client, NewToolExecutor(func(_ context.Context, call chmctx.ToolCall) chmctx.Message {
		return chmctx.Message{Role: chmctx.RoleTool, ToolCallID: call.ID, ToolName: call.Name, Content: "ok"}
	})).WithObserver(observer)
}

func TestRuntimeLifecycleAndObserver(t *testing.T) {
	observer := &fakeObserver{enabled: true}
	runtime := observedRuntime(observer)
	if snapshot := runtime.Snapshot(); snapshot.Phase != PhaseIdle || !snapshot.Connected || runtime.Active() {
		t.Fatalf("initial snapshot = %#v", snapshot)
	}
	if _, ok := runtime.LiveContextSize("p"); ok {
		t.Fatal("unexpected context hint")
	}
	runtime.SetLiveContextSize("p", 0)
	runtime.SetLiveContextSize("p", 4096)
	if contextSize, ok := runtime.LiveContextSize("p"); !ok || contextSize != 4096 {
		t.Fatal("context hint")
	}
	runtime.SetConnected(false)
	if runtime.Snapshot().Connected {
		t.Fatal("connection snapshot")
	}
	runtime.ObserveUser("hello")
	runtime.ObserveCancel()
	runtime.ObserveSlash("/models")
	runtime.ObserveSession("v", "p", "m", "u", 100, 10)
	runtime.BeginTurn(time.Now().Add(-time.Second))
	runtime.AppendUser("hello")
	runtime.Stream.Retrying = true
	if !runtime.EndTurn() || runtime.Stream.Phase != PhaseIdle {
		t.Fatal("end turn")
	}
	runtime.BeginTurn(time.Now().Add(-time.Second))
	runtime.Stream.StreamingEstimate = 4
	if wall, tokens := runtime.FinalizeTurn(time.Now().Add(-time.Second), time.Now()); wall <= 0 || tokens != 4 {
		t.Fatalf("finalize = %s %d", wall, tokens)
	}
	runtime.ObserveTurnEnd("4 tok", "4 tok", time.Second, " · 4 tok/s avg")
	runtime.ResetConversation()
	if len(runtime.Turn.History) != 0 || runtime.Stream.SessionTokens != 0 || runtime.Loop.FailStreak != 0 {
		t.Fatal("reset conversation")
	}
	if len(observer.records) < 3 {
		t.Fatal("missing lifecycle records")
	}
}

func TestRuntimeStreamObservation(t *testing.T) {
	observer := &fakeObserver{enabled: true}
	runtime := observedRuntime(observer)
	runtime.BeginTurn(time.Now())
	runtime.Turn.History = []chmctx.Message{{Role: chmctx.RoleUser, Content: "───── truncated: x"}}
	stream, summary := runtime.StartRound("system", "model", 200)
	if summary.Truncated != 1 || stream == nil || runtime.CurrentStream() != stream {
		t.Fatal("request summary")
	}
	runtime.ApplyEvent("p", 100, llm.Event{Kind: llm.EventRetry, Content: "wait", Err: errors.New("retry")})
	runtime.ApplyEvent("p", 100, llm.Event{Kind: llm.EventReasoning, Content: "reason"})
	final := chmctx.Message{Role: chmctx.RoleAssistant, Content: "done"}
	runtime.ApplyEvent("p", 100, llm.Event{Kind: llm.EventDone, Final: &final, Tokens: 2, PromptTokens: 96, ContextWindow: 100, Elapsed: time.Second})
	runtime.ApplyEvent("p", 100, llm.Event{Kind: llm.EventError, Err: errors.New("boom")})

	closed := StreamDelivery{turnID: runtime.Turn.ID, roundID: stream.roundID, closed: true}
	if effect := runtime.ApplyDelivery(stream, "p", 100, closed); !effect.Accepted || !effect.Closed {
		t.Fatal("closed delivery")
	}
	stale := StreamDelivery{turnID: runtime.Turn.ID + 1, roundID: stream.roundID}
	if effect := runtime.ApplyDelivery(stream, "p", 100, stale); effect.Accepted {
		t.Fatal("stale delivery")
	}
	current := StreamDelivery{turnID: runtime.Turn.ID, roundID: stream.roundID, event: llm.Event{Kind: llm.EventContent, Content: "x"}}
	if effect := runtime.ApplyDelivery(stream, "p", 100, current); !effect.Accepted || effect.Stream.Content != "x" {
		t.Fatal("current delivery")
	}

	disabled := &fakeObserver{}
	disabledRuntime := observedRuntime(disabled)
	disabledRuntime.BeginTurn(time.Now())
	disabledRuntime.ApplyEvent("p", 100, llm.Event{Kind: llm.EventReasoning, Content: "private"})
	if disabledRuntime.reasoning.Len() != 0 {
		t.Fatal("disabled observer retained reasoning")
	}
	if len(observer.messages) != 1 || !hasCategory(observer.records, "ctx_pressure") || !hasCategory(observer.records, "error") {
		t.Fatal("missing stream records")
	}
}

func TestRuntimeToolAndCloseObservation(t *testing.T) {
	observer := &fakeObserver{enabled: true}
	runtime := observedRuntime(observer)
	runtime.BeginTurn(time.Now())
	if work, ok := runtime.NextTool(); ok || work != nil {
		t.Fatal("empty tool queue")
	}
	runtime.Stream.Pending = []chmctx.ToolCall{{ID: "1", Name: "read_file"}}
	work, ok := runtime.NextTool()
	if !ok || work == nil {
		t.Fatal("tool dispatch")
	}
	if effect := runtime.ApplyToolResult(work.Run()); !effect.Accepted || effect.Failed {
		t.Fatal("successful result")
	}

	runtime.Loop.LastToolKey = "read_file|x"
	runtime.Loop.FailKey = runtime.Loop.LastToolKey
	runtime.Loop.FailStreak = MaxToolFailStreak - 1
	runtime.Loop.ToolRounds = MaxToolRounds
	failure := ToolDelivery{TurnID: runtime.Turn.ID, Message: chmctx.Message{Role: chmctx.RoleTool, ToolName: "read_file", Content: "(read error: x)"}}
	if effect := runtime.ApplyToolResult(failure); !effect.FailureNudged || !effect.RunawayNudged {
		t.Fatal("failure policy records")
	}
	runtime.EndTurn()
	if effect := runtime.ApplyToolResult(failure); effect.Accepted {
		t.Fatal("inactive result")
	}

	for _, setup := range []func(Runtime){
		func(r Runtime) { r.Turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant}} },
		func(r Runtime) {
			r.Loop.EmptyNudged = true
			r.Turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant}}
		},
		func(r Runtime) {
			r.Turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "<tool_call>bad"}}
		},
		func(r Runtime) {
			r.Loop.ToolRounds = VerifyNudgeMinRounds
			r.Turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "done"}}
		},
		func(r Runtime) { r.Turn.History = []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "done"}} },
		func(r Runtime) { r.Stream.Pending = []chmctx.ToolCall{{Name: "read_file"}} },
	} {
		r := observedRuntime(observer)
		r.BeginTurn(time.Now())
		setup(r)
		r.DecideClose()
	}
	for _, category := range []string{"tool_start", "tool_result", "tool_outcome", "nudge", "leak"} {
		if !hasCategory(observer.records, category) {
			t.Fatalf("missing %s record", category)
		}
	}
}

func hasCategory(records []observerRecord, category string) bool {
	for _, record := range records {
		if record.category == category || strings.Contains(record.body, category) {
			return true
		}
	}
	return false
}
