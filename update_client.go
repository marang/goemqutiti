package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/goemqutiti/ui"
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

// handleClientMouse processes mouse events in client mode.
func (m *model) handleClientMouse(msg tea.MouseMsg) tea.Cmd {
	var cmds []tea.Cmd
	if msg.Action == tea.MouseActionPress && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
		if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
			var hCmd tea.Cmd
			m.history.list, hCmd = m.history.list.Update(msg)
			cmds = append(cmds, hCmd)
		} else if m.ui.focusOrder[m.ui.focusIndex] == idTopics {
			delta := -1
			if msg.Button == tea.MouseButtonWheelDown {
				delta = 1
			}
			m.scrollTopics(delta)
		}
		return tea.Batch(cmds...)
	}
	if msg.Type == tea.MouseLeft {
		cmds = append(cmds, m.focusFromMouse(msg.Y))
		if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
			idx := m.historyIndexAt(msg.Y)
			if idx >= 0 {
				m.history.list.Select(idx)
				if msg.Shift {
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
	switch t := msg.(type) {
	case statusMessage:
		return m.handleStatusMessage(t)
	case MQTTMessage:
		return m.handleMQTTMessage(t)
	case tea.KeyMsg:
		cmd := m.handleClientKey(t)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tea.MouseMsg:
		cmd := m.handleClientMouse(t)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if m.currentMode() == modeConfirmDelete {
		cmds = append(cmds, listenStatus(m.connections.statusChan))
		return tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	m.topics.input, cmd = m.topics.input.Update(msg)
	cmds = append(cmds, cmd)
	var cmdMsg tea.Cmd
	m.message.input, cmdMsg = m.message.input.Update(msg)
	cmds = append(cmds, cmdMsg)
	var vpCmd tea.Cmd
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
	if !skipVP {
		m.ui.viewport, vpCmd = m.ui.viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	var histCmd tea.Cmd
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
		m.history.list, histCmd = m.history.list.Update(msg)
		cmds = append(cmds, histCmd)
	}

	if st := m.history.list.FilterState(); st == list.Filtering || st == list.FilterApplied {
		q := m.history.list.FilterInput.Value()
		topics, start, end, text := parseHistoryQuery(q)
		m.history.filterQuery = q
		var msgs []Message
		if m.history.showArchived {
			msgs = m.history.store.SearchArchived(topics, start, end, text)
		} else {
			msgs = m.history.store.Search(topics, start, end, text)
		}
		_, items := messagesToHistoryItems(msgs)
		m.history.list.SetItems(items)
	} else if m.history.filterQuery != "" {
		topics, start, end, text := parseHistoryQuery(m.history.filterQuery)
		var msgs []Message
		if m.history.showArchived {
			msgs = m.history.store.SearchArchived(topics, start, end, text)
		} else {
			msgs = m.history.store.Search(topics, start, end, text)
		}
		_, items := messagesToHistoryItems(msgs)
		m.history.list.SetItems(items)
	} else {
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			items[i] = it
		}
		m.history.list.SetItems(items)
	}
	cmds = append(cmds, listenStatus(m.connections.statusChan))
	if m.mqttClient != nil {
		cmds = append(cmds, listenMessages(m.mqttClient.MessageChan))
	}
	return tea.Batch(cmds...)
}
