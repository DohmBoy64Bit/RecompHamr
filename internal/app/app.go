// Package app composes RecompHamr services and owns the application lifecycle.
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/logging"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tui"
)

var (
	getWorkingDirectory = os.Getwd
	bootstrapConfig     = config.Bootstrap
	absolutePath        = filepath.Abs
	getEnvironment      = os.Getenv
	newClient           = llm.New
	newAgentRuntime     = func(client *llm.Client) agent.Runtime {
		return agent.NewRuntime(client, agent.LocalToolExecutor()).WithObserver(logging.NewObserver())
	}
	newFrontend = func(cfg *config.Config, client *llm.Client, runtime agent.Runtime, projectDir, version string) tea.Model {
		return tui.New(cfg, client, runtime, projectDir, version)
	}
	openDebugLog      = logging.Open
	closeDebugLog     = logging.Close
	printFrontendHelp = tui.PrintHelp
	runTeaProgram     = executeTeaProgram
	newTeaProgram     = createTeaProgram
)

type teaProgram interface {
	Run() (tea.Model, error)
}

// Run bootstraps configuration and services, constructs the terminal frontend,
// and owns its logging and Bubble Tea lifecycle until the program exits.
func Run(stdout io.Writer, version string) error {
	cwd, err := getWorkingDirectory()
	if err != nil {
		return err
	}
	cfg, _, err := bootstrapConfig(cwd)
	if err != nil {
		return err
	}
	applyEnvOverrides(cfg)

	if cfg.Logging {
		openDebugLog(cfg.Dir)
		defer closeDebugLog()
	}

	profile := cfg.ActiveProfile()
	client := newClient(cfg.ActiveURL(), profile.LLM, profile.ResolvedKey())
	runtime := newAgentRuntime(client)
	projectDir, err := absolutePath(cwd)
	if err != nil {
		projectDir = cwd
	}
	frontend := newFrontend(cfg, client, runtime, projectDir, version)

	// Preserve the accepted inline behavior: no alternate screen, and terminal
	// scrollback is still owned by Bubble Tea through tea.Println.
	if _, err := io.WriteString(stdout, "\x1b[2J\x1b[3J\x1b[H"); err != nil {
		return err
	}
	return runTeaProgram(frontend)
}

// PrintHelp writes the application-owned command, key, configuration, and
// environment contract used by the process entrypoint.
func PrintHelp(stdout io.Writer) error {
	if _, err := fmt.Fprintln(stdout, strings.TrimSpace(`
recomphamr - barebones local-first coding-agent baseline

Usage:
  recomphamr             start the inherited TUI
  recomphamr --version   print version

Slash commands:`)); err != nil {
		return err
	}
	printFrontendHelp(stdout)
	_, err := fmt.Fprintln(stdout, strings.TrimSpace(`

Keys:
  ctrl+l   clear the screen while keeping the conversation
  ctrl+c   cancel an active turn; press again while idle to quit
  ctrl+d   quit on empty input

Config:
  .rehamr/config.yaml

Environment:
  RECOMPHAMR_URL           override the active profile URL for this process
  RECOMPHAMR_IDLE_TIMEOUT  stream idle timeout, e.g. 90m or 1h`))
	return err
}

func applyEnvOverrides(cfg *config.Config) {
	if envURL := getEnvironment("RECOMPHAMR_URL"); envURL != "" {
		cfg.URLOverride = envURL
	}
}

func executeTeaProgram(model tea.Model) error {
	_, err := newTeaProgram(model).Run()
	return err
}

func createTeaProgram(model tea.Model) teaProgram {
	return tea.NewProgram(model, tea.WithReportFocus())
}
