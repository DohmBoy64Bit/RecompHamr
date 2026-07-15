package tools

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
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

func TestPowerShellExecutionShapes(t *testing.T) {
	if got := PowerShell(context.Background(), "   ", time.Second); got != "(empty script)" {
		t.Fatalf("empty = %q", got)
	}
	if got := PowerShell(context.Background(), "Write-Output ok", time.Second); !strings.Contains(got, "ok") {
		t.Fatalf("success = %q", got)
	}
	if got := PowerShell(context.Background(), "exit 7", time.Second); !strings.Contains(got, "exit status 7") {
		t.Fatalf("exit = %q", got)
	}
	if got := PowerShell(context.Background(), "Start-Sleep -Seconds 2", 10*time.Millisecond); !strings.Contains(got, "timeout after") {
		t.Fatalf("timeout = %q", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := PowerShell(ctx, "Write-Output no", time.Second); got != "(cancelled)" {
		t.Fatalf("pre-cancel = %q", got)
	}
}

func TestPowerShellCapsActualProcessOutput(t *testing.T) {
	if _, err := findPowerShell(); err != nil {
		t.Skip(err)
	}
	got := PowerShell(context.Background(), `"x" * 2200000`, 10*time.Second)
	if !strings.Contains(got, "output capped at capture") || !strings.Contains(got, "bytes OMITTED") {
		t.Fatalf("large process output was not visibly bounded: len=%d", len(got))
	}
}

func TestFindPowerShellUnavailable(t *testing.T) {
	old := lookPath
	t.Cleanup(func() { lookPath = old })
	lookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	if _, err := findPowerShell(); err == nil {
		t.Fatal("missing PowerShell should fail")
	}
	if got := PowerShell(context.Background(), "echo x", time.Second); !strings.Contains(got, "powershell unavailable") {
		t.Fatalf("result = %q", got)
	}
}

func TestFinishPowerShellResult(t *testing.T) {
	boom := errors.New("boom")
	if got := finishPowerShellResult("x", nil, nil, nil, time.Second); got != "x" {
		t.Fatal(got)
	}
	if got := finishPowerShellResult("x", boom, context.DeadlineExceeded, nil, time.Second); !strings.Contains(got, "timeout") {
		t.Fatal(got)
	}
	if got := finishPowerShellResult("x", boom, nil, context.Canceled, time.Second); !strings.Contains(got, "cancelled") {
		t.Fatal(got)
	}
	if got := finishPowerShellResult("x", exec.ErrWaitDelay, nil, nil, time.Second); got != "x" {
		t.Fatal(got)
	}
	if got := finishPowerShellResult("x", boom, nil, nil, time.Second); !strings.Contains(got, "boom") {
		t.Fatal(got)
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

func TestHeadTailBufferSmallAndWrappedWrites(t *testing.T) {
	b := &headTailBuffer{}
	if n, err := b.Write([]byte("small")); err != nil || n != 5 || b.String() != "small" || b.droppedBytes() != 0 {
		t.Fatalf("small n=%d text=%q dropped=%d err=%v", n, b.String(), b.droppedBytes(), err)
	}
	b = &headTailBuffer{}
	_, _ = b.Write([]byte(strings.Repeat("h", powerShellOutputHead)))
	_, _ = b.Write([]byte(strings.Repeat("a", powerShellOutputTail-2)))
	_, _ = b.Write([]byte("WXYZ"))
	if !strings.HasSuffix(b.String(), "WXYZ") {
		t.Fatal("ring wrap lost tail")
	}
	b = &headTailBuffer{head: []byte("head"), ring: []byte("abcdef"), pos: 2, tailBytes: 9}
	if got := b.String(); got != "head\n───── 3 bytes OMITTED here (capture cap) ─────\ncdefab" {
		t.Fatalf("non-zero ring String() = %q", got)
	}
	b = &headTailBuffer{head: []byte("head"), ring: []byte("xyz000"), tailBytes: 3}
	if got := b.String(); got != "headxyz" {
		t.Fatalf("partial ring String() = %q", got)
	}
}

func TestRunRawAndInlineStatusRemainingShapes(t *testing.T) {
	parse := runRaw(context.Background(), chmctx.ToolCall{Arguments: map[string]any{"_parse_error": "bad"}})
	if !strings.Contains(parse, "not valid JSON") {
		t.Fatal(parse)
	}
	ps := runRaw(context.Background(), chmctx.ToolCall{Name: PowerShellName, Arguments: map[string]any{"script": "Write-Output ok", "timeout_seconds": float64(999999)}})
	if !strings.Contains(ps, "ok") {
		t.Fatal(ps)
	}
	for _, call := range []chmctx.ToolCall{
		{Name: PowerShellName, Arguments: map[string]any{"script": "one\ntwo"}},
		{Name: "custom", Arguments: map[string]any{"x": "detail"}},
		{Name: "custom", Arguments: map[string]any{}},
	} {
		if got := InlineStatus(call); !strings.HasPrefix(got, "▶") {
			t.Fatalf("status = %q", got)
		}
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
