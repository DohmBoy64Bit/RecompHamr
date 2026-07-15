// Command recomphamr starts the barebones RecompHamr terminal application.
package main

import (
	"fmt"
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

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("recomphamr", version)
			return
		case "-h", "--help", "help":
			printHelp()
			return
		}
	}

	cwd := mustCwd()
	cfg, _, err := config.Bootstrap(cwd)
	if err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
	applyEnvOverrides(cfg)

	if cfg.Logging {
		tui.OpenDebugLog(cfg.Dir)
		defer tui.CloseDebugLog()
	}

	profile := cfg.ActiveProfile()
	client := llm.New(cfg.ActiveURL(), profile.LLM, profile.ResolvedKey())
	projectDir, err := filepath.Abs(cwd)
	if err != nil {
		projectDir = cwd
	}
	model := tui.New(cfg, client, projectDir, version)

	// Preserve the inherited inline TUI behavior: no alternate screen and no
	// layout redesign. Scrollback remains terminal-native via tea.Println.
	_, _ = os.Stdout.WriteString("\x1b[2J\x1b[3J\x1b[H")
	if _, err := tea.NewProgram(model, tea.WithReportFocus()).Run(); err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
}

func printHelp() {
	fmt.Println(strings.TrimSpace(`
recomphamr - barebones local-first coding-agent baseline

Usage:
  recomphamr             start the inherited TUI
  recomphamr --version   print version

Slash commands:`))
	tui.PrintHelp(os.Stdout)
	fmt.Println(strings.TrimSpace(`

Keys:
  ctrl+l   clear the screen while keeping the conversation
  ctrl+c   cancel an active turn; press again while idle to quit
  ctrl+d   quit on empty input

Config:
  .rehamr/config.yaml

Environment:
  RECOMPHAMR_URL           override the active profile URL for this process
  RECOMPHAMR_IDLE_TIMEOUT  stream idle timeout, e.g. 90m or 1h`))
}

func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
	return cwd
}

func applyEnvOverrides(cfg *config.Config) {
	if envURL := os.Getenv("RECOMPHAMR_URL"); envURL != "" {
		cfg.URLOverride = envURL
	}
}
