package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
)

// handleProbe consumes an activation-time Probe result. Success stores the live
// context window and prints the activation line; failure surfaces the error
// inline and leaves the active profile set. Late probes for a no-longer-active
// profile still update liveContextSize, ready for when the user switches back.
//
// Connection-state mutations are gated on the probe profile still being active.
// A late result for a previous profile cannot overwrite the current indicator.

func (m Model) applyFrontendTransition(transition frontend.Transition) (tea.Model, tea.Cmd) {
	for _, event := range transition.Events {
		switch event.Kind {
		case frontend.EventWarning:
			m.appendLine(styleWarn.Render("⚠ " + event.Text))
		case frontend.EventProfileActivated:
			if transition.Snapshot.ActiveKeyed {
				m.appendLine(styleDim.Render(fmt.Sprintf("▶ probing %s · %s @ %s", event.Profile, event.Model, event.URL)))
			} else {
				m.appendLine(styleOK.Render(fmt.Sprintf("✓ active: %s · %s @ %s", event.Profile, event.Model, event.URL)))
			}
		case frontend.EventProbe:
			m.applyProbeEvent(event, transition.Snapshot)
		}
	}
	return m, runFrontendWork(transition.Work)
}

func (m *Model) applyProbeEvent(event frontend.Event, facts frontend.Snapshot) {
	active := event.Profile == facts.Active
	if !event.OK {
		m.appendLine(styleError.Render("⚠ probe " + event.Profile + ": " + event.Text))
		return
	}
	p, ok := facts.Profile(event.Profile)
	if !ok || !active {
		return
	}
	// Don't print "✓ active: <profile>" for a stale probe whose profile is no
	// longer active. (liveContextSize is set above.)
	suffix := ""
	if event.ContextWindow > 0 {
		suffix = fmt.Sprintf(" · ctx: %s", humanInt(event.ContextWindow))
	}
	m.appendLine(styleOK.Render(fmt.Sprintf(
		"✓ active: %s · %s @ %s%s", event.Profile, p.Model, p.URL, suffix)))
}

// probeErrorMessage maps provider errors to human hints for the
// activation line. Falls back to the raw error string for anything else.
func probeErrorMessage(err error) string {
	return agent.ProbeErrorMessage(err)
}
