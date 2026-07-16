// Package controller owns the application/frontend orchestration boundary.
package controller

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
	"github.com/DohmBoy64Bit/RecompHamr/internal/session"
)

const (
	reachabilityTimeout = 2 * time.Second
	probeTimeout        = 15 * time.Second
)

// Controller translates backend-neutral frontend intents into exactly-once
// application actions while retaining all backend capabilities privately.
type Controller struct {
	session   *session.Runtime
	agent     agent.Runtime
	system    func() string
	version   string
	nextID    uint64
	pending   map[uint64]struct{}
	turnStart time.Time
	now       func() time.Time
}

// NewController constructs the stable application/frontend boundary.
func NewController(sessionRuntime *session.Runtime, agentRuntime agent.Runtime, system func() string, version string) *Controller {
	return &Controller{
		session: sessionRuntime,
		agent:   agentRuntime,
		system:  system,
		version: version,
		pending: make(map[uint64]struct{}),
		now:     time.Now,
	}
}

// Snapshot returns a deep-copied non-secret application state.
func (c *Controller) Snapshot() frontend.Snapshot {
	sessionSnapshot := c.session.Snapshot()
	agentSnapshot := c.agent.Snapshot()
	profiles := make([]frontend.Profile, 0, len(sessionSnapshot.Profiles))
	for _, profile := range sessionSnapshot.Profiles {
		profiles = append(profiles, frontend.Profile{
			Name: profile.Name, Model: profile.Model, URL: profile.URL,
			ContextSize: profile.ContextSize, Active: profile.Active, Keyed: profile.Keyed,
		})
	}
	contextSize := sessionSnapshot.ContextSize
	if live, ok := c.agent.LiveContextSize(sessionSnapshot.Active); ok {
		contextSize = live
	}
	return frontend.Snapshot{
		Phase:             frontend.Phase(agentSnapshot.Phase),
		Connected:         agentSnapshot.Connected,
		Retrying:          agentSnapshot.Retrying,
		SessionTokens:     agentSnapshot.SessionTokens,
		StreamingEstimate: agentSnapshot.StreamingEstimate,
		Active:            sessionSnapshot.Active,
		ActiveURL:         sessionSnapshot.ActiveURL,
		ActiveModel:       sessionSnapshot.ActiveModel,
		ContextSize:       contextSize,
		ActiveKeyed:       sessionSnapshot.ActiveKeyed,
		Profiles:          profiles,
	}
}

// Bootstrap records the initial backend contract, loads prompt history, and
// captures the accepted keyed-probe or keyless-reachability startup work.
func (c *Controller) Bootstrap() frontend.Transition {
	snapshot := c.Snapshot()
	contextSize := c.activeContextSize(snapshot)
	c.agent.ObserveSession(c.version, snapshot.Active, snapshot.ActiveModel, snapshot.ActiveURL, contextSize, chmctx.Tokens(c.system()))
	events := []frontend.Event{{Kind: frontend.EventHistory, Values: c.session.LoadHistory()}}
	return frontend.Transition{Snapshot: c.Snapshot(), Events: events, Work: c.connectivityWork(snapshot, true)}
}

// Dispatch applies one valid intent or opaque completion. Invalid, foreign,
// duplicate, and already-consumed completions are safe no-ops.
func (c *Controller) Dispatch(intent frontend.Intent) frontend.Transition {
	transition := frontend.Transition{}
	switch intent.Kind() {
	case frontend.IntentObserveSlash:
		c.agent.ObserveSlash(intent.Text())
	case frontend.IntentAppendHistory:
		_ = c.session.AppendHistory(intent.Text())
	case frontend.IntentReload:
		if _, _, err := c.session.Reload(); err != nil {
			transition.Events = []frontend.Event{{Kind: frontend.EventWarning, Text: err.Error()}}
		}
	case frontend.IntentActivate:
		if snapshot, err := c.session.Activate(intent.ProfileName()); err != nil {
			transition.Events = []frontend.Event{{Kind: frontend.EventWarning, Text: err.Error()}}
		} else {
			transition.Events = []frontend.Event{{Kind: frontend.EventProfileActivated, Profile: snapshot.Active, Model: snapshot.ActiveModel, URL: snapshot.ActiveURL}}
			transition.Work = c.connectivityWork(c.Snapshot(), false)
		}
	case frontend.IntentClearConversation:
		c.agent.ResetConversation()
		c.turnStart = time.Time{}
		_ = c.session.ClearHistory()
	case frontend.IntentComplete:
		c.applyCompletion(intent.WorkCompletion(), &transition)
	case frontend.IntentSubmitGoal:
		if !c.agent.Active() {
			c.agent.ObserveUser(intent.Text())
			c.agent.BeginTurn(intent.Time())
			c.agent.AppendUser(intent.Text())
			c.turnStart = intent.Time()
			transition.Events = []frontend.Event{{Kind: frontend.EventTurnStarted, At: intent.Time()}}
			transition.Work = c.startRound()
		}
	case frontend.IntentCancel:
		if c.agent.Active() {
			c.agent.ObserveCancel()
			transition.Events = c.finishTurn(false, true, false, "✗ cancelled", intent.Time())
		}
	case frontend.IntentInvalid:
		// The zero-value malformed intent is a permanent no-op.
	}
	transition.Snapshot = c.Snapshot()
	return transition
}

func (c *Controller) applyCompletion(value frontend.Completion, transition *frontend.Transition) {
	completed, ok := value.(completion)
	if !ok {
		return
	}
	if _, pending := c.pending[completed.id]; !pending {
		return
	}
	delete(c.pending, completed.id)
	switch result := completed.result.(type) {
	case session.ReachabilityResult:
		if result.URL != c.Snapshot().ActiveURL {
			return
		}
		ok := result.Err == nil
		c.agent.SetConnected(ok)
		transition.Events = []frontend.Event{{Kind: frontend.EventReachability, URL: result.URL, OK: ok}}
	case probeCompletion:
		active := result.result.Profile == c.Snapshot().Active
		if active {
			c.agent.SetConnected(result.result.Err == nil)
		}
		if result.result.Err == nil && result.result.ContextWindow > 0 {
			if _, exists := c.Snapshot().Profile(result.result.Profile); exists {
				c.agent.SetLiveContextSize(result.result.Profile, result.result.ContextWindow)
			}
		}
		if !result.silent {
			event := frontend.Event{Kind: frontend.EventProbe, Profile: result.result.Profile, ContextWindow: result.result.ContextWindow, OK: result.result.Err == nil}
			if result.result.Err != nil {
				event.Text = agent.ProbeErrorMessage(result.result.Err)
			}
			transition.Events = []frontend.Event{event}
		}
	case modelCompletion:
		c.applyModelCompletion(result, transition)
	case toolCompletion:
		c.applyToolCompletion(result, transition)
	}
}

type modelCompletion struct {
	stream   *agent.Stream
	delivery agent.StreamDelivery
}

type toolCompletion struct{ delivery agent.ToolDelivery }

func (c *Controller) startRound() frontend.Work {
	snapshot := c.Snapshot()
	stream, _ := c.agent.StartRound(c.system(), snapshot.ActiveModel, c.activeContextSize(snapshot))
	return c.readStream(stream)
}

func (c *Controller) readStream(stream *agent.Stream) frontend.Work {
	return c.capture(func() any { return modelCompletion{stream: stream, delivery: stream.Read()} })
}

func (c *Controller) applyModelCompletion(result modelCompletion, transition *frontend.Transition) {
	snapshot := c.Snapshot()
	delivery := c.agent.ApplyDelivery(result.stream, snapshot.Active, c.activeContextSize(snapshot), result.delivery)
	if !delivery.Accepted {
		if !result.delivery.Closed() {
			transition.Work = c.readStream(result.stream)
		}
		return
	}
	if delivery.Closed {
		c.applyClose(transition)
		return
	}
	effect := delivery.Stream
	if effect.RetryCleared {
		transition.Events = append(transition.Events, frontend.Event{Kind: frontend.EventStatus})
	}
	if effect.RetryText != "" {
		transition.Events = append(transition.Events, frontend.Event{Kind: frontend.EventStatus, Text: effect.RetryText})
	}
	if effect.Content != "" {
		transition.Events = append(transition.Events, frontend.Event{Kind: frontend.EventContent, Text: effect.Content})
	}
	if effect.Flush || effect.Done {
		transition.Events = append(transition.Events, frontend.Event{Kind: frontend.EventFlush})
	}
	if effect.Error != nil {
		message := agent.StreamErrorMessage(effect.Error, snapshot.Active, snapshot.ActiveURL)
		transition.Events = append(transition.Events, c.finishTurn(false, false, false, message, c.now())...)
		transition.Work = c.readStream(result.stream)
		return
	}
	transition.Work = c.readStream(result.stream)
}

func (c *Controller) applyClose(transition *frontend.Transition) {
	if !c.agent.Active() {
		return
	}
	decision := c.agent.DecideClose()
	switch decision.Action {
	case agent.CloseRunTool:
		c.nextTool(transition)
	case agent.CloseRestartModel:
		transition.Work = c.startRound()
	case agent.CloseFinishStopped:
		message := "⚠ a tool call leaked into the reply as text instead of running - your model server isn't parsing tool calls. Enable its OpenAI tool-call parser server-side (e.g. vLLM `--tool-call-parser`, llama.cpp `--jinja`)."
		if decision.Reason == agent.CloseEmptyStall {
			message = "⚠ the model ended its turn with no reply and no tool call - it stalled, or your server dropped the call. If thinking is on, its reasoning parser may be swallowing calls - enable one (e.g. vLLM `--reasoning-parser`) or disable thinking for tool turns."
		}
		transition.Events = append(transition.Events, c.finishTurn(false, false, true, message, c.now())...)
	default:
		transition.Events = append(transition.Events, c.finishTurn(true, false, true, "", c.now())...)
	}
}

func (c *Controller) nextTool(transition *frontend.Transition) {
	work, ok := c.agent.NextTool()
	if !ok {
		return
	}
	transition.Events = append(transition.Events, frontend.Event{Kind: frontend.EventToolStatus, Text: work.Status()})
	transition.Work = c.capture(func() any { return toolCompletion{delivery: work.Run()} })
}

func (c *Controller) applyToolCompletion(result toolCompletion, transition *frontend.Transition) {
	effect := c.agent.ApplyToolResult(result.delivery)
	if !effect.Accepted {
		return
	}
	if effect.ContinueTools {
		c.nextTool(transition)
		return
	}
	transition.Work = c.startRound()
}

func (c *Controller) finishTurn(ok, cancelled, natural bool, message string, now time.Time) []frontend.Event {
	wall, tokens := time.Duration(0), 0
	if !c.turnStart.IsZero() {
		wall, tokens = c.agent.FinalizeTurn(c.turnStart, now)
	}
	average := humanRate(tokens, wall)
	if average != "" {
		average = " · " + average + " avg"
	}
	c.agent.ObserveTurnEnd(humanTokens(tokens), humanTokens(c.agent.Snapshot().SessionTokens), wall, average)
	if cancelled {
		c.agent.CancelTurn()
	} else {
		c.agent.EndTurn()
	}
	c.turnStart = time.Time{}
	return []frontend.Event{{Kind: frontend.EventTurnFinished, Text: message, Elapsed: wall, Tokens: tokens, OK: ok, Cancelled: cancelled, Natural: natural}}
}

func humanTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d tok", n)
	}
	return fmt.Sprintf("%.1fk tok", float64(n)/1000)
}

func humanRate(tokens int, elapsed time.Duration) string {
	seconds := elapsed.Seconds()
	if tokens <= 0 || seconds <= 0 {
		return ""
	}
	rate := float64(tokens) / seconds
	if rate < 10 {
		return fmt.Sprintf("%.1f tok/s", rate)
	}
	return fmt.Sprintf("%d tok/s", int(math.Round(rate)))
}

func (c *Controller) activeContextSize(snapshot frontend.Snapshot) int {
	if value, ok := c.agent.LiveContextSize(snapshot.Active); ok {
		return value
	}
	if snapshot.ContextSize > 0 {
		return snapshot.ContextSize
	}
	return 16177
}

type completion struct {
	id     uint64
	result any
}

type work struct {
	id  uint64
	run func() any
}

func (w work) Run() frontend.Completion { return completion{id: w.id, result: w.run()} }

func (c *Controller) capture(run func() any) frontend.Work {
	c.nextID++
	id := c.nextID
	c.pending[id] = struct{}{}
	return work{id: id, run: run}
}

func (c *Controller) connectivityWork(snapshot frontend.Snapshot, silent bool) frontend.Work {
	if snapshot.ActiveKeyed {
		probe := c.session.Probe(snapshot.Active)
		return c.capture(func() any {
			ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
			defer cancel()
			return probeCompletion{result: probe.Run(ctx), silent: silent}
		})
	}
	reachability := c.session.Reachability()
	return c.capture(func() any {
		ctx, cancel := context.WithTimeout(context.Background(), reachabilityTimeout)
		defer cancel()
		return reachability.Run(ctx)
	})
}

type probeCompletion struct {
	result session.ProbeResult
	silent bool
}
