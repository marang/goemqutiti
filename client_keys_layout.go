package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleTabKey moves focus forward.
func (m *model) handleTabKey() tea.Cmd {
	if len(m.ui.focusOrder) > 0 {
		m.focus.Next()
		m.ui.focusIndex = m.focus.Index()
		id := m.ui.focusOrder[m.ui.focusIndex]
		m.setFocus(id)
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
	}
	return nil
}

// handleShiftTabKey moves focus backward.
func (m *model) handleShiftTabKey() tea.Cmd {
	if len(m.ui.focusOrder) > 0 {
		m.focus.Prev()
		m.ui.focusIndex = m.focus.Index()
		id := m.ui.focusOrder[m.ui.focusIndex]
		m.setFocus(id)
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
	}
	return nil
}

// handleResizeUpKey reduces the height of the focused pane.
func (m *model) handleResizeUpKey() tea.Cmd {
	id := m.ui.focusOrder[m.ui.focusIndex]
	if id == idMessage {
		if m.layout.message.height > 1 {
			m.layout.message.height--
			m.message.Input().SetHeight(m.layout.message.height)
		}
	} else if id == idHistory {
		if m.layout.history.height > 1 {
			m.layout.history.height--
			m.history.List().SetSize(m.ui.width-4, m.layout.history.height)
		}
	} else if id == idTopics {
		if m.layout.topics.height > 1 {
			m.layout.topics.height--
		}
	}
	return nil
}

// handleResizeDownKey increases the height of the focused pane.
func (m *model) handleResizeDownKey() tea.Cmd {
	id := m.ui.focusOrder[m.ui.focusIndex]
	if id == idMessage {
		m.layout.message.height++
		m.message.Input().SetHeight(m.layout.message.height)
	} else if id == idHistory {
		m.layout.history.height++
		m.history.List().SetSize(m.ui.width-4, m.layout.history.height)
	} else if id == idTopics {
		m.layout.topics.height++
	}
	return nil
}

// handleModeSwitchKey switches application modes for special key combos.
func (m *model) handleModeSwitchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+b":
		if err := m.connections.Manager.LoadProfiles(""); err != nil {
			m.history.Append("", err.Error(), "log", err.Error())
		}
		m.connections.RefreshConnectionItems()
		m.connections.SaveCurrent(m.topics.Snapshot(), m.payloads.Snapshot())
		m.traces.SavePlannedTraces()
		return m.setMode(modeConnections)
	case "ctrl+t":
		m.topics.SetActivePane(0)
		m.topics.RebuildActiveTopicList()
		m.topics.SetSelected(0)
		return m.setMode(modeTopics)
	case "ctrl+p":
		m.payloads.List().SetSize(m.ui.width-4, m.ui.height-4)
		return m.setMode(modePayloads)
	case "ctrl+r":
		m.traces.List().SetSize(m.ui.width-4, m.ui.height-4)
		return m.setMode(modeTracer)
	default:
		return nil
	}
}
