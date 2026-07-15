// Package terminal owns the concrete Bubble Tea frontend adapter and terminal lifecycle.
package terminal

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/app"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tui"
)

var (
	bootstrapApplication func(string) (applicationRuntime, error) = func(version string) (applicationRuntime, error) { return app.Bootstrap(version) }
	newFrontend                                                   = func(controller frontend.Controller, version string) tea.Model { return tui.New(controller, version) }
	newTeaProgram                                                 = createTeaProgram
)

type applicationRuntime interface {
	Controller() frontend.Controller
	Close()
}

type teaProgram interface{ Run() (tea.Model, error) }

// Run composes the application runtime with the concrete terminal frontend and
// owns inline clearing, Bubble Tea focus reporting, execution, and cleanup.
func Run(stdout io.Writer, version string) error {
	runtime, err := bootstrapApplication(version)
	if err != nil {
		return err
	}
	defer runtime.Close()
	model := newFrontend(runtime.Controller(), version)
	if _, err := io.WriteString(stdout, "\x1b[2J\x1b[3J\x1b[H"); err != nil {
		return err
	}
	_, err = newTeaProgram(model).Run()
	return err
}

// PrintHelp writes the terminal command, key, configuration, and environment contract.
func PrintHelp(stdout io.Writer) error {
	if _, err := fmt.Fprintln(stdout, strings.TrimSpace(`
recomphamr - barebones local-first coding-agent baseline

Usage:
  recomphamr             start the inherited TUI
  recomphamr --version   print version

Slash commands:`)); err != nil {
		return err
	}
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	for _, command := range frontend.Commands() {
		_, _ = fmt.Fprintf(w, "  %s\t%s\n", command.Name, command.Description)
	}
	if err := w.Flush(); err != nil {
		return err
	}
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

func createTeaProgram(model tea.Model) teaProgram {
	return tea.NewProgram(model, tea.WithReportFocus())
}
