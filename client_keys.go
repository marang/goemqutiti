package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleClientKey processes keyboard events in client mode.
func (m *model) handleClientKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+d":
		m.saveCurrent()
		m.savePlannedTraces()
		return tea.Quit
	case "ctrl+c":
		return m.handleCopyKey()
	case "ctrl+x":
		if m.mqttClient != nil {
			m.mqttClient.Disconnect()
			m.connections.manager.Statuses[m.connections.active] = "disconnected"
			m.connections.manager.Errors[m.connections.active] = ""
			m.refreshConnectionItems()
			m.connections.connection = ""
			m.connections.active = ""
			m.mqttClient = nil
		}
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
		return m.handleScrollKeys(msg)
	case "ctrl+s", "ctrl+enter":
		if m.ui.focusOrder[m.ui.focusIndex] == idMessage {
			payload := m.message.input.Value()
			for _, t := range m.topics.items {
				if t.subscribed {
					m.message.payloads = append(m.message.payloads, payloadItem{topic: t.title, payload: payload})
					m.appendHistory(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
					if m.mqttClient != nil {
						m.mqttClient.Publish(t.title, 0, false, payload)
					}
				}
			}
		}
	case "enter":
		return m.handleEnterKey()
	case "a":
		return m.handleArchiveKey()
	case "delete":
		return m.handleDeleteKey()
	default:
		return m.handleModeSwitchKey(msg)
	}
	return nil
}

// handleScrollKeys dispatches scroll keys based on focus.
func (m *model) handleScrollKeys(msg tea.KeyMsg) tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
		return m.handleHistoryScrollKeys(msg)
	}
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics {
		return m.handleTopicsScrollKeys(msg)
	}
	return nil
}

// handleDeleteKey dispatches delete key handling based on focus.
func (m *model) handleDeleteKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		return m.handleHistoryDeleteKey()
	}
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics {
		return m.handleTopicsDeleteKey()
	}
	return nil
}
