// Debug instrumentation, self-contained so it can be ripped out cleanly.
// Activated by `logging: true` in config.yaml; log.txt is truncated on
// every start so a session never appends onto a stale run.
package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

var (
	dbgMu             sync.Mutex
	dbgFile           *os.File
	restrictDebugPath = config.RestrictPrivatePath
)

// OpenDebugLog truncates <dir>/log.txt and opens it for writing. On failure
// it reports once on stderr and disables logging: the log must never block
// the TUI from starting.
//
// 0o600 because the log captures every prompt: PowerShell arguments can carry secrets.
// Owner-only is the only honest answer.
func OpenDebugLog(dir string) {
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "log.txt")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠ debuglog:", err)
		return
	}
	if err := restrictDebugPath(path, false); err != nil {
		_ = f.Close()
		fmt.Fprintln(os.Stderr, "⚠ debuglog:", err)
		return
	}
	dbgMu.Lock()
	dbgFile = f
	dbgMu.Unlock()
	dbgWritef("session", "recomphamr started · project=%s", dir)
}

// CloseDebugLog flushes and closes the log. Idempotent.
func CloseDebugLog() {
	dbgMu.Lock()
	defer dbgMu.Unlock()
	if dbgFile != nil {
		_ = dbgFile.Close()
		dbgFile = nil
	}
}

// dbgEnabled reports whether logging is on. Callers use it to skip building
// expensive log payloads (e.g. accumulating a round's reasoning) when the log
// is off. dbgWritef itself is already a no-op, but the work feeding it isn't.
func dbgEnabled() bool {
	dbgMu.Lock()
	defer dbgMu.Unlock()
	return dbgFile != nil
}

// dbgWritef appends one timestamped record. No-op when logging is off. The
// timestamp carries the date too (not just the clock) so a shared log is
// unambiguous across day boundaries and correlatable with other tooling.
func dbgWritef(category, format string, args ...any) {
	dbgMu.Lock()
	defer dbgMu.Unlock()
	if dbgFile == nil {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	body := fmt.Sprintf(format, args...)
	fmt.Fprintf(dbgFile, "[%s] %s\n%s\n\n", ts, category, body)
}

// dbgWriteSession records the active backend and context budget once at startup.
// Behaviour differs sharply by model (different model families fail in
// different ways) and by context window, so a shared log must name exactly what
// produced it. The system prompt itself isn't dumped: it's the embedded
// PROMPT_SYS.md plus the working-dir anchor, both reconstructable from the repo;
// only its size (which feeds the packing budget) is worth recording.
func dbgWriteSession(version, profile, model, url string, ctxSize, sysTokens int, tools []string) {
	dbgWritef("session",
		"recomphamr %s · profile=%s · model=%s @ %s\ncontext_size=%d tokens · system_prompt≈%d tokens · tools=[%s]",
		version, profile, model, url, ctxSize, sysTokens, strings.Join(tools, ", "))
}

// dbgWriteRequest records, per LLM round, what newest-first packing actually
// sent: how much of history survived the budget and how many tool outputs are
// truncated. The message bodies are already captured as user/assistant/
// tool_result records, so this logs only the packing decisions those per-message
// records cannot show: what the model saw versus what was dropped. packed
// includes the prepended system message; historyLen is the pre-pack history.
func dbgWriteRequest(model string, summary agent.RequestSummary) {
	if !dbgEnabled() {
		return
	}
	note := ""
	if summary.Truncated > 0 {
		note = fmt.Sprintf(" · %d tool output(s) truncated", summary.Truncated)
	}
	dbgWritef("request",
		"model=%s · ctx=%d (history budget=%d) · history=%d msgs → packed=%d msgs (~%d tokens) · dropped=%d oldest%s",
		model, summary.ContextSize, summary.Budget, summary.History, summary.Packed, summary.Tokens, summary.Dropped, note)
}

// dbgWriteMessage records a chmctx.Message readably: content and tool calls
// each get a labeled section. No-op when logging is off, so callers needn't guard.
func dbgWriteMessage(category string, msg chmctx.Message) {
	dbgMu.Lock()
	enabled := dbgFile != nil
	dbgMu.Unlock()
	if !enabled {
		return
	}
	var b strings.Builder
	if msg.Content != "" {
		b.WriteString("CONTENT:\n")
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	for _, tc := range msg.ToolCalls {
		args, _ := json.Marshal(tc.Arguments)
		fmt.Fprintf(&b, "TOOL_CALL %s id=%s args=%s\n", tc.Name, tc.ID, args)
	}
	if msg.ToolCallID != "" {
		fmt.Fprintf(&b, "tool=%s id=%s\n", msg.ToolName, msg.ToolCallID)
	}
	dbgWritef(category, "%s", strings.TrimRight(b.String(), "\n"))
}
