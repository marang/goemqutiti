package emqutiti

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleClientKey processes keyboard events in client mode.
func (m *model) handleClientKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+d":
		return m.handleQuitKey()
	case "ctrl+c":
		return m.handleCopyKey()
	case "ctrl+x":
		return m.handleDisconnectKey()
	case "/":
		return m.handleHistoryFilterKey()
	case "ctrl+f":
		return m.handleClearFilterKey()
	case "space":
		return m.handleSpaceKey()
	case "shift+up":
		return m.handleShiftUpKey()
	case "shift+down":
		return m.handleShiftDownKey()
	case "tab":
		return m.handleTabKey()
	case "shift+tab":
		return m.handleShiftTabKey()
	case "left":
		return m.handleLeftKey()
	case "right":
		return m.handleRightKey()
	case "ctrl+shift+up":
		return m.handleResizeUpKey()
	case "ctrl+shift+down":
		return m.handleResizeDownKey()
	case "ctrl+a":
		return m.handleSelectAllKey()
	case "ctrl+l":
		return m.handleToggleArchiveKey()
	case "up", "down", "k", "j":
		return m.handleScrollKeys(msg.String())
	case "ctrl+s", "ctrl+enter":
		return m.handlePublishKey()
	case "enter":
		return m.handleEnterKey()
	case "a":
		return m.handleArchiveKey()
	case "delete":
		return m.handleDeleteKey()
	default:
		return m.handleModeSwitchKey(msg)
	}
}

// handleQuitKey saves current state and quits the application.
func (m *model) handleQuitKey() tea.Cmd {
	m.history.SaveCurrent()
	m.traces.savePlannedTraces()
	return tea.Quit
}

// handleDisconnectKey disconnects from the active broker.
func (m *model) handleDisconnectKey() tea.Cmd {
	if m.mqttClient != nil {
		m.mqttClient.Disconnect()
		m.connections.SetDisconnected(m.connections.active, "")
		m.refreshConnectionItems()
		m.connections.connection = ""
		m.connections.active = ""
		m.mqttClient = nil
	}
	return nil
}

// handleScrollKeys dispatches scroll events based on focus.
func (m *model) handleScrollKeys(key string) tea.Cmd {
	switch m.ui.focusOrder[m.ui.focusIndex] {
	case idHistory:
		return m.handleHistoryScroll(key)
	case idTopics:
		return m.handleTopicScroll(key)
	default:
		return nil
	}
}

// handlePublishKey publishes the current message to subscribed topics.
func (m *model) handlePublishKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idMessage {
		payload := m.message.input.Value()
		for _, t := range m.topics.items {
			if t.subscribed {
				m.payloads.Add(t.title, payload)
				m.history.Append(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
				if m.mqttClient != nil {
					m.mqttClient.Publish(t.title, 0, false, payload)
				}
			}
		}
	}
	return nil
}

// handleDeleteKey dispatches deletion based on focus.
func (m *model) handleDeleteKey() tea.Cmd {
	switch m.ui.focusOrder[m.ui.focusIndex] {
	case idHistory:
		if !m.history.showArchived {
			return m.handleDeleteHistoryKey()
		}
	case idTopics:
		sel := m.topics.Selected()
		if sel >= 0 && sel < len(m.topics.items) {
			return m.handleDeleteTopicKey()
		}
	}
	return nil
}
