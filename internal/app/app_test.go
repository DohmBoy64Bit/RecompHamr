package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

type fakeController struct{}

func (fakeController) Bootstrap() frontend.Transition               { return frontend.Transition{} }
func (fakeController) Dispatch(frontend.Intent) frontend.Transition { return frontend.Transition{} }
func (fakeController) Snapshot() frontend.Snapshot                  { return frontend.Snapshot{} }

func restoreAppHooks(t *testing.T) {
	origCwd, origBootstrap, origAbs, origEnv := getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment
	origSession, origAgent, origController := newSessionRuntime, newAgentRuntime, newController
	origOpen, origClose := openDebugLog, closeDebugLog
	t.Cleanup(func() {
		getWorkingDirectory, bootstrapConfig, absolutePath, getEnvironment = origCwd, origBootstrap, origAbs, origEnv
		newSessionRuntime, newAgentRuntime, newController = origSession, origAgent, origController
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
	newAgentRuntime = func(client agent.ChatClient) agent.Runtime {
		return agent.NewRuntime(client, agent.LocalToolExecutor())
	}
	createdController := false
	newController = func(sessionRuntime *session.Runtime, runtime agent.Runtime, system, version string) frontend.Controller {
		createdController = sessionRuntime != nil && runtime.Client == sessionRuntime && strings.Contains(system, "Working directory: "+root) && version == "test"
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
}
