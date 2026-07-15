package agent

import (
	"fmt"
	"strings"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

const (
	// NudgeOrigin identifies deterministic policy notes as application-generated
	// rather than as new user instructions.
	NudgeOrigin = "[Automated RecompHamr check - not a message from your user.] "

	// MaxToolFailStreak is the consecutive same-target failure threshold.
	MaxToolFailStreak = 5

	// MaxToolRounds is the per-turn runaway self-check threshold.
	MaxToolRounds = 75

	// VerifyNudgeMinRounds is the substantial-turn verification threshold.
	VerifyNudgeMinRounds = 8
)

// NewestAssistant returns the most recent assistant-role message.
func NewestAssistant(history []chmctx.Message) (chmctx.Message, bool) {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == chmctx.RoleAssistant {
			return history[i], true
		}
	}
	return chmctx.Message{}, false
}

// NewestAssistantEmpty reports whether the newest assistant message contains
// neither visible content nor a structured tool call.
func NewestAssistantEmpty(history []chmctx.Message) bool {
	message, ok := NewestAssistant(history)
	return ok && strings.TrimSpace(message.Content) == "" && len(message.ToolCalls) == 0
}

// NewestAssistantUnverified reports whether the newest assistant message
// contains the case-insensitive marker used to suppress the verification nudge.
func NewestAssistantUnverified(history []chmctx.Message) bool {
	message, ok := NewestAssistant(history)
	return ok && strings.Contains(strings.ToLower(message.Content), "unverified")
}

// HasToolCallLeak reports whether the newest assistant emitted raw tool-call
// markup as text without also emitting a structured tool call.
func HasToolCallLeak(history []chmctx.Message) bool {
	message, ok := NewestAssistant(history)
	return ok && len(message.ToolCalls) == 0 && strings.Contains(message.Content, "<tool_call>")
}

// ToolTargetKey returns the stable name-and-target identity used by repeated-
// failure detection. File tools use their path and PowerShell uses the trimmed
// first script line.
func ToolTargetKey(call chmctx.ToolCall) string {
	switch call.Name {
	case tools.WriteFileName, tools.EditFileName, tools.ReadFileName:
		path, _ := call.Arguments["path"].(string)
		return call.Name + "|" + path
	case tools.PowerShellName:
		script, _ := call.Arguments["script"].(string)
		if i := strings.IndexByte(script, '\n'); i >= 0 {
			script = script[:i]
		}
		return call.Name + "|" + strings.TrimSpace(script)
	default:
		return call.Name
	}
}

// ToolResultFailed classifies the accepted result formats for repeated-failure
// policy. User cancellation is never counted as a model failure.
func ToolResultFailed(name, result string) bool {
	if strings.Contains(result, "(cancelled)") {
		return false
	}
	trimmed := strings.TrimSpace(result)
	if strings.HasPrefix(trimmed, "(tool arguments were not valid JSON") || strings.HasPrefix(trimmed, "(unknown tool:") {
		return true
	}
	switch name {
	case tools.WriteFileName, tools.EditFileName:
		return strings.HasPrefix(trimmed, "(")
	case tools.ReadFileName:
		return strings.HasPrefix(trimmed, "(read error:") || trimmed == "(empty path)"
	case tools.PowerShellName:
		return strings.Contains(result, "\n(exit: ") || strings.Contains(result, "(timeout after ") || trimmed == "(empty script)"
	default:
		return false
	}
}

// FailureNudge returns the existing repeated-failure policy message.
func FailureNudge(streak int) string {
	return NudgeOrigin + fmt.Sprintf(
		"The last %d tool calls to the same target failed the same way. Stop repeating it - read the error, change your approach, or tell the user what's blocking you.",
		streak)
}

// RunawayNudge returns the existing long-turn self-check message.
func RunawayNudge(rounds int) string {
	return NudgeOrigin + fmt.Sprintf(
		"%d tool calls so far this turn without finishing. If you're still making real progress, keep going. If you're repeating a step that can't work here - a blocked install, a missing tool, a path failing the same way - stop chasing it (that loop burns the turn); verify another way. If you're stuck or unsure you're converging, tell the user where things stand and what's blocking you.",
		rounds)
}

// EmptyReplyNudge is the existing bounded recovery instruction for a round that
// ended with neither content nor a structured tool call.
const EmptyReplyNudge = NudgeOrigin + "Your last turn ended with no reply and no tool call. If you meant to call a tool and it did not run, issue it again now as a proper tool call. If you are still working, continue. If the task is done, check it against the original request - actually run or drive what proves it works - then reply with a one-line summary."

// VerifyNudge is the existing substantial-turn re-grounding instruction.
const VerifyNudge = NudgeOrigin + "Before you finish: re-read the original request and walk its acceptance criteria one at a time. For each, name the check you actually ran and what it showed. Anything runnable you built or changed is proven only by running it - build or type-check it, run the test, execute the script, or for a page or UI load it in a headless browser and drive the primary interaction (click Start, press the keys, submit the form) and confirm the state changed - then fix what breaks and re-run. If a check seems to need a runtime or browser this environment lacks, prove the lack with one read-only probe (`command -v node`, `command -v chromium chromium-browser google-chrome`, `ls ~/.cache/ms-playwright`) instead of assuming; the fire-once browser install your instructions allow is the ONE install worth attempting, and only if you haven't tried it this turn - never re-try a failed install or hunt missing libs, and if the probe comes up empty with no network, stop hunting. Only then mark the check `unverified: <what> - <why>` and lead your summary with it, not with a confident \"works\"; never dress up a static check (a brace count, a grep, an HTTP 200) as proof, and never report a check you didn't run. Then reply with your one-line summary."
