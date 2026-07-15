package agent

// Runtime groups the mutable agent-turn components and their injected model
// and tool boundaries. The application composition root constructs it; a
// frontend adapter carries it while Stage C narrows the remaining contract.
type Runtime struct {
	Turn     TurnState
	Stream   StreamState
	Loop     LoopState
	Client   ChatClient
	Executor ToolExecutor
}

// NewRuntime constructs an idle agent runtime from application-owned
// dependencies.
func NewRuntime(client ChatClient, executor ToolExecutor) Runtime {
	return Runtime{
		Turn:     NewTurnState(nil),
		Stream:   NewStreamState(),
		Client:   client,
		Executor: executor,
	}
}
