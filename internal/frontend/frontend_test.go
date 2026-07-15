package frontend

import (
	"testing"
	"time"
)

func TestSnapshotProfilePhaseAndCommands(t *testing.T) {
	snapshot := Snapshot{Profiles: []Profile{{Name: "a", Active: true}}}
	if profile, ok := snapshot.Profile("a"); !ok || !profile.Active {
		t.Fatalf("profile = %#v %v", profile, ok)
	}
	if _, ok := snapshot.Profile("missing"); ok {
		t.Fatal("missing profile found")
	}
	if PhaseIdle.Active() || !PhaseThinking.Active() || !PhaseStreaming.Active() || !PhaseRunning.Active() {
		t.Fatal("phase activity contract")
	}
	for phase, want := range map[Phase]string{PhaseIdle: "", PhaseThinking: "thinking", PhaseStreaming: "generating", PhaseRunning: "running", Phase(99): ""} {
		if got := phase.Label(); got != want {
			t.Fatalf("phase %d = %q", phase, got)
		}
	}
	commands := Commands()
	commands[0].Name = "changed"
	if got := Commands(); len(got) != 2 || got[0].Name != "/clear" || got[1].Name != "/models" {
		t.Fatalf("commands = %#v", got)
	}
}

func TestIntentConstructors(t *testing.T) {
	now := time.Now()
	completion := struct{ value string }{"done"}
	cases := []struct {
		intent  Intent
		kind    IntentKind
		text    string
		profile string
	}{
		{ObserveSlash("/models"), IntentObserveSlash, "/models", ""},
		{AppendHistory("prompt"), IntentAppendHistory, "prompt", ""},
		{Reload(), IntentReload, "", ""},
		{Activate("other"), IntentActivate, "", "other"},
		{ClearConversation(), IntentClearConversation, "", ""},
		{Cancel(now), IntentCancel, "", ""},
		{SubmitGoal("goal", now), IntentSubmitGoal, "goal", ""},
		{Complete(completion), IntentComplete, "", ""},
	}
	for _, tc := range cases {
		if tc.intent.kind != tc.kind || tc.intent.text != tc.text || tc.intent.profile != tc.profile {
			t.Fatalf("intent = %#v", tc.intent)
		}
	}
	if !cases[4].intent.now.IsZero() || cases[5].intent.now != now || cases[6].intent.now != now {
		t.Fatal("intent time contract")
	}
	if got := cases[7].intent.completion; got != completion {
		t.Fatalf("completion = %#v", got)
	}
	if cases[0].intent.Kind() != IntentObserveSlash || cases[0].intent.Text() != "/models" || cases[3].intent.ProfileName() != "other" || cases[5].intent.Time() != now || cases[7].intent.WorkCompletion() != completion {
		t.Fatal("intent accessors")
	}
}
