package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func restoreAppHooks(t *testing.T) {
	origCwd, origBootstrap, origAbs, origEnv := getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment
	origClient, origFrontend := newClient, newFrontend
	origOpen, origClose, origHelp := openDebugLog, closeDebugLog, printFrontendHelp
	origRun, origNew := runTeaProgram, newTeaProgram
	t.Cleanup(func() {
		getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment = origCwd, origBootstrap, origAbs, origEnv
		newClient, newFrontend = origClient, origFrontend
		openDebugLog, closeDebugLog, printFrontendHelp = origOpen, origClose, origHelp
		runTeaProgram, newTeaProgram = origRun, origNew
	})
}

type inertModel struct{}

func (inertModel) Init() tea.Cmd                         { return tea.Quit }
func (m inertModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (inertModel) View() string                          { return "" }

type fakeProgram struct{ err error }

func (p fakeProgram) Run() (tea.Model, error) { return inertModel{}, p.err }

func TestRunCompositionAndLogging(t *testing.T) {
	restoreAppHooks(t)
	root := t.TempDir()
	getWorkingDirectory = func() (string, error) { return root, nil }
	cfg := config.Default()
	cfg.Dir = filepath.Join(root, config.DirName)
	cfg.Logging = true
	bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
	getEnvironment = func(name string) string {
		if name == "RECOMPHAMR_URL" {
			return "http://override"
		}
		return ""
	}
	absolutePath = func(string) (string, error) { return "", errors.New("abs") }
	clientCreated := false
	newClient = func(baseURL, model, key string) *llm.Client {
		clientCreated = true
		if baseURL != "http://override" || model == "" || key != "" {
			t.Fatalf("client args = %q %q %q", baseURL, model, key)
		}
		return llm.New(baseURL, model, key)
	}
	frontendCreated := false
	newFrontend = func(gotCfg *config.Config, client *llm.Client, projectDir, version string) tea.Model {
		frontendCreated = true
		if gotCfg != cfg || client == nil || projectDir != root || version != "test" {
			t.Fatalf("frontend args = %p %v %q %q", gotCfg, client, projectDir, version)
		}
		return inertModel{}
	}
	opened, closed := "", 0
	openDebugLog = func(dir string) { opened = dir }
	closeDebugLog = func() { closed++ }
	runTeaProgram = func(model tea.Model) error {
		if model == nil {
			t.Fatal("nil frontend")
		}
		return nil
	}
	var out bytes.Buffer
	if err := Run(&out, "test"); err != nil {
		t.Fatal(err)
	}
	if !clientCreated || !frontendCreated || opened != cfg.Dir || closed != 1 {
		t.Fatalf("composition = client:%v frontend:%v open:%q close:%d", clientCreated, frontendCreated, opened, closed)
	}
	if !strings.Contains(out.String(), "\x1b[2J") {
		t.Fatalf("clear sequence = %q", out.String())
	}
}

func TestRunFailuresAndNoLogging(t *testing.T) {
	boom := errors.New("boom")
	t.Run("cwd", func(t *testing.T) {
		restoreAppHooks(t)
		getWorkingDirectory = func() (string, error) { return "", boom }
		if err := Run(io.Discard, "test"); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("bootstrap", func(t *testing.T) {
		restoreAppHooks(t)
		getWorkingDirectory = func() (string, error) { return t.TempDir(), nil }
		bootstrapConfig = func(string) (*config.Config, bool, error) { return nil, false, boom }
		if err := Run(io.Discard, "test"); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	for _, tc := range []struct {
		name      string
		configure func()
	}{
		{"output", func() { runTeaProgram = func(tea.Model) error { return nil } }},
		{"program", func() { runTeaProgram = func(tea.Model) error { return boom } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreAppHooks(t)
			root := t.TempDir()
			getWorkingDirectory = func() (string, error) { return root, nil }
			cfg := config.Default()
			cfg.Dir = root
			bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
			newFrontend = func(*config.Config, *llm.Client, string, string) tea.Model { return inertModel{} }
			tc.configure()
			writer := io.Writer(io.Discard)
			if tc.name == "output" {
				writer = failingWriter{boom}
			}
			if err := Run(writer, "test"); !errors.Is(err, boom) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	restoreAppHooks(t)
	printFrontendHelp = func(w io.Writer) { _, _ = io.WriteString(w, "commands\n") }
	var out bytes.Buffer
	if err := PrintHelp(&out); err != nil || !strings.Contains(out.String(), "Slash commands:\ncommands") || !strings.Contains(out.String(), "RECOMPHAMR_URL") {
		t.Fatalf("help = %q %v", out.String(), err)
	}
	boom := errors.New("boom")
	if err := PrintHelp(failingWriter{boom}); !errors.Is(err, boom) {
		t.Fatalf("first write = %v", err)
	}
	printFrontendHelp = func(io.Writer) {}
	writes := 0
	writer := writerFunc(func(p []byte) (int, error) {
		writes++
		if writes == 2 {
			return 0, boom
		}
		return len(p), nil
	})
	if err := PrintHelp(writer); !errors.Is(err, boom) {
		t.Fatalf("second write = %v", err)
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

func TestTeaProgramBoundary(t *testing.T) {
	restoreAppHooks(t)
	if model := newFrontend(config.Default(), llm.New("http://localhost", "model", ""), t.TempDir(), "test"); model == nil {
		t.Fatal("default frontend factory returned nil")
	}
	if createTeaProgram(inertModel{}) == nil {
		t.Fatal("factory returned nil")
	}
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{} }
	if err := executeTeaProgram(inertModel{}); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("tea")
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{err: boom} }
	if err := executeTeaProgram(inertModel{}); !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
}

func TestMain(m *testing.M) { os.Exit(m.Run()) }
