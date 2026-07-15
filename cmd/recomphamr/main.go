// Command recomphamr starts the barebones RecompHamr terminal application.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tui"
)

var version = "dev"

var (
	getWorkingDirectory = os.Getwd
	bootstrapConfig     = config.Bootstrap
	absolutePath        = filepath.Abs
	getEnvironment      = os.Getenv
	runTeaProgram       = executeTeaProgram
	newTeaProgram       = createTeaProgram
	exitProcess         = os.Exit
)

type teaProgram interface {
	Run() (tea.Model, error)
}

func main() {
	handleRunResult(run(os.Args[1:], os.Stdout))
}

func executeTeaProgram(model tea.Model) error {
	_, err := newTeaProgram(model).Run()
	return err
}

func createTeaProgram(model tea.Model) teaProgram {
	return tea.NewProgram(model, tea.WithReportFocus())
}

func handleRunResult(err error) {
	if err != nil {
		log.Printf("recomphamr: %v", err)
		exitProcess(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "-v", "--version", "version":
			_, err := fmt.Fprintln(stdout, "recomphamr", version)
			return err
		case "-h", "--help", "help":
			return printHelp(stdout)
		}
	}

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
		tui.OpenDebugLog(cfg.Dir)
		defer tui.CloseDebugLog()
	}

	profile := cfg.ActiveProfile()
	client := llm.New(cfg.ActiveURL(), profile.LLM, profile.ResolvedKey())
	projectDir, err := absolutePath(cwd)
	if err != nil {
		projectDir = cwd
	}
	model := tui.New(cfg, client, projectDir, version)

	// Preserve the inherited inline TUI behavior: no alternate screen and no
	// layout redesign. Scrollback remains terminal-native via tea.Println.
	if _, err := io.WriteString(stdout, "\x1b[2J\x1b[3J\x1b[H"); err != nil {
		return err
	}
	return runTeaProgram(model)
}

func printHelp(stdout io.Writer) error {
	if _, err := fmt.Fprintln(stdout, strings.TrimSpace(`
recomphamr - barebones local-first coding-agent baseline

Usage:
  recomphamr             start the inherited TUI
  recomphamr --version   print version

Slash commands:`)); err != nil {
		return err
	}
	tui.PrintHelp(stdout)
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
