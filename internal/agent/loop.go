package agent

import (
	"context"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

// LoopState owns tool ordering and all deterministic turn-policy latches.
type LoopState struct {
	LastToolKey   string
	FailKey       string
	FailStreak    int
	ToolRounds    int
	RunawayNudged bool
	EmptyNudged   bool
	VerifyNudged  bool
}

// ResetUserGoal clears failure history when a new user goal begins.
func (s *LoopState) ResetUserGoal() {
	s.FailKey, s.FailStreak = "", 0
}

// ResetTurn clears per-turn counters and latches after completion or abort.
func (s *LoopState) ResetTurn() {
	s.ToolRounds = 0
	s.RunawayNudged = false
	s.EmptyNudged = false
	s.VerifyNudged = false
}

// ToolExecutor is an injected local-tool execution boundary.
type ToolExecutor struct {
	execute func(context.Context, chmctx.ToolCall) chmctx.Message
}

// NewToolExecutor constructs an executor from a deterministic implementation.
func NewToolExecutor(execute func(context.Context, chmctx.ToolCall) chmctx.Message) ToolExecutor {
	return ToolExecutor{execute: execute}
}

// LocalToolExecutor returns the production executor backed by internal/tools.
func LocalToolExecutor() ToolExecutor {
	return NewToolExecutor(tools.Execute)
}

// ToolWork is an opaque, cancellable local-tool action.
type ToolWork struct {
	turnID   TurnID
	context  context.Context
	call     chmctx.ToolCall
	executor ToolExecutor
	status   string
}

// Status returns the accepted argument-bounded inline tool description.
func (w *ToolWork) Status() string { return w.status }

// ToolDelivery is one completed tool result tagged with its originating turn.
type ToolDelivery struct {
	TurnID  TurnID
	Message chmctx.Message
}

// Run executes the opaque tool action. Bubble Tea schedules this outside its
// Update function; cancellation flows through the turn context.
func (w *ToolWork) Run() ToolDelivery {
	return ToolDelivery{TurnID: w.turnID, Message: w.executor.execute(w.context, w.call)}
}

// NextTool pops the oldest pending call, advances policy state, and returns an
// opaque tool action. Calls remain strictly sequential in model emission order.
func (s *LoopState) NextTool(turn *TurnState, stream *StreamState, executor ToolExecutor) (*ToolWork, bool) {
	if len(stream.Pending) == 0 {
		return nil, false
	}
	call := stream.Pending[0]
	stream.Pending = stream.Pending[1:]
	s.LastToolKey = ToolTargetKey(call)
	s.ToolRounds++
	stream.Phase = PhaseRunning
	return &ToolWork{turnID: turn.ID, context: turn.context, call: call, executor: executor, status: tools.InlineStatus(call)}, true
}

// ToolResultEffect records the accepted result and any policy notes injected
// only after the complete pending-call group has drained.
type ToolResultEffect struct {
	Accepted      bool
	Message       chmctx.Message
	Failed        bool
	FailStreak    int
	FailKey       string
	FailureNudged bool
	FailureCount  int
	FailureKey    string
	RunawayNudged bool
	ToolRounds    int
	ContinueTools bool
}

// ApplyToolResult appends a current-turn result, updates repeated-failure state,
// and injects failure/runaway notes only after all sibling calls are paired.
func (s *LoopState) ApplyToolResult(turn *TurnState, stream *StreamState, delivery ToolDelivery) ToolResultEffect {
	if delivery.TurnID != turn.ID || !turn.Active() {
		return ToolResultEffect{}
	}
	effect := ToolResultEffect{Accepted: true, Message: delivery.Message}
	turn.Append(delivery.Message)
	effect.Failed = ToolResultFailed(delivery.Message.ToolName, delivery.Message.Content)
	if !effect.Failed {
		s.FailKey, s.FailStreak = "", 0
	} else if s.LastToolKey == s.FailKey && s.FailKey != "" {
		s.FailStreak++
	} else {
		s.FailKey, s.FailStreak = s.LastToolKey, 1
	}
	effect.FailStreak, effect.FailKey = s.FailStreak, s.FailKey
	if len(stream.Pending) > 0 {
		effect.ContinueTools = true
		return effect
	}
	if s.FailStreak >= MaxToolFailStreak {
		effect.FailureNudged = true
		effect.FailureCount, effect.FailureKey = s.FailStreak, s.FailKey
		turn.Append(chmctx.Message{Role: chmctx.RoleSystem, Content: FailureNudge(s.FailStreak)})
		s.FailKey, s.FailStreak = "", 0
	}
	if !s.RunawayNudged && s.ToolRounds >= MaxToolRounds {
		s.RunawayNudged = true
		effect.RunawayNudged, effect.ToolRounds = true, s.ToolRounds
		turn.Append(chmctx.Message{Role: chmctx.RoleSystem, Content: RunawayNudge(s.ToolRounds)})
	}
	stream.Phase = PhaseThinking
	return effect
}

// CloseAction tells the adapter how to continue after one model stream closes.
type CloseAction int

const (
	// CloseFinishDone completes a clean turn.
	CloseFinishDone CloseAction = iota
	// CloseFinishStopped completes a stalled or leaked turn.
	CloseFinishStopped
	// CloseRestartModel starts another model round after a policy nudge.
	CloseRestartModel
	// CloseRunTool dispatches the next pending tool.
	CloseRunTool
)

// CloseReason identifies the exact policy branch for logging and presentation.
type CloseReason int

const (
	CloseClean CloseReason = iota
	CloseToolPending
	CloseEmptyNudge
	CloseEmptyStall
	CloseToolLeak
	CloseVerifyNudge
)

// CloseDecision is the presentation-neutral result of a current stream close.
type CloseDecision struct {
	Action     CloseAction
	Reason     CloseReason
	ToolRounds int
}

// DecideClose applies the accepted empty/leak/verification policy after a
// current stream closes and appends any required system note.
func (s *LoopState) DecideClose(turn *TurnState, stream *StreamState) CloseDecision {
	stream.EndStream()
	if len(stream.Pending) > 0 {
		s.EmptyNudged = false
		return CloseDecision{Action: CloseRunTool, Reason: CloseToolPending}
	}
	if NewestAssistantEmpty(turn.History) {
		if !s.EmptyNudged {
			s.EmptyNudged = true
			turn.Append(chmctx.Message{Role: chmctx.RoleSystem, Content: EmptyReplyNudge})
			stream.Phase = PhaseThinking
			return CloseDecision{Action: CloseRestartModel, Reason: CloseEmptyNudge}
		}
		return CloseDecision{Action: CloseFinishStopped, Reason: CloseEmptyStall}
	}
	if HasToolCallLeak(turn.History) {
		return CloseDecision{Action: CloseFinishStopped, Reason: CloseToolLeak}
	}
	if !s.VerifyNudged && s.ToolRounds >= VerifyNudgeMinRounds && !NewestAssistantUnverified(turn.History) {
		s.VerifyNudged = true
		turn.Append(chmctx.Message{Role: chmctx.RoleSystem, Content: VerifyNudge})
		stream.Phase = PhaseThinking
		return CloseDecision{Action: CloseRestartModel, Reason: CloseVerifyNudge, ToolRounds: s.ToolRounds}
	}
	return CloseDecision{Action: CloseFinishDone, Reason: CloseClean}
}
