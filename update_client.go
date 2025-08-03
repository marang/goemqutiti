package emqutiti

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// handleStatusMessage processes broker status updates.
func (m *model) handleStatusMessage(msg statusMessage) tea.Cmd {
	m.history.Append("", string(msg), "log", string(msg))
	if strings.HasPrefix(string(msg), "Connected") && m.connections.active != "" {
		m.connections.SetConnected(m.connections.active)
		m.connections.RefreshConnectionItems()
		m.connectionsAPI().SubscribeActiveTopics()
	} else if strings.HasPrefix(string(msg), "Connection lost") && m.connections.active != "" {
		m.connections.SetDisconnected(m.connections.active, "")
		m.connections.RefreshConnectionItems()
	}
	return m.connections.ListenStatus()
}

// handleMQTTMessage appends received MQTT messages to history.
func (m *model) handleMQTTMessage(msg MQTTMessage) tea.Cmd {
	m.history.Append(msg.Topic, msg.Payload, "sub", fmt.Sprintf("Received on %s: %s", msg.Topic, msg.Payload))
	return listenMessages(m.mqttClient.MessageChan)
}

// handleTopicToggle subscribes or unsubscribes from a topic and logs the action.
func (m *model) handleTopicToggle(msg topicToggleMsg) tea.Cmd {
	if m.mqttClient != nil {
		if msg.subscribed {
			m.mqttClient.Subscribe(msg.topic, 0, nil)
			m.history.Append(msg.topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", msg.topic))
		} else {
			m.mqttClient.Unsubscribe(msg.topic)
			m.history.Append(msg.topic, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", msg.topic))
		}
	}
	return nil
}

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

// updateClient updates the UI when in client mode.
func (m *model) updateClient(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	if cmd, done := m.handleClientMsg(msg); done {
		return cmd
	} else if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if m.currentMode() != modeConfirmDelete {
		cmds = append(cmds, m.updateClientInputs(msg)...)
		m.filterHistoryList()
	}

	cmds = append(cmds, m.updateClientStatus()...)
	return tea.Batch(cmds...)
}

// updateClientStatus returns commands to listen for connection and message updates.
func (m *model) updateClientStatus() []tea.Cmd {
	cmds := []tea.Cmd{m.connections.ListenStatus()}
	if m.mqttClient != nil {
		cmds = append(cmds, listenMessages(m.mqttClient.MessageChan))
	}
	return cmds
}

// handleClientMsg dispatches client messages and returns a command.
// The boolean indicates if processing should stop after the command.
func (m *model) handleClientMsg(msg tea.Msg) (tea.Cmd, bool) {
	switch t := msg.(type) {
	case statusMessage:
		return m.handleStatusMessage(t), true
	case MQTTMessage:
		return m.handleMQTTMessage(t), true
	case tea.KeyMsg:
		return m.handleClientKey(t), false
	case tea.MouseMsg:
		return m.handleClientMouse(t), false
	}
	return nil, false
}

// updateClientInputs updates form inputs, viewport and history list.
func (m *model) updateClientInputs(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	if cmd := m.topics.UpdateInput(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if mCmd := m.message.Update(msg); mCmd != nil {
		cmds = append(cmds, mCmd)
	}
	if vpCmd := m.updateViewport(msg); vpCmd != nil {
		cmds = append(cmds, vpCmd)
	}
	if m.FocusedID() == idHistory {
		if histCmd := m.history.Update(msg); histCmd != nil {
			cmds = append(cmds, histCmd)
		}
	}
	return cmds
}

// updateViewport updates the main viewport unless history handles the scroll.
func (m *model) updateViewport(msg tea.Msg) tea.Cmd {
	skipVP := false
	if m.FocusedID() == idHistory {
		switch mt := msg.(type) {
		case tea.KeyMsg:
			s := mt.String()
			if s == "up" || s == "down" || s == "pgup" || s == "pgdown" || s == "k" || s == "j" {
				skipVP = true
			}
		case tea.MouseMsg:
			if mt.Action == tea.MouseActionPress && (mt.Button == tea.MouseButtonWheelUp || mt.Button == tea.MouseButtonWheelDown) {
				skipVP = true
			}
		}
	}
	if skipVP {
		return nil
	}
	var cmd tea.Cmd
	m.ui.viewport, cmd = m.ui.viewport.Update(msg)
	return cmd
}

// filterHistoryList refreshes history items based on the current filter state.
func (m *model) filterHistoryList() {
	if st := m.history.List().FilterState(); st == list.Filtering || st == list.FilterApplied {
		q := m.history.List().FilterInput.Value()
		hitems, litems := history.ApplyFilter(q, m.history.Store(), m.history.ShowArchived())
		m.history.SetItems(hitems)
		m.history.SetFilterQuery(q)
		m.history.List().SetItems(litems)
	} else if m.history.FilterQuery() != "" {
		hitems, litems := history.ApplyFilter(m.history.FilterQuery(), m.history.Store(), m.history.ShowArchived())
		m.history.SetItems(hitems)
		m.history.List().SetItems(litems)
	} else {
		items := make([]list.Item, len(m.history.Items()))
		for i, it := range m.history.Items() {
			items[i] = it
		}
		m.history.List().SetItems(items)
	}
}
