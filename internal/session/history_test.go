package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryRoundTripBoundsAndClear(t *testing.T) {
	dir := t.TempDir()
	history := NewHistory(dir)
	if got := history.Load(); len(got) != 0 {
		t.Fatalf("fresh history = %d", len(got))
	}
	want := []string{"first", "second\nwith newline", `third "quoted"`, "café 🐹"}
	for _, value := range want {
		if err := history.Append(value); err != nil {
			t.Fatal(err)
		}
	}
	got := history.Load()
	if len(got) != len(want) {
		t.Fatalf("history = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %q, want %q", i, got[i], want[i])
		}
	}
	if err := history.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := history.Clear(); err != nil {
		t.Fatalf("idempotent clear = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, historyFileName)); !os.IsNotExist(err) {
		t.Fatalf("history remains: %v", err)
	}
}

func TestHistoryTrimMalformedAndOversizeRecovery(t *testing.T) {
	dir := t.TempDir()
	history := NewHistory(dir)
	if err := history.Append(""); err != nil {
		t.Fatal(err)
	}
	if err := history.Append(strings.Repeat("x", historyMaxEntryBytes+1)); err != nil {
		t.Fatal(err)
	}
	if err := history.Append(strings.Repeat("\x01", historyMaxEntryBytes)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < historyMaxEntries+50; i++ {
		if err := history.Append(fmt.Sprintf("p-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	got := history.Load()
	if len(got) != historyMaxEntries || got[0] != "p-50" || got[len(got)-1] != "p-549" {
		t.Fatalf("trimmed history boundary = %d %q %q", len(got), got[0], got[len(got)-1])
	}
	path := filepath.Join(dir, historyFileName)
	if err := os.WriteFile(path, []byte("not quoted\n\"valid\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := history.Load(); len(got) != 1 || got[0] != "valid" {
		t.Fatalf("malformed recovery = %#v", got)
	}
	atCap := strings.Repeat("y", historyMaxEntryBytes)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := history.Append(atCap); err != nil {
		t.Fatal(err)
	}
	if got := history.Load(); len(got) != 1 || got[0] != atCap {
		t.Fatalf("at-cap round trip = %d", len(got))
	}
}

func TestHistoryConcurrentAppendsKeepBoth(t *testing.T) {
	history := NewHistory(t.TempDir())
	const count = 50
	done := make(chan struct{}, 2)
	for _, prefix := range []string{"alpha", "beta"} {
		go func(prefix string) {
			for i := 0; i < count; i++ {
				_ = history.Append(fmt.Sprintf("%s-%d", prefix, i))
			}
			done <- struct{}{}
		}(prefix)
	}
	<-done
	<-done
	got := history.Load()
	if len(got) != 2*count {
		t.Fatalf("concurrent history = %d/%d", len(got), 2*count)
	}
	seen := map[string]bool{}
	for _, value := range got {
		seen[value] = true
	}
	for _, prefix := range []string{"alpha", "beta"} {
		for i := 0; i < count; i++ {
			if !seen[fmt.Sprintf("%s-%d", prefix, i)] {
				t.Fatalf("missing %s-%d", prefix, i)
			}
		}
	}
}

type fakeHistoryAppender struct {
	writeErr error
	closeErr error
	closed   bool
}

func (f *fakeHistoryAppender) WriteString(string) (int, error) { return 0, f.writeErr }
func (f *fakeHistoryAppender) Close() error                    { f.closed = true; return f.closeErr }

func restoreHistoryHooks(t *testing.T) {
	originalOpen, originalRestrict, originalCount, originalRemove := openHistoryAppend, restrictHistoryPath, countStoredHistory, removeHistoryPath
	t.Cleanup(func() {
		openHistoryAppend, restrictHistoryPath, countStoredHistory = originalOpen, originalRestrict, originalCount
		removeHistoryPath = originalRemove
	})
}

func TestHistoryFailureAndScannerBoundaries(t *testing.T) {
	boom := errors.New("boom")
	t.Run("open", func(t *testing.T) {
		history := NewHistory(filepath.Join(t.TempDir(), "missing"))
		if err := history.Append("x"); err == nil {
			t.Fatal("expected open failure")
		}
	})
	for _, tc := range []struct {
		name     string
		appender *fakeHistoryAppender
		restrict error
	}{
		{name: "restrict", appender: &fakeHistoryAppender{}, restrict: boom},
		{name: "write", appender: &fakeHistoryAppender{writeErr: boom}},
		{name: "close", appender: &fakeHistoryAppender{closeErr: boom}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restoreHistoryHooks(t)
			openHistoryAppend = func(string) (historyAppender, error) { return tc.appender, nil }
			restrictHistoryPath = func(string, bool) error { return tc.restrict }
			err := NewHistory("x").Append("x")
			if !errors.Is(err, boom) || (tc.name != "close" && !tc.appender.closed) {
				t.Fatalf("error=%v closed=%v", err, tc.appender.closed)
			}
		})
	}
	t.Run("count best effort", func(t *testing.T) {
		restoreHistoryHooks(t)
		openHistoryAppend = func(string) (historyAppender, error) { return &fakeHistoryAppender{}, nil }
		restrictHistoryPath = func(string, bool) error { return nil }
		countStoredHistory = func(string) (int, error) { return 0, boom }
		if err := NewHistory("x").Append("x"); err != nil {
			t.Fatal(err)
		}
	})
	if _, err := countHistoryLines(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected missing-file count error")
	}
	path := filepath.Join(t.TempDir(), "huge")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", historyScannerMax+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := countHistoryLines(path); err == nil {
		t.Fatal("expected scanner overflow")
	}
	if err := NewHistory(t.TempDir()).Clear(); err != nil {
		t.Fatal(err)
	}
	restoreHistoryHooks(t)
	removeHistoryPath = func(string) error { return boom }
	if err := NewHistory("x").Clear(); !errors.Is(err, boom) {
		t.Fatalf("clear failure = %v", err)
	}
}
