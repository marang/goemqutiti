package emqutiti

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/topics"
)

// logTopicAction appends a log entry for a topic action.
// action should be "subscribe" or "unsubscribe".
func (m *model) logTopicAction(topic, action string, err error) {
	if len(action) == 0 {
		return
	}
	act := strings.ToUpper(action[:1]) + action[1:]
	if err != nil {
		m.history.Append(topic, "", "log", fmt.Sprintf("%s error for %s: %v", act, topic, err))
		return
	}
	switch action {
	case "subscribe":
		m.history.Append(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
	case "unsubscribe":
		m.history.Append(topic, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", topic))
	}
}

// handleTopicToggle subscribes or unsubscribes from a topic and logs the action.
// If the MQTT client is nil, the action is logged as an error.
func (m *model) handleTopicToggle(msg topics.ToggleMsg) tea.Cmd {
	action := "unsubscribe"
	if msg.Subscribed {
		action = "subscribe"
	}

	if m.mqttClient == nil {
		m.logTopicAction(msg.Topic, action, fmt.Errorf("no mqtt client"))
		return nil
	}

	var err error
	if msg.Subscribed {
		err = m.mqttClient.Subscribe(msg.Topic, 0, nil)
	} else {
		err = m.mqttClient.Unsubscribe(msg.Topic)
	}
	m.logTopicAction(msg.Topic, action, err)
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

	if m.CurrentMode() != constants.ModeConfirmDelete {
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
