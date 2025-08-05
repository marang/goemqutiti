package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// handleWindowSize adjusts the layout when the terminal is resized.
func (m *model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.ui.width = msg.Width
	m.ui.height = msg.Height
	m.connections.Manager.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
	// textinput.View() renders the prompt and cursor in addition
	// to the configured width. Reduce the width slightly so the
	// surrounding box stays within the terminal boundaries.
	m.topics.Input.Width = msg.Width - 7
	m.message.Input().SetWidth(msg.Width - 4)
	m.message.Input().SetHeight(m.layout.message.height)
	if m.layout.history.height == 0 {
		m.layout.history.height = (msg.Height-1)/3 + 10
	}
	m.history.List().SetSize(msg.Width-4, m.layout.history.height)
	if m.layout.trace.height == 0 {
		m.layout.trace.height = msg.Height - 6
	}
	m.traces.ViewList().SetSize(msg.Width-4, m.layout.trace.height)
	m.traces.List().SetSize(msg.Width-4, msg.Height-4)
	m.topics.List().SetSize(msg.Width/2-4, msg.Height-4)
	m.help.SetSize(msg.Width, msg.Height)
	m.history.Detail().Width = msg.Width - 4
	m.history.Detail().Height = msg.Height - 4
	m.ui.viewport.Width = msg.Width
	// Reserve two lines for the info header at the top of the view.
	m.ui.viewport.Height = msg.Height - 2
	return nil
}

// handleKeyNav processes global navigation key presses.
func (m *model) handleKeyNav(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+up", "ctrl+k":
		m.ui.viewport.ScrollUp(1)
		return nil, true
	case "ctrl+down", "ctrl+j":
		m.ui.viewport.ScrollDown(1)
		return nil, true
	case "tab":
		if m.CurrentMode() == modeHistoryFilter {
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
	case "shift+tab":
		if m.CurrentMode() == modeHistoryFilter {
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

	if m.CurrentMode() != modeHistoryFilter &&
		(msg.String() == "enter" || msg.String() == " " || msg.String() == "space") &&
		m.help.Focused() {
		return m.SetMode(modeHelp), true
	}
	return nil, false
}
