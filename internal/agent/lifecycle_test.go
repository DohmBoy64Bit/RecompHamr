package agent

import (
	"context"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

func TestTurnStateLifecycle(t *testing.T) {
	s := NewTurnState([]chmctx.Message{{Role: chmctx.RoleSystem, Content: "old"}})
	if s.Active() || len(s.History) != 1 {
		t.Fatal("new state")
	}
	now := time.Unix(10, 0)
	first := s.Begin(nil, now)
	firstContext := s.context
	if first == 0 || !s.Active() || s.StartedAt != now {
		t.Fatal("first begin")
	}
	s.Append(chmctx.Message{Role: chmctx.RoleUser, Content: "hi"})
	second := s.Begin(context.Background(), now.Add(time.Second))
	if second == first || firstContext.Err() != context.Canceled || len(s.History) != 2 {
		t.Fatal("replacement begin")
	}
	secondContext := s.context
	s.End()
	if s.Active() || s.context != nil || !s.StartedAt.IsZero() || secondContext.Err() != context.Canceled || s.ID != second {
		t.Fatal("end")
	}
	s.End()
	s.Reset()
	if len(s.History) != 0 {
		t.Fatal("reset")
	}
}
