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
	"github.com/DohmBoy64Bit/RecompHamr/internal/skills"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
	"github.com/DohmBoy64Bit/RecompHamr/internal/workspace"
)

var (
	getWorkingDirectory = os.Getwd
	bootstrapConfig     = config.Bootstrap
	absolutePath        = filepath.Abs
	getEnvironment      = os.Getenv
	getUserHome         = os.UserHomeDir
	newSessionRuntime   = session.NewRuntime
	newAgentRuntime     = func(client agent.ChatClient, privateRoot string, skillRuntime *skills.Runtime) agent.Runtime {
		toolSet := tools.NewSet(privateRoot, config.RestrictPrivatePath)
		entries := skillRuntime.Entries()
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name)
		}
		toolSet = toolSet.WithSkillActivator(skillActivator(skillRuntime))
		toolSet = toolSet.WithSkillResourceReader(skillRuntime.ReadResource)
		return agent.NewRuntime(client, agent.NewToolExecutor(toolSet.Execute)).WithSkillTool(names).WithObserver(logging.NewObserver())
	}
	newController = func(sessionRuntime *session.Runtime, runtime agent.Runtime, skillRuntime *skills.Runtime, refresh func() skills.Catalog, initEvidence func() error, evidenceStatus func() (string, error), system func() string, version string) frontend.Controller {
		return appcontroller.NewControllerWithSkills(sessionRuntime, runtime, skillRuntime, refresh, initEvidence, evidenceStatus, system, version)
	}
	newWorkspace   = workspace.Open
	installBundled = skills.InstallBundled
	openDebugLog   = logging.Open
	closeDebugLog  = logging.Close
)

// skillActivator adapts conversation-scoped skill activation to the bounded
// text contract exposed by the model tool without exposing the skills runtime.
func skillActivator(runtime *skills.Runtime) func(string) (string, error) {
	return func(name string) (string, error) {
		_, fresh, err := runtime.Activate(name)
		if err != nil {
			return "", err
		}
		if !fresh {
			return "skill already active: " + name, nil
		}
		return "activated skill: " + name, nil
	}
}

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
	home, _ := getUserHome()
	bundledRoot, err := installBundled(cfg.Dir)
	if err != nil {
		return nil, err
	}
	discover := func() skills.Catalog {
		return discoverSkillCatalog(bundledRoot, home, projectDir, sessionRuntime.SkillSettings())
	}
	catalog := discover()
	skillRuntime := skills.NewRuntime(catalog)
	agentRuntime := newAgentRuntime(sessionRuntime, cfg.Dir, skillRuntime)
	system := func() string {
		prompt, promptErr := projectWorkspace.SystemPrompt(config.DefaultSystemPrompt)
		if promptErr != nil {
			prompt = config.DefaultSystemPrompt + "\n\nWorking directory: " + projectWorkspace.Root()
		}
		skillText := skillRuntime.SystemText()
		if skillText != "" {
			prompt += "\n\n" + skillText
		}
		return prompt
	}
	return &Runtime{controller: newController(sessionRuntime, agentRuntime, skillRuntime, discover, projectWorkspace.InitializeEvidence, projectWorkspace.EvidenceStatus, system, version), close: close}, nil
}

func discoverSkillCatalog(bundledRoot, home, projectDir string, settings config.SkillsConfig) skills.Catalog {
	return skills.Discover([]skills.Root{
		{Path: bundledRoot, Scope: skills.ScopeBundled, Trusted: true},
		{Path: filepath.Join(home, ".agents", "skills"), Scope: skills.ScopeUser, Trusted: true},
		{Path: filepath.Join(projectDir, ".agents", "skills"), Scope: skills.ScopeProject, Trusted: settings.TrustProject},
	}).Without(settings.Disabled)
}

func applyEnvOverrides(cfg *config.Config) {
	if envURL := getEnvironment("RECOMPHAMR_URL"); envURL != "" {
		cfg.URLOverride = envURL
	}
}
