package main

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
	m.appendHistory("", string(msg), "log", string(msg))
	if strings.HasPrefix(string(msg), "Connected") && m.connections.active != "" {
		m.connections.manager.Statuses[m.connections.active] = "connected"
		m.refreshConnectionItems()
		m.subscribeActiveTopics()
	} else if strings.HasPrefix(string(msg), "Connection lost") && m.connections.active != "" {
		m.connections.manager.Statuses[m.connections.active] = "disconnected"
		m.refreshConnectionItems()
	}
	return listenStatus(m.connections.statusChan)
}

// handleMQTTMessage appends received MQTT messages to history.
func (m *model) handleMQTTMessage(msg MQTTMessage) tea.Cmd {
	m.appendHistory(msg.Topic, msg.Payload, "sub", fmt.Sprintf("Received on %s: %s", msg.Topic, msg.Payload))
	return listenMessages(m.mqttClient.MessageChan)
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
	if m.topics.selected < 0 || m.topics.selected >= len(m.topics.items) {
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
	if m.topics.selected >= len(bounds) {
		return
	}
	b := bounds[m.topics.selected]
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
		m.handleTopicsClick(msg)
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// handleTopicsClick processes mouse events within the topics area.
// The mouse coordinates are adjusted for the viewport offset and compared
// against precomputed chip bounds.
func (m *model) handleTopicsClick(msg tea.MouseMsg) {
	y := msg.Y + m.ui.viewport.YOffset
	idx := m.topicAtPosition(msg.X, y)
	if idx < 0 {
		return
	}
	m.topics.selected = idx
	if msg.Type == tea.MouseLeft {
		m.toggleTopic(idx)
		if m.currentMode() == modeTopics {
			m.rebuildActiveTopicList()
		}
	} else if msg.Type == tea.MouseRight {
		name := m.topics.items[idx].title
		m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
		m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
			m.removeTopic(idx)
			if m.currentMode() == modeTopics {
				m.rebuildActiveTopicList()
			}
		})
	}
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
	cmds := []tea.Cmd{listenStatus(m.connections.statusChan)}
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
	var cmd tea.Cmd
	m.topics.input, cmd = m.topics.input.Update(msg)
	cmds = append(cmds, cmd)
	m.message.input, cmd = m.message.input.Update(msg)
	cmds = append(cmds, cmd)
	if vpCmd := m.updateViewport(msg); vpCmd != nil {
		cmds = append(cmds, vpCmd)
	}
	if histCmd := m.updateHistoryList(msg); histCmd != nil {
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

// updateHistoryList updates the history list when focused.
func (m *model) updateHistoryList(msg tea.Msg) tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	var cmd tea.Cmd
	m.history.list, cmd = m.history.list.Update(msg)
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
