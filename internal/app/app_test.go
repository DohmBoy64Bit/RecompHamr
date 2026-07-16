package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
	"github.com/DohmBoy64Bit/RecompHamr/internal/workspace"
)

type fakeController struct{}

func (fakeController) Bootstrap() frontend.Transition               { return frontend.Transition{} }
func (fakeController) Dispatch(frontend.Intent) frontend.Transition { return frontend.Transition{} }
func (fakeController) Snapshot() frontend.Snapshot                  { return frontend.Snapshot{} }

func restoreAppHooks(t *testing.T) {
	origCwd, origBootstrap, origAbs, origEnv := getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment
	origSession, origAgent, origController, origWorkspace := newSessionRuntime, newAgentRuntime, newController, newWorkspace
	origOpen, origClose := openDebugLog, closeDebugLog
	t.Cleanup(func() {
		getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment = origCwd, origBootstrap, origAbs, origEnv
		newSessionRuntime, newAgentRuntime, newController, newWorkspace = origSession, origAgent, origController, origWorkspace
		openDebugLog, closeDebugLog = origOpen, origClose
	})
}

func TestBootstrapCompositionAndClose(t *testing.T) {
	restoreAppHooks(t)
	root := t.TempDir()
	getWorkingDirectory = func() (string, error) { return root, nil }
	cfg := config.Default()
	cfg.Dir = filepath.Join(root, config.DirName)
	cfg.Logging = true
	if err := os.MkdirAll(cfg.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(cfg.Dir, workspace.StateFileName)
	if err := os.WriteFile(statePath, []byte("first state"), 0o600); err != nil {
		t.Fatal(err)
	}
	bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
	getEnvironment = func(name string) string {
		if name == "RECOMPHAMR_URL" {
			return "http://override"
		}
		return ""
	}
	absolutePath = func(string) (string, error) { return "", errors.New("abs") }
	createdSession := false
	newSessionRuntime = func(got *config.Config) *session.Runtime {
		createdSession = true
		if got.ActiveURL() != "http://override" {
			t.Fatalf("override = %q", got.ActiveURL())
		}
		return session.NewRuntime(got)
	}
	newAgentRuntime = func(client agent.ChatClient, privateRoot string) agent.Runtime {
		if privateRoot != cfg.Dir {
			t.Fatalf("private root = %q", privateRoot)
		}
		return agent.NewRuntime(client, agent.LocalToolExecutor())
	}
	createdController := false
	newController = func(sessionRuntime *session.Runtime, runtime agent.Runtime, system func() string, version string) frontend.Controller {
		first := system()
		if err := os.WriteFile(statePath, []byte("second state"), 0o600); err != nil {
			t.Fatal(err)
		}
		second := system()
		createdController = sessionRuntime != nil && runtime.Client == sessionRuntime && strings.Contains(first, "Working directory: "+root) &&
			strings.Contains(first, "first state") && strings.Contains(second, "second state") && !strings.Contains(second, "first state") && version == "test"
		return fakeController{}
	}
	opened, closed := "", 0
	openDebugLog = func(dir string) { opened = dir }
	closeDebugLog = func() { closed++ }
	runtime, err := Bootstrap("test")
	if err != nil || !createdSession || !createdController || opened != cfg.Dir || runtime.Controller() == nil {
		t.Fatalf("bootstrap runtime=%#v err=%v session=%v controller=%v open=%q", runtime, err, createdSession, createdController, opened)
	}
	runtime.Close()
	runtime.Close()
	if closed != 1 {
		t.Fatalf("close count = %d", closed)
	}
}

func TestBootstrapFailuresAndNoLogging(t *testing.T) {
	boom := errors.New("boom")
	t.Run("cwd", func(t *testing.T) {
		restoreAppHooks(t)
		getWorkingDirectory = func() (string, error) { return "", boom }
		if runtime, err := Bootstrap("test"); runtime != nil || !errors.Is(err, boom) {
			t.Fatalf("runtime=%#v err=%v", runtime, err)
		}
	})
	t.Run("config", func(t *testing.T) {
		restoreAppHooks(t)
		getWorkingDirectory = func() (string, error) { return t.TempDir(), nil }
		bootstrapConfig = func(string) (*config.Config, bool, error) { return nil, false, boom }
		if runtime, err := Bootstrap("test"); runtime != nil || !errors.Is(err, boom) {
			t.Fatalf("runtime=%#v err=%v", runtime, err)
		}
	})
	t.Run("no logging", func(t *testing.T) {
		restoreAppHooks(t)
		root := t.TempDir()
		getWorkingDirectory = func() (string, error) { return root, nil }
		cfg := config.Default()
		cfg.Dir = root
		bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
		opened := false
		openDebugLog = func(string) { opened = true }
		runtime, err := Bootstrap("test")
		if err != nil || opened {
			t.Fatalf("err=%v opened=%v", err, opened)
		}
		runtime.Close()
	})
	t.Run("workspace", func(t *testing.T) {
		restoreAppHooks(t)
		root := t.TempDir()
		getWorkingDirectory = func() (string, error) { return root, nil }
		cfg := config.Default()
		cfg.Dir = filepath.Join(root, config.DirName)
		bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
		newWorkspace = func(string) (*workspace.Workspace, error) { return nil, boom }
		if runtime, err := Bootstrap("test"); runtime != nil || !errors.Is(err, boom) {
			t.Fatalf("runtime=%#v err=%v", runtime, err)
		}
	})
	t.Run("invalid optional state falls back", func(t *testing.T) {
		restoreAppHooks(t)
		root := t.TempDir()
		getWorkingDirectory = func() (string, error) { return root, nil }
		cfg := config.Default()
		cfg.Dir = filepath.Join(root, config.DirName)
		if err := os.MkdirAll(filepath.Join(cfg.Dir, workspace.StateFileName), 0o700); err != nil {
			t.Fatal(err)
		}
		bootstrapConfig = func(string) (*config.Config, bool, error) { return cfg, false, nil }
		newController = func(_ *session.Runtime, _ agent.Runtime, system func() string, _ string) frontend.Controller {
			prompt := system()
			if !strings.Contains(prompt, "Working directory: "+root) || strings.Contains(prompt, "Persistent Memory") {
				t.Fatalf("fallback prompt = %q", prompt)
			}
			return fakeController{}
		}
		runtime, err := Bootstrap("test")
		if err != nil {
			t.Fatal(err)
		}
		runtime.Close()
	})
}
