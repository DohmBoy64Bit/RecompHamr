package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

// probeTimeout caps the activation hello-world request: long enough for a cold
// remote endpoint, short enough that a stuck backend doesn't hang "▶ probing".
const probeTimeout = 15 * time.Second

// probeMsg carries the outcome of an activation-time Probe (hello-world chat):
// validates URL+model+key in one round trip and harvests the live context
// window from response headers. profile is tagged explicitly so a late probe
// can't overwrite the wrong profile's window after a /models switch.
type probeMsg struct {
	profile       string
	contextWindow int
	silent        bool // suppress the "✓ active" line; startup probe only
	err           error
}

// probeBackend wraps llm.Client.Probe in a tea.Cmd, bounded by probeTimeout so
// a hung backend never freezes activation. silent=true (startup probe) skips
// the "✓ active" banner, just seeding the optional live context value.
func probeBackend(cli *llm.Client, profileName string, silent bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		defer cancel()
		res, err := cli.Probe(ctx)
		return probeMsg{
			profile:       profileName,
			contextWindow: res.ContextWindow,
			silent:        silent,
			err:           err,
		}
	}
}

// handleProbe consumes an activation-time Probe result. Success stores the live
// context window and prints the activation line; failure surfaces the error
// inline and leaves the active profile set. Late probes for a no-longer-active
// profile still update liveContextSize, ready for when the user switches back.
//
// Connection-state mutations are gated on the probe profile still being active.
// A late result for a previous profile cannot overwrite the current indicator.
func (m Model) handleProbe(msg probeMsg) (tea.Model, tea.Cmd) {
	active := msg.profile == m.cfg.Active
	if msg.err != nil {
		if active {
			m.runtime.Connected = false
		}
		// Silent startup probes print no banner either way; an offline launch
		// shouldn't greet the user with "⚠ probe". connected=false suffices;
		// the next user action surfaces the real failure.
		if !msg.silent {
			m.appendLine(styleError.Render("⚠ probe " + msg.profile + ": " + probeErrorMessage(msg.err)))
		}
		return m, nil
	}
	if active {
		m.runtime.Connected = true
	}
	p, ok := m.cfg.Models[msg.profile]
	if !ok {
		// Profile vanished between dispatch and return (hand-edited config or
		// pruned by /models). Skip the cache write: an orphan key would
		// accumulate across a long session.
		return m, nil
	}
	if msg.contextWindow > 0 {
		m.runtime.LiveContextSize[msg.profile] = msg.contextWindow
	}
	// Don't print "✓ active: <profile>" for a stale probe whose profile is no
	// longer active. (liveContextSize is set above.)
	if msg.silent || !active {
		return m, nil
	}
	suffix := ""
	if msg.contextWindow > 0 {
		suffix = fmt.Sprintf(" · ctx: %s", humanInt(msg.contextWindow))
	}
	m.appendLine(styleOK.Render(fmt.Sprintf(
		"✓ active: %s · %s @ %s%s", msg.profile, p.LLM, p.URL, suffix)))
	return m, nil
}

// probeErrorMessage maps provider errors to human hints for the
// activation line. Falls back to the raw error string for anything else.
func probeErrorMessage(err error) string {
	switch {
	case errors.Is(err, provider.ErrUnauthorized):
		return "key rejected"
	}
	if un, ok := errors.AsType[provider.ErrUnreachable](err); ok {
		return "unreachable (" + un.Err.Error() + ")"
	}
	return err.Error()
}
