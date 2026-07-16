package tui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	appcontroller "github.com/DohmBoy64Bit/RecompHamr/internal/app/controller"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/logging"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

type snapshotController struct {
	frontend.Controller
	snapshot frontend.Snapshot
}

func (c *snapshotController) Snapshot() frontend.Snapshot { return c.snapshot }
func (c *snapshotController) Dispatch(intent frontend.Intent) frontend.Transition {
	switch intent.Kind() {
	case frontend.IntentCancel:
		c.snapshot.Phase = frontend.PhaseIdle
		return frontend.Transition{Snapshot: c.snapshot, Events: []frontend.Event{{Kind: frontend.EventTurnFinished, Text: "✗ cancelled", Cancelled: true}}}
	case frontend.IntentSubmitGoal:
		c.snapshot.Phase = frontend.PhaseThinking
		return frontend.Transition{Snapshot: c.snapshot, Events: []frontend.Event{{Kind: frontend.EventTurnStarted, At: intent.Time()}}}
	default:
		transition := c.Controller.Dispatch(intent)
		c.snapshot = transition.Snapshot
		return transition
	}
}

func mutateTestSnapshot(m *Model, mutate func(*frontend.Snapshot)) {
	snapshot := m.controller.Snapshot()
	mutate(&snapshot)
	m.controller = &snapshotController{Controller: m.controller, snapshot: snapshot}
}

func setTestPhase(m *Model, phase frontend.Phase) {
	mutateTestSnapshot(m, func(snapshot *frontend.Snapshot) { snapshot.Phase = phase })
}

func newModelWithRuntime(sessionRuntime *session.Runtime, runtime agent.Runtime, system, version string) Model {
	return New(appcontroller.NewController(sessionRuntime, runtime, func() string { return system }, version), version)
}

// newTestModel wires a model against a mock OpenAI-compatible SSE server so
// focused TUI tests can exercise submit -> stream -> done without a real backend.
func newTestModel(t *testing.T, handler http.HandlerFunc) Model {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg, _, err := config.Bootstrap(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg.ActiveProfile().URL = srv.URL
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	sessionRuntime := session.NewRuntime(cfg)
	m := newModelWithRuntime(sessionRuntime, agent.NewRuntime(sessionRuntime, agent.LocalToolExecutor()).WithObserver(logging.NewObserver()), "test system", "test")
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return sized.(Model)
}

// drain executes a Bubble Tea command chain until it produces no more commands.
func drain(m tea.Model, cmd tea.Cmd) (tea.Model, []tea.Msg) {
	var seen []tea.Msg
	queue := []tea.Cmd{cmd}
	for len(queue) > 0 {
		cmd, queue = queue[0], queue[1:]
		if cmd == nil {
			continue
		}
		msg := cmd()
		if msg == nil {
			continue
		}
		seen = append(seen, msg)
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c == nil {
					continue
				}
				bm, bcmd := m.Update(c())
				m = bm
				queue = append(queue, bcmd)
			}
			continue
		}
		var nextCmd tea.Cmd
		m, nextCmd = m.Update(msg)
		queue = append(queue, nextCmd)
	}
	return m, seen
}

// stripANSI removes CSI escape sequences from rendered TUI text for assertions.
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			for j := i + 2; j < len(s); j++ {
				if s[j] >= 0x40 && s[j] <= 0x7e {
					i = j
					break
				}
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func testModel(t *testing.T, cfg *config.Config, client *llm.Client) Model {
	t.Helper()
	cfg.ActiveProfile().URL = client.BaseURL
	cfg.ActiveProfile().LLM = client.Model
	sessionRuntime := session.NewRuntime(cfg)
	return newModelWithRuntime(sessionRuntime, agent.NewRuntime(sessionRuntime, agent.LocalToolExecutor()).WithObserver(logging.NewObserver()), "test system", "test")
}

func TestNewLoadsInjectedPromptHistory(t *testing.T) {
	cfg, _, err := config.Bootstrap(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	history := session.NewHistory(cfg.Dir)
	if err := history.Append("persisted\nUnicode 🐹"); err != nil {
		t.Fatal(err)
	}
	sessionRuntime := session.NewRuntime(cfg)
	m := newModelWithRuntime(sessionRuntime, agent.NewRuntime(sessionRuntime, agent.LocalToolExecutor()), "test system", "test")
	if len(m.promptHistory) != 1 || m.promptHistory[0].display != "persisted\nUnicode 🐹" {
		t.Fatalf("loaded history = %#v", m.promptHistory)
	}
}
