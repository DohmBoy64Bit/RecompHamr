package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
			m.appendActivation(event, transition.Snapshot)
		case frontend.EventProbe:
			m.applyProbeEvent(event, transition.Snapshot)
		case frontend.EventTurnStarted:
			m.turnStart = event.At
			m.lastOutcome = outcomeNone
		case frontend.EventContent:
			m.streaming.WriteString(strings.ReplaceAll(event.Text, "\t", "    "))
		case frontend.EventFlush:
			m.flushStreaming()
		case frontend.EventStatus:
			m.status = event.Text
		case frontend.EventToolStatus:
			m.appendLine(styleDim.Render(event.Text))
		case frontend.EventTurnFinished:
			m.applyTurnFinished(event)
			if event.Natural {
				m.fireQueuedAfterTransition = true
			}
		}
	}
	return m, runFrontendWork(transition.Work)
}

func (m *Model) applyTurnFinished(event frontend.Event) {
	m.flushStreaming()
	if event.Text != "" {
		if event.Cancelled {
			m.appendLine(styleWarn.Render(event.Text))
		} else {
			m.appendLine(styleError.Render(event.Text))
		}
	}
	if !event.Natural && m.queued != nil {
		if m.ta.Value() == "" {
			m.setPromptText(m.queued.send)
		}
		m.queued = nil
	}
	m.lastElapsed = event.Elapsed
	m.lastTokens = event.Tokens
	if event.OK {
		m.lastOutcome = outcomeDone
	} else {
		m.lastOutcome = outcomeStopped
	}
	m.turnStart = time.Time{}
	m.status = ""
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
