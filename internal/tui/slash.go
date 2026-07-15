package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	tea "github.com/charmbracelet/bubbletea"
)

// argOption is one popover entry, used at command-level (one row per command)
// and argument-level (one row per accepted value for the active command).
type argOption struct {
	value       string // what gets inserted / committed
	description string // right-aligned help text
	current     bool   // rendered bold; default-selected when the popover opens
}

// command is one row in the popover, --help, and the dispatch table.
// args, if non-nil, supplies the argument-level popover entries.
type command struct {
	name        string
	description string
	handler     func(Model, []string) (tea.Model, tea.Cmd)
	args        func(Model) []argOption
}

// commands lists every slash command, in popover/--help order. Keep it short.
var commands = []command{
	{
		name:        "/clear",
		description: "reset the conversation",
		handler:     (Model).cmdClear,
	},
	{
		name:        "/models",
		description: "list · <name> set (Tab cycles in the popover)",
		handler:     (Model).cmdModel,
		args: func(m Model) []argOption {
			facts := m.controller.Snapshot()
			out := make([]argOption, 0, len(facts.Profiles))
			for _, p := range facts.Profiles {
				out = append(out, argOption{
					value:       p.Name,
					description: p.Model + " @ " + p.URL,
					current:     p.Active,
				})
			}
			return out
		},
	},
}

// commandByName returns the registered command for a slash name, or nil.
// Centralises the linear scan shared by completion, dispatch, and runSlash.
func commandByName(name string) *command {
	for i := range commands {
		if commands[i].name == name {
			return &commands[i]
		}
	}
	return nil
}

// runSlash dispatches a slash-prefixed submission; unknown commands produce a
// quiet hint, not an error. config.yaml is re-read before every slash so
// hand-edits take effect without a restart (see reloadConfigFromDisk).
func (m Model) runSlash(text string) (tea.Model, tea.Cmd) {
	if err := m.reloadConfigFromDisk(); err != nil {
		m.appendLine(styleWarn.Render("⚠ " + err.Error()))
	}
	fields := strings.Fields(text)
	if c := commandByName(fields[0]); c != nil {
		return c.handler(m, fields[1:])
	}
	m.appendLine(styleWarn.Render("unknown command - type / to see options"))
	return m, nil
}

// reloadConfigFromDisk asks the session owner to re-bootstrap config.yaml so
// hand-edits between slash commands take effect immediately.
//
// Returns the Bootstrap error verbatim; callers decide whether to surface it
// (runSlash warns on submit; the popover-refresh path ignores it so a broken
// file doesn't spam a warning on every keystroke).
func (m *Model) reloadConfigFromDisk() error {
	transition := m.controller.Dispatch(frontend.Reload())
	for _, event := range transition.Events {
		if event.Kind == frontend.EventWarning {
			return errors.New(event.Text)
		}
	}
	return nil
}

// PrintHelp writes the canonical human-readable command list. Used by --help.
func PrintHelp(out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, c := range frontend.Commands() {
		fmt.Fprintf(w, "  %s\t%s\n", c.Name, c.Description)
	}
	w.Flush()
}

// --- handlers ---------------------------------------------------------------

// cmdModel: `/models` lists, `/models <name>` sets. Cycling is Tab/Shift+Tab
// in the popover, no separate "next" command.
func (m Model) cmdModel(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.printModelList()
		return m, nil
	}
	transition := m.controller.Dispatch(frontend.Activate(args[0]))
	for _, event := range transition.Events {
		if event.Kind == frontend.EventWarning {
			m.appendLine(styleError.Render("⚠ " + event.Text))
			return m, nil
		}
		if event.Kind == frontend.EventProfileActivated {
			m.appendActivation(event, transition.Snapshot)
		}
	}
	return m, runFrontendWork(transition.Work)
}

// printModelList writes the "▸ active, name, llm @ url" rollup to scroll.
func (m *Model) printModelList() {
	m.appendLine(styleDim.Render("models (▸ active, /models <name> to switch):"))
	for _, p := range m.controller.Snapshot().Profiles {
		mark := "  "
		if p.Active {
			mark = "▸ "
		}
		m.appendLine(fmt.Sprintf("%s%s  %s",
			mark, p.Name, styleDim.Render(p.Model+" @ "+p.URL)))
	}
}

// confirmActive emits the activation line for the active profile and returns
// its reachability command. Keyed profiles use a minimal probe so credentials
// and an optional live context window are validated; keyless profiles use the
// cheaper reachability ping.
func (m *Model) confirmActive(profile string) tea.Cmd {
	transition := m.controller.Dispatch(frontend.Activate(profile))
	for _, event := range transition.Events {
		if event.Kind == frontend.EventProfileActivated {
			m.appendActivation(event, transition.Snapshot)
		}
	}
	return runFrontendWork(transition.Work)
}

func (m *Model) appendActivation(event frontend.Event, snapshot frontend.Snapshot) {
	if snapshot.ActiveKeyed {
		m.appendLine(styleDim.Render(fmt.Sprintf("▶ probing %s · %s @ %s", event.Profile, event.Model, event.URL)))
		return
	}
	m.appendLine(styleOK.Render(fmt.Sprintf("✓ active: %s · %s @ %s", event.Profile, event.Model, event.URL)))
}

func (m Model) cmdClear(_ []string) (tea.Model, tea.Cmd) {
	m.controller.Dispatch(frontend.ClearConversation())
	m.scroll.Reset()
	// Drop any queued follow-up: it targeted the conversation just wiped.
	m.queued = nil
	// Reset the repeated-failure streak so the next turn starts clean.
	// Wipe prompt recall too: in-memory ring and on-disk .rehamr/history,
	// or leftover history would contradict the "fresh start" promise.
	m.promptHistory = nil
	m.histIdx = -1
	// Full wipe (unlike Ctrl+L, which redraws but keeps scrollback).
	// tea.ClearScreen emits \x1b[2J, which only wipes the viewport; the
	// saved-lines buffer needs eraseScrollback (DECSED 3) too, or old replies
	// stay scrollable above the reset line. tea.Sequence keeps the print from
	// racing past the clear (tea.Batch runs both concurrently and the print
	// could land first, then get wiped). scroll keeps the line for resize
	// replay; outbox is cleared because the Sequence owns the print now.
	line := styleOK.Render("✓ conversation reset")
	m.scroll.WriteString(line + "\n")
	m.outbox = nil
	// Wrap like the outbox drain would: this Println bypasses it (the Sequence
	// owns the print), and every string handed to tea.Println must be wrapped
	// or an over-width line drifts the renderer's cursor math.
	return m, tea.Sequence(tea.ClearScreen, eraseScrollback, tea.Println(wrapForScrollback(line, m.width)))
}
