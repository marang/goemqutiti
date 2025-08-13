package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

const (
	focusNext = 1
	focusPrev = -1
)

// handleWindowSize adjusts the layout when the terminal is resized.
func (m *model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.ui.width = msg.Width
	m.ui.height = msg.Height
	cw, ch := m.layout.ConnectionsSize(msg.Width, msg.Height)
	m.connections.Manager.ConnectionsList.SetSize(cw, ch)
	// textinput.View() renders the prompt and cursor in addition
	// to the configured width. Reduce the width slightly so the
	// surrounding box stays within the terminal boundaries.
	m.topics.Input.Width = m.layout.TopicsInputWidth(msg.Width)
	m.message.Input().SetWidth(m.layout.MessageWidth(msg.Width))
	m.message.Input().SetHeight(m.layout.Message.Height)
	hw, hh := m.layout.HistorySize(msg.Width, msg.Height)
	m.history.List().SetSize(hw, hh)
	th := m.layout.TraceHeight(msg.Height)
	m.traces.ViewList().SetSize(m.layout.MessageWidth(msg.Width), th)
	tw, tlh := m.layout.TraceListSize(msg.Width, msg.Height)
	m.traces.List().SetSize(tw, tlh)
	lw, lh := m.layout.TopicsListSize(msg.Width, msg.Height)
	m.topics.List().SetSize(lw, lh)
	m.help.SetSize(msg.Width, msg.Height)
	m.logs.SetSize(msg.Width, msg.Height)
	dw, dh := m.layout.DetailSize(msg.Width, msg.Height)
	m.history.Detail().Width = dw
	m.history.Detail().Height = dh
	m.ui.viewport.Width = msg.Width
	// Reserve two lines for the info header at the top of the view.
	m.ui.viewport.Height = m.layout.ViewportHeight(msg.Height)
	return nil
}

// cycleFocus moves focus forward or backward through the focus order.
// It updates the focus index and ensures the topics list selection is valid.
// A non-zero return indicates the focus changed.
func (m *model) cycleFocus(direction int) (tea.Cmd, bool) {
	if len(m.ui.focusOrder) == 0 {
		return nil, false
	}
	switch direction {
	case focusNext:
		m.focus.Next()
	case focusPrev:
		m.focus.Prev()
	default:
		return nil, false
	}
	m.ui.focusIndex = m.focus.Index()
	id := m.ui.focusOrder[m.ui.focusIndex]
	cmd := m.SetFocus(id)
	if id == idTopics {
		if len(m.topics.Items) > 0 {
			sel := m.topics.Selected()
			if sel < 0 || sel >= len(m.topics.Items) {
				m.topics.SetSelected(0)
			}
			m.topics.EnsureVisible(m.ui.width - 4)
		} else {
			m.topics.SetSelected(-1)
		}
	}
	return cmd, true
}

// handleKeyNav processes global navigation key presses.
func (m *model) handleKeyNav(msg tea.KeyMsg) (tea.Cmd, bool) {
	key := msg.String()
	switch key {
	case constants.KeyCtrlUp, constants.KeyCtrlK:
		m.ui.viewport.ScrollUp(1)
		return nil, true
	case constants.KeyCtrlDown, constants.KeyCtrlJ:
		m.ui.viewport.ScrollDown(1)
		return nil, true
	case constants.KeyTab:
		if m.CurrentMode() == constants.ModeHistoryFilter {
			return m.history.UpdateFilter(msg), true
		}
		if cmd, ok := m.cycleFocus(focusNext); ok {
			return cmd, true
		}
	case constants.KeyShiftTab:
		if m.CurrentMode() == constants.ModeHistoryFilter {
			return m.history.UpdateFilter(msg), true
		}
		if cmd, ok := m.cycleFocus(focusPrev); ok {
			return cmd, true
		}
	}

	if m.CurrentMode() != constants.ModeHistoryFilter &&
		(key == constants.KeyEnter || key == constants.KeySpaceBar || key == constants.KeySpace) &&
		m.help.Focused() {
		return m.SetMode(constants.ModeHelp), true
	}
	return nil, false
}
