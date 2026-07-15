package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

type fakeChatClient struct {
	ctx      context.Context
	messages []chmctx.Message
	tools    []llm.Tool
	events   chan llm.Event
}

func (f *fakeChatClient) Chat(ctx context.Context, messages []chmctx.Message, tools []llm.Tool) <-chan llm.Event {
	f.ctx, f.messages, f.tools = ctx, messages, tools
	return f.events
}

func TestPhaseAndStreamState(t *testing.T) {
	cases := []struct {
		phase  Phase
		active bool
		label  string
	}{
		{PhaseIdle, false, ""},
		{PhaseThinking, true, "thinking"},
		{PhaseStreaming, true, "generating"},
		{PhaseRunning, true, "running"},
		{Phase(99), true, ""},
	}
	for _, tc := range cases {
		if tc.phase.Active() != tc.active || tc.phase.Label() != tc.label {
			t.Fatalf("phase %d", tc.phase)
		}
	}
	s := NewStreamState()
	if !s.Connected || s.LiveContextSize == nil || s.Phase.Active() {
		t.Fatal("new state")
	}
	channel := make(chan llm.Event, 1)
	stream := s.BeginStream(3, channel)
	if !s.Current(stream) || s.Current(nil) {
		t.Fatal("current stream")
	}
	channel <- llm.Event{Kind: llm.EventContent, Content: "x"}
	delivery := stream.Read()
	if delivery.turnID != 3 || delivery.roundID == 0 || delivery.Closed() || delivery.event.Content != "x" {
		t.Fatal("stream event")
	}
	close(channel)
	if delivery = stream.Read(); !delivery.Closed() {
		t.Fatal("stream close")
	}
	other := s.BeginStream(3, make(chan llm.Event))
	if s.Current(stream) || !s.Current(other) {
		t.Fatal("replaced stream")
	}
	s.EndStream()
	if s.Stream != nil || s.Current(other) {
		t.Fatal("end stream")
	}
	client := &fakeChatClient{events: make(chan llm.Event)}
	started := s.StartRound(&TurnState{ID: 8, Context: context.Background()}, client, []chmctx.Message{{Content: "m"}}, []llm.Tool{{Type: "function"}})
	if !s.Current(started) || client.ctx == nil || len(client.messages) != 1 || len(client.tools) != 1 {
		t.Fatal("start round")
	}
	s.Phase = PhaseRunning
	s.Retrying = true
	s.Pending = []chmctx.ToolCall{{Name: "x"}}
	s.ResetTurn()
	if s.Phase != PhaseIdle || s.Retrying || s.Pending != nil || !s.Connected || s.LiveContextSize == nil {
		t.Fatal("reset")
	}
	s.TurnTokens = 2
	s.StreamingEstimate = 3
	start := time.Unix(20, 0)
	elapsed, tokens := s.Finalize(start, start.Add(2*time.Second))
	if elapsed != 2*time.Second || tokens != 5 || s.TurnTokens != 0 || s.SessionTokens != 3 || s.StreamingEstimate != 0 {
		t.Fatal("finalize")
	}
	s.TurnTokens, s.SessionTokens, s.StreamingEstimate = 1, 2, 3
	s.ResetSession()
	if s.TurnTokens != 0 || s.SessionTokens != 0 || s.StreamingEstimate != 0 {
		t.Fatal("reset session")
	}
}

func TestApplyStreamEvents(t *testing.T) {
	turn := NewTurnState(nil)
	turn.Begin(nil, time.Now())
	s := NewStreamState()
	s.Phase = PhaseThinking

	retryErr := errors.New("retry")
	effect := s.Apply(&turn, "p", llm.Event{Kind: llm.EventRetry, Content: "wait", Err: retryErr})
	if !s.Retrying || effect.RetryText != "wait" || effect.RetryError != retryErr {
		t.Fatal("retry")
	}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventReasoning, Content: "1234"})
	if !effect.RetryCleared || effect.Reasoning != "1234" || s.StreamingEstimate != 1 || s.Phase != PhaseThinking {
		t.Fatal("reasoning")
	}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventToolArgs, Content: "12345678"})
	if effect.ToolArgs == "" || s.Phase != PhaseStreaming || s.StreamingEstimate != 3 {
		t.Fatal("tool args")
	}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventContent, Content: "abcd"})
	if effect.Content != "abcd" || s.StreamingEstimate != 4 {
		t.Fatal("content")
	}
	call := chmctx.ToolCall{Name: "x"}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventToolCall, ToolCall: &call})
	if !effect.Flush || len(s.Pending) != 1 {
		t.Fatal("tool call")
	}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventToolCall})
	if !errors.Is(effect.Error, ErrMalformedToolCall) {
		t.Fatal("malformed tool call")
	}
	final := chmctx.Message{Role: chmctx.RoleAssistant, Content: "done"}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventDone, Final: &final, ContextWindow: 4096})
	if !effect.Done || !effect.Flush || effect.CountedTokens != 4 || s.TurnTokens != 4 || s.SessionTokens != 4 || s.StreamingEstimate != 0 || s.LiveContextSize["p"] != 4096 || !s.Connected || len(turn.History) != 1 {
		t.Fatal("estimated done")
	}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventDone, Tokens: 7})
	if effect.CountedTokens != 7 || s.TurnTokens != 11 || s.SessionTokens != 11 {
		t.Fatal("counted done")
	}
	s.Connected = true
	unreachable := provider.ErrUnreachable{Err: errors.New("down")}
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventError, Err: unreachable})
	if effect.Error == nil || s.Connected {
		t.Fatal("unreachable")
	}
	s.Connected = true
	plain := errors.New("plain")
	effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventError, Err: plain})
	if effect.Error != plain || !s.Connected {
		t.Fatal("plain error")
	}
	if effect = s.Apply(&turn, "p", llm.Event{Kind: llm.EventKind(99)}); effect != (StreamEffect{}) {
		t.Fatal("unknown event")
	}
	s2 := NewStreamState()
	s2.Phase = PhaseThinking
	effect = s2.Apply(&turn, "p", llm.Event{Kind: llm.EventContent, Content: "x"})
	if s2.Phase != PhaseStreaming || effect.Content != "x" {
		t.Fatal("first content transition")
	}
}

func TestApplyDeliveryValidatesAndReducesOpaqueEvents(t *testing.T) {
	turn := NewTurnState(nil)
	turn.ID = 7
	state := NewStreamState()
	stream := state.BeginStream(turn.ID, make(chan llm.Event))

	stale := StreamDelivery{turnID: turn.ID + 1, roundID: stream.roundID, event: llm.Event{Kind: llm.EventContent, Content: "stale"}}
	if effect := state.ApplyDelivery(&turn, stream, "p", stale); effect.Accepted {
		t.Fatal("stale delivery accepted")
	}
	current := StreamDelivery{turnID: turn.ID, roundID: stream.roundID, event: llm.Event{Kind: llm.EventContent, Content: "current"}}
	effect := state.ApplyDelivery(&turn, stream, "p", current)
	if !effect.Accepted || effect.Closed || effect.Stream.Content != "current" {
		t.Fatalf("current delivery = %#v", effect)
	}
	closed := StreamDelivery{turnID: turn.ID, roundID: stream.roundID, closed: true}
	effect = state.ApplyDelivery(&turn, stream, "p", closed)
	if !effect.Accepted || !effect.Closed || effect.Stream != (StreamEffect{}) {
		t.Fatalf("closed delivery = %#v", effect)
	}
}
