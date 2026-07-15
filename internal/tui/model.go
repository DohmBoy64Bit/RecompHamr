package tui

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

const (
	defaultWidth = 80              // bootstrap width before the first WindowSizeMsg
	minViewport  = 5               // rows reserved above the prompt for streaming tokens
	popoverCap   = 6               // max rows the popover may claim
	pingTimeout  = 2 * time.Second // backend reachability probe budget
)

// phase aliases the presentation-neutral agent state during the mechanical
// Stage C adapter migration.
type phase = agent.Phase

const (
	phaseIdle      = agent.PhaseIdle
	phaseThinking  = agent.PhaseThinking
	phaseStreaming = agent.PhaseStreaming
	phaseRunning   = agent.PhaseRunning
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

	cfg *config.Config
	cli *llm.Client

	turn         *agent.TurnState   // test-visible alias; private turn capabilities remain inside internal/agent
	runtime      *agent.StreamState // test-visible alias; production paths use agentRuntime methods
	loop         *agent.LoopState   // test-visible alias; production paths use agentRuntime methods
	executor     agent.ToolExecutor // test-visible injected executor alias
	agentRuntime agent.Runtime
	system       string // embedded system prompt + working-directory anchor

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

	status string // transient status-bar hint; cleared by the event that obsoletes it (keypress, quit-arm timer, endTurn)

}

func New(cfg *config.Config, cli *llm.Client, runtime agent.Runtime, projectDir, version string) Model {
	ta := newPromptInput()

	// Fixed dark style: WithAutoStyle queries the terminal (OSC 11) before
	// bubbletea grabs raw stdin, so the reply bytes leak into the textarea as
	// "1;rgb:1e1e/1e1e/1e1e" garbage. Dev containers are dark: no query, no leak.
	r, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(defaultWidth-4))

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = styleSpinner

	m := Model{
		Version:  version,
		cfg:      cfg,
		cli:      cli,
		system:   buildSystem(projectDir),
		ta:       ta,
		renderer: r,
		spinner:  sp,
		// width/height left at 0; View() returns "" until the first
		// WindowSizeMsg, so we don't flash an 80×24 frame then resize.
		streaming:    new(strings.Builder),
		scroll:       new(strings.Builder),
		histIdx:      -1,
		turn:         runtime.Turn,
		runtime:      runtime.Stream,
		loop:         runtime.Loop,
		executor:     runtime.Executor,
		agentRuntime: runtime,
	}
	// Record the active backend once, before any turn, so a shared log
	// names exactly which model/endpoint/context window produced the behaviour.
	m.agentRuntime.ObserveSession(version, cfg.Active, cfg.ActiveProfile().LLM, cfg.ActiveURL(),
		m.activeContextSize(), chmctx.Tokens(m.system))
	// Seed prompt history from .rehamr/history so ↑ recalls prompts from
	// earlier sessions. Loaded entries carry no chip metadata (the on-disk
	// format stores expanded text only), so a recalled multi-line paste
	// appears uncollapsed, the right tradeoff for a cat-friendly history file.
	m.promptHistory = loadPromptHistory(cfg.Dir)
	return m
}

// activeContextSize returns the context window the packer should aim at: the
// live server-reported value for the active profile if known, else the on-disk
// ContextSize, else defaultPackFallback, so providers before their first
// response (and any missing/zero value) still get a sensible budget.
func (m *Model) activeContextSize() int {
	if v, ok := m.agentRuntime.LiveContextSize(m.cfg.Active); ok {
		return v
	}
	if v := m.cfg.ActiveProfile().ContextSize; v > 0 {
		return v
	}
	return defaultPackFallback
}

// defaultPackFallback is the conservative window used until the server reports
// a real value. Matches config.defaultContextSize so profiles behave like
// a fresh local one until X-Context-Window arrives on the next response.
const defaultPackFallback = 16177

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

// pingMsg carries a backend-reachability result. baseURL is the URL probed;
// Update drops the message when it no longer matches the live client's URL,
// else a stale ping from the prior profile (a mid-flight /models switch) would
// overwrite connected state with the wrong endpoint's reachability.
type pingMsg struct {
	ok      bool
	baseURL string
}

// quitArmResetMsg fires ~3s after Ctrl+C arms the quit: if not already quit or
// re-armed, clear the hint from the status bar.
type quitArmResetMsg struct{}

func (m Model) Init() tea.Cmd {
	// Keyed profiles get a silent Probe at startup so credentials are validated
	// and an optional live context window can be harvested. Keyless profiles use
	// the cheaper reachability probe.
	connectivity := pingBackend(m.cli.BaseURL)
	if p := m.cfg.ActiveProfile(); p != nil && p.ResolvedKey() != "" {
		connectivity = probeBackend(m.cli, m.cfg.Active, true)
	}
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		connectivity,
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

	case pingMsg:
		// Drop stale pings from a prior backend (a /models switch while a ping
		// was in flight). The live client's URL is the source of truth.
		if msg.baseURL != m.cli.BaseURL {
			return m, nil
		}
		m.agentRuntime.SetConnected(msg.ok)
		return m, nil

	case probeMsg:
		return m.handleProbe(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case streamEventMsg:
		// Stale event from a stream the current turn no longer owns (Ctrl+C →
		// fresh submit while the prior readEvent was in flight). Keep draining
		// the channel so the producer goroutine exits cleanly, but never let
		// the event mutate the now-active turn's state.
		delivery := m.agentRuntime.ApplyDelivery(msg.stream, m.cfg.Active, m.activeContextSize(), msg.delivery)
		if !delivery.Accepted {
			return m, readEvent(msg.stream)
		}
		return m.applyStreamEffect(delivery.Stream)

	case streamClosedMsg:
		// Stale close from the prior turn's channel; running handleStreamClosed
		// would release the live agent stream and, worse, finalizeTurn + endTurn the
		// active turn, killing the user's request out from under them.
		delivery := m.agentRuntime.ApplyDelivery(msg.stream, m.cfg.Active, m.activeContextSize(), msg.delivery)
		if !delivery.Accepted || !delivery.Closed {
			return m, nil
		}
		return m.handleStreamClosed()

	case toolResultMsg:
		// Stale result from a turn the user already cancelled. Without this
		// drop, the orphan tool message gets appended to the live turn's history
		// (no preceding assistant.tool_calls → the next /v1 request 400s) and
		// startChat would abandon the in-flight stream. The turnCtx tag was
		// captured at runToolCall time; endTurn deactivates that identity and a
		// fresh beginTurn installs a new one that cannot match.
		effect := m.agentRuntime.ApplyToolResult(msg.delivery)
		if !effect.Accepted {
			return m, nil
		}
		// Drain every remaining call before re-entering chat: OpenAI rejects an
		// assistant.tool_calls message followed by fewer tool messages than
		// calls issued, so a partial dispatch 400s and loses the rest.
		// Sequential dispatch in emit order keeps the pairing intact.
		if effect.ContinueTools {
			return m.dispatchNextTool()
		}
		return m, m.startChat()

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
	_ = appendPromptHistory(m.cfg.Dir, safeText)

	if strings.HasPrefix(sendText, "/") {
		m.agentRuntime.ObserveSlash(safeText)
		return m.runSlash(sendText)
	}
	// A new user message is a new goal: drop any in-progress failure streak so
	// a stale count can't trip the nudge early. History persists; only the
	// counter resets.
	m.agentRuntime.ObserveUser(sendText)
	return m, m.appendUserTurn(sendText)
}

func (m *Model) startChat() tea.Cmd {
	m.agentRuntime.Client = m.cli
	stream, _ := m.agentRuntime.StartRound(m.system, m.cfg.ActiveProfile().LLM, m.activeContextSize())
	return readEvent(stream)
}

// installTurnContext cancels any in-flight turn context and installs a fresh
// opaque per-turn root. Cancel-old-then-install-new keeps Ctrl+C consistent:
// one agent-owned cancellation capability unwinds the whole current cascade.
func (m *Model) installTurnContext() {
	m.agentRuntime.BeginTurn(time.Now())
}

// beginTurn installs a fresh per-turn context, flips phase to thinking, and
// returns the chat stream-reader Cmd. Every path starting a new LLM round
// funnels through here so one agent-owned cancellation root cancels the cascade.
func (m *Model) beginTurn() tea.Cmd {
	m.installTurnContext()
	m.turnStart = time.Now()
	m.lastOutcome = outcomeNone // the new run replaces the prior frozen summary
	return m.startChat()
}

// appendUserTurn appends a user-role message to history and starts a turn.
// The only path that does so; used by submit.
func (m *Model) appendUserTurn(content string) tea.Cmd {
	m.agentRuntime.AppendUser(content)
	return m.beginTurn()
}

// endTurn zeroes per-turn state after a turn finishes or aborts. Pair to
// beginTurn. Cancels the per-turn context unconditionally to release the
// CancelFunc; Background-rooted contexts otherwise leak one child cancelCtx
// per turn until the process exits. Drops pending tool calls so a turn cut
// short mid-dispatch (Ctrl+C or error) can't leak a leftover call into the next
// turn, which would dispatch with stale args and append an orphan tool_result
// whose tool_call_id no longer pairs the latest assistant message. Does NOT
// touch scrollback; callers decide whether to flush streaming or emit a banner.
func (m *Model) endTurn() {
	wasRetrying := m.agentRuntime.EndTurn()
	// The queue-refusal hint says "send it when the turn ends"; that moment is
	// now, so the advice would be stale from the next render on.
	if m.status == queueSlashHint {
		m.status = ""
	}
	// A retry hint dies with its turn: a Ctrl+C during the backoff wait would
	// otherwise leave "retry 1/3 in 15s" stranded in the idle status bar.
	if wasRetrying {
		m.status = ""
	}
}

func (m Model) applyStreamEffect(effect agent.StreamEffect) (tea.Model, tea.Cmd) {
	if effect.RetryCleared {
		m.status = ""
	}
	if effect.RetryText != "" {
		m.status = effect.RetryText
	}
	if effect.Content != "" {
		// Display-only tab expansion keeps authoritative model history intact.
		m.streaming.WriteString(strings.ReplaceAll(effect.Content, "\t", "    "))
	}
	if effect.Done {
		m.applyDoneEffect(effect)
	} else if effect.Flush {
		m.flushStreaming()
	}
	if effect.Error != nil {
		return m, m.applyError(effect.Error)
	}
	return m, readEvent(m.agentRuntime.CurrentStream())
}

func (m *Model) applyDoneEffect(effect agent.StreamEffect) {
	m.flushStreaming()
}

// applyError unwinds the turn on a stream error: preserve content streamed
// before the error (so the user keeps failure context), emit the one-line hint,
// drop the pending queue, reset turn state.
func (m *Model) applyError(err error) tea.Cmd {
	m.abortTurn(styleError.Render(m.errorMessage(err)))
	return nil
}

// abortTurn winds down a turn that did not complete normally: flush in-flight
// text so the partial block lands in scrollback, post the explanatory banner,
// drop pending tool calls, reset per-turn counters and context. Pair to
// applyDone for the happy path.
func (m *Model) abortTurn(banner string) {
	m.flushStreaming()
	if banner != "" {
		m.appendLine(banner)
	}
	// A prompt queued mid-turn must NOT auto-fire on an abort: the user took back
	// control, so its follow-up may no longer be wanted. Restore it to the
	// textarea instead (editable, one Enter to send), which also avoids leaving an
	// idle "queued" box that would orphan-fire after the next turn. Only when the
	// textarea is empty, so a draft typed mid-turn isn't clobbered.
	if m.queued != nil {
		if m.ta.Value() == "" {
			m.setPromptText(m.queued.send)
		}
		m.queued = nil
	}
	// finalizeTurn folds the in-flight estimate into the counters and zeroes it,
	// so the avg counts what was generated up to the interrupt; don't drop it here.
	m.finalizeTurn(outcomeStopped)
	m.endTurn() // drops pending tool calls along with the rest of the turn state
}

// finalizeTurn freezes the finished turn's wall-clock summary into the status
// bar (shown at idle until the next submit) and logs the totals. outcome is
// the finish glyph: ✓ clean, ✗ abort/stall. The bar's avg tok/s divides
// lastTokens by lastElapsed (wall-clock), so it stays self-verifying against
// the duration shown right beside it. There is no scrollback banner: the footer
// owns the run summary, and the precise wall time lands in the turn_end log.
// Common to every wind-down (handleStreamClosed and abortTurn both call it).
func (m *Model) finalizeTurn(outcome turnOutcome) {
	if m.turnStart.IsZero() {
		return // defensive: finalizeTurn only runs inside a turn beginTurn started
	}
	wall, tokens := m.agentRuntime.FinalizeTurn(m.turnStart, time.Now())
	m.lastElapsed = wall
	m.lastTokens = tokens
	m.lastOutcome = outcome
	avg := humanRate(tokens, wall)
	if avg != "" {
		avg = " · " + avg + " avg"
	}
	m.agentRuntime.ObserveTurnEnd(humanTokens(tokens), humanTokens(m.agentRuntime.Snapshot().SessionTokens), wall, avg)
	m.turnStart = time.Time{}
}

// handleStreamClosed drives what happens after one round's stream finishes:
// dispatch the next pending tool call, or, if none, finalize the turn and
// hand control back. A turn ends precisely when the assistant emits no tool
// calls; there is no loop tool to land on.
func (m Model) handleStreamClosed() (tea.Model, tea.Cmd) {
	if !m.agentRuntime.Snapshot().Phase.Active() {
		return m, nil
	}
	decision := m.agentRuntime.DecideClose()
	switch decision.Action {
	case agent.CloseRunTool:
		return m.dispatchNextTool()
	case agent.CloseRestartModel:
		return m, m.startChat()
	case agent.CloseFinishStopped:
		if decision.Reason == agent.CloseEmptyStall {
			m.appendLine(styleError.Render("⚠ the model ended its turn with no reply and no tool call - it stalled, or your server dropped the call. If thinking is on, its reasoning parser may be swallowing calls - enable one (e.g. vLLM `--reasoning-parser`) or disable thinking for tool turns."))
		} else {
			m.appendLine(styleError.Render("⚠ a tool call leaked into the reply as text instead of running - your model server isn't parsing tool calls. Enable its OpenAI tool-call parser server-side (e.g. vLLM `--tool-call-parser`, llama.cpp `--jinja`)."))
		}
		m.finalizeTurn(outcomeStopped)
	default:
		m.finalizeTurn(outcomeDone)
	}
	m.endTurn()
	return m.fireQueued()
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

// dispatchNextTool pops the next pending tool call and runs it. Every tool
// flows through runToolCall; none are special-cased. lastToolKey records this
// call's target so the failure nudge can tell when the model keeps retrying the
// same failing operation (see recordToolOutcome).
func (m Model) dispatchNextTool() (tea.Model, tea.Cmd) {
	work, _ := m.agentRuntime.NextTool()
	m.appendLine(styleDim.Render(work.Status()))
	return m, runToolCall(work)
}

// cursorOnFirstLine: true when ↑ should walk prompt history instead of moving
// the textarea's own cursor. cursorOnLastLine is the mirror for ↓.
func (m Model) cursorOnFirstLine() bool { return m.ta.Line() == 0 }
func (m Model) cursorOnLastLine() bool  { return m.ta.Line() == m.ta.LineCount()-1 }

// buildSystem appends the working-directory anchor to the embedded system
// prompt so "hier" / "here" resolves to a concrete path.
func buildSystem(projectDir string) string {
	return config.DefaultSystemPrompt + "\n\nWorking directory: " + projectDir
}

// pingBackend issues a short GET to baseURL/v1/models via provider.Reachable. Any
// HTTP response counts as reachable; transport errors and timeouts mean
// disconnected. The result carries the URL it was issued against so Update can
// drop late results arriving after a /models switch.
func pingBackend(baseURL string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()
		return pingMsg{ok: provider.Reachable(ctx, baseURL) == nil, baseURL: baseURL}
	}
}
