// Package logging owns the private runtime debug-log sink independently of
// terminal presentation.
package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

var (
	mu           sync.Mutex
	file         *os.File
	restrictPath = config.RestrictPrivatePath
)

// Observer is the stateless adapter injected into agent orchestration.
type Observer struct{}

// NewObserver constructs a private runtime observer backed by the current
// process log lifecycle.
func NewObserver() Observer { return Observer{} }

// Enabled reports whether private logging is active.
func (Observer) Enabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return file != nil
}

// Writef appends one timestamped category record when logging is active.
func (Observer) Writef(category, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	body := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(file, "[%s] %s\n%s\n\n", timestamp, category, body)
}

// WriteMessage records one model-facing message readably. Arguments and
// content remain only in the owner-protected log and never enter reports.
func (o Observer) WriteMessage(category string, message chmctx.Message) {
	if !o.Enabled() {
		return
	}
	var builder strings.Builder
	if message.Content != "" {
		builder.WriteString("CONTENT:\n")
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	for _, call := range message.ToolCalls {
		arguments, _ := json.Marshal(call.Arguments)
		_, _ = fmt.Fprintf(&builder, "TOOL_CALL %s id=%s args=%s\n", call.Name, call.ID, arguments)
	}
	if message.ToolCallID != "" {
		_, _ = fmt.Fprintf(&builder, "tool=%s id=%s\n", message.ToolName, message.ToolCallID)
	}
	o.Writef(category, "%s", strings.TrimRight(builder.String(), "\n"))
}

// Open truncates dir/log.txt and enables private logging. Failure is reported
// once on stderr and never prevents application startup.
func Open(dir string) {
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "log.txt")
	opened, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "⚠ debuglog:", err)
		return
	}
	if err := restrictPath(path, false); err != nil {
		_ = opened.Close()
		_, _ = fmt.Fprintln(os.Stderr, "⚠ debuglog:", err)
		return
	}
	mu.Lock()
	file = opened
	mu.Unlock()
	NewObserver().Writef("session", "recomphamr started · project=%s", dir)
}

// Close flushes and closes the private log. It is idempotent.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}
