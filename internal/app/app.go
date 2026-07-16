// Package app composes RecompHamr backend services and owns their lifecycle.
package app

import (
	"os"
	"path/filepath"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	appcontroller "github.com/DohmBoy64Bit/RecompHamr/internal/app/controller"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/logging"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
	"github.com/DohmBoy64Bit/RecompHamr/internal/workspace"
)

var (
	getWorkingDirectory = os.Getwd
	bootstrapConfig     = config.Bootstrap
	absolutePath        = filepath.Abs
	getEnvironment      = os.Getenv
	newSessionRuntime   = session.NewRuntime
	newAgentRuntime     = func(client agent.ChatClient, privateRoot string) agent.Runtime {
		toolSet := tools.NewSet(privateRoot, config.RestrictPrivatePath)
		return agent.NewRuntime(client, agent.NewToolExecutor(toolSet.Execute)).WithObserver(logging.NewObserver())
	}
	newController = func(sessionRuntime *session.Runtime, runtime agent.Runtime, system func() string, version string) frontend.Controller {
		return appcontroller.NewController(sessionRuntime, runtime, system, version)
	}
	newWorkspace  = workspace.Open
	openDebugLog  = logging.Open
	closeDebugLog = logging.Close
)

// Runtime is the application-owned backend lifetime exposed to concrete
// frontend adapters. It reveals only the neutral controller and idempotent
// cleanup, never concrete session, agent, logging, or credential capabilities.
type Runtime struct {
	controller frontend.Controller
	close      func()
}

// Controller returns the backend-neutral presentation boundary.
func (r *Runtime) Controller() frontend.Controller { return r.controller }

// Close releases application-owned resources. Calling Close more than once is
// safe and closes the private debug log at most once.
func (r *Runtime) Close() {
	if r.close != nil {
		r.close()
		r.close = nil
	}
}

// Bootstrap loads configuration, applies process overrides, composes session
// and agent services, and returns their neutral controller lifetime.
func Bootstrap(version string) (*Runtime, error) {
	cwd, err := getWorkingDirectory()
	if err != nil {
		return nil, err
	}
	cfg, _, err := bootstrapConfig(cwd)
	if err != nil {
		return nil, err
	}
	applyEnvOverrides(cfg)
	projectDir, err := absolutePath(cwd)
	if err != nil {
		projectDir = cwd
	}
	projectWorkspace, err := newWorkspace(projectDir)
	if err != nil {
		return nil, err
	}

	close := func() {}
	if cfg.Logging {
		openDebugLog(cfg.Dir)
		close = closeDebugLog
	}
	sessionRuntime := newSessionRuntime(cfg)
	agentRuntime := newAgentRuntime(sessionRuntime, cfg.Dir)
	system := func() string {
		prompt, promptErr := projectWorkspace.SystemPrompt(config.DefaultSystemPrompt)
		if promptErr != nil {
			return config.DefaultSystemPrompt + "\n\nWorking directory: " + projectWorkspace.Root()
		}
		return prompt
	}
	return &Runtime{controller: newController(sessionRuntime, agentRuntime, system, version), close: close}, nil
}

func applyEnvOverrides(cfg *config.Config) {
	if envURL := getEnvironment("RECOMPHAMR_URL"); envURL != "" {
		cfg.URLOverride = envURL
	}
}
