package agent

import (
	"strings"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

// Runtime groups the mutable agent-turn components and their injected model
// and tool boundaries. The application composition root constructs it; a
// frontend adapter carries it while Stage C narrows the remaining contract.
type Runtime struct {
	Turn     *TurnState
	Stream   *StreamState
	Loop     *LoopState
	Client   ChatClient
	Executor ToolExecutor
}

// NewRuntime constructs an idle agent runtime from application-owned
// dependencies.
func NewRuntime(client ChatClient, executor ToolExecutor) Runtime {
	return Runtime{
		Turn:     newTurnState(nil),
		Stream:   newStreamState(),
		Loop:     &LoopState{},
		Client:   client,
		Executor: executor,
	}
}

// RequestSummary contains content-free context-packing facts suitable for
// presentation diagnostics and private operational logging.
type RequestSummary struct {
	ContextSize int
	Budget      int
	History     int
	Packed      int
	Tokens      int
	Truncated   int
	Dropped     int
}

// StartRound packs current history, opens exactly one model request with the
// retained tool set, and returns its opaque reader plus content-free facts.
func (r Runtime) StartRound(system string, contextSize int) (*Stream, RequestSummary) {
	messages := BuildMessages(system, r.Turn.History, contextSize)
	summary := RequestSummary{
		ContextSize: contextSize,
		Budget:      chmctx.Budget(contextSize),
		History:     len(r.Turn.History),
		Packed:      len(messages) - 1,
	}
	for _, message := range messages[1:] {
		summary.Tokens += message.Tokens()
		if strings.Contains(message.Content, "───── truncated:") {
			summary.Truncated++
		}
	}
	summary.Dropped = summary.History - summary.Packed
	return r.Stream.StartRound(r.Turn, r.Client, messages, Tools()), summary
}

func newTurnState(history []chmctx.Message) *TurnState {
	state := NewTurnState(history)
	return &state
}

func newStreamState() *StreamState {
	state := NewStreamState()
	return &state
}
