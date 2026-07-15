package logging

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

func TestLifecycleAndPayloads(t *testing.T) {
	Close()
	Open("")
	observer := NewObserver()
	if observer.Enabled() {
		t.Fatal("empty path enabled logging")
	}
	dir := t.TempDir()
	Open(dir)
	if !observer.Enabled() {
		t.Fatal("logging disabled")
	}
	observer.Writef("request", "model=%s", "m")
	observer.WriteMessage("assistant", chmctx.Message{Role: chmctx.RoleAssistant, Content: "answer", ToolCalls: []chmctx.ToolCall{{ID: "1", Name: "read_file", Arguments: map[string]any{"path": "x"}}}})
	observer.WriteMessage("tool", chmctx.Message{Role: chmctx.RoleTool, ToolName: "read_file", ToolCallID: "1"})
	Close()
	Close()
	data, err := os.ReadFile(filepath.Join(dir, "log.txt"))
	if err != nil || !strings.Contains(string(data), "recomphamr started") || !strings.Contains(string(data), "TOOL_CALL read_file") || !strings.Contains(string(data), "tool=read_file") {
		t.Fatalf("log = %q %v", data, err)
	}
	observer.Writef("off", "ignored")
	observer.WriteMessage("off", chmctx.Message{})
	Open(filepath.Join(dir, "missing", "child"))
}

func TestRestrictionFailure(t *testing.T) {
	original := restrictPath
	t.Cleanup(func() { restrictPath = original; Close() })
	restrictPath = func(string, bool) error { return errors.New("boom") }
	Open(t.TempDir())
	if NewObserver().Enabled() {
		t.Fatal("restriction failure left logging enabled")
	}
}
