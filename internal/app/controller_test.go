package app

import (
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
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

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
	for _, intent := range []frontend.Intent{frontend.Cancel(time.Now()), frontend.SubmitGoal("later", time.Now()), {}} {
		if got := controller.Dispatch(intent); len(got.Events) != 0 || got.Work != nil {
			t.Fatalf("transitional no-op = %#v", got)
		}
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
