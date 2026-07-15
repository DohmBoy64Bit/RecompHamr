package main

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
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func restoreMainHooks(t *testing.T) {
	origCwd, origBootstrap, origAbs := getWorkingDirectory, bootstrapConfig, absolutePath
	origEnv, origRun, origNew, origExit, origArgs := getEnvironment, runTeaProgram, newTeaProgram, exitProcess, os.Args
	t.Cleanup(func() {
		getWorkingDirectory, bootstrapConfig, absolutePath = origCwd, origBootstrap, origAbs
		getEnvironment, runTeaProgram, newTeaProgram, exitProcess, os.Args = origEnv, origRun, origNew, origExit, origArgs
	})
}

func TestRunHelpAndVersionAliases(t *testing.T) {
	for _, arg := range []string{"-v", "--version", "version"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil || !strings.Contains(out.String(), "recomphamr") {
			t.Fatalf("%s: %q %v", arg, out.String(), err)
		}
	}
	for _, arg := range []string{"-h", "--help", "help"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil || !strings.Contains(out.String(), "Slash commands") {
			t.Fatalf("%s: %q %v", arg, out.String(), err)
		}
	}
}

func TestRunStartupAndEnvironmentContracts(t *testing.T) {
	restoreMainHooks(t)
	root := t.TempDir()
	getWorkingDirectory = func() (string, error) { return root, nil }
	getEnvironment = func(name string) string {
		if name == "RECOMPHAMR_URL" {
			return "http://override"
		}
		return ""
	}
	cfg := config.Default()
	cfg.Dir = filepath.Join(root, config.DirName)
	if err := os.Mkdir(cfg.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
	absolutePath = func(string) (string, error) { return "", errors.New("abs") }
	runTeaProgram = func(model tea.Model) error {
		if model == nil {
			t.Fatal("nil model")
		}
		return nil
	}
	var out bytes.Buffer
	if err := run([]string{"unknown-is-start"}, &out); err != nil {
		t.Fatal(err)
	}
	if cfg.URLOverride != "http://override" {
		t.Fatalf("override = %q", cfg.URLOverride)
	}
	if !strings.Contains(out.String(), "\x1b[2J") {
		t.Fatalf("clear sequence = %q", out.String())
	}

	cfg.Logging = true
	if err := run(nil, io.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestRunPropagatesStartupFailures(t *testing.T) {
	boom := errors.New("boom")
	t.Run("cwd", func(t *testing.T) {
		restoreMainHooks(t)
		getWorkingDirectory = func() (string, error) { return "", boom }
		if err := run(nil, io.Discard); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("bootstrap", func(t *testing.T) {
		restoreMainHooks(t)
		getWorkingDirectory = func() (string, error) { return t.TempDir(), nil }
		bootstrapConfig = func(string) (*config.Config, bool, error) { return nil, false, boom }
		if err := run(nil, io.Discard); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("output", func(t *testing.T) {
		restoreMainHooks(t)
		root := t.TempDir()
		getWorkingDirectory = func() (string, error) { return root, nil }
		cfg := config.Default()
		cfg.Dir = root
		bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
		if err := run(nil, failingWriter{boom}); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("program", func(t *testing.T) {
		restoreMainHooks(t)
		root := t.TempDir()
		getWorkingDirectory = func() (string, error) { return root, nil }
		cfg := config.Default()
		cfg.Dir = root
		bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
		runTeaProgram = func(tea.Model) error { return boom }
		if err := run(nil, io.Discard); !errors.Is(err, boom) {
			t.Fatalf("error = %v", err)
		}
	})
	if err := run([]string{"--version"}, failingWriter{boom}); !errors.Is(err, boom) {
		t.Fatalf("version write = %v", err)
	}
	if err := printHelp(failingWriter{boom}); !errors.Is(err, boom) {
		t.Fatalf("help write = %v", err)
	}
}

func TestMainVersionPath(t *testing.T) {
	restoreMainHooks(t)
	os.Args = []string{"recomphamr", "--version"}
	main()
}

func TestHandleRunResultExitContract(t *testing.T) {
	restoreMainHooks(t)
	code := 0
	exitProcess = func(got int) { code = got }
	handleRunResult(nil)
	if code != 0 {
		t.Fatalf("success exit = %d", code)
	}
	handleRunResult(errors.New("startup"))
	if code != 1 {
		t.Fatalf("failure exit = %d", code)
	}
}

type quittingModel struct{}

func (quittingModel) Init() tea.Cmd                         { return tea.Quit }
func (m quittingModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (quittingModel) View() string                          { return "" }

type fakeProgram struct{ err error }

func (p fakeProgram) Run() (tea.Model, error) { return quittingModel{}, p.err }

func TestExecuteTeaProgramBoundary(t *testing.T) {
	restoreMainHooks(t)
	if createTeaProgram(quittingModel{}) == nil {
		t.Fatal("factory returned nil")
	}
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{} }
	if err := executeTeaProgram(quittingModel{}); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("tea")
	newTeaProgram = func(tea.Model) teaProgram { return fakeProgram{err: boom} }
	if err := executeTeaProgram(quittingModel{}); !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
}
