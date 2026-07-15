package terminal

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
)

type fakeRuntime struct {
	controller frontend.Controller
	closes     int
}

func (r *fakeRuntime) Controller() frontend.Controller { return r.controller }
func (r *fakeRuntime) Close()                          { r.closes++ }

type fakeController struct{}

func (fakeController) Bootstrap() frontend.Transition               { return frontend.Transition{} }
func (fakeController) Dispatch(frontend.Intent) frontend.Transition { return frontend.Transition{} }
func (fakeController) Snapshot() frontend.Snapshot                  { return frontend.Snapshot{} }

type inertModel struct{}

func (inertModel) Init() tea.Cmd                         { return tea.Quit }
func (m inertModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (inertModel) View() string                          { return "" }

type fakeProgram struct{ err error }

func (p fakeProgram) Run() (tea.Model, error) { return inertModel{}, p.err }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func restoreHooks(t *testing.T) {
	origBootstrap, origFrontend, origProgram := bootstrapApplication, newFrontend, newTeaProgram
	t.Cleanup(func() { bootstrapApplication, newFrontend, newTeaProgram = origBootstrap, origFrontend, origProgram })
}

func TestRunLifecycleAndFailures(t *testing.T) {
	restoreHooks(t)
	runtime := &fakeRuntime{controller: fakeController{}}
	bootstrapApplication = func(version string) (applicationRuntime, error) {
		if version != "test" {
			t.Fatalf("version = %q", version)
		}
		return runtime, nil
	}
	newFrontend = func(controller frontend.Controller, version string) tea.Model {
		if controller != runtime.controller || version != "test" {
			t.Fatal("frontend arguments")
		}
		return inertModel{}
	}
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{} }
	var out bytes.Buffer
	if err := Run(&out, "test"); err != nil || !strings.Contains(out.String(), "\x1b[2J") || runtime.closes != 1 {
		t.Fatalf("out=%q closes=%d err=%v", out.String(), runtime.closes, err)
	}

	boom := errors.New("boom")
	bootstrapApplication = func(string) (applicationRuntime, error) { return nil, boom }
	if err := Run(io.Discard, "test"); !errors.Is(err, boom) {
		t.Fatalf("bootstrap err = %v", err)
	}
	bootstrapApplication = func(string) (applicationRuntime, error) { return runtime, nil }
	if err := Run(failingWriter{boom}, "test"); !errors.Is(err, boom) {
		t.Fatalf("write err = %v", err)
	}
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{err: boom} }
	if err := Run(io.Discard, "test"); !errors.Is(err, boom) {
		t.Fatalf("program err = %v", err)
	}
	if createTeaProgram(inertModel{}) == nil {
		t.Fatal("nil Tea program")
	}
}

func TestPrintHelp(t *testing.T) {
	var out bytes.Buffer
	if err := PrintHelp(&out); err != nil || !strings.Contains(out.String(), "Slash commands:\n  /clear") || !strings.Contains(out.String(), "RECOMPHAMR_URL") {
		t.Fatalf("help=%q err=%v", out.String(), err)
	}
	boom := errors.New("boom")
	if err := PrintHelp(failingWriter{boom}); !errors.Is(err, boom) {
		t.Fatalf("first write = %v", err)
	}
	writes := 0
	writer := writerFunc(func(p []byte) (int, error) {
		writes++
		if writes == 2 {
			return 0, boom
		}
		return len(p), nil
	})
	if err := PrintHelp(writer); !errors.Is(err, boom) {
		t.Fatalf("table write = %v", err)
	}
	writes = 0
	writer = writerFunc(func(p []byte) (int, error) {
		writes++
		if writes == 3 {
			return 0, boom
		}
		return len(p), nil
	})
	if err := PrintHelp(writer); !errors.Is(err, boom) {
		t.Fatalf("final write = %v", err)
	}
}

func TestDefaultFactories(t *testing.T) {
	runtime, err := bootstrapApplication("test")
	if err != nil || runtime.Controller() == nil {
		t.Fatalf("runtime=%#v err=%v", runtime, err)
	}
	runtime.Close()
	if model := newFrontend(fakeController{}, "test"); model == nil {
		t.Fatal("nil frontend")
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
