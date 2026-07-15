package agent

import (
	"strings"
	"testing"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

func TestAssistantPolicy(t *testing.T) {
	if _, ok := NewestAssistant([]chmctx.Message{{Role: chmctx.RoleUser}}); ok {
		t.Fatal("found assistant")
	}
	history := []chmctx.Message{{Role: chmctx.RoleAssistant, Content: "UNVERIFIED"}}
	if NewestAssistantEmpty(history) || !NewestAssistantUnverified(history) || HasToolCallLeak(history) {
		t.Fatal("unverified classification")
	}
	history = append(history, chmctx.Message{Role: chmctx.RoleAssistant, Content: " <tool_call> "})
	if NewestAssistantEmpty(history) || NewestAssistantUnverified(history) || !HasToolCallLeak(history) {
		t.Fatal("leak classification")
	}
	history[len(history)-1].ToolCalls = []chmctx.ToolCall{{Name: "real"}}
	if HasToolCallLeak(history) {
		t.Fatal("structured call classified as leak")
	}
	history = append(history, chmctx.Message{Role: chmctx.RoleAssistant, Content: " \n "})
	if !NewestAssistantEmpty(history) {
		t.Fatal("empty assistant not classified")
	}
}

func TestToolPolicy(t *testing.T) {
	calls := []struct {
		call chmctx.ToolCall
		key  string
	}{
		{chmctx.ToolCall{Name: tools.ReadFileName, Arguments: map[string]any{"path": "a"}}, tools.ReadFileName + "|a"},
		{chmctx.ToolCall{Name: tools.PowerShellName, Arguments: map[string]any{"script": " echo hi \nnext"}}, tools.PowerShellName + "|echo hi"},
		{chmctx.ToolCall{Name: "other"}, "other"},
	}
	for _, tc := range calls {
		if got := ToolTargetKey(tc.call); got != tc.key {
			t.Fatalf("key = %q, want %q", got, tc.key)
		}
	}

	cases := []struct {
		name, result string
		failed       bool
	}{
		{tools.WriteFileName, "(write error: no)", true},
		{tools.EditFileName, "edited a", false},
		{tools.ReadFileName, "(read error: no)", true},
		{tools.ReadFileName, "(valid file content)", false},
		{tools.ReadFileName, "(empty path)", true},
		{tools.PowerShellName, "x\n(exit: 1)", true},
		{tools.PowerShellName, "(timeout after 1s)", true},
		{tools.PowerShellName, "(empty script)", true},
		{tools.PowerShellName, "(cancelled)", false},
		{"other", "(tool arguments were not valid JSON: x)", true},
		{"other", "(unknown tool: x)", true},
		{"other", "ok", false},
	}
	for _, tc := range cases {
		if got := ToolResultFailed(tc.name, tc.result); got != tc.failed {
			t.Fatalf("ToolResultFailed(%q, %q) = %v", tc.name, tc.result, got)
		}
	}
	if !strings.Contains(FailureNudge(MaxToolFailStreak), "last 5 tool calls") || !strings.Contains(RunawayNudge(MaxToolRounds), "75 tool calls") || !strings.HasPrefix(EmptyReplyNudge, NudgeOrigin) || !strings.HasPrefix(VerifyNudge, NudgeOrigin) {
		t.Fatal("nudge contract")
	}
}
