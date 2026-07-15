package agent

import (
	"context"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

// TurnID is a process-local monotonically increasing identity. Async results
// carry it across the presentation boundary so cancelled work cannot mutate a
// later turn.
type TurnID uint64

// TurnState is the mechanically extracted mutable root of one agent turn.
// Call Begin, End, and Reset to preserve cancellation and identity invariants.
// Its fields remain exported during Stage C so the existing value-model adapter
// can move incrementally; presentation must treat Context and CancelFunc as
// opaque and must not expose them in snapshots or logs.
type TurnState struct {
	History   []chmctx.Message
	context   context.Context
	cancel    context.CancelFunc
	StartedAt time.Time
	ID        TurnID
	nextID    TurnID
}

// NewTurnState constructs an idle turn root with optional existing history.
func NewTurnState(history []chmctx.Message) TurnState {
	return TurnState{History: history}
}

// Begin cancels any prior turn, creates a fresh context and stable identity,
// records its start time, and returns the new identity.
func (s *TurnState) Begin(parent context.Context, now time.Time) TurnID {
	s.End()
	if parent == nil {
		parent = context.Background()
	}
	s.nextID++
	s.ID = s.nextID
	s.context, s.cancel = context.WithCancel(parent)
	s.StartedAt = now
	return s.ID
}

// Active reports whether a turn context is installed.
func (s *TurnState) Active() bool {
	return s.cancel != nil
}

// Append adds a model-facing conversation message in causal order.
func (s *TurnState) Append(message chmctx.Message) {
	s.History = append(s.History, message)
}

// discardCurrentGoal removes the newest user goal and everything the model
// produced for it. It is used only when cancellation interrupts a running tool:
// retaining that unresolved instruction would invite the next model turn to
// execute the cancelled side effect again.
func (s *TurnState) discardCurrentGoal() {
	for i := len(s.History) - 1; i >= 0; i-- {
		if s.History[i].Role == chmctx.RoleUser {
			s.History = s.History[:i]
			return
		}
	}
}

// End cancels and releases the active context while retaining history and the
// last identity for stale-result comparison.
func (s *TurnState) End() {
	if s.cancel != nil {
		s.cancel()
	}
	s.context = nil
	s.cancel = nil
	s.StartedAt = time.Time{}
}

// Reset clears model-facing history after ending any active turn.
func (s *TurnState) Reset() {
	s.End()
	s.History = nil
}
