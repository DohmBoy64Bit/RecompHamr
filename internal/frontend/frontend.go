// Package frontend defines the backend-neutral contract between application
// orchestration and concrete presentation adapters.
package frontend

import "time"

// Phase is the display-safe lifecycle state of the current agent turn.
type Phase uint8

const (
	// PhaseIdle means no turn is active.
	PhaseIdle Phase = iota
	// PhaseThinking means a model round is pending its first visible output.
	PhaseThinking
	// PhaseStreaming means model output is arriving.
	PhaseStreaming
	// PhaseRunning means an agent-selected tool is executing.
	PhaseRunning
)

// Active reports whether the phase belongs to a live turn.
func (p Phase) Active() bool { return p != PhaseIdle }

// Profile is a non-secret immutable model-profile fact for presentation.
type Profile struct {
	Name        string
	Model       string
	URL         string
	ContextSize int
	Active      bool
	Keyed       bool
}

// Snapshot is an immutable, display-safe application state. Profiles returns
// a fresh slice on every controller snapshot so presentation cannot mutate
// application state through shared backing storage.
type Snapshot struct {
	Phase             Phase
	Connected         bool
	Retrying          bool
	SessionTokens     int
	StreamingEstimate int
	Active            string
	ActiveURL         string
	ActiveModel       string
	ContextSize       int
	ActiveKeyed       bool
	Profiles          []Profile
}

// Profile finds a named profile in the immutable snapshot.
func (s Snapshot) Profile(name string) (Profile, bool) {
	for _, profile := range s.Profiles {
		if profile.Name == name {
			return profile, true
		}
	}
	return Profile{}, false
}

// IntentKind identifies one application action requested by presentation.
type IntentKind uint8

const (
	// IntentInvalid is the zero-value malformed intent and is always a no-op.
	IntentInvalid IntentKind = iota
	// IntentObserveSlash records a slash submission.
	IntentObserveSlash
	// IntentAppendHistory persists prompt recall best-effort.
	IntentAppendHistory
	// IntentReload reloads configuration.
	IntentReload
	// IntentActivate activates a named profile.
	IntentActivate
	// IntentClearConversation resets conversation and recall.
	IntentClearConversation
	// IntentCancel cancels an active turn.
	IntentCancel
	// IntentSubmitGoal starts a user goal.
	IntentSubmitGoal
	// IntentComplete returns opaque asynchronous work.
	IntentComplete
)

// Completion is an opaque asynchronous result. Presentation may carry it back
// to Controller.Dispatch but must not inspect or manufacture backend payloads.
type Completion interface{}

// Intent is a constructor-backed presentation request. Its fields are private
// so only the contract's valid variants can be emitted by production adapters.
type Intent struct {
	kind       IntentKind
	text       string
	profile    string
	now        time.Time
	completion Completion
}

// ObserveSlash records one accepted slash-command submission.
func ObserveSlash(text string) Intent { return Intent{kind: IntentObserveSlash, text: text} }

// AppendHistory requests best-effort durable prompt recall.
func AppendHistory(text string) Intent { return Intent{kind: IntentAppendHistory, text: text} }

// Reload requests configuration reload using the retained on-disk contract.
func Reload() Intent { return Intent{kind: IntentReload} }

// Activate requests atomic activation of a named model profile.
func Activate(profile string) Intent { return Intent{kind: IntentActivate, profile: profile} }

// ClearConversation requests agent and prompt-history reset.
func ClearConversation() Intent { return Intent{kind: IntentClearConversation} }

// Cancel requests cancellation of the active turn at now.
func Cancel(now time.Time) Intent { return Intent{kind: IntentCancel, now: now} }

// SubmitGoal requests one new model-facing user goal at now.
func SubmitGoal(text string, now time.Time) Intent {
	return Intent{kind: IntentSubmitGoal, text: text, now: now}
}

// Complete returns one opaque asynchronous result to the controller.
func Complete(completion Completion) Intent {
	return Intent{kind: IntentComplete, completion: completion}
}

// Kind returns the intent discriminator.
func (i Intent) Kind() IntentKind { return i.kind }

// Text returns the intent's user text, when applicable.
func (i Intent) Text() string { return i.text }

// ProfileName returns the requested profile, when applicable.
func (i Intent) ProfileName() string { return i.profile }

// Time returns the intent timestamp, when applicable.
func (i Intent) Time() time.Time { return i.now }

// WorkCompletion returns the opaque completion, when applicable.
func (i Intent) WorkCompletion() Completion { return i.completion }

// EventKind identifies one ordered display-safe application event.
type EventKind uint8

const (
	// EventHistory seeds prompt recall with Values in oldest-first order.
	EventHistory EventKind = iota + 1
	// EventWarning carries a user-visible warning in Text.
	EventWarning
	// EventProfileActivated reports a persisted profile selection.
	EventProfileActivated
	// EventReachability reports a captured endpoint check.
	EventReachability
	// EventProbe reports an authenticated profile probe.
	EventProbe
	// EventContent carries streamed visible assistant content.
	EventContent
	// EventFlush requests presentation to commit its live content block.
	EventFlush
	// EventStatus replaces transient presentation status text.
	EventStatus
	// EventToolStatus carries a bounded display-safe tool action label.
	EventToolStatus
	// EventTurnFinished carries frozen duration/token outcome facts.
	EventTurnFinished
)

// Event is an ordered presentation fact. It deliberately contains no
// credential, model history, reasoning, raw transport, tool payload, process
// handle, cancellation capability, observer, or log content.
type Event struct {
	Kind          EventKind
	Text          string
	Values        []string
	Profile       string
	Model         string
	URL           string
	ContextWindow int
	Silent        bool
	OK            bool
	Elapsed       time.Duration
	Tokens        int
}

// Work is captured asynchronous application work. Run owns its timeout and
// returns an opaque completion that presentation can only dispatch back.
type Work interface {
	Run() Completion
}

// Transition is the ordered result of one controller operation.
type Transition struct {
	Snapshot Snapshot
	Events   []Event
	Work     Work
}

// Controller is the complete presentation-facing application boundary.
type Controller interface {
	Bootstrap() Transition
	Dispatch(Intent) Transition
	Snapshot() Snapshot
}

// Command describes one canonical slash-command help row.
type Command struct {
	Name        string
	Description string
}

// Commands returns the canonical command order and text shared by CLI help
// and concrete presentation adapters.
func Commands() []Command {
	return []Command{
		{Name: "/clear", Description: "reset the conversation"},
		{Name: "/models", Description: "list · <name> set (Tab cycles in the popover)"},
	}
}
