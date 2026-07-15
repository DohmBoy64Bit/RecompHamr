package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
)

func TestPromptInternalBoundaryContracts(t *testing.T) {
	p := newPromptInput()
	p.ta.SetValue("abc")
	p.deleteSpan(chipSpan{id: 1, start: 0, end: 99})
	if p.ta.Value() != "abc" {
		t.Fatal("invalid delete changed value")
	}

	label := chipLabel(2)
	p.ta.SetValue(label)
	p.store = map[int]chipContent{1: {content: "one\ntwo", lines: 2}}
	p.spans = []chipSpan{{id: 99}, {id: 1}}
	p.reconcile()
	if len(p.spans) != 1 {
		t.Fatalf("missing store not dropped: %#v", p.spans)
	}
	p.store[2] = chipContent{content: "x\ny", lines: 2}
	p.spans = append(p.spans, chipSpan{id: 2})
	p.reconcile()
	if len(p.spans) != 0 {
		t.Fatal("ambiguous identical labels retained")
	}
	p.store = map[int]chipContent{3: {content: "a\nb\nc", lines: 3}}
	p.spans = []chipSpan{{id: 3}}
	p.ta.SetValue("unrelated")
	p.reconcile()
	if len(p.spans) != 0 {
		t.Fatal("missing label retained")
	}
	label2, label3 := chipLabel(2), chipLabel(3)
	p.ta.SetValue(label2 + label3)
	p.store = map[int]chipContent{2: {lines: 2}, 3: {lines: 3}}
	p.spans = []chipSpan{{id: 3}, {id: 2}}
	p.reconcile()
	if len(p.spans) != 1 {
		t.Fatalf("out-of-order spans = %#v", p.spans)
	}

	if runeCount([]rune("abc"), nil) != 0 || runeIndex([]rune("abc"), nil) != 0 || runeIndex([]rune("a"), []rune("ab")) != -1 {
		t.Fatal("rune empty/oversize")
	}
	if runeIndex([]rune("abc"), []rune("bc")) != 1 || runeIndex([]rune("abc"), []rune("z")) != -1 {
		t.Fatal("rune search")
	}
	if row, col := runeOffsetToRowCol("a\nb", 2); row != 1 || col != 0 {
		t.Fatalf("newline = %d,%d", row, col)
	}
	if row, col := runeOffsetToRowCol("a", 99); row != 0 || col != 1 {
		t.Fatalf("past end = %d,%d", row, col)
	}

	p.ta.SetValue("first\nsecond")
	p.ta.CursorUp()
	p.ta.CursorStart()
	p.setCursorRuneOffset(8)
	if p.cursorRuneOffset() != 8 {
		t.Fatalf("cursor = %d", p.cursorRuneOffset())
	}
	p.setCursorRuneOffset(1)
	if p.cursorRuneOffset() != 1 {
		t.Fatalf("cursor up = %d", p.cursorRuneOffset())
	}
	_ = p.Line()
	_ = p.LineCount()
	origDown := movePromptCursorDown
	t.Cleanup(func() { movePromptCursorDown = origDown })
	p.ta.SetValue("a\nb")
	p.ta.CursorUp()
	movePromptCursorDown = func(*textarea.Model) {}
	p.setCursorRuneOffset(2)

	p.ta.SetValue("label")
	p.spans = []chipSpan{{id: 404, start: 0, end: 5}}
	p.store = map[int]chipContent{}
	if got := p.Value(); got != "label" {
		t.Fatalf("missing-store expansion = %q", got)
	}
}
