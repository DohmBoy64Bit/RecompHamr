package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/RecompHamr/internal/agent"
)

// streamEventMsg and streamClosedMsg tag their originating channel so the model
// can drop events from a stream the current turn no longer owns. After Ctrl+C →
// fresh submit, the prior turn's readEvent Cmd is still scheduled; without the
// tag its tokens leak into the new turn, or its close runs endTurn against it.
type streamEventMsg struct {
	stream   *agent.Stream
	delivery agent.StreamDelivery
}

type streamClosedMsg struct {
	stream   *agent.Stream
	delivery agent.StreamDelivery
}

// toolResultMsg carries one finished tool call back to Update, tagged with the
// turn identity it was dispatched against. Update drops it when that turn is
// no longer active: the originating work was Ctrl+C'd or superseded.
// Otherwise the orphan result appends to the new turn's history with no
// preceding assistant.tool_calls and abandons that turn's live stream.
type toolResultMsg struct {
	delivery agent.ToolDelivery
}

// readEvent drains one event from the LLM stream as a tea.Msg, re-scheduled
// until the channel closes. Tags ch so Update can spot stale prior-turn events.
func readEvent(stream *agent.Stream) tea.Cmd {
	return func() tea.Msg {
		delivery := stream.Read()
		if delivery.Closed() {
			return streamClosedMsg{stream: stream, delivery: delivery}
		}
		return streamEventMsg{stream: stream, delivery: delivery}
	}
}

// runToolCall executes one tool call off the UI goroutine. parent is the
// per-turn root: Ctrl+C aborts the tool mid-run, and the toolResultMsg carries
// that ctx so Update can drop it if the turn has moved on.
//
// No outer timeout: the PowerShell tool owns its model-set per-call timeout
// (capped at 3600s by the schema), while write_file/edit_file are filesystem-fast. An outer cap would
// silently override the model's PowerShell timeout: a 30-min build dying at 3 min.
func runToolCall(work *agent.ToolWork) tea.Cmd {
	return func() tea.Msg {
		return toolResultMsg{delivery: work.Run()}
	}
}

// errorMessage maps a stream error into a one-line TUI hint, same format across
// all profiles.
func (m Model) errorMessage(err error) string {
	facts := m.controller.Snapshot()
	return agent.StreamErrorMessage(err, facts.Active, facts.ActiveURL)
}
