package agent

import "testing"

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
