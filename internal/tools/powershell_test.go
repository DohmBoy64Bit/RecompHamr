package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

func TestPowerShellSchemaNameAndRequiredScript(t *testing.T) {
	fn := PowerShellSchema()["function"].(map[string]any)
	if fn["name"] != PowerShellName {
		t.Fatalf("name = %v, want %q", fn["name"], PowerShellName)
	}
	params := fn["parameters"].(map[string]any)
	required := params["required"].([]string)
	if len(required) != 1 || required[0] != "script" {
		t.Fatalf("required = %#v", required)
	}
}

func TestExecuteRejectsUnknownTool(t *testing.T) {
	msg := Execute(context.Background(), chmctx.ToolCall{Name: "future_extension", Arguments: map[string]any{}})
	if !strings.Contains(msg.Content, "unknown tool") {
		t.Fatalf("unexpected result: %q", msg.Content)
	}
}

func TestPowerShellPreCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := PowerShell(ctx, "Write-Output ok", 1); got != "(cancelled)" {
		t.Fatalf("got %q", got)
	}
}

func TestExecuteWriteFileRefusesMissingContent(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/must-not-be-truncated.txt"
	if err := os.WriteFile(path, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	msg := Execute(context.Background(), chmctx.ToolCall{
		Name:      WriteFileName,
		Arguments: map[string]any{"path": path},
	})
	if !strings.Contains(msg.Content, "missing content argument") {
		t.Fatalf("unexpected result: %q", msg.Content)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep me" {
		t.Fatalf("file changed after missing content: %q", got)
	}
}

func TestExecuteEditFileRefusesMissingNewString(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/must-not-be-edited.txt"
	if err := os.WriteFile(path, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	msg := Execute(context.Background(), chmctx.ToolCall{
		Name: EditFileName,
		Arguments: map[string]any{
			"path":       path,
			"old_string": "keep",
		},
	})
	if !strings.Contains(msg.Content, "missing new_string argument") {
		t.Fatalf("unexpected result: %q", msg.Content)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep me" {
		t.Fatalf("file changed after missing new_string: %q", got)
	}
}

func TestHeadTailBufferBoundsCapture(t *testing.T) {
	buf := &headTailBuffer{}
	head := strings.Repeat("H", powerShellOutputHead)
	middle := strings.Repeat("M", 257)
	tail := strings.Repeat("T", powerShellOutputTail)
	for _, part := range []string{head, middle, tail} {
		if _, err := buf.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if got := buf.droppedBytes(); got != int64(len(middle)) {
		t.Fatalf("droppedBytes = %d, want %d", got, len(middle))
	}
	if got := buf.totalBytes(); got != int64(len(head)+len(middle)+len(tail)) {
		t.Fatalf("totalBytes = %d", got)
	}
	out := buf.String()
	if !strings.HasPrefix(out, head[:128]) || !strings.HasSuffix(out, tail[len(tail)-128:]) {
		t.Fatal("bounded capture did not preserve head and tail")
	}
	if !strings.Contains(out, "OMITTED here") {
		t.Fatal("bounded capture did not mark the omitted middle")
	}
}

func TestFirstLineHandlesCRAndUTF8Boundary(t *testing.T) {
	if got := firstLine("first\rsecond"); got != "first" {
		t.Fatalf("CR first line = %q", got)
	}
	input := strings.Repeat("a", 116) + "🙂" + strings.Repeat("z", 20)
	got := firstLine(input)
	if !utf8.ValidString(got) {
		t.Fatalf("firstLine returned invalid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncation suffix: %q", got)
	}
}
