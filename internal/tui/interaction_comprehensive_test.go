package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
)

func TestPopoverFilteringSelectionAndRendering(t *testing.T) {
	m := baselineModel(t)
	if m.popoverOpen() {
		t.Fatal("popover initially open")
	}
	if _, ok := m.currentSuggestion(); ok {
		t.Fatal("closed suggestion")
	}
	m.ta.SetValue("/")
	m.refreshSuggest()
	if !m.popoverOpen() || len(m.suggest) != len(commands) {
		t.Fatalf("suggestions = %#v", m.suggest)
	}
	next, _ := m.popoverMoveSelection(-1)
	m = next.(Model)
	if m.suggestIdx != len(m.suggest)-1 {
		t.Fatalf("wrapped index = %d", m.suggestIdx)
	}
	m.width = 50
	if got := m.renderPopover(); !strings.Contains(got, "/models") {
		t.Fatalf("render = %q", got)
	}
	m.setPopover([]argOption{{value: "a"}, {value: "b", current: true}}, true, "/models")
	if m.suggestIdx != 1 || selectInitialIdx([]argOption{{value: "a"}}) != 0 {
		t.Fatal("initial selection")
	}
	m.setPopover([]argOption{{value: "selected"}, {value: "current", current: true}}, false, "")
	m.suggestIdx = 0
	_ = m.renderPopover()
	m.suggest = append(m.suggest, make([]argOption, popoverCap+2)...)
	m.suggestIdx = len(m.suggest) - 1
	_ = m.renderPopover()
	m.ta.SetValue("plain")
	m.refreshSuggest()
	if m.popoverOpen() {
		t.Fatal("plain text popover")
	}
	m.ta.SetValue("/clear ")
	m.refreshSuggest()
	if m.popoverOpen() {
		t.Fatal("no-arg command popover")
	}
	m.setPopover(nil, false, "")
	if m.popoverOpen() {
		t.Fatal("empty set stayed open")
	}
	if got, _ := m.popoverMoveSelection(1); got.(Model).suggestIdx != 0 {
		t.Fatal("empty move")
	}
	m.setPopover([]argOption{{value: "plain", description: "row"}}, false, "")
	m.suggestIdx = -1
	m.suggestIdx = 0
	_ = m.renderPopover()
	if got, ok := m.currentSuggestion(); !ok || got.value != "plain" {
		t.Fatal("current suggestion")
	}
	m.closePopover()
	if m.popoverHeight() != 0 || m.renderPopover() != "" {
		t.Fatal("closed rendering")
	}
}

func TestHistoryTabEscapeAndQuitKeys(t *testing.T) {
	m := baselineModel(t)
	if m.historyUp().histIdx != -1 || m.historyDown().histIdx != -1 {
		t.Fatal("empty history moved")
	}
	m.promptHistory = []promptEntry{{display: "old"}, {display: "new"}}
	m.ta.SetValue("draft")
	m = m.historyUp()
	if m.ta.Value() != "new" {
		t.Fatalf("up = %q", m.ta.Value())
	}
	m = m.historyUp()
	m = m.historyUp()
	if m.ta.Value() != "old" {
		t.Fatalf("oldest = %q", m.ta.Value())
	}
	m = m.historyDown()
	m = m.historyDown()
	if m.ta.Value() != "draft" {
		t.Fatalf("draft = %q", m.ta.Value())
	}

	m.ta.Reset()
	next, _ := m.handleTab(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(Model)
	if m.ta.Value() != "/" || !m.popoverOpen() {
		t.Fatal("tab did not seed slash")
	}
	m.ta.SetValue("/mod")
	m.refreshSuggest()
	next, _ = m.handleTab(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(Model)
	if m.ta.Value() != "/models " || !m.suggestArgLevel {
		t.Fatalf("completion = %q", m.ta.Value())
	}
	next, _ = m.handleEscInPopover()
	m = next.(Model)
	if m.ta.Value() != "/models" {
		t.Fatalf("arg escape = %q", m.ta.Value())
	}
	next, _ = m.handleEscInPopover()
	m = next.(Model)
	if m.ta.Value() != "" || m.popoverOpen() {
		t.Fatal("command escape")
	}

	next, cmd := m.handleCtrlC()
	m = next.(Model)
	if cmd == nil || m.status != quitArmText {
		t.Fatal("quit not armed")
	}
	next, cmd = m.handleCtrlC()
	if cmd == nil {
		t.Fatal("second ctrl-c did not quit")
	}
	m = next.(Model)
	m.quitArmedAt = time.Now().Add(time.Second)
	m.status = quitArmText
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = next.(Model)
	if !m.quitArmedAt.IsZero() || m.status != "" {
		t.Fatal("ordinary key did not disarm")
	}
}

func TestEnterQueueAndSlashContracts(t *testing.T) {
	m := baselineModel(t)
	m.phase = phaseThinking
	next, _ := m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.queued != nil {
		t.Fatal("empty prompt queued")
	}
	m.ta.SetValue("first")
	next, _ = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	m.ta.SetValue("second")
	next, _ = m.queuePrompt()
	m = next.(Model)
	if m.queued.send != "first\nsecond" {
		t.Fatalf("queue = %q", m.queued.send)
	}
	next, _ = m.unqueuePrompt()
	m = next.(Model)
	if m.queued != nil || m.ta.Value() != "first\nsecond" {
		t.Fatal("unqueue")
	}
	m.ta.SetValue("/clear")
	next, _ = m.queuePrompt()
	m = next.(Model)
	m.ta.SetValue("prose")
	next, _ = m.queuePrompt()
	m = next.(Model)
	if m.status != queueSlashHint || m.ta.Value() != "prose" {
		t.Fatal("slash boundary")
	}

	m.phase = phaseIdle
	m.ta.Reset()
	next, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Fatal("empty ctrl-d")
	}
	m = next.(Model)
	m.ta.SetValue("x")
	_, cmd = m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd != nil {
		t.Fatal("draft ctrl-d")
	}
	next, _ = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter, Alt: true})
	m = next.(Model)
	if !strings.Contains(m.ta.Value(), "\n") {
		t.Fatalf("alt enter = %q", m.ta.Value())
	}
}

func TestSlashCommandsModelSwitchAndHelp(t *testing.T) {
	m := baselineModel(t)
	m.cfg.Models["other"] = &config.Profile{LLM: "other-model", URL: "http://other", ContextSize: 4096}
	if err := m.cfg.Save(); err != nil {
		t.Fatal(err)
	}
	if commandByName("/clear") == nil || commandByName("/missing") != nil {
		t.Fatal("command lookup")
	}
	var help bytes.Buffer
	PrintHelp(&help)
	if !strings.Contains(help.String(), "/models") {
		t.Fatal("help")
	}
	next, _ := m.runSlash("/models")
	m = next.(Model)
	if !strings.Contains(m.scroll.String(), "models") {
		t.Fatal("model list")
	}
	next, _ = m.cmdModel([]string{"missing"})
	m = next.(Model)
	if !strings.Contains(m.scroll.String(), "unknown model") {
		t.Fatal("unknown switch")
	}
	next, cmd := m.cmdModel([]string{"other"})
	m = next.(Model)
	if m.cfg.Active != "other" || cmd == nil || m.cli.Model != "other-model" {
		t.Fatal("switch")
	}
	next, _ = m.runSlash("/unknown")
	m = next.(Model)
	if !strings.Contains(m.scroll.String(), "unknown command") {
		t.Fatal("unknown slash")
	}
	m.cfg.Models["keyed"] = &config.Profile{LLM: "keyed", URL: "http://keyed", Key: "secret", ContextSize: 4096}
	m.cfg.Active = "keyed"
	m.rebuildClient()
	if cmd := m.confirmActive("keyed"); cmd == nil || !strings.Contains(m.scroll.String(), "probing") {
		t.Fatal("keyed confirm")
	}
	// A hand-edited endpoint must rebuild the client on reload.
	m.cfg.Active = "other"
	if err := m.cfg.Save(); err != nil {
		t.Fatal(err)
	}
	rawPath := filepath.Join(m.cfg.Dir, "config.yaml")
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rawPath, []byte(strings.Replace(string(raw), "http://other", "http://changed", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := m.reloadConfigFromDisk(); err != nil || m.cli.BaseURL != "http://changed" {
		t.Fatalf("reload = %v url=%s", err, m.cli.BaseURL)
	}
	if err := os.WriteFile(rawPath, []byte("not: [valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	next, _ = m.runSlash("/models")
	m = next.(Model)
	if !strings.Contains(m.scroll.String(), "config.yaml") {
		t.Fatal("reload error not surfaced")
	}
}

func TestHandleKeyAllInteractionBranches(t *testing.T) {
	m := baselineModel(t)
	m.width, m.height = 80, 24
	m.ta.SetValue("clear me")
	next, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlL})
	m = next.(Model)
	if cmd == nil || m.ta.Value() != "" {
		t.Fatal("ctrl-l")
	}
	m.phase = phaseThinking
	m.queued = &queuedPrompt{send: "edit me", echo: "edit me"}
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(Model)
	if m.queued != nil || m.ta.Value() != "edit me" {
		t.Fatal("backspace unqueue")
	}
	_, cmd = m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd != nil {
		t.Fatal("active ctrl-d")
	}
	m.phase = phaseIdle
	m.ta.SetValue("one\ntwo")
	m.ta.ta.CursorUp()
	m.ta.ta.CursorStart()
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m.ta.ta.CursorDown()
	m.ta.CursorEnd()
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m.setPopover([]argOption{{value: "a"}, {value: "b"}}, false, "")
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(Model)
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = next.(Model)
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	m.ta.SetValue("text")
	_, _ = m.handleTab(tea.KeyMsg{Type: tea.KeyTab})
	m.setPopover([]argOption{{value: "/clear"}, {value: "/models"}}, false, "")
	_, _ = m.handleTab(tea.KeyMsg{Type: tea.KeyTab})

	m.cancel = func() {}
	m.phase = phaseThinking
	m.turnStart = time.Now().Add(-time.Second)
	next, _ = m.handleCtrlC()
	m = next.(Model)
	if m.phase != phaseIdle {
		t.Fatal("active ctrl-c")
	}
	m.setPopover([]argOption{{value: "x"}}, false, "")
	next, _ = m.handleCtrlC()
	m = next.(Model)
	if m.popoverOpen() {
		t.Fatal("popover ctrl-c")
	}
}

func TestHandleEnterSelectionAndSlashCommit(t *testing.T) {
	m := baselineModel(t)
	m.promptHistory = []promptEntry{{display: "recall"}}
	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(Model)
	if m.ta.Value() != "recall" {
		t.Fatal("key history up")
	}
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	m.setPopover([]argOption{{value: "/models"}}, false, "")
	next, _ = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if !m.suggestArgLevel {
		t.Fatal("command did not advance")
	}
	m.cfg.Models["other"] = &config.Profile{LLM: "m", URL: "http://x", ContextSize: 10}
	if err := m.cfg.Save(); err != nil {
		t.Fatal(err)
	}
	m.setPopover([]argOption{{value: "other"}}, true, "/models")
	next, _ = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.cfg.Active != "other" {
		t.Fatal("argument selection not committed")
	}
	m.ta.SetValue("   ")
	next, cmd := m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || next.(Model).ta.Value() == "" { /* empty is a no-op and whitespace remains */
	}
	m.ta.SetValue("/clear")
	next, cmd = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd == nil || len(m.history) != 0 {
		t.Fatal("plain slash commit")
	}
	if !m.cursorOnFirstLine() || !m.cursorOnLastLine() || m.ta.Line() != 0 || m.ta.LineCount() < 1 {
		t.Fatal("cursor helpers")
	}
	m.setPopover([]argOption{{value: "/clear"}}, false, "")
	next, _ = m.handleEnter(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)
	_, tick := m.handleCtrlC()
	if tick != nil {
		if _, ok := tick().(quitArmResetMsg); !ok {
			t.Fatal("quit tick")
		}
	}
}
