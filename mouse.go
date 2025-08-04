package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// isHistoryFocused reports if the history list has focus.
func (m *model) isHistoryFocused() bool {
	return m.FocusedID() == idHistory
}

// isTopicsFocused reports if the topics view has focus.
func (m *model) isTopicsFocused() bool {
	return m.FocusedID() == idTopics
}

// handleMouseScroll processes scroll wheel events.
// It returns a command and a boolean indicating if the event was handled.
func (m *model) handleMouseScroll(msg tea.MouseMsg) (tea.Cmd, bool) {
	if msg.Action == tea.MouseActionPress && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
		if m.isHistoryFocused() && !m.history.ShowArchived() {
			return m.history.Scroll(msg), true
		}
		if m.isTopicsFocused() {
			delta := -1
			if msg.Button == tea.MouseButtonWheelDown {
				delta = 1
			}
			m.topics.Scroll(delta)
			return nil, true
		}
		return nil, true
	}
	return nil, false
}

// handleMouseLeft manages left-click focus and selection.
func (m *model) handleMouseLeft(msg tea.MouseMsg) tea.Cmd {
	cmd := m.focusFromMouse(msg.Y)
	if m.isHistoryFocused() && !m.history.ShowArchived() {
		m.history.HandleClick(msg, m.ui.elemPos[idHistory], m.ui.viewport.YOffset)
	}
	return cmd
}

// handleClientMouse processes mouse events in client mode.
func (m *model) handleClientMouse(msg tea.MouseMsg) tea.Cmd {
	if cmd, handled := m.handleMouseScroll(msg); handled {
		return cmd
	}
	var cmds []tea.Cmd
	if msg.Type == tea.MouseLeft {
		if cmd := m.handleMouseLeft(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if msg.Type == tea.MouseLeft || msg.Type == tea.MouseRight {
		if cmd := m.topics.HandleClick(msg, m.ui.viewport.YOffset); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}
