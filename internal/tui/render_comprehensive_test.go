package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderQueuedViewAndStatusStates(t *testing.T) {
	m := baselineModel(t)
	if m.View() != "" || m.renderQueued() != "" || m.queuedHeight() != 0 {
		t.Fatal("zero-size render")
	}
	m.width, m.height = 80, 24
	m.queued = &queuedPrompt{echo: strings.Repeat("long queued words ", 30), send: "x"}
	q := m.renderQueued()
	if !strings.Contains(q, "+") || m.queuedHeight() < 2 {
		t.Fatalf("queued = %q", q)
	}
	m.streaming.WriteString("live")
	m.setPopover([]argOption{{value: "x", description: "desc"}}, false, "")
	if view := m.View(); !strings.Contains(view, "live") || !strings.Contains(view, "queued") {
		t.Fatalf("view = %q", view)
	}
	m.suppressView = true
	if m.View() != "" {
		t.Fatal("suppressed view")
	}
	m.suppressView = false

	m.runtime.SessionTokens = 1234
	m.runtime.StreamingEstimate = 4
	m.runtime.Phase = phaseThinking
	m.turnStart = time.Now().Add(-2 * time.Second)
	m.status = "status"
	if bar := m.renderStatusBar(); !strings.Contains(bar, "thinking") || !strings.Contains(bar, "status") || !strings.Contains(bar, "1.2k") {
		t.Fatalf("active bar = %q", bar)
	}
	m.runtime.Phase = phaseIdle
	m.lastOutcome = outcomeDone
	m.lastElapsed = 2 * time.Second
	m.lastTokens = 20
	if bar := m.renderStatusBar(); !strings.Contains(bar, "avg") || !strings.Contains(bar, "✓") {
		t.Fatalf("done bar = %q", bar)
	}
	m.lastOutcome = outcomeStopped
	m.lastTokens = 0
	if bar := m.renderStatusBar(); !strings.Contains(bar, "✗") {
		t.Fatalf("stopped bar = %q", bar)
	}
}

func TestRenderResizeAndBoundaryHelpers(t *testing.T) {
	m := baselineModel(t)
	next, _ := m.handleWindowSize(windowSize(80, 24))
	m = next.(Model)
	next, _ = m.handleWindowSize(windowSize(80, 20))
	m = next.(Model)
	next, cmd := m.handleWindowSize(windowSize(40, 20))
	m = next.(Model)
	if cmd == nil || !m.suppressView {
		t.Fatal("width resize")
	}
	next, _ = m.handleResizeSettle(resizeSettleMsg{gen: m.resizeGen - 1})
	if !next.(Model).suppressView {
		t.Fatal("stale settle")
	}
	m.scroll.WriteString("history\n")
	m.outbox = []string{"pending"}
	next, cmd = m.handleResizeSettle(resizeSettleMsg{gen: m.resizeGen})
	m = next.(Model)
	if cmd == nil || m.suppressView || m.outbox != nil {
		t.Fatal("settle replay")
	}

	m.streaming.WriteString("raw")
	m.renderer = nil
	// A nil renderer is not a production state; keep the ordinary renderer path.
	m = baselineModel(t)
	m.streaming.WriteString("render me")
	m.flushStreaming()
	if m.streaming.Len() != 0 {
		t.Fatal("flush")
	}
	m.flushStreaming()
	if m.maxTextareaHeight() != 1 {
		t.Fatal("zero-height cap")
	}
	m.width = 2
	m.ta.SetValue("x")
	if m.visualPromptLines() != 1 {
		t.Fatal("narrow visual rows")
	}
	if wrapRows("x", 0) != 1 || wrapRows("ab cd", 2) < 2 {
		t.Fatal("wrap boundaries")
	}
	if wrapped := wrapForScrollback("a\tb", 3); wrapForScrollback("x", 0) != "x" || strings.Contains(wrapped, "\t") {
		t.Fatal("scroll wrapping")
	}
}

func windowSize(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }
