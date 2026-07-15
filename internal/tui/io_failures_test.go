package tui

import (
	"errors"
	"strings"
	"testing"
)

func TestRenderFallback(t *testing.T) {
	boom := errors.New("boom")
	m := baselineModel(t)
	m.renderer = nil
	if _, err := renderMarkdown(&m, "x"); err == nil {
		t.Fatal("nil renderer")
	}
	m.streaming.WriteString("raw fallback")
	origRender := renderMarkdown
	t.Cleanup(func() { renderMarkdown = origRender })
	renderMarkdown = func(*Model, string) (string, error) { return "", boom }
	m.flushStreaming()
	if !strings.Contains(m.scroll.String(), "raw fallback") {
		t.Fatal("raw fallback lost")
	}
}
