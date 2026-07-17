package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
)

func TestNewRuntimeComposesDependencies(t *testing.T) {
	client := &fakeChatClient{}
	executor := NewToolExecutor(nil)
	runtime := NewRuntime(client, executor)
	if runtime.Client != client || runtime.Executor.execute != nil {
		t.Fatal("runtime dependencies were not preserved")
	}
	if runtime.Turn.Active() || runtime.Stream.Phase != PhaseIdle || !runtime.Stream.Connected {
		t.Fatal("runtime did not start idle and optimistically connected")
	}
}

func TestRuntimeSkillToolIsOptionalAndConstrained(t *testing.T) {
	runtime := NewRuntime(&fakeChatClient{}, NewToolExecutor(nil))
	if got := runtime.WithSkillTool(nil); len(got.tools) != 6 || len(got.toolNames) != 6 {
		t.Fatalf("empty skill tool = %d %v", len(got.tools), got.toolNames)
	}
	names := []string{"alpha"}
	got := runtime.WithSkillTool(names)
	names[0] = "source-mutated"
	if len(got.tools) != 8 || len(got.toolNames) != 8 || got.toolNames[6] != "activate_skill" || got.toolNames[7] != "read_skill_resource" || len(runtime.tools) != 6 {
		t.Fatalf("skill tools = %d %v", len(got.tools), got.toolNames)
	}
	properties := got.tools[6].Function.Parameters["properties"].(map[string]any)
	values := properties["name"].(map[string]any)["enum"].([]string)
	resourceProperties := got.tools[7].Function.Parameters["properties"].(map[string]any)
	resourceValues := resourceProperties["name"].(map[string]any)["enum"].([]string)
	if values[0] != "alpha" || resourceValues[0] != "alpha" {
		t.Fatalf("schema names = %v %v", values, resourceValues)
	}
	removed := got.WithSkillTool(nil)
	if len(removed.tools) != 6 || len(removed.toolNames) != 6 {
		t.Fatalf("removed skill tools = %d %v", len(removed.tools), removed.toolNames)
	}
}

func TestRuntimeStartsPackedRound(t *testing.T) {
	events := make(chan llm.Event)
	client := &fakeChatClient{events: events}
	runtime := NewRuntime(client, NewToolExecutor(nil))
	runtime.Turn.Begin(context.Background(), time.Now())
	runtime.Turn.History = []chmctx.Message{
		{Role: chmctx.RoleUser, Content: "───── truncated: " + strings.Repeat("x", 20)},
	}
	stream, summary := runtime.StartRound("system", "model", 200)
	if stream == nil || runtime.Stream.Stream != stream || len(client.messages) == 0 || len(client.tools) != 6 {
		t.Fatal("round dependencies were not applied")
	}
	if summary.ContextSize != 200 || summary.Budget != chmctx.Budget(200) || summary.History != 1 || summary.Packed != len(client.messages)-1 || summary.Dropped != summary.History-summary.Packed || summary.Truncated != 1 {
		t.Fatalf("summary = %#v", summary)
	}
	if summary.Packed > 0 && summary.Tokens == 0 {
		t.Fatalf("packed token summary = %#v", summary)
	}
}

func TestRuntimeCancelTurnDropsOnlyRunningToolGoal(t *testing.T) {
	runtime := NewRuntime(&fakeChatClient{}, NewToolExecutor(nil))
	prior := chmctx.Message{Role: chmctx.RoleAssistant, Content: "prior answer"}
	runtime.Turn.History = []chmctx.Message{prior}
	runtime.BeginTurn(time.Now())
	runtime.AppendUser("run a delayed side effect")
	runtime.Turn.Append(chmctx.Message{Role: chmctx.RoleAssistant, ToolCalls: []chmctx.ToolCall{{ID: "call-1", Name: "powershell"}}})
	runtime.Stream.Phase = PhaseRunning
	if runtime.CancelTurn() {
		t.Fatal("cancel unexpectedly reported a retry")
	}
	if runtime.Active() || runtime.Stream.Phase != PhaseIdle || len(runtime.Turn.History) != 1 || runtime.Turn.History[0].Content != prior.Content {
		t.Fatalf("running-tool cancellation retained an incomplete goal: %#v", runtime.Turn.History)
	}

	runtime.BeginTurn(time.Now())
	runtime.AppendUser("keep an interrupted streaming goal")
	runtime.Stream.Phase = PhaseStreaming
	runtime.CancelTurn()
	if len(runtime.Turn.History) != 2 || runtime.Turn.History[1].Content != "keep an interrupted streaming goal" {
		t.Fatalf("non-tool cancellation discarded retained history: %#v", runtime.Turn.History)
	}
}
