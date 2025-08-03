package emqutiti

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"
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

// scrollTopics scrolls the topics viewport by the given number of rows.
func (m *model) scrollTopics(delta int) {
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	if delta > 0 {
		m.topics.vp.ScrollDown(delta * rowH)
	} else if delta < 0 {
		m.topics.vp.ScrollUp(-delta * rowH)
	}
}

// ensureTopicVisible keeps the selected topic within the visible viewport.
func (m *model) ensureTopicVisible() {
	sel := m.topics.Selected()
	if sel < 0 || sel >= len(m.topics.items) {
		return
	}
	var chips []string
	for _, t := range m.topics.items {
		st := ui.ChipStyle
		if !t.subscribed {
			st = ui.ChipInactive
		}
		chips = append(chips, st.Render(t.title))
	}
	_, bounds := layoutChips(chips, m.ui.width-4)
	if sel >= len(bounds) {
		return
	}
	b := bounds[sel]
	if b.yPos < m.topics.vp.YOffset {
		m.topics.vp.SetYOffset(b.yPos)
	} else if b.yPos+b.height > m.topics.vp.YOffset+m.topics.vp.Height {
		m.topics.vp.SetYOffset(b.yPos + b.height - m.topics.vp.Height)
	}
}

// isHistoryFocused reports if the history list has focus.
func (m *model) isHistoryFocused() bool {
	return m.ui.focusOrder[m.ui.focusIndex] == idHistory
}

// isTopicsFocused reports if the topics view has focus.
func (m *model) isTopicsFocused() bool {
	return m.ui.focusOrder[m.ui.focusIndex] == idTopics
}

// historyScroll forwards scroll events to the history list.
func (m *model) historyScroll(msg tea.MouseMsg) tea.Cmd {
	var hCmd tea.Cmd
	m.history.list, hCmd = m.history.list.Update(msg)
	return hCmd
}

// topicsScroll adjusts the topics viewport based on the mouse wheel.
func (m *model) topicsScroll(msg tea.MouseMsg) {
	delta := -1
	if msg.Button == tea.MouseButtonWheelDown {
		delta = 1
	}
	m.scrollTopics(delta)
}

// handleMouseScroll processes scroll wheel events.
// It returns a command and a boolean indicating if the event was handled.
func (m *model) handleMouseScroll(msg tea.MouseMsg) (tea.Cmd, bool) {
	if msg.Action == tea.MouseActionPress && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
		if m.isHistoryFocused() && !m.history.showArchived {
			return m.historyScroll(msg), true
		}
		if m.isTopicsFocused() {
			m.topicsScroll(msg)
			return nil, true
		}
		return nil, true
	}
	return nil, false
}

// handleHistorySelection updates history selection based on index and shift key.
func (m *model) handleHistorySelection(idx int, shift bool) {
	m.history.list.Select(idx)
	if shift {
		if m.history.selectionAnchor == -1 {
			m.history.selectionAnchor = m.history.list.Index()
			if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
				v := true
				m.history.items[m.history.selectionAnchor].isSelected = &v
			}
		}
		m.updateSelectionRange(idx)
	} else {
		for i := range m.history.items {
			m.history.items[i].isSelected = nil
		}
		m.history.selectionAnchor = -1
	}
}

// handleMouseLeft manages left-click focus and selection.
func (m *model) handleMouseLeft(msg tea.MouseMsg) tea.Cmd {
	cmd := m.focusFromMouse(msg.Y)
	if m.isHistoryFocused() && !m.history.showArchived {
		m.handleHistoryClick(msg)
	}
	return cmd
}

// handleHistoryClick selects a history item based on the mouse position.
func (m *model) handleHistoryClick(msg tea.MouseMsg) {
	idx := m.historyIndexAt(msg.Y)
	if idx >= 0 {
		m.handleHistorySelection(idx, msg.Shift)
	}
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
		if cmd := m.handleTopicsClick(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// handleTopicsClick processes mouse events within the topics area.
// The mouse coordinates are adjusted for the viewport offset and compared
// against precomputed chip bounds.
func (m *model) handleTopicsClick(msg tea.MouseMsg) tea.Cmd {
	y := msg.Y + m.ui.viewport.YOffset
	idx := m.topics.TopicAtPosition(msg.X, y)
	if idx < 0 {
		return nil
	}
	m.topics.SetSelected(idx)
	if msg.Type == tea.MouseLeft {
		cmd := m.topics.ToggleTopic(idx)
		if m.currentMode() == modeTopics {
			m.topics.RebuildActiveTopicList()
		}
		return cmd
	} else if msg.Type == tea.MouseRight {
		name := m.topics.items[idx].title
		rf := func() tea.Cmd { return m.setFocus(m.ui.focusOrder[m.ui.focusIndex]) }
		m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", rf, func() tea.Cmd {
			cmd := m.topics.RemoveTopic(idx)
			if m.currentMode() == modeTopics {
				m.topics.RebuildActiveTopicList()
			}
			return cmd
		}, nil)
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
	if histCmd := m.history.Update(msg); histCmd != nil {
		cmds = append(cmds, histCmd)
	}
	return cmds
}

// updateViewport updates the main viewport unless history handles the scroll.
func (m *model) updateViewport(msg tea.Msg) tea.Cmd {
	skipVP := false
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
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
	if st := m.history.list.FilterState(); st == list.Filtering || st == list.FilterApplied {
		q := m.history.list.FilterInput.Value()
		var items []list.Item
		m.history.items, items = applyHistoryFilter(q, m.history.store, m.history.showArchived)
		m.history.filterQuery = q
		m.history.list.SetItems(items)
	} else if m.history.filterQuery != "" {
		var items []list.Item
		m.history.items, items = applyHistoryFilter(m.history.filterQuery, m.history.store, m.history.showArchived)
		m.history.list.SetItems(items)
	} else {
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			items[i] = it
		}
		m.history.list.SetItems(items)
	}
}
