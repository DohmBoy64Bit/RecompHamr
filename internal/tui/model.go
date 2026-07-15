package tui

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"

	"github.com/DohmBoy64Bit/RecompHamr/internal/frontend"
)

const (
	defaultWidth = 80 // bootstrap width before the first WindowSizeMsg
	minViewport  = 5  // rows reserved above the prompt for streaming tokens
	popoverCap   = 6  // max rows the popover may claim
)

// phase aliases the presentation-neutral agent state during the mechanical
// Stage C adapter migration.
type phase = frontend.Phase

const (
	phaseIdle      = frontend.PhaseIdle
	phaseThinking  = frontend.PhaseThinking
	phaseStreaming = frontend.PhaseStreaming
	phaseRunning   = frontend.PhaseRunning
)

// turnOutcome is how a finished turn ended, frozen into the status bar until the
// next submit. outcomeNone is the zero value (no turn has finished yet, or the
// frozen summary was cleared at the start of a new turn).
type turnOutcome int

const (
	outcomeNone turnOutcome = iota
	outcomeDone
	outcomeStopped
)

// marker is the status-bar glyph for the frozen finish: ✓ for a clean finish,
// ✗ for an abort (cancel, error, or a stalled/leaked end). "" suppresses the
// frozen segment for outcomeNone.
func (o turnOutcome) marker() string {
	switch o {
	case outcomeDone:
		return "✓"
	case outcomeStopped:
		return "✗"
	}
	return ""
}

// queuedPrompt is a prompt the user committed while a turn was running, held
// until the turn ends and then auto-submitted. send is the chip-expanded text
// for the LLM; echo is the collapsed, trimmed form shown in the queued box and
// the scrollback echo. Chip state is not preserved: an unqueued or recalled
// paste reappears expanded (matching disk-loaded history), never as a chip, but
// its content is always intact.
type queuedPrompt struct {
	send string
	echo string
}

type Model struct {
	Version string

	controller  frontend.Controller
	startupWork frontend.Work

	// streaming is the live raw token buffer for the current content block,
	// rendered above the prompt by View() while the model talks. On flush
	// (block end, tool call, cancel, error) it's rendered through glamour,
	// queued into outbox for tea.Println, then reset.
	streaming *strings.Builder

	// outbox holds lines bound for terminal scrollback via tea.Println on the
	// next Update cycle. The Update wrapper drains it every cycle, so handlers
	// can call appendLine / flushStreaming without threading a Cmd back.
	outbox []string

	// scroll is the in-memory transcript of every appendLine / flushStreaming
	// line. The real scrollback lives in the user's terminal; this copy is
	// replayed in handleResizeSettle after a width change wipes the terminal,
	// and tests read it to verify what was emitted.
	scroll *strings.Builder

	ta       promptInput
	renderer *glamour.TermRenderer
	spinner  spinner.Model

	// turnStart stamps the wall-clock start of the current turn, set in beginTurn
	// (the user-submit path only; tool re-entry bypasses it), so it spans every
	// tool round rather than resetting per round. The status bar ticks
	// liveElapsed(time.Since(turnStart)) while a turn runs.
	turnStart time.Time
	// last* hold the finished turn's frozen footer summary, shown at idle until
	// the next submit: outcome marker, wall-clock duration, and the token count
	// the avg tok/s divides by. lastTokens ÷ lastElapsed IS the displayed rate,
	// so it stays self-verifying against the shown duration.
	lastElapsed time.Duration
	lastTokens  int
	lastOutcome turnOutcome
	// streamingEstimate is a live char/4 estimate of tokens for the current
	// round (reasoning + content). The server reports the authoritative count
	// only in the final usage block, so without this the footer would freeze
	// through the whole reasoning phase then jump. Reset to 0 on
	// EventDone/Error, where the real count takes over.
	width  int
	height int

	// View() returns "" while suppressView is on, so bubbletea's async ticker
	// can't commit a stale frame mid-drag.
	suppressView bool
	// resizeGen is bumped per width change; settle ticks act only on the
	// matching gen, so older debounces self-discard.
	resizeGen int

	// splashShown guards first-frame emission; later resizes re-emit via the
	// settle handler.
	splashShown bool

	// arrow-key history: every successful submit is appended; histIdx tracks
	// the ↑/↓ walker position (-1 = current draft, 0 = newest). Entries carry
	// display text and chip state so ↑ reconstructs the original atomic-chip
	// prompt, not just its visible text.
	promptHistory []promptEntry
	histIdx       int
	// histDraft stashes the unsent draft when ↑ first leaves the live line, so
	// ↓ back to histIdx -1 restores it instead of clearing the user's typing.
	histDraft promptEntry

	// queued holds a single prompt the user committed while a turn was still
	// running, auto-submitted when the turn finishes naturally (see
	// handleStreamClosed; a Ctrl+C/error abort leaves it untouched for manual
	// send). nil = nothing queued. Enter mid-turn fills/appends to it; Backspace
	// on an empty prompt pulls it back for editing (see queuePrompt/unqueuePrompt).
	queued *queuedPrompt

	// slash-autocomplete popover state. suggest holds command rows (when
	// suggestArgLevel is false) or argument rows for activeCmd (when true):
	// same renderer, same keybindings.
	suggest         []argOption
	suggestIdx      int
	suggestOpen     bool
	suggestArgLevel bool
	activeCmd       string

	quitArmedAt time.Time // first Ctrl+C in idle arms; second within 3s quits

	status                    string // transient status-bar hint; cleared by the event that obsoletes it (keypress, quit-arm timer, endTurn)
	fireQueuedAfterTransition bool
}

func New(controller frontend.Controller, version string) Model {
	ta := newPromptInput()

	// Fixed dark style: WithAutoStyle queries the terminal (OSC 11) before
	// bubbletea grabs raw stdin, so the reply bytes leak into the textarea as
	// "1;rgb:1e1e/1e1e/1e1e" garbage. Dev containers are dark: no query, no leak.
	r, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(defaultWidth-4))

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = styleSpinner

	m := Model{
		Version:    version,
		controller: controller,
		ta:         ta,
		renderer:   r,
		spinner:    sp,
		// width/height left at 0; View() returns "" until the first
		// WindowSizeMsg, so we don't flash an 80×24 frame then resize.
		streaming: new(strings.Builder),
		scroll:    new(strings.Builder),
		histIdx:   -1,
	}
	bootstrap := controller.Bootstrap()
	m.startupWork = bootstrap.Work
	// Seed prompt history from .rehamr/history so ↑ recalls prompts from
	// earlier sessions. Loaded entries carry no chip metadata (the on-disk
	// format stores expanded text only), so a recalled multi-line paste
	// appears uncollapsed, the right tradeoff for a cat-friendly history file.
	for _, event := range bootstrap.Events {
		if event.Kind == frontend.EventHistory {
			for _, value := range event.Values {
				m.promptHistory = append(m.promptHistory, promptEntry{display: value})
			}
		}
	}
	return m
}

// resizeSettleDelay debounces width-resize bursts: longer than typical drag
// SIGWINCH cadence (10-50ms) so a continuous drag collapses to one settle,
// short enough that a one-off resize feels instant.
const resizeSettleDelay = 150 * time.Millisecond

type resizeSettleMsg struct{ gen int }

// eraseScrollback wipes the terminal's saved-lines buffer (DECSED 3); no
// tea.ClearScreen equivalent clears scrollback.
var eraseScrollback tea.Cmd = func() tea.Msg {
	os.Stdout.WriteString(ansi.EraseDisplay(3))
	return nil
}

// frontendCompletionMsg carries an opaque application completion back through
// Bubble Tea without exposing or interpreting its backend payload.
type frontendCompletionMsg struct{ completion frontend.Completion }

// quitArmResetMsg fires ~3s after Ctrl+C arms the quit: if not already quit or
// re-armed, clear the hint from the status bar.
type quitArmResetMsg struct{}

func (m Model) Init() tea.Cmd {
	// Keyed profiles get a silent Probe at startup so credentials are validated
	// and an optional live context window can be harvested. Keyless profiles use
	// the cheaper reachability probe.
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		runFrontendWork(m.startupWork),
	)
}

// Update is the bubbletea entry point: it dispatches to update()'s typed
// handlers then drains the outbox into a single tea.Println, so lines land in
// scrollback in the exact order appendLine / flushStreaming queued them. One
// Println per cycle, never a Batch; Batch runs children concurrently, leaving
// arrival order undefined, so splash lines and tool-call banners would shuffle.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	nm := next.(Model)
	if len(nm.outbox) > 0 {
		printCmd := tea.Println(wrapForScrollback(strings.Join(nm.outbox, "\n"), nm.width))
		nm.outbox = nil
		cmd = tea.Batch(printCmd, cmd)
	}
	return nm, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.FocusMsg, tea.BlurMsg:
		// Terminal focus reports (CSI I / CSI O) arrive as these typed msgs
		// under tea.WithReportFocus. Swallow them so they never reach
		// textarea.Update; otherwise the escape fragments get parsed as
		// printable runes, inserted into the prompt, and bloat textarea height
		// on every focus switch.
		return m, nil

	case tea.KeyMsg:
		// An empty-runes key can surface when the parser chokes mid-escape.
		// Drop it before recomputeLayout wastes cycles.
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 0 {
			return m, nil
		}
		// Pre-grow the textarea before bubbles processes the key. bubbles'
		// repositionView() runs at the end of textarea.Update and scrolls the
		// viewport down whenever the cursor crosses below current Height, but
		// our recomputeLayout() grows Height only AFTER handleKey returns. So
		// a char that wraps to a new visual row leaves YOffset>0 with the first
		// wrap row clipped off the top, which recomputeLayout can't reclaim.
		// Inflating Height to the screen-cap first keeps the cursor inside the
		// visible band for any normal keystroke, so repositionView doesn't
		// scroll and YOffset stays 0; recomputeLayout then trims Height back to
		// visualPromptLines so the live region doesn't bloat empty rows.
		m.preGrowTextarea()
		next, cmd := m.handleKey(msg)
		nm := next.(Model)
		nm.recomputeLayout()
		return nm, cmd

	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case resizeSettleMsg:
		return m.handleResizeSettle(msg)

	case frontendCompletionMsg:
		next, cmd := m.applyFrontendTransition(m.controller.Dispatch(frontend.Complete(msg.completion)))
		m = next.(Model)
		if m.fireQueuedAfterTransition {
			m.fireQueuedAfterTransition = false
			return m.fireQueued()
		}
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case quitArmResetMsg:
		if !m.quitArmedAt.IsZero() && time.Now().After(m.quitArmedAt) {
			m.quitArmedAt = time.Time{}
			if m.status == quitArmText {
				m.status = ""
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	m.recomputeLayout()
	return m, cmd
}

// handleWindowSize tracks new dimensions, rebuilds the glamour renderer on a
// wrap-width change, emits the splash on the first frame, and on a true width
// change starts the debounced resize-settle cycle (suppressView until the
// settle tick lands at the matching gen).
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	first := !m.splashShown
	widthChanged := m.width > 0 && m.width != msg.Width
	m.width, m.height = msg.Width, msg.Height
	m.ta.SetWidth(msg.Width - 2)
	if first || widthChanged {
		// Glamour compiles a stylesheet + template tree per build, so rebuild
		// only on a real wrap-width change; height-only events and intra-drag
		// duplicates reuse the existing renderer.
		if r, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(max(msg.Width-4, 1))); err == nil {
			m.renderer = r
		}
	}
	m.recomputeLayout()
	if first {
		m.splashShown = true
		m.outbox = append(m.outbox, m.splashLines()...)
		return m, nil
	}
	if !widthChanged {
		return m, nil
	}
	m.suppressView = true
	m.resizeGen++
	gen := m.resizeGen
	return m, tea.Tick(resizeSettleDelay, resizeSettled(gen))
}

func resizeSettled(gen int) func(time.Time) tea.Msg {
	return func(time.Time) tea.Msg { return resizeSettleMsg{gen: gen} }
}

// handleResizeSettle fires once the post-resize debounce expires for the
// matching gen. Wipes the terminal (so previous-width rows can't soft-wrap into
// stair-steps), then re-emits splash, replayed scroll, and any pending outbox
// at the new width. Older debounces self-discard on the gen check.
func (m Model) handleResizeSettle(msg resizeSettleMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.resizeGen {
		return m, nil
	}
	m.suppressView = false
	// tea.Sequence keeps order strict; Batch would race the clears with the
	// writes. After the wipe every line below is emitted at the current width,
	// so no previous-width row can soft-wrap into stair-steps.
	cmds := []tea.Cmd{tea.ClearScreen, eraseScrollback}
	if splash := strings.Join(m.splashLines(), "\n"); splash != "" {
		cmds = append(cmds, tea.Println(wrapForScrollback(splash, m.width)))
	}
	if scroll := strings.TrimRight(m.scroll.String(), "\n"); scroll != "" {
		cmds = append(cmds, tea.Println(wrapForScrollback(scroll, m.width)))
	}
	if len(m.outbox) > 0 {
		cmds = append(cmds, tea.Println(wrapForScrollback(strings.Join(m.outbox, "\n"), m.width)))
		m.outbox = nil
	}
	return m, tea.Sequence(cmds...)
}

// submit commits a user prompt. sendText is the expanded form sent to the LLM
// (chip labels replaced by their original paste content); echoText is the
// collapsed form shown in scrollback so the chat doesn't swallow 80 lines of
// pasted log every turn; entry is the history snapshot replayed by ↑/↓,
// including chip state.
func (m Model) submit(sendText, echoText string, entry promptEntry) (tea.Model, tea.Cmd) {
	safeText := sendText
	// Echo to scrollback with the same accent ▌ the textarea uses, one visual
	// language for "your voice" across live input and history.
	m.appendLine(stylePrompt.Render("▌ ") + styleUser.Render(echoText))
	m.promptHistory = append(m.promptHistory, entry)
	m.histIdx = -1
	// Persist the (redacted) prompt so ↑ finds it after a restart. Errors are
	// swallowed: a transient failure isn't worth derailing submit, and a
	// permanent one (read-only .rehamr/) would just be noise on every prompt.
	m.controller.Dispatch(frontend.AppendHistory(safeText))

	if strings.HasPrefix(sendText, "/") {
		m.controller.Dispatch(frontend.ObserveSlash(safeText))
		return m.runSlash(sendText)
	}
	// A new user message is a new goal: drop any in-progress failure streak so
	// a stale count can't trip the nudge early. History persists; only the
	// counter resets.
	return m.applyFrontendTransition(m.controller.Dispatch(frontend.SubmitGoal(sendText, time.Now())))
}

// fireQueued auto-submits a prompt the user queued mid-turn, once the turn has
// wound down to idle. Reached only from the natural finish path (here, after
// finalizeTurn/endTurn); a Ctrl+C or stream-error abort routes through abortTurn
// and never gets here, so an interrupt leaves the slot for a manual send. No-op
// when nothing is queued. The expanded send goes to the LLM, the collapsed echo
// to scrollback, exactly as a typed submit; the recall entry carries the expanded
// text (no chip), matching disk-loaded history.
func (m Model) fireQueued() (tea.Model, tea.Cmd) {
	if m.queued == nil {
		return m, nil
	}
	q := m.queued
	m.queued = nil
	return m.submit(q.send, q.echo, promptEntry{display: q.send})
}

// cursorOnFirstLine: true when ↑ should walk prompt history instead of moving
// the textarea's own cursor. cursorOnLastLine is the mirror for ↓.
func (m Model) cursorOnFirstLine() bool { return m.ta.Line() == 0 }
func (m Model) cursorOnLastLine() bool  { return m.ta.Line() == m.ta.LineCount()-1 }

// runFrontendWork executes captured application work without inspecting its
// backend payload and returns only the opaque completion to the controller.
func runFrontendWork(work frontend.Work) tea.Cmd {
	if work == nil {
		return nil
	}
	return func() tea.Msg {
		return frontendCompletionMsg{completion: work.Run()}
	}
}
