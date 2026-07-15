package tui

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestQueueStoresPromptMidTurn: Enter while a turn is running stashes the typed
// prompt in the queue slot and clears the textarea, instead of the old silent
// drop. The turn keeps running (phase unchanged) and nothing is submitted yet.
func TestQueueStoresPromptMidTurn(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.ta.SetValue("run the tests next")

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	om := out.(Model)

	if om.queued == nil || om.queued.send != "run the tests next" {
		t.Fatalf("mid-turn Enter must queue the prompt, got %+v", om.queued)
	}
	if om.ta.Value() != "" {
		t.Fatalf("textarea must be cleared after queuing, got %q", om.ta.Value())
	}
	if om.controller.Snapshot().Phase != phaseThinking {
		t.Fatalf("queuing must not change the running phase, got %v", om.controller.Snapshot().Phase)
	}
	if cmd != nil {
		t.Fatal("queuing must not start a turn (nil Cmd)")
	}
}

// TestQueueEmptyMidTurnIsNoOp: Enter on a blank prompt mid-turn stays silent, so
// a reflexive Enter while watching the agent doesn't queue an empty slot.
func TestQueueEmptyMidTurnIsNoOp(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	if m.ta.Value() != "" {
		t.Fatal("precondition: textarea empty")
	}
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if out.(Model).queued != nil {
		t.Fatalf("empty mid-turn Enter must not queue, got %+v", out.(Model).queued)
	}
}

// TestQueueSecondEnterAppends: a second mid-turn Enter appends (newline-joined)
// to the existing slot, so a multi-part instruction builds up in one queued
// prompt that fires as a single turn.
func TestQueueSecondEnterAppends(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)

	m.ta.SetValue("run the tests")
	o1, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m1 := o1.(Model)

	m1.ta.SetValue("then commit if green")
	o2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := o2.(Model)

	want := "run the tests\nthen commit if green"
	if m2.queued == nil || m2.queued.send != want {
		t.Fatalf("second Enter must append newline-joined, got %+v", m2.queued)
	}
}

// TestQueueRefusesSlashMix: a slash command never newline-joins with a queued
// prompt, in either order. Joined slash-first, the whole slot would fire as ONE
// slash command whose Fields-split swallows the prose as bogus args (a queued
// /clear plus a follow-up instruction wipes the conversation AND silently drops
// the instruction); joined prose-first, the slash line ships to the LLM as
// prose. The refused draft must stay in the textarea, nothing lost.
func TestQueueRefusesSlashMix(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.ta.SetValue("/clear")
	o1, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m1 := o1.(Model)
	if m1.queued == nil || m1.queued.send != "/clear" {
		t.Fatalf("precondition: a slash prompt queues alone, got %+v", m1.queued)
	}

	m1.ta.SetValue("run the tests")
	o2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := o2.(Model)
	if m2.queued == nil || m2.queued.send != "/clear" {
		t.Fatalf("append onto a queued slash command must be refused, got %+v", m2.queued)
	}
	if got := m2.ta.Value(); got != "run the tests" {
		t.Fatalf("the refused draft must stay in the textarea, got %q", got)
	}

	// Reverse order: prose queued first, slash appended.
	mr := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&mr, phaseThinking)
	mr.ta.SetValue("run the tests")
	o3, _ := mr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := o3.(Model)
	m3.ta.SetValue("/clear")
	o4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := o4.(Model)
	if m4.queued == nil || m4.queued.send != "run the tests" {
		t.Fatalf("a slash append onto a queued prompt must be refused, got %+v", m4.queued)
	}
	if got := m4.ta.Value(); got != "/clear" {
		t.Fatalf("the refused slash draft must stay in the textarea, got %q", got)
	}
}

// TestQueueAutoSubmitsAfterTurn drives a full turn with a prompt queued mid-flight
// and asserts the queued prompt auto-fires a second request when the turn ends,
// then the slot is cleared. round==2 is the proof the follow-up actually ran.
func TestQueueAutoSubmitsAfterTurn(t *testing.T) {
	var round int
	var bodies []string
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		round++
		buf, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(buf))
		fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"reply"}}]}`)
		fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"completion_tokens":1}}`)
		fmt.Fprint(w, "data: [DONE]\n\n")
	}
	m := newTestModel(t, handler)
	mm, cmd := m.submit("first", "first", promptEntry{display: "first"})
	// Queue a follow-up while the first turn is in flight (before draining it).
	m2 := mm.(Model)
	m2.queued = &queuedPrompt{send: "second please", echo: "second please"}

	out, _ := drain(m2, cmd)
	final := out.(Model)

	if round != 2 {
		t.Fatalf("queued prompt must auto-fire a second request, got %d round(s)", round)
	}
	if final.queued != nil {
		t.Fatalf("slot must clear after auto-fire, got %+v", final.queued)
	}
	if !strings.Contains(strings.Join(bodies, ""), "second please") {
		t.Fatalf("the queued prompt never reached the server:\n%s", strings.Join(bodies, "\n"))
	}
	if final.controller.Snapshot().Phase.Active() {
		t.Fatalf("both turns done → idle, got phase %v", final.controller.Snapshot().Phase)
	}
}

// TestQueueRestoredOnCtrlC: a Ctrl+C abort never fires the queued prompt (the
// user took back control); it returns the text to the textarea as an editable
// draft and clears the slot, so there's no idle "queued" box that would
// orphan-fire after the next turn.
func TestQueueRestoredOnCtrlC(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.queued = &queuedPrompt{send: "later task", echo: "later task"}

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	om := out.(Model)

	if om.controller.Snapshot().Phase.Active() {
		t.Fatalf("Ctrl+C must return to idle, got %v", om.controller.Snapshot().Phase)
	}
	if om.queued != nil {
		t.Fatalf("Ctrl+C must clear the slot, got %+v", om.queued)
	}
	if om.ta.Value() != "later task" {
		t.Fatalf("Ctrl+C must restore the queued text to the textarea, got %q", om.ta.Value())
	}
}

// TestQueueRestoreKeepsExistingDraft: if the user was typing a new prompt when
// they Ctrl+C, that draft wins; the queued prompt is dropped rather than
// clobbering the in-progress text.
func TestQueueRestoreKeepsExistingDraft(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.queued = &queuedPrompt{send: "queued one", echo: "queued one"}
	m.ta.SetValue("a fresh draft")

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	om := out.(Model)

	if om.ta.Value() != "a fresh draft" {
		t.Fatalf("an in-progress draft must survive Ctrl+C, got %q", om.ta.Value())
	}
	if om.queued != nil {
		t.Fatalf("the queued slot must clear on abort, got %+v", om.queued)
	}
}

// TestUnqueueRestoresToTextarea: Backspace on an empty prompt while a turn runs
// pulls the queued prompt back into the textarea and clears the slot, a
// reversible way to edit or drop it.
func TestUnqueueRestoresToTextarea(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.queued = &queuedPrompt{send: "hello there", echo: "hello there"}

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	om := out.(Model)

	if om.ta.Value() != "hello there" {
		t.Fatalf("Backspace must restore the queued text, got %q", om.ta.Value())
	}
	if om.queued != nil {
		t.Fatalf("unqueue must clear the slot, got %+v", om.queued)
	}
}

// TestUnqueueOnlyWhenTextareaEmpty: with a draft already in the textarea,
// Backspace must delete a character as usual and leave the queued slot untouched,
// so editing a second prompt never clobbers the first.
func TestUnqueueOnlyWhenTextareaEmpty(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)
	m.queued = &queuedPrompt{send: "queued one", echo: "queued one"}
	m.ta.SetValue("draft")

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	om := out.(Model)

	if om.queued == nil || om.queued.send != "queued one" {
		t.Fatalf("Backspace on a non-empty prompt must not unqueue, got %+v", om.queued)
	}
	if om.ta.Value() != "draf" {
		t.Fatalf("Backspace on a non-empty prompt must delete a char, got %q", om.ta.Value())
	}
}

// TestClearWipesQueue: /clear resets the conversation, so a queued follow-up must
// go with it (it would target a conversation that no longer exists).
func TestClearWipesQueue(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	m.queued = &queuedPrompt{send: "stale", echo: "stale"}
	out, _ := m.runSlash("/clear")
	if out.(Model).queued != nil {
		t.Fatalf("/clear must wipe the queued slot, got %+v", out.(Model).queued)
	}
}

// TestQueuedPromptRendersInView: while something is queued the prompt area shows a
// labeled box with the echo text, so the user can see what will auto-submit;
// nothing renders when the slot is empty.
func TestQueuedPromptRendersInView(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)

	m.queued = &queuedPrompt{send: "run the tests", echo: "run the tests"}
	view := stripANSI(m.View())
	if !strings.Contains(view, "run the tests") {
		t.Fatalf("View must show the queued text:\n%s", view)
	}
	if !strings.Contains(strings.ToLower(view), "queued") {
		t.Fatalf("View must label the queued region:\n%s", view)
	}

	m.queued = nil
	if v := stripANSI(m.View()); strings.Contains(strings.ToLower(v), "queued") {
		t.Fatalf("an empty slot must render no queued box:\n%s", v)
	}
}

// TestQueueExpandsChipsOnFire: a chip-bearing paste queued mid-turn keeps its
// expanded content for the LLM (send) while the box echo stays collapsed, the
// same Value()/DisplayValue() split submit uses for a typed prompt.
func TestQueueExpandsChipsOnFire(t *testing.T) {
	m := newTestModel(t, func(http.ResponseWriter, *http.Request) {})
	setTestPhase(&m, phaseThinking)

	big := strings.Repeat("payload line\n", 6) // ≥5 lines → collapses into a chip
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(big), Paste: true})
	m1 := out.(Model)
	if !strings.Contains(m1.ta.DisplayValue(), "Pasted text") {
		t.Fatalf("precondition: a large paste must chip, display=%q", m1.ta.DisplayValue())
	}

	o2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := o2.(Model)
	if m2.queued == nil {
		t.Fatal("the paste must queue mid-turn")
	}
	if !strings.Contains(m2.queued.send, "payload line") {
		t.Fatalf("send must carry the EXPANDED paste, got %q", m2.queued.send)
	}
	if !strings.Contains(m2.queued.echo, "Pasted text") {
		t.Fatalf("echo must stay COLLAPSED, got %q", m2.queued.echo)
	}
}
