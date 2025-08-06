package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// calcHistorySize returns the width and height for the history list.
// It defaults the height when the current value is zero.
func calcHistorySize(width, height, currentHeight int) (int, int) {
	if currentHeight == 0 {
		currentHeight = (height-1)/3 + 10
	}
	return calcMessageWidth(width), currentHeight
}

// calcMessageWidth returns the width for message inputs and lists.
func calcMessageWidth(width int) int {
	return width - 4
}

// calcConnectionsSize returns the width and height for the connections list.
func calcConnectionsSize(width, height int) (int, int) {
	return calcMessageWidth(width), height - 6
}

// calcTopicsInputWidth returns the width for the topics input.
func calcTopicsInputWidth(width int) int {
	return width - 7
}

// calcTraceHeight returns the height for trace views, defaulting if zero.
func calcTraceHeight(height, currentHeight int) int {
	if currentHeight == 0 {
		return height - 6
	}
	return currentHeight
}

// calcTraceListSize returns size for trace lists.
func calcTraceListSize(width, height int) (int, int) {
	return calcMessageWidth(width), height - 4
}

// calcTopicsListSize returns size for the topics list.
func calcTopicsListSize(width, height int) (int, int) {
	return width/2 - 4, height - 4
}

// calcDetailSize returns size for the history detail view.
func calcDetailSize(width, height int) (int, int) {
	return calcMessageWidth(width), height - 4
}

// calcViewportHeight returns the viewport height, reserving two lines for headers.
func calcViewportHeight(height int) int {
	return height - 2
}

// handleWindowSize adjusts the layout when the terminal is resized.
func (m *model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.ui.width = msg.Width
	m.ui.height = msg.Height
	cw, ch := calcConnectionsSize(msg.Width, msg.Height)
	m.connections.Manager.ConnectionsList.SetSize(cw, ch)
	// textinput.View() renders the prompt and cursor in addition
	// to the configured width. Reduce the width slightly so the
	// surrounding box stays within the terminal boundaries.
	m.topics.Input.Width = calcTopicsInputWidth(msg.Width)
	m.message.Input().SetWidth(calcMessageWidth(msg.Width))
	m.message.Input().SetHeight(m.layout.message.height)
	hw, hh := calcHistorySize(msg.Width, msg.Height, m.layout.history.height)
	m.layout.history.height = hh
	m.history.List().SetSize(hw, hh)
	m.layout.trace.height = calcTraceHeight(msg.Height, m.layout.trace.height)
	m.traces.ViewList().SetSize(calcMessageWidth(msg.Width), m.layout.trace.height)
	tw, th := calcTraceListSize(msg.Width, msg.Height)
	m.traces.List().SetSize(tw, th)
	lw, lh := calcTopicsListSize(msg.Width, msg.Height)
	m.topics.List().SetSize(lw, lh)
	m.help.SetSize(msg.Width, msg.Height)
	dw, dh := calcDetailSize(msg.Width, msg.Height)
	m.history.Detail().Width = dw
	m.history.Detail().Height = dh
	m.ui.viewport.Width = msg.Width
	// Reserve two lines for the info header at the top of the view.
	m.ui.viewport.Height = calcViewportHeight(msg.Height)
	return nil
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
		if len(m.ui.focusOrder) > 0 {
			m.focus.Next()
			m.ui.focusIndex = m.focus.Index()
			id := m.ui.focusOrder[m.ui.focusIndex]
			m.SetFocus(id)
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
			return nil, true
		}
	case constants.KeyShiftTab:
		if m.CurrentMode() == constants.ModeHistoryFilter {
			return m.history.UpdateFilter(msg), true
		}
		if len(m.ui.focusOrder) > 0 {
			m.focus.Prev()
			m.ui.focusIndex = m.focus.Index()
			id := m.ui.focusOrder[m.ui.focusIndex]
			m.SetFocus(id)
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
			return nil, true
		}
	}

	if m.CurrentMode() != constants.ModeHistoryFilter &&
		(key == constants.KeyEnter || key == constants.KeySpaceBar || key == constants.KeySpace) &&
		m.help.Focused() {
		return m.SetMode(constants.ModeHelp), true
	}
	return nil, false
}
