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
			if len(m.topics.items) > 0 {
				m.topics.selected = 0
				m.ensureTopicVisible()
			} else {
				m.topics.selected = -1
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
			if len(m.topics.items) > 0 {
				m.topics.selected = 0
				m.ensureTopicVisible()
			} else {
				m.topics.selected = -1
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
			m.message.input.SetHeight(m.layout.message.height)
		}
	} else if id == idHistory {
		if m.layout.history.height > 1 {
			m.layout.history.height--
			m.history.list.SetSize(m.ui.width-4, m.layout.history.height)
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
		m.message.input.SetHeight(m.layout.message.height)
	} else if id == idHistory {
		m.layout.history.height++
		m.history.list.SetSize(m.ui.width-4, m.layout.history.height)
	} else if id == idTopics {
		m.layout.topics.height++
	}
	return nil
}

// handleModeSwitchKey switches application modes for special key combos.
func (m *model) handleModeSwitchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+b":
		if err := m.connections.manager.LoadProfiles(""); err != nil {
			m.appendHistory("", err.Error(), "log", err.Error())
		}
		m.refreshConnectionItems()
		m.saveCurrent()
		m.traces.savePlannedTraces()
		return m.setMode(modeConnections)
	case "ctrl+t":
		m.topics.panes.subscribed = paneState{sel: 0, page: 0, index: 0, m: m}
		m.topics.panes.unsubscribed = paneState{sel: 0, page: 0, index: 1, m: m}
		m.topics.panes.active = 0
		m.topics.list.SetSize(m.ui.width/2-2, m.ui.height-4)
		m.rebuildActiveTopicList()
		return m.setMode(modeTopics)
	case "ctrl+p":
		m.payloads.list.SetSize(m.ui.width-4, m.ui.height-4)
		return m.setMode(modePayloads)
	case "ctrl+r":
		m.traces.list.SetSize(m.ui.width-4, m.ui.height-4)
		return m.setMode(modeTracer)
	default:
		return nil
	}
}
