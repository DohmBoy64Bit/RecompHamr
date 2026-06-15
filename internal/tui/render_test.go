package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// A prompt echo (or any content line) wider than the terminal must be wrapped
// before it reaches tea.Println: bubbletea dumps queued Println lines verbatim,
// so an over-width line soft-wraps in the terminal into rows the renderer never
// counted, drifting its cursor math and leaving a duplicated prompt fragment on
// screen. wrapForScrollback is the guard; every emitted line must fit the width.
func TestWrapForScrollbackCapsEveryLineToWidth(t *testing.T) {
	const width = 80
	longEcho := stylePrompt.Render("▌ ") + styleUser.Render(strings.Repeat("word ", 60))
	for line := range strings.SplitSeq(wrapForScrollback(longEcho, width), "\n") {
		if w := ansi.StringWidth(line); w > width {
			t.Fatalf("wrapped line width %d exceeds terminal width %d: %q", w, width, line)
		}
	}
}

// A long unbroken token (no spaces) must still be hard-wrapped, or it would
// soft-wrap and re-trigger the drift the helper exists to prevent.
func TestWrapForScrollbackHardWrapsUnbrokenToken(t *testing.T) {
	const width = 40
	out := wrapForScrollback(strings.Repeat("z", 200), width)
	for line := range strings.SplitSeq(out, "\n") {
		if w := ansi.StringWidth(line); w > width {
			t.Fatalf("unbroken token line width %d exceeds %d", w, width)
		}
	}
	if got := len(strings.Split(out, "\n")); got < 5 {
		t.Fatalf("expected the 200-rune token split across multiple rows, got %d", got)
	}
}

// width <= 0 (before the first WindowSizeMsg) is a no-op passthrough.
func TestWrapForScrollbackNoWidthIsNoop(t *testing.T) {
	s := "▌ some text"
	if got := wrapForScrollback(s, 0); got != s {
		t.Fatalf("zero width should pass through unchanged, got %q", got)
	}
}



