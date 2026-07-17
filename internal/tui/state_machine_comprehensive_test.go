package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
)

type fakeFrontendController struct {
	snapshot    frontend.Snapshot
	bootstrap   frontend.Transition
	transitions []frontend.Transition
	intents     []frontend.Intent
}

func (f *fakeFrontendController) Bootstrap() frontend.Transition { return f.bootstrap }
func (f *fakeFrontendController) Snapshot() frontend.Snapshot    { return f.snapshot }
func (f *fakeFrontendController) Dispatch(intent frontend.Intent) frontend.Transition {
	f.intents = append(f.intents, intent)
	if len(f.transitions) == 0 {
		return frontend.Transition{Snapshot: f.snapshot}
	}
	transition := f.transitions[0]
	f.transitions = f.transitions[1:]
	f.snapshot = transition.Snapshot
	return transition
}

type fakeFrontendWork struct{ completion frontend.Completion }

func (w fakeFrontendWork) Run() frontend.Completion { return w.completion }

func presentationModel(controller *fakeFrontendController) Model {
	model := New(controller, "test")
	model.width, model.height = 100, 30
	return model
}

func TestFrontendEventsReproducePresentationState(t *testing.T) {
	now := time.Now().Add(-time.Second)
	controller := &fakeFrontendController{snapshot: frontend.Snapshot{Active: "local", ActiveModel: "model", Connected: true}}
	m := presentationModel(controller)
	m.queued = &queuedPrompt{send: "queued", echo: "queued"}
	transition := frontend.Transition{Snapshot: controller.snapshot, Events: []frontend.Event{
		{Kind: frontend.EventTurnStarted, At: now},
		{Kind: frontend.EventStatus, Text: "retry"},
		{Kind: frontend.EventContent, Text: "a\tb"},
		{Kind: frontend.EventFlush},
		{Kind: frontend.EventToolStatus, Text: "read_file x"},
		{Kind: frontend.EventTurnFinished, Text: "✗ cancelled", Elapsed: time.Second, Tokens: 8, Cancelled: true},
	}}
	next, cmd := m.applyFrontendTransition(transition)
	m = next.(Model)
	if cmd != nil || m.turnStart != (time.Time{}) || m.lastOutcome != outcomeStopped || m.lastTokens != 8 || m.queued != nil || m.ta.Value() != "queued" {
		t.Fatalf("finished state = %#v", m)
	}
	scroll := stripANSI(m.scroll.String())
	if !strings.Contains(scroll, "a    b") || !strings.Contains(scroll, "read_file x") || !strings.Contains(scroll, "cancelled") || m.status != "" {
		t.Fatalf("events scroll=%q status=%q", scroll, m.status)
	}
}

func TestSkillCommandsRemainPresentationOnly(t *testing.T) {
	snapshot := frontend.Snapshot{Skills: []frontend.Skill{{Name: "alpha", Description: "Use alpha.", Active: true}}, SkillDiagnostics: []string{"bad: invalid"}}
	controller := &fakeFrontendController{snapshot: snapshot, transitions: []frontend.Transition{
		{Snapshot: snapshot},
		{Snapshot: snapshot},
		{Snapshot: snapshot},
		{Snapshot: snapshot, Events: []frontend.Event{{Kind: frontend.EventSkillActivated, Text: "loaded skill: alpha"}}},
		{Snapshot: snapshot},
		{Snapshot: snapshot, Events: []frontend.Event{{Kind: frontend.EventWarning, Text: "unknown skill: missing"}}},
		{Snapshot: snapshot},
	}}
	m := presentationModel(controller)
	args := skillArgs(m)
	if len(args) != 1 || args[0].value != "alpha" || !args[0].current || args[0].description != "Use alpha." {
		t.Fatalf("skill args = %#v", args)
	}
	next, _ := m.runSlash("/skills")
	m = next.(Model)
	if scroll := stripANSI(m.scroll.String()); !strings.Contains(scroll, "Agent Skills (* active):") || !strings.Contains(scroll, "* alpha") || !strings.Contains(scroll, "bad: invalid") {
		t.Fatalf("skills output = %q", scroll)
	}
	next, _ = m.runSlash("/skill")
	m = next.(Model)
	if !strings.Contains(stripANSI(m.scroll.String()), "usage: /skill <name>") {
		t.Fatal("missing skill usage absent")
	}
	next, _ = m.runSlash("/skill alpha")
	m = next.(Model)
	next, _ = m.runSlash("/skill missing")
	m = next.(Model)
	next, _ = m.runSlash("/help")
	m = next.(Model)
	activateCount := 0
	for _, intent := range controller.intents {
		if intent.Kind() == frontend.IntentActivateSkill {
			activateCount++
		}
	}
	if activateCount != 2 || !strings.Contains(stripANSI(m.scroll.String()), "loaded skill: alpha") || !strings.Contains(stripANSI(m.scroll.String()), "unknown skill: missing") || !strings.Contains(stripANSI(m.scroll.String()), "/skills") {
		t.Fatalf("skill command intents=%#v scroll=%q", controller.intents, stripANSI(m.scroll.String()))
	}

	emptyController := &fakeFrontendController{}
	empty := presentationModel(emptyController)
	next, _ = empty.runSlash("/skills")
	if !strings.Contains(stripANSI(next.(Model).scroll.String()), "No Agent Skills discovered.") {
		t.Fatal("empty skill catalog message absent")
	}
}

func TestEvidenceCommandsRemainPresentationOnly(t *testing.T) {
	controller := &fakeFrontendController{transitions: []frontend.Transition{
		{Events: []frontend.Event{{Kind: frontend.EventWorkspace, Text: ".rehamr/ evidence workspace initialized"}}},
		{Events: []frontend.Event{{Kind: frontend.EventWorkspace, Text: "status text"}}},
		{Events: []frontend.Event{{Kind: frontend.EventWarning, Text: "init-re: denied"}}},
	}}
	m := presentationModel(controller)
	next, _ := m.cmdInitEvidence(nil)
	m = next.(Model)
	next, _ = m.cmdEvidenceStatus(nil)
	m = next.(Model)
	next, _ = m.cmdInitEvidence(nil)
	m = next.(Model)
	output := stripANSI(m.scroll.String())
	if !strings.Contains(output, "evidence workspace initialized") || !strings.Contains(output, "status text") || !strings.Contains(output, "init-re: denied") {
		t.Fatalf("evidence output = %q", output)
	}
	if len(controller.intents) != 3 || controller.intents[0].Kind() != frontend.IntentInitializeEvidence || controller.intents[1].Kind() != frontend.IntentEvidenceStatus {
		t.Fatalf("evidence intents = %#v", controller.intents)
	}
}

func TestFrontendCompletionAndNaturalQueueFireExactlyOnce(t *testing.T) {
	controller := &fakeFrontendController{snapshot: frontend.Snapshot{Active: "local", Phase: frontend.PhaseThinking}}
	controller.bootstrap = frontend.Transition{Snapshot: controller.snapshot, Events: []frontend.Event{{Kind: frontend.EventHistory, Values: []string{"remember"}}}, Work: fakeFrontendWork{completion: "startup"}}
	controller.transitions = []frontend.Transition{
		{Snapshot: frontend.Snapshot{Active: "local"}, Events: []frontend.Event{{Kind: frontend.EventTurnFinished, OK: true, Natural: true}}},
		{Snapshot: frontend.Snapshot{Active: "local"}},
		{Snapshot: frontend.Snapshot{Active: "local", Phase: frontend.PhaseThinking}, Events: []frontend.Event{{Kind: frontend.EventTurnStarted, At: time.Now()}}, Work: fakeFrontendWork{completion: "queued-work"}},
	}
	m := presentationModel(controller)
	if len(m.promptHistory) != 1 || m.promptHistory[0].display != "remember" {
		t.Fatalf("history = %#v", m.promptHistory)
	}
	m.queued = &queuedPrompt{send: "next", echo: "next"}
	msg := runFrontendWork(controller.bootstrap.Work)()
	next, cmd := m.update(msg)
	m = next.(Model)
	if cmd == nil || len(controller.intents) != 3 || controller.intents[0].Kind() != frontend.IntentComplete || controller.intents[1].Kind() != frontend.IntentAppendHistory || controller.intents[2].Kind() != frontend.IntentSubmitGoal {
		t.Fatalf("intents = %#v cmd=%v", controller.intents, cmd)
	}
	if m.queued != nil || m.lastOutcome != outcomeNone || m.controller.Snapshot().Phase != frontend.PhaseThinking {
		t.Fatalf("natural finish = %#v", m)
	}
	if completion := cmd().(frontendCompletionMsg); completion.completion != "queued-work" {
		t.Fatalf("queued completion = %#v", completion)
	}

	controller.transitions = []frontend.Transition{{Snapshot: controller.snapshot, Events: []frontend.Event{{Kind: frontend.EventWarning, Text: "warn"}, {Kind: frontend.EventProfileActivated, Profile: "p", Model: "m", URL: "u"}}}}
	next, _ = m.update(frontendCompletionMsg{completion: "other"})
	m = next.(Model)
	if !strings.Contains(stripANSI(m.scroll.String()), "warn") || !strings.Contains(stripANSI(m.scroll.String()), "active: p") {
		t.Fatal("completion events missing")
	}

	if _, cmd = m.update(tea.FocusMsg{}); cmd != nil {
		t.Fatal("focus produced command")
	}
}

func TestFrontendWorkNilAndPhaseLabels(t *testing.T) {
	if runFrontendWork(nil) != nil {
		t.Fatal("nil work")
	}
	for phase, label := range map[frontend.Phase]string{
		frontend.PhaseIdle: "", frontend.PhaseThinking: "thinking", frontend.PhaseStreaming: "generating", frontend.PhaseRunning: "running", frontend.Phase(99): "",
	} {
		if phase.Label() != label {
			t.Fatalf("phase %d = %q", phase, phase.Label())
		}
	}
}

func TestFrontendUpdateAndProbeBranches(t *testing.T) {
	profiles := []frontend.Profile{{Name: "active", Model: "model", URL: "http://active", Active: true}}
	controller := &fakeFrontendController{snapshot: frontend.Snapshot{Active: "active", Profiles: profiles}}
	controller.bootstrap = frontend.Transition{Snapshot: controller.snapshot, Work: fakeFrontendWork{completion: "startup"}}
	m := presentationModel(controller)
	if batch, ok := m.Init()().(tea.BatchMsg); !ok || len(batch) != 3 {
		t.Fatalf("init = %#v", batch)
	}
	for _, msg := range []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes},
		tea.WindowSizeMsg{Width: 100, Height: 30},
		resizeSettleMsg{gen: 99},
		spinner.TickMsg{},
		struct{}{},
	} {
		next, _ := m.update(msg)
		m = next.(Model)
	}
	m.quitArmedAt = time.Now().Add(-time.Second)
	m.status = quitArmText
	next, _ := m.update(quitArmResetMsg{})
	m = next.(Model)
	if m.status != "" || quitArmReset(time.Now()) != (quitArmResetMsg{}) || resizeSettled(4)(time.Now()) != (resizeSettleMsg{gen: 4}) {
		t.Fatal("timer branches")
	}

	transition := frontend.Transition{Snapshot: controller.snapshot, Events: []frontend.Event{
		{Kind: frontend.EventProbe, Profile: "active", Text: "failed"},
		{Kind: frontend.EventProbe, Profile: "missing", OK: true},
		{Kind: frontend.EventProbe, Profile: "active", OK: true, ContextWindow: 8192},
	}}
	next, _ = m.applyFrontendTransition(transition)
	m = next.(Model)
	scroll := stripANSI(m.scroll.String())
	if !strings.Contains(scroll, "probe active: failed") || !strings.Contains(scroll, "ctx: 8,192") {
		t.Fatalf("probe scroll = %q", scroll)
	}
	m.ta.SetValue("draft")
	m.queued = &queuedPrompt{send: "queued", echo: "queued"}
	m.applyTurnFinished(frontend.Event{Kind: frontend.EventTurnFinished})
	if m.ta.Value() != "draft" || m.queued != nil {
		t.Fatal("finish clobbered draft")
	}
	m.applyTurnFinished(frontend.Event{Kind: frontend.EventTurnFinished, Text: "error"})
	if !strings.Contains(stripANSI(m.scroll.String()), "error") {
		t.Fatal("error finish missing")
	}
}
