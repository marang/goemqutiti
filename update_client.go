package emqutiti

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/topics"
)

// handleTopicToggle subscribes or unsubscribes from a topic and logs the action.
func (m *model) handleTopicToggle(msg topics.ToggleMsg) tea.Cmd {
	if m.mqttClient != nil {
		if msg.Subscribed {
			if err := m.mqttClient.Subscribe(msg.Topic, 0, nil); err != nil {
				m.history.Append(msg.Topic, "", "log", fmt.Sprintf("Subscribe error for %s: %v", msg.Topic, err))
			} else {
				m.history.Append(msg.Topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", msg.Topic))
			}
		} else {
			if err := m.mqttClient.Unsubscribe(msg.Topic); err != nil {
				m.history.Append(msg.Topic, "", "log", fmt.Sprintf("Unsubscribe error for %s: %v", msg.Topic, err))
			} else {
				m.history.Append(msg.Topic, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", msg.Topic))
			}
		}
	}
	return nil
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

// handleClientMsg dispatches client messages and returns a command.
// The boolean indicates if processing should stop after the command.
func (m *model) handleClientMsg(msg tea.Msg) (tea.Cmd, bool) {
	switch t := msg.(type) {
	case connections.StatusMessage:
		return m.handleStatusMessage(t), true
	case MQTTMessage:
		return m.handleMQTTMessage(t), true
	case tea.KeyMsg:
		return HandleClientKey(m, t), false
	case tea.MouseMsg:
		return m.handleClientMouse(t), false
	}
	return nil, false
}
