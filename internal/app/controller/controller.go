// Package controller owns the application/frontend orchestration boundary.
package controller

import (
	"context"
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
	session *session.Runtime
	agent   agent.Runtime
	system  string
	version string
	nextID  uint64
	pending map[uint64]struct{}
}

// NewController constructs the stable application/frontend boundary.
func NewController(sessionRuntime *session.Runtime, agentRuntime agent.Runtime, system, version string) *Controller {
	return &Controller{
		session: sessionRuntime,
		agent:   agentRuntime,
		system:  system,
		version: version,
		pending: make(map[uint64]struct{}),
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
	c.agent.ObserveSession(c.version, snapshot.Active, snapshot.ActiveModel, snapshot.ActiveURL, contextSize, chmctx.Tokens(c.system))
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
		_ = c.session.ClearHistory()
	case frontend.IntentComplete:
		c.applyCompletion(intent.WorkCompletion(), &transition)
	case frontend.IntentCancel, frontend.IntentSubmitGoal, frontend.IntentInvalid:
		// Agent-turn intents migrate in Checkpoint 4D; until then the accepted
		// TUI path remains authoritative. Invalid intents are permanent no-ops.
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
	}
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
