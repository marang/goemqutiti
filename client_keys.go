package emqutiti

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleClientKey processes keyboard events in client mode.
func (m *model) HandleClientKey(msg tea.KeyMsg) tea.Cmd {
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
	m.connections.SaveCurrent(m.topics.Snapshot(), m.payloads.Snapshot())
	m.traces.SavePlannedTraces()
	return tea.Quit
}

// handleDisconnectKey disconnects from the active broker.
func (m *model) handleDisconnectKey() tea.Cmd {
	if m.mqttClient != nil {
		m.mqttClient.Disconnect()
		m.connections.SetDisconnected(m.connections.Active, "")
		m.connections.RefreshConnectionItems()
		m.connections.Connection = ""
		m.connections.Active = ""
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
		for _, t := range m.topics.Items {
			if t.Subscribed {
				m.payloads.Add(t.Name, payload)
				m.history.Append(t.Name, payload, "pub", fmt.Sprintf("Published to %s: %s", t.Name, payload))
				if m.mqttClient != nil {
					m.mqttClient.Publish(t.Name, 0, false, payload)
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
		if !m.history.ShowArchived() {
			return m.handleDeleteHistoryKey()
		}
	case idTopics:
		sel := m.topics.Selected()
		if sel >= 0 && sel < len(m.topics.Items) {
			return m.handleDeleteTopicKey()
		}
	}
	return nil
}
