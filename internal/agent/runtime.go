package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
)

// Runtime groups the mutable agent-turn components and their injected model
// and tool boundaries. The application composition root constructs it; a
// frontend adapter carries it while Stage C narrows the remaining contract.
type Runtime struct {
	Turn      *TurnState
	Stream    *StreamState
	Loop      *LoopState
	Client    ChatClient
	Executor  ToolExecutor
	observer  Observer
	reasoning *strings.Builder
}

// Observer receives private, causal agent-runtime records. Implementations
// own storage and redaction; orchestration owns when each record is emitted.
type Observer interface {
	Enabled() bool
	Writef(category, format string, args ...any)
	WriteMessage(category string, message chmctx.Message)
}

// Snapshot is an immutable, presentation-safe view of current runtime facts.
// It excludes history, contexts, cancellation, streams, tool calls, arguments,
// credentials, reasoning, and observer state.
type Snapshot struct {
	Phase             Phase
	Connected         bool
	Retrying          bool
	SessionTokens     int
	StreamingEstimate int
}

// NewRuntime constructs an idle agent runtime from application-owned
// dependencies.
func NewRuntime(client ChatClient, executor ToolExecutor) Runtime {
	return Runtime{
		Turn:      newTurnState(nil),
		Stream:    newStreamState(),
		Loop:      &LoopState{},
		Client:    client,
		Executor:  executor,
		reasoning: &strings.Builder{},
	}
}

// WithObserver returns runtime with an injected private event observer.
func (r Runtime) WithObserver(observer Observer) Runtime {
	r.observer = observer
	return r
}

// Snapshot returns current presentation facts by value.
func (r Runtime) Snapshot() Snapshot {
	return Snapshot{
		Phase:             r.Stream.Phase,
		Connected:         r.Stream.Connected,
		Retrying:          r.Stream.Retrying,
		SessionTokens:     r.Stream.SessionTokens,
		StreamingEstimate: r.Stream.StreamingEstimate,
	}
}

// Active reports whether an agent turn is active.
func (r Runtime) Active() bool { return r.Turn.Active() }

// LiveContextSize returns the positive runtime context hint for profile.
func (r Runtime) LiveContextSize(profile string) (int, bool) {
	value, ok := r.Stream.LiveContextSize[profile]
	return value, ok && value > 0
}

// SetConnected updates the presentation-safe backend connectivity fact.
func (r Runtime) SetConnected(connected bool) { r.Stream.Connected = connected }

// SetLiveContextSize records a positive probe result for a known profile.
func (r Runtime) SetLiveContextSize(profile string, contextSize int) {
	if contextSize > 0 {
		r.Stream.LiveContextSize[profile] = contextSize
	}
}

// CurrentStream returns the opaque reader currently owned by the runtime.
func (r Runtime) CurrentStream() *Stream { return r.Stream.Stream }

func (r Runtime) observef(category, format string, args ...any) {
	if r.observer != nil {
		r.observer.Writef(category, format, args...)
	}
}

func (r Runtime) observeMessage(category string, message chmctx.Message) {
	if r.observer != nil {
		r.observer.WriteMessage(category, message)
	}
}

// ObserveUser records a submitted user prompt at the agent boundary.
func (r Runtime) ObserveUser(content string) { r.observef("user", "%s", content) }

// ObserveCancel records the accepted user cancellation intent.
func (r Runtime) ObserveCancel() { r.observef("cancel", "user cancelled the turn (Ctrl+C)") }

// ObserveSlash records a slash-command submission without treating it as a
// model-facing user goal.
func (r Runtime) ObserveSlash(content string) { r.observef("user_slash", "%s", content) }

// ObserveSession records the active backend and context/tool contract without
// copying the system prompt or credentials into the private log.
func (r Runtime) ObserveSession(version, profile, model, url string, contextSize, systemTokens int) {
	r.observef("session", "recomphamr %s · profile=%s · model=%s @ %s\ncontext_size=%d tokens · system_prompt≈%d tokens · tools=[%s]",
		version, profile, model, url, contextSize, systemTokens, strings.Join(ToolNames(), ", "))
}

// BeginTurn installs the agent-owned cancellation root and initializes the
// presentation-neutral active phase.
func (r Runtime) BeginTurn(now time.Time) TurnID {
	id := r.Turn.Begin(context.Background(), now)
	r.Stream.Phase = PhaseThinking
	return id
}

// AppendUser starts a new goal and appends its model-facing user message.
func (r Runtime) AppendUser(content string) {
	r.Loop.ResetUserGoal()
	r.Turn.Append(chmctx.Message{Role: chmctx.RoleUser, Content: content})
}

// EndTurn releases the cancellation root and resets all per-turn stream and
// policy state. It reports whether a retry hint was active for presentation.
func (r Runtime) EndTurn() bool {
	r.Turn.End()
	wasRetrying := r.Stream.Retrying
	r.Stream.ResetTurn()
	r.Loop.ResetTurn()
	return wasRetrying
}

// CancelTurn ends an interrupted turn. When a local tool was running, it also
// removes that incomplete model-facing goal so a later request cannot reissue
// the cancelled side effect. Visible transcript and prompt recall are owned by
// presentation and remain unchanged.
func (r Runtime) CancelTurn() bool {
	if r.Stream.Phase == PhaseRunning {
		r.Turn.discardCurrentGoal()
	}
	return r.EndTurn()
}

// ResetConversation clears agent history, accounting, and failure state.
func (r Runtime) ResetConversation() {
	r.Turn.Reset()
	r.Stream.ResetSession()
	r.Loop.ResetUserGoal()
}

// FinalizeTurn freezes interrupted estimates into turn and session totals.
func (r Runtime) FinalizeTurn(start, now time.Time) (time.Duration, int) {
	return r.Stream.Finalize(start, now)
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
func (r Runtime) StartRound(system, model string, contextSize int) (*Stream, RequestSummary) {
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
	note := ""
	if summary.Truncated > 0 {
		note = fmt.Sprintf(" · %d tool output(s) truncated", summary.Truncated)
	}
	r.observef("request", "model=%s · ctx=%d (history budget=%d) · history=%d msgs → packed=%d msgs (~%d tokens) · dropped=%d oldest%s",
		model, summary.ContextSize, summary.Budget, summary.History, summary.Packed, summary.Tokens, summary.Dropped, note)
	return r.Stream.StartRound(r.Turn, r.Client, messages, Tools()), summary
}

// ApplyDelivery validates and reduces one model delivery while emitting the
// retained causal runtime records exactly once.
func (r Runtime) ApplyDelivery(stream *Stream, profile string, contextSize int, delivery StreamDelivery) DeliveryEffect {
	result := r.Stream.ApplyDelivery(r.Turn, stream, profile, delivery)
	if !result.Accepted || result.Closed {
		return result
	}
	r.observeStreamEffect(contextSize, result.Stream)
	return result
}

// ApplyEvent reduces a direct event through the same observer path. It exists
// for deterministic adapter tests; production consumes opaque deliveries.
func (r Runtime) ApplyEvent(profile string, contextSize int, event llm.Event) StreamEffect {
	effect := r.Stream.Apply(r.Turn, profile, event)
	r.observeStreamEffect(contextSize, effect)
	return effect
}

func (r Runtime) observeStreamEffect(contextSize int, effect StreamEffect) {
	if effect.RetryText != "" {
		r.observef("retry", "%s (%v)", effect.RetryText, effect.RetryError)
	}
	if effect.Reasoning != "" && r.observer != nil && r.observer.Enabled() {
		r.reasoning.WriteString(effect.Reasoning)
	}
	if effect.Done {
		if reasoning := r.reasoning.String(); reasoning != "" {
			r.observef("reasoning", "%s", reasoning)
		}
		r.reasoning.Reset()
		if effect.Final != nil {
			r.observeMessage("assistant", *effect.Final)
		}
		r.observef("round_done", "tokens=%d (counted=%d) · prompt_tokens=%d · elapsed=%s · ctx_window=%d",
			effect.ReportedTokens, effect.CountedTokens, effect.PromptTokens, effect.Elapsed.Round(time.Millisecond), effect.ContextWindow)
		if effect.PromptTokens > 0 && effect.PromptTokens >= contextSize-contextSize/20 {
			r.observef("ctx_pressure", "prompt_tokens=%d at >=95%% of ctx=%d; real prompt has outgrown the packer's estimate, next request risks silent server-side truncation", effect.PromptTokens, contextSize)
		}
	}
	if effect.Error != nil {
		r.observef("error", "%v", effect.Error)
	}
}

// NextTool dispatches the next opaque tool action and records its bounded
// presentation-safe status.
func (r Runtime) NextTool() (*ToolWork, bool) {
	work, ok := r.Loop.NextTool(r.Turn, r.Stream, r.Executor)
	if ok {
		r.observef("tool_start", "%s", work.Status())
	}
	return work, ok
}

// ApplyToolResult reduces one tool delivery and records accepted outcome and
// nudge facts exactly once.
func (r Runtime) ApplyToolResult(delivery ToolDelivery) ToolResultEffect {
	effect := r.Loop.ApplyToolResult(r.Turn, r.Stream, delivery)
	if !effect.Accepted {
		return effect
	}
	r.observeMessage("tool_result", effect.Message)
	if effect.Failed {
		r.observef("tool_outcome", "tool=%s FAILED · same-target streak=%d/%d · key=%s", effect.Message.ToolName, effect.FailStreak, MaxToolFailStreak, effect.FailKey)
	}
	if effect.FailureNudged {
		r.observef("nudge", "repeated-failure nudge injected after %d same-target failures (key=%s)", effect.FailureCount, effect.FailureKey)
	}
	if effect.RunawayNudged {
		r.observef("nudge", "runaway-iteration nudge injected at %d tool calls this turn", effect.ToolRounds)
	}
	return effect
}

// DecideClose applies stream-close policy and records its policy branch.
func (r Runtime) DecideClose() CloseDecision {
	decision := r.Loop.DecideClose(r.Turn, r.Stream)
	switch decision.Reason {
	case CloseEmptyNudge:
		r.observef("nudge", "empty-reply nudge injected (turn ended with no content and no tool call)")
	case CloseVerifyNudge:
		r.observef("nudge", "finish re-grounding nudge injected at %d tool calls this turn", decision.ToolRounds)
	case CloseEmptyStall:
		r.observef("leak", "turn ended with an empty assistant message after a re-prompt (model stalled or the call was swallowed server-side)")
	case CloseToolLeak:
		r.observef("leak", "turn ended with tool-call text leaked into the reply (server-side parser misconfigured)")
	}
	return decision
}

// ObserveTurnEnd records the presentation-formatted frozen turn summary.
func (r Runtime) ObserveTurnEnd(tokens, session string, wall time.Duration, average string) {
	r.observef("turn_end", "%s · %s wall%s · session_total=%s", tokens, wall.Round(time.Millisecond), average, session)
}

func newTurnState(history []chmctx.Message) *TurnState {
	state := NewTurnState(history)
	return &state
}

func newStreamState() *StreamState {
	state := NewStreamState()
	return &state
}
