// Package tools holds the four baseline local executors: powershell,
// read_file, write_file, and edit_file.
package tools

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

const (
	PowerShellName = "powershell"
	WriteFileName  = "write_file"
	EditFileName   = "edit_file"
	ReadFileName   = "read_file"
)

const maxPowerShellTimeoutSeconds = 3600

// PowerShell runs one script in a fresh, non-interactive PowerShell process.
// It prefers PowerShell 7 (pwsh) and falls back to Windows PowerShell where
// available. It never invokes WSL, bash, or /bin/sh.
func PowerShell(parent context.Context, script string, timeout time.Duration) string {
	if parent.Err() != nil {
		return "(cancelled)"
	}
	if strings.TrimSpace(script) == "" {
		return "(empty script)"
	}
	exe, err := findPowerShell()
	if err != nil {
		return "(powershell unavailable: " + err.Error() + ")"
	}

	ctxT, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	args := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script}
	cmd := exec.CommandContext(ctxT, exe, args...)
	configureProcessTree(cmd)
	cmd.WaitDelay = 100 * time.Millisecond

	// Bound capture before conversation truncation. A firehose command can emit
	// far more than the model will ever see; retaining the first and last 1 MiB
	// prevents the TUI from allocating unbounded output while preserving useful
	// diagnostics at both ends.
	buf := &headTailBuffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	runErr := cmd.Run()
	text := buf.String()
	if dropped := buf.droppedBytes(); dropped > 0 {
		text += fmt.Sprintf("\n(output capped at capture: %d bytes total, %d bytes dropped mid-stream)", buf.totalBytes(), dropped)
	}

	if runErr != nil {
		switch {
		case ctxT.Err() == context.DeadlineExceeded:
			return text + fmt.Sprintf("\n(timeout after %s)", timeout)
		case parent.Err() == context.Canceled || ctxT.Err() == context.Canceled:
			return text + "\n(cancelled)"
		case errors.Is(runErr, exec.ErrWaitDelay):
			return text
		default:
			text += fmt.Sprintf("\n(exit: %v)", runErr)
		}
	}
	return text
}

func findPowerShell() (string, error) {
	candidates := []string{"pwsh"}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, "powershell.exe", "powershell")
	}
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("install PowerShell 7 (pwsh) or use Windows PowerShell")
}

func PowerShellSchema() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        PowerShellName,
			"description": "Run a native PowerShell script in the current project directory. Combined stdout and stderr are returned. Use targeted commands to avoid conversation-output truncation. No WSL or bash dependency.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"script": map[string]any{
						"type":        "string",
						"description": "PowerShell script or command to execute.",
					},
					"timeout_seconds": map[string]any{
						"type":        "integer",
						"description": "Optional timeout in seconds. Default 120; hard capped at 3600.",
					},
				},
				"required": []string{"script"},
			},
		},
	}
}

// Execute dispatches one assistant tool call. The baseline deliberately has no
// extension fallback: unknown names fail explicitly instead of reaching MCP,
// skills, plugins, or deferred feature hooks.
func Execute(parent context.Context, call chmctx.ToolCall) chmctx.Message {
	raw := runRaw(parent, call)
	return chmctx.Message{
		Role:       chmctx.RoleTool,
		Content:    chmctx.Truncate(raw),
		ToolCallID: call.ID,
		ToolName:   call.Name,
	}
}

func runRaw(parent context.Context, call chmctx.ToolCall) string {
	if msg, ok := call.Arguments["_parse_error"].(string); ok {
		return fmt.Sprintf("(tool arguments were not valid JSON: %s, most likely the content was too large and the server truncated the call at its output-token limit. Do NOT retry the same oversized whole-file call. Build the file in smaller verified write_file/edit_file steps, or use PowerShell here-strings with Set-Content/Add-Content when appropriate, then verify the result.)", msg)
	}

	switch call.Name {
	case PowerShellName:
		script, _ := call.Arguments["script"].(string)
		timeout := 2 * time.Minute
		if secs, ok := call.Arguments["timeout_seconds"].(float64); ok && secs > 0 {
			// Clamp before multiplying by time.Second to avoid duration overflow.
			secs = min(max(secs, 1), maxPowerShellTimeoutSeconds)
			timeout = time.Duration(secs) * time.Second
		}
		return PowerShell(parent, script, timeout)
	case WriteFileName:
		path, _ := call.Arguments["path"].(string)
		content, ok := call.Arguments["content"].(string)
		if !ok {
			return `(missing content argument: the call carried no string "content", refusing to write - resend with the full content; an intentionally empty file needs an explicit "content": "")`
		}
		return WriteFile(path, content)
	case EditFileName:
		path, _ := call.Arguments["path"].(string)
		oldString, _ := call.Arguments["old_string"].(string)
		newString, ok := call.Arguments["new_string"].(string)
		if !ok {
			return `(missing new_string argument: the call carried no string "new_string", refusing to edit - resend it; deleting the match needs an explicit "new_string": "")`
		}
		return EditFile(path, oldString, newString)
	case ReadFileName:
		path, _ := call.Arguments["path"].(string)
		return ReadFile(path)
	default:
		return fmt.Sprintf("(unknown tool: %s)", call.Name)
	}
}

// InlineStatus is the one-line tool status shown by the inherited TUI.
func InlineStatus(call chmctx.ToolCall) string {
	switch call.Name {
	case PowerShellName:
		script, _ := call.Arguments["script"].(string)
		return "▶ powershell: " + firstLine(script)
	case WriteFileName:
		path, _ := call.Arguments["path"].(string)
		return "▶ write_file: " + path
	case EditFileName:
		path, _ := call.Arguments["path"].(string)
		return "▶ edit_file: " + path
	case ReadFileName:
		path, _ := call.Arguments["path"].(string)
		return "▶ read_file: " + path
	default:
		for _, value := range call.Arguments {
			if s, ok := value.(string); ok && s != "" {
				return fmt.Sprintf("▶ %s: %s", call.Name, firstLine(s))
			}
		}
		return "▶ " + call.Name
	}
}

const (
	powerShellOutputHead = 1 << 20
	powerShellOutputTail = 1 << 20
)

// headTailBuffer keeps the first and last bounded regions of process output.
// The tail is a fixed ring, so capture memory remains bounded for firehose
// commands while callers still receive useful beginning and ending context.
type headTailBuffer struct {
	head      []byte
	ring      []byte
	pos       int
	tailBytes int64
}

func (w *headTailBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if room := powerShellOutputHead - len(w.head); room > 0 {
		take := min(room, len(p))
		w.head = append(w.head, p[:take]...)
		p = p[take:]
	}
	if len(p) == 0 {
		return n, nil
	}
	if w.ring == nil {
		w.ring = make([]byte, powerShellOutputTail)
	}
	w.tailBytes += int64(len(p))
	if len(p) >= powerShellOutputTail {
		copy(w.ring, p[len(p)-powerShellOutputTail:])
		w.pos = 0
		return n, nil
	}
	k := copy(w.ring[w.pos:], p)
	w.pos = (w.pos + k) % powerShellOutputTail
	if k < len(p) {
		w.pos = copy(w.ring, p[k:])
	}
	return n, nil
}

func (w *headTailBuffer) droppedBytes() int64 {
	if dropped := w.tailBytes - int64(len(w.ring)); dropped > 0 {
		return dropped
	}
	return 0
}

func (w *headTailBuffer) totalBytes() int64 {
	return int64(len(w.head)) + w.tailBytes
}

func (w *headTailBuffer) String() string {
	switch {
	case w.tailBytes == 0:
		return string(w.head)
	case w.tailBytes <= int64(len(w.ring)):
		return string(w.head) + string(w.ring[:w.tailBytes])
	default:
		return string(w.head) + fmt.Sprintf("\n───── %d bytes OMITTED here (capture cap) ─────\n", w.droppedBytes()) + string(w.ring[w.pos:]) + string(w.ring[:w.pos])
	}
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 120 {
		cut := 117
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut] + "..."
	}
	return s
}
