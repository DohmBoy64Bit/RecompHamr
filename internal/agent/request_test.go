package agent

import (
	"testing"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

func TestBuildMessagesAndTools(t *testing.T) {
	history := []chmctx.Message{{Role: chmctx.RoleUser, Content: "hello"}}
	messages := BuildMessages("system", history, 16177)
	if len(messages) != 2 || messages[0].Role != chmctx.RoleSystem || messages[0].Content != "system" || messages[1].Content != "hello" {
		t.Fatalf("messages = %#v", messages)
	}
	definitions := Tools()
	want := []string{tools.PowerShellName, tools.ReadFileName, tools.WriteFileName, tools.EditFileName, tools.RepomixrName, tools.RecompReferenceName}
	if len(definitions) != len(want) {
		t.Fatalf("tools = %d", len(definitions))
	}
	for i, name := range want {
		if definitions[i].Type != "function" || definitions[i].Function.Name != name || definitions[i].Function.Description == "" || definitions[i].Function.Parameters == nil {
			t.Fatalf("tool %d = %#v", i, definitions[i])
		}
	}
	for i, name := range ToolNames() {
		if name != want[i] {
			t.Fatalf("tool name %d = %q", i, name)
		}
	}
}

func TestBuildMessagesAccountsForActivatedSkillInstructions(t *testing.T) {
	largeSystem := string(make([]byte, (chmctx.FixedSystem+100)*4))
	history := []chmctx.Message{
		{Role: chmctx.RoleUser, Content: "anchor"},
		{Role: chmctx.RoleAssistant, Content: string(make([]byte, 22_288))},
		{Role: chmctx.RoleUser, Content: "latest"},
	}
	messages := BuildMessages(largeSystem, history, 20_000)
	if len(messages) != 2 || messages[1].Content != "latest" {
		t.Fatalf("skill-adjusted message count = %d, latest = %q", len(messages), messages[len(messages)-1].Content)
	}
}
