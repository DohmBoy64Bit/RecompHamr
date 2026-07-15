package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

type scriptedClient struct{ rounds [][]llm.Event }

func (s *scriptedClient) Chat(context.Context, []chmctx.Message, []llm.Tool) <-chan llm.Event {
	ch := make(chan llm.Event, len(s.rounds[0]))
	for _, event := range s.rounds[0] {
		ch <- event
	}
	close(ch)
	s.rounds = s.rounds[1:]
	return ch
}

func scriptedController(t *testing.T, rounds [][]llm.Event, execute func(context.Context, chmctx.ToolCall) chmctx.Message) *Controller {
	t.Helper()
	cfg := config.Default()
	cfg.Dir = t.TempDir()
	sessionRuntime := session.NewRuntime(cfg)
	if execute == nil {
		execute = func(context.Context, chmctx.ToolCall) chmctx.Message { return chmctx.Message{Role: chmctx.RoleTool} }
	}
	runtime := agent.NewRuntime(&scriptedClient{rounds: rounds}, agent.NewToolExecutor(execute))
	controller := NewController(sessionRuntime, runtime, "system", "test")
	controller.now = func() time.Time { return time.Unix(20, 0) }
	return controller
}

func driveTransition(controller *Controller, transition frontend.Transition, limit int) ([]frontend.Event, frontend.Transition) {
	events := append([]frontend.Event(nil), transition.Events...)
	for transition.Work != nil && limit > 0 {
		transition = controller.Dispatch(frontend.Complete(transition.Work.Run()))
		events = append(events, transition.Events...)
		limit--
	}
	return events, transition
}

func controllerFixture(t *testing.T) (*Controller, *config.Config, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Context-Window", "8192")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	}))
	t.Cleanup(server.Close)
	root := t.TempDir()
	cfg, _, err := config.Bootstrap(root)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ActiveProfile().URL = server.URL
	cfg.Models["other"] = &config.Profile{LLM: "other", URL: server.URL, Key: "secret", ContextSize: 4096}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	sessionRuntime := session.NewRuntime(cfg)
	agentRuntime := agent.NewRuntime(sessionRuntime, agent.LocalToolExecutor())
	return NewController(sessionRuntime, agentRuntime, "system", "test"), cfg, server
}

func TestControllerSnapshotBootstrapAndCompletions(t *testing.T) {
	controller, _, _ := controllerFixture(t)
	if got := controller.activeContextSize(frontend.Snapshot{Active: "missing", ContextSize: 10}); got != 10 {
		t.Fatalf("configured context = %d", got)
	}
	if got := controller.activeContextSize(frontend.Snapshot{Active: "missing"}); got != 16177 {
		t.Fatalf("fallback context = %d", got)
	}
	controller.agent.BeginTurn(time.Now())
	controller.agent.SetConnected(true)
	controller.agent.SetLiveContextSize("local", 1234)
	snapshot := controller.Snapshot()
	if snapshot.Phase != frontend.PhaseThinking || !snapshot.Connected || snapshot.Active != "local" || len(snapshot.Profiles) != 2 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	snapshot.Profiles[0].Name = "mutated"
	if controller.Snapshot().Profiles[0].Name == "mutated" {
		t.Fatal("snapshot profiles share backing storage")
	}
	controller.agent.EndTurn()
	if err := controller.session.AppendHistory("remember"); err != nil {
		t.Fatal(err)
	}
	transition := controller.Bootstrap()
	if len(transition.Events) != 1 || transition.Events[0].Kind != frontend.EventHistory || transition.Events[0].Values[0] != "remember" || transition.Work == nil {
		t.Fatalf("bootstrap = %#v", transition)
	}
	completion := transition.Work.Run()
	transition = controller.Dispatch(frontend.Complete(completion))
	if len(transition.Events) != 1 || transition.Events[0].Kind != frontend.EventReachability || !transition.Events[0].OK || !transition.Snapshot.Connected {
		t.Fatalf("reachability = %#v", transition)
	}
	if duplicate := controller.Dispatch(frontend.Complete(completion)); len(duplicate.Events) != 0 {
		t.Fatalf("duplicate = %#v", duplicate)
	}
	if foreign := controller.Dispatch(frontend.Complete(struct{}{})); len(foreign.Events) != 0 {
		t.Fatalf("foreign = %#v", foreign)
	}
	unknown := controller.capture(func() any { return "unknown" }).Run()
	if got := controller.Dispatch(frontend.Complete(unknown)); len(got.Events) != 0 {
		t.Fatalf("unknown completion = %#v", got)
	}
}

func TestControllerSessionIntentsAndProbe(t *testing.T) {
	controller, cfg, _ := controllerFixture(t)
	if got := controller.Dispatch(frontend.ObserveSlash("/models")); got.Snapshot.Active != "local" {
		t.Fatalf("observe slash = %#v", got)
	}
	controller.Dispatch(frontend.AppendHistory("persisted"))
	if got := controller.session.LoadHistory(); len(got) != 1 || got[0] != "persisted" {
		t.Fatalf("history = %#v", got)
	}
	if got := controller.Dispatch(frontend.Reload()); len(got.Events) != 0 {
		t.Fatalf("reload = %#v", got)
	}
	if got := controller.Dispatch(frontend.Activate("missing")); len(got.Events) != 1 || !strings.Contains(got.Events[0].Text, "unknown model") {
		t.Fatalf("missing activation = %#v", got)
	}
	transition := controller.Dispatch(frontend.Activate("other"))
	if transition.Work == nil || len(transition.Events) != 1 || transition.Events[0].Kind != frontend.EventProfileActivated || transition.Snapshot.Active != "other" {
		t.Fatalf("activation = %#v", transition)
	}
	probeCompletion := transition.Work.Run()
	transition = controller.Dispatch(frontend.Complete(probeCompletion))
	if len(transition.Events) != 1 || transition.Events[0].Kind != frontend.EventProbe || !transition.Events[0].OK || transition.Events[0].ContextWindow != 8192 || !transition.Snapshot.Connected {
		t.Fatalf("probe = %#v", transition)
	}
	silent := controller.connectivityWork(controller.Snapshot(), true).Run()
	if got := controller.Dispatch(frontend.Complete(silent)); len(got.Events) != 0 {
		t.Fatalf("silent probe = %#v", got)
	}
	controller.Dispatch(frontend.ClearConversation())
	if len(controller.session.LoadHistory()) != 0 {
		t.Fatal("clear retained history")
	}
	for _, intent := range []frontend.Intent{frontend.Cancel(time.Now()), {}} {
		if got := controller.Dispatch(intent); len(got.Events) != 0 || got.Work != nil {
			t.Fatalf("inactive no-op = %#v", got)
		}
	}
	start := time.Now().Add(-time.Second)
	transition = controller.Dispatch(frontend.SubmitGoal("later", start))
	if len(transition.Events) != 1 || transition.Events[0].Kind != frontend.EventTurnStarted || transition.Work == nil {
		t.Fatalf("submit = %#v", transition)
	}
	var events []frontend.Event
	for transition.Work != nil {
		transition = controller.Dispatch(frontend.Complete(transition.Work.Run()))
		events = append(events, transition.Events...)
	}
	if len(events) < 3 || events[0].Kind != frontend.EventContent || events[len(events)-1].Kind != frontend.EventTurnFinished || !events[len(events)-1].OK {
		t.Fatalf("turn events = %#v", events)
	}
	transition = controller.Dispatch(frontend.SubmitGoal("cancel", start))
	staleWork := transition.Work
	transition = controller.Dispatch(frontend.Cancel(start.Add(time.Second)))
	if len(transition.Events) != 1 || !transition.Events[0].Cancelled || transition.Events[0].Natural || transition.Snapshot.Phase.Active() {
		t.Fatalf("cancel = %#v", transition)
	}
	if got := controller.Dispatch(frontend.Complete(staleWork.Run())); len(got.Events) != 0 {
		t.Fatalf("stale stream drain = %#v", got)
	}
	if humanTokens(999) != "999 tok" || humanTokens(1500) != "1.5k tok" || humanRate(0, time.Second) != "" || humanRate(5, time.Second) != "5.0 tok/s" || humanRate(15, time.Second) != "15 tok/s" {
		t.Fatal("turn formatting")
	}

	path := filepath.Join(cfg.Dir, "config.yaml")
	if err := os.WriteFile(path, []byte("not: [valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := controller.Dispatch(frontend.Reload()); len(got.Events) != 1 || !strings.Contains(got.Events[0].Text, "config.yaml") {
		t.Fatalf("reload error = %#v", got)
	}
}

func TestControllerStaleAndFailedWork(t *testing.T) {
	controller, cfg, _ := controllerFixture(t)
	stale := controller.connectivityWork(controller.Snapshot(), false)
	cfg.Models["other"].URL = "http://different"
	if _, err := controller.session.Activate("other"); err != nil {
		t.Fatal(err)
	}
	if got := controller.Dispatch(frontend.Complete(stale.Run())); len(got.Events) != 0 {
		t.Fatalf("stale reachability = %#v", got)
	}
	failed := controller.capture(func() any {
		return probeCompletion{result: session.ProbeResult{Profile: "other", Err: fmt.Errorf("probe failed")}}
	}).Run()
	got := controller.Dispatch(frontend.Complete(failed))
	if len(got.Events) != 1 || got.Events[0].Kind != frontend.EventProbe || got.Events[0].OK || got.Events[0].Text != "probe failed" || got.Snapshot.Connected {
		t.Fatalf("failed probe = %#v", got)
	}
}

func TestControllerStreamTransitionsAndErrors(t *testing.T) {
	final := chmctx.Message{Role: chmctx.RoleAssistant, Content: "hello"}
	controller := scriptedController(t, [][]llm.Event{{
		{Kind: llm.EventRetry, Content: "retrying"},
		{Kind: llm.EventContent, Content: "hello"},
		{Kind: llm.EventDone, Final: &final, Tokens: 12},
	}}, nil)
	start := time.Unix(10, 0)
	events, transition := driveTransition(controller, controller.Dispatch(frontend.SubmitGoal("goal", start)), 10)
	if transition.Work != nil || transition.Snapshot.Phase.Active() {
		t.Fatalf("unfinished transition = %#v", transition)
	}
	kinds := make([]frontend.EventKind, 0, len(events))
	for _, event := range events {
		kinds = append(kinds, event.Kind)
	}
	want := []frontend.EventKind{frontend.EventTurnStarted, frontend.EventStatus, frontend.EventStatus, frontend.EventContent, frontend.EventFlush, frontend.EventTurnFinished}
	if fmt.Sprint(kinds) != fmt.Sprint(want) || !events[len(events)-1].OK || events[len(events)-1].Tokens != 12 {
		t.Fatalf("stream events = %#v", events)
	}
	activeController := scriptedController(t, [][]llm.Event{{{Kind: llm.EventContent, Content: "pending"}}}, nil)
	if got := activeController.Dispatch(frontend.SubmitGoal("one", start)); got.Work == nil {
		t.Fatal("fresh submit rejected")
	} else if duplicate := activeController.Dispatch(frontend.SubmitGoal("two", start)); duplicate.Work != nil || len(duplicate.Events) != 0 {
		t.Fatalf("active duplicate = %#v", duplicate)
	}

	errController := scriptedController(t, [][]llm.Event{{{Kind: llm.EventError, Err: fmt.Errorf("broken")}}}, nil)
	events, transition = driveTransition(errController, errController.Dispatch(frontend.SubmitGoal("error", start)), 10)
	if transition.Work != nil || events[len(events)-1].Kind != frontend.EventTurnFinished || events[len(events)-1].Natural || !strings.Contains(events[len(events)-1].Text, "broken") {
		t.Fatalf("error events = %#v transition=%#v", events, transition)
	}
}

func TestControllerSequentialToolsAndStaleTool(t *testing.T) {
	calls := []chmctx.ToolCall{
		{ID: "1", Name: "read_file", Arguments: map[string]any{"path": "one"}},
		{ID: "2", Name: "write_file", Arguments: map[string]any{"path": "two", "content": "ok"}},
	}
	finalTools := chmctx.Message{Role: chmctx.RoleAssistant, ToolCalls: calls}
	finalDone := chmctx.Message{Role: chmctx.RoleAssistant, Content: "done"}
	var executed []string
	controller := scriptedController(t, [][]llm.Event{
		{{Kind: llm.EventToolCall, ToolCall: &calls[0]}, {Kind: llm.EventToolCall, ToolCall: &calls[1]}, {Kind: llm.EventDone, Final: &finalTools}},
		{{Kind: llm.EventContent, Content: "done"}, {Kind: llm.EventDone, Final: &finalDone}},
	}, func(_ context.Context, call chmctx.ToolCall) chmctx.Message {
		executed = append(executed, call.ID)
		return chmctx.Message{Role: chmctx.RoleTool, ToolCallID: call.ID, ToolName: call.Name, Content: "ok"}
	})
	events, transition := driveTransition(controller, controller.Dispatch(frontend.SubmitGoal("tools", time.Unix(10, 0))), 20)
	if transition.Work != nil || fmt.Sprint(executed) != "[1 2]" || events[len(events)-1].Kind != frontend.EventTurnFinished || !events[len(events)-1].OK {
		t.Fatalf("tool loop events=%#v executed=%v transition=%#v", events, executed, transition)
	}
	toolStatuses := 0
	for _, event := range events {
		if event.Kind == frontend.EventToolStatus {
			toolStatuses++
		}
	}
	if toolStatuses != 2 {
		t.Fatalf("tool statuses = %d", toolStatuses)
	}

	stale := scriptedController(t, [][]llm.Event{{{Kind: llm.EventToolCall, ToolCall: &calls[0]}, {Kind: llm.EventDone, Final: &finalTools}}}, nil)
	transition = stale.Dispatch(frontend.SubmitGoal("cancel tool", time.Unix(10, 0)))
	for transition.Work != nil && transition.Snapshot.Phase != frontend.PhaseRunning {
		transition = stale.Dispatch(frontend.Complete(transition.Work.Run()))
	}
	toolWork := transition.Work
	stale.Dispatch(frontend.Cancel(time.Unix(11, 0)))
	if got := stale.Dispatch(frontend.Complete(toolWork.Run())); got.Work != nil || len(got.Events) != 0 {
		t.Fatalf("stale tool = %#v", got)
	}
	empty := frontend.Transition{}
	stale.nextTool(&empty)
	if empty.Work != nil || len(empty.Events) != 0 {
		t.Fatalf("empty next tool = %#v", empty)
	}
}

func TestControllerClosePolicies(t *testing.T) {
	empty := chmctx.Message{Role: chmctx.RoleAssistant}
	controller := scriptedController(t, [][]llm.Event{
		{{Kind: llm.EventDone, Final: &empty}},
		{{Kind: llm.EventDone, Final: &empty}},
	}, nil)
	events, transition := driveTransition(controller, controller.Dispatch(frontend.SubmitGoal("stall", time.Unix(10, 0))), 20)
	if transition.Work != nil || events[len(events)-1].Natural != true || !strings.Contains(events[len(events)-1].Text, "stalled") {
		t.Fatalf("empty policy = %#v", events)
	}

	leak := chmctx.Message{Role: chmctx.RoleAssistant, Content: "<tool_call>bad"}
	controller = scriptedController(t, [][]llm.Event{{{Kind: llm.EventDone, Final: &leak}}}, nil)
	events, _ = driveTransition(controller, controller.Dispatch(frontend.SubmitGoal("leak", time.Unix(10, 0))), 10)
	if !strings.Contains(events[len(events)-1].Text, "leaked") {
		t.Fatalf("leak policy = %#v", events)
	}

	inactive := frontend.Transition{}
	controller.applyClose(&inactive)
	if len(inactive.Events) != 0 || inactive.Work != nil {
		t.Fatalf("inactive close = %#v", inactive)
	}
	controller.applyToolCompletion(toolCompletion{}, &inactive)
	if len(inactive.Events) != 0 || inactive.Work != nil {
		t.Fatalf("malformed tool completion = %#v", inactive)
	}

	controller.agent.BeginTurn(time.Unix(30, 0))
	oldEvents := make(chan llm.Event, 1)
	oldEvents <- llm.Event{Kind: llm.EventContent, Content: "stale"}
	oldStream := controller.agent.Stream.BeginStream(controller.agent.Turn.ID, oldEvents)
	delivery := oldStream.Read()
	controller.agent.Stream.BeginStream(controller.agent.Turn.ID, make(chan llm.Event))
	staleTransition := frontend.Transition{}
	controller.applyModelCompletion(modelCompletion{stream: oldStream, delivery: delivery}, &staleTransition)
	if staleTransition.Work == nil || len(staleTransition.Events) != 0 {
		t.Fatalf("stale nonclosed stream = %#v", staleTransition)
	}
	controller.agent.EndTurn()
}
