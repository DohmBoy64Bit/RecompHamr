package agent

import (
	"context"
	"errors"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

// Phase is the presentation-neutral state of an agent turn.
type Phase int

const (
	// PhaseIdle means no turn is active.
	PhaseIdle Phase = iota
	// PhaseThinking means the model is active without user-facing output yet.
	PhaseThinking
	// PhaseStreaming means the model is producing content or tool arguments.
	PhaseStreaming
	// PhaseRunning means a local tool is executing.
	PhaseRunning
)

// Active reports whether the phase belongs to an in-flight turn.
func (p Phase) Active() bool { return p != PhaseIdle }

// Label returns the accepted user-facing status label for a phase.
func (p Phase) Label() string {
	switch p {
	case PhaseThinking:
		return "thinking"
	case PhaseStreaming:
		return "generating"
	case PhaseRunning:
		return "running"
	default:
		return ""
	}
}

// StreamState owns mutable model-stream, pending-call, accounting, connection,
// and live-context facts. Stage C's TUI adapter reads these facts for rendering
// but does not define their policy.
type StreamState struct {
	Stream            *Stream
	Pending           []chmctx.ToolCall
	Phase             Phase
	Retrying          bool
	TurnTokens        int
	SessionTokens     int
	StreamingEstimate int
	Connected         bool
	LiveContextSize   map[string]int
	nextRoundID       RoundID
}

// RoundID identifies one model request within a turn.
type RoundID uint64

// Stream is an opaque model-event reader tagged with its owning turn and round.
// Presentation can schedule Read but cannot access the transport channel.
type Stream struct {
	events  <-chan llm.Event
	turnID  TurnID
	roundID RoundID
}

// ChatClient is the provider-neutral model boundary required by orchestration.
type ChatClient interface {
	Chat(context.Context, []chmctx.Message, []llm.Tool) <-chan llm.Event
}

// StreamDelivery is one event or closure notification from an opaque Stream.
type StreamDelivery struct {
	turnID  TurnID
	roundID RoundID
	event   llm.Event
	closed  bool
}

// Closed reports whether the opaque stream ended instead of yielding an event.
func (d StreamDelivery) Closed() bool { return d.closed }

// BeginStream installs a newly opened model channel and returns its opaque,
// stably identified reader.
func (s *StreamState) BeginStream(turnID TurnID, events <-chan llm.Event) *Stream {
	s.nextRoundID++
	s.Stream = &Stream{events: events, turnID: turnID, roundID: s.nextRoundID}
	return s.Stream
}

// StartRound opens one model request against the active turn and installs its
// opaque stream reader.
func (s *StreamState) StartRound(turn *TurnState, client ChatClient, messages []chmctx.Message, tools []llm.Tool) *Stream {
	return s.BeginStream(turn.ID, client.Chat(turn.context, messages, tools))
}

// Current reports whether stream is the active turn/round reader.
func (s *StreamState) Current(stream *Stream) bool {
	return stream != nil && s.Stream != nil && stream.turnID == s.Stream.turnID && stream.roundID == s.Stream.roundID
}

// EndStream releases the active reader after its close delivery is accepted.
func (s *StreamState) EndStream() {
	s.Stream = nil
}

// Read blocks for one model event. Bubble Tea invokes it outside Update.
func (s *Stream) Read() StreamDelivery {
	event, ok := <-s.events
	return StreamDelivery{turnID: s.turnID, roundID: s.roundID, event: event, closed: !ok}
}

// DeliveryEffect is the typed result of validating and reducing one opaque
// stream delivery. Stale deliveries are rejected without exposing transport
// events or runtime identity to presentation.
type DeliveryEffect struct {
	Accepted bool
	Closed   bool
	Stream   StreamEffect
}

// ApplyDelivery validates stable turn/round identity and reduces a current
// event. A current closure is reported without ending the stream because close
// policy owns that transition; stale deliveries never mutate agent state.
func (s *StreamState) ApplyDelivery(turn *TurnState, stream *Stream, profile string, delivery StreamDelivery) DeliveryEffect {
	if !s.Current(stream) || delivery.turnID != turn.ID || delivery.roundID != stream.roundID {
		return DeliveryEffect{}
	}
	if delivery.closed {
		return DeliveryEffect{Accepted: true, Closed: true}
	}
	return DeliveryEffect{Accepted: true, Stream: s.Apply(turn, profile, delivery.event)}
}

// NewStreamState constructs an idle, optimistically connected stream state.
func NewStreamState() StreamState {
	return StreamState{Connected: true, LiveContextSize: map[string]int{}}
}

// ResetTurn clears state that must never cross a turn boundary while retaining
// session totals, connectivity, and per-profile context hints.
func (s *StreamState) ResetTurn() {
	s.Stream = nil
	s.Pending = nil
	s.Phase = PhaseIdle
	s.Retrying = false
}

// Finalize folds any interrupted live estimate into token totals, clears the
// per-turn counters, and returns the frozen wall-clock summary facts.
func (s *StreamState) Finalize(start, now time.Time) (time.Duration, int) {
	s.TurnTokens += s.StreamingEstimate
	s.SessionTokens += s.StreamingEstimate
	s.StreamingEstimate = 0
	elapsed := now.Sub(start)
	tokens := s.TurnTokens
	s.TurnTokens = 0
	return elapsed, tokens
}

// ResetSession clears token accounting for the `/clear` presentation intent.
func (s *StreamState) ResetSession() {
	s.TurnTokens = 0
	s.SessionTokens = 0
	s.StreamingEstimate = 0
}

// ErrMalformedToolCall is returned when transport labels an event as a tool
// call but supplies no resolved call. The transport currently guarantees a
// non-nil call; the boundary still fails safely instead of panicking.
var ErrMalformedToolCall = errors.New("model stream emitted a tool-call event without a tool call")

// StreamEffect contains presentation and logging facts produced by one stream
// transition. It intentionally excludes contexts, credentials, and tool
// arguments; resolved calls remain inside StreamState.Pending.
type StreamEffect struct {
	Content        string
	Reasoning      string
	ToolArgs       string
	RetryText      string
	RetryError     error
	RetryCleared   bool
	Flush          bool
	Done           bool
	Final          *chmctx.Message
	CountedTokens  int
	ReportedTokens int
	PromptTokens   int
	Elapsed        time.Duration
	ContextWindow  int
	Error          error
}

// Apply reduces one transport event into agent state and a presentation-safe
// effect. Callers must keep draining inactive/stale streams but must not call
// Apply for them.
func (s *StreamState) Apply(turn *TurnState, profile string, event llm.Event) StreamEffect {
	effect := StreamEffect{}
	if s.Retrying && event.Kind != llm.EventRetry {
		s.Retrying = false
		effect.RetryCleared = true
	}
	switch event.Kind {
	case llm.EventRetry:
		s.Retrying = true
		effect.RetryText = event.Content
		effect.RetryError = event.Err
	case llm.EventContent:
		if s.Phase == PhaseThinking {
			s.Phase = PhaseStreaming
		}
		s.StreamingEstimate += len(event.Content) / 4
		effect.Content = event.Content
	case llm.EventReasoning:
		s.StreamingEstimate += len(event.Content) / 4
		effect.Reasoning = event.Content
	case llm.EventToolArgs:
		if s.Phase == PhaseThinking {
			s.Phase = PhaseStreaming
		}
		s.StreamingEstimate += len(event.Content) / 4
		effect.ToolArgs = event.Content
	case llm.EventToolCall:
		if event.ToolCall == nil {
			effect.Error = ErrMalformedToolCall
			return effect
		}
		s.Pending = append(s.Pending, *event.ToolCall)
		effect.Flush = true
	case llm.EventDone:
		if event.ContextWindow > 0 {
			s.LiveContextSize[profile] = event.ContextWindow
		}
		delta := event.Tokens
		if delta == 0 {
			delta = s.StreamingEstimate
		}
		s.TurnTokens += delta
		s.SessionTokens += delta
		s.StreamingEstimate = 0
		s.Connected = true
		if event.Final != nil {
			turn.Append(*event.Final)
		}
		effect.Done = true
		effect.Final = event.Final
		effect.CountedTokens = delta
		effect.ReportedTokens = event.Tokens
		effect.PromptTokens = event.PromptTokens
		effect.Elapsed = event.Elapsed
		effect.ContextWindow = event.ContextWindow
		effect.Flush = true
	case llm.EventError:
		if _, ok := errors.AsType[provider.ErrUnreachable](event.Err); ok {
			s.Connected = false
		}
		effect.Error = event.Err
	}
	return effect
}
