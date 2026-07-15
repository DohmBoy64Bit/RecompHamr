package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeHistoryAppender struct {
	writeErr, closeErr error
	closed             bool
}

func (f *fakeHistoryAppender) WriteString(string) (int, error) { return 0, f.writeErr }
func (f *fakeHistoryAppender) Close() error                    { f.closed = true; return f.closeErr }

func TestHistoryAppendFailureAndTrimBoundaries(t *testing.T) {
	boom := errors.New("boom")
	origOpen, origRestrict, origCount := openHistoryAppend, restrictHistoryPath, countStoredHistory
	t.Cleanup(func() { openHistoryAppend, restrictHistoryPath, countStoredHistory = origOpen, origRestrict, origCount })
	if err := appendPromptHistory(filepath.Join(t.TempDir(), "missing"), "x"); err == nil {
		t.Fatal("open failure")
	}

	f := &fakeHistoryAppender{}
	openHistoryAppend = func(string) (historyAppender, error) { return f, nil }
	restrictHistoryPath = func(string, bool) error { return boom }
	if err := appendPromptHistory("x", "x"); !errors.Is(err, boom) || !f.closed {
		t.Fatalf("restrict = %v closed=%v", err, f.closed)
	}
	restrictHistoryPath = func(string, bool) error { return nil }
	f = &fakeHistoryAppender{writeErr: boom}
	openHistoryAppend = func(string) (historyAppender, error) { return f, nil }
	if err := appendPromptHistory("x", "x"); !errors.Is(err, boom) || !f.closed {
		t.Fatalf("write = %v", err)
	}
	f = &fakeHistoryAppender{closeErr: boom}
	openHistoryAppend = func(string) (historyAppender, error) { return f, nil }
	if err := appendPromptHistory("x", "x"); !errors.Is(err, boom) {
		t.Fatalf("close = %v", err)
	}
	f = &fakeHistoryAppender{}
	openHistoryAppend = func(string) (historyAppender, error) { return f, nil }
	countStoredHistory = func(string) (int, error) { return 0, boom }
	if err := appendPromptHistory("x", "x"); err != nil {
		t.Fatalf("count best effort = %v", err)
	}
	if _, err := countHistoryLines(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("count missing")
	}
}

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

func TestHistoryOversizeAndScannerError(t *testing.T) {
	if err := appendPromptHistory(t.TempDir(), strings.Repeat("x", historyMaxEntryBytes+1)); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "huge")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", historyScannerMax+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := countHistoryLines(path); err == nil {
		t.Fatal("scanner overflow")
	}
}
