package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
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
	m.saveCurrent()
	m.savePlannedTraces()
	return tea.Quit
}

// handleCopyKey copies selected or current history items to the clipboard.
func (m *model) handleCopyKey() tea.Cmd {
	var idxs []int
	for i, it := range m.history.items {
		if it.isSelected != nil && *it.isSelected {
			idxs = append(idxs, i)
		}
	}
	if len(idxs) > 0 {
		sort.Ints(idxs)
		var parts []string
		items := m.history.list.Items()
		for _, i := range idxs {
			if i >= 0 && i < len(items) {
				hi := items[i].(historyItem)
				txt := hi.payload
				if hi.kind != "log" {
					txt = fmt.Sprintf("%s: %s", hi.topic, hi.payload)
				}
				parts = append(parts, txt)
			}
		}
		if err := clipboard.WriteAll(strings.Join(parts, "\n")); err != nil {
			m.appendHistory("", err.Error(), "log", err.Error())
		}
	} else if len(m.history.list.Items()) > 0 {
		idx := m.history.list.Index()
		if idx >= 0 {
			hi := m.history.list.Items()[idx].(historyItem)
			text := hi.payload
			if hi.kind != "log" {
				text = fmt.Sprintf("%s: %s", hi.topic, hi.payload)
			}
			if err := clipboard.WriteAll(text); err != nil {
				m.appendHistory("", err.Error(), "log", err.Error())
			}
		}
	}
	return nil
}

// handleDisconnectKey disconnects from the active broker.
func (m *model) handleDisconnectKey() tea.Cmd {
	if m.mqttClient != nil {
		m.mqttClient.Disconnect()
		m.connections.manager.Statuses[m.connections.active] = "disconnected"
		m.connections.manager.Errors[m.connections.active] = ""
		m.refreshConnectionItems()
		m.connections.connection = ""
		m.connections.active = ""
		m.mqttClient = nil
	}
	return nil
}

// handleHistoryFilterKey opens the history filter when focused on history.
func (m *model) handleHistoryFilterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
		return m.startHistoryFilter()
	}
	return nil
}

// handleClearFilterKey clears any active history filters.
func (m *model) handleClearFilterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	m.history.filterQuery = ""
	m.history.list.FilterInput.SetValue("")
	m.history.list.SetFilterState(list.Unfiltered)
	var msgs []Message
	if m.history.showArchived {
		msgs = m.history.store.SearchArchived(nil, time.Time{}, time.Time{}, "")
	} else {
		msgs = m.history.store.Search(nil, time.Time{}, time.Time{}, "")
	}
	m.history.items = make([]historyItem, len(msgs))
	items := make([]list.Item, len(msgs))
	for i, mm := range msgs {
		hi := historyItem{timestamp: mm.Timestamp, topic: mm.Topic, payload: mm.Payload, kind: mm.Kind, archived: mm.Archived}
		m.history.items[i] = hi
		items[i] = hi
	}
	m.history.list.SetItems(items)
	return nil
}

// handleSpaceKey toggles selection in history.
func (m *model) handleSpaceKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		idx := m.history.list.Index()
		if idx >= 0 && idx < len(m.history.items) {
			if m.history.items[idx].isSelected != nil && *m.history.items[idx].isSelected {
				m.history.items[idx].isSelected = nil
			} else {
				v := true
				m.history.items[idx].isSelected = &v
			}
			m.history.selectionAnchor = idx
		}
	}
	return nil
}

// handleShiftUpKey extends selection upward in history.
func (m *model) handleShiftUpKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if m.history.selectionAnchor == -1 {
			m.history.selectionAnchor = m.history.list.Index()
			if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
				v := true
				m.history.items[m.history.selectionAnchor].isSelected = &v
			}
		}
		if m.history.list.Index() > 0 {
			m.history.list.CursorUp()
			idx := m.history.list.Index()
			m.updateSelectionRange(idx)
		}
	}
	return nil
}

// handleShiftDownKey extends selection downward in history.
func (m *model) handleShiftDownKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if m.history.selectionAnchor == -1 {
			m.history.selectionAnchor = m.history.list.Index()
			if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
				v := true
				m.history.items[m.history.selectionAnchor].isSelected = &v
			}
		}
		if m.history.list.Index() < len(m.history.list.Items())-1 {
			m.history.list.CursorDown()
			idx := m.history.list.Index()
			m.updateSelectionRange(idx)
		}
	}
	return nil
}

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

// handleLeftKey moves topic selection left.
func (m *model) handleLeftKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected - 1 + len(m.topics.items)) % len(m.topics.items)
		m.ensureTopicVisible()
	}
	return nil
}

// handleRightKey moves topic selection right.
func (m *model) handleRightKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected + 1) % len(m.topics.items)
		m.ensureTopicVisible()
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

// handleSelectAllKey selects or clears all history items.
func (m *model) handleSelectAllKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		allSelected := true
		for i := range m.history.items {
			if m.history.items[i].isSelected == nil || !*m.history.items[i].isSelected {
				allSelected = false
				break
			}
		}
		if allSelected {
			for i := range m.history.items {
				m.history.items[i].isSelected = nil
			}
			m.history.selectionAnchor = -1
		} else {
			for i := range m.history.items {
				v := true
				m.history.items[i].isSelected = &v
			}
			if len(m.history.items) > 0 {
				m.history.selectionAnchor = 0
			}
		}
	}
	return nil
}

// handleToggleArchiveKey toggles between active and archived history.
func (m *model) handleToggleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && m.history.store != nil {
		m.history.showArchived = !m.history.showArchived
		var msgs []Message
		if m.history.showArchived {
			msgs = m.history.store.SearchArchived(nil, time.Time{}, time.Time{}, "")
		} else {
			msgs = m.history.store.Search(nil, time.Time{}, time.Time{}, "")
		}
		m.history.items = make([]historyItem, len(msgs))
		items := make([]list.Item, len(msgs))
		for i, mm := range msgs {
			hi := historyItem{timestamp: mm.Timestamp, topic: mm.Topic, payload: mm.Payload, kind: mm.Kind, archived: mm.Archived}
			m.history.items[i] = hi
			items[i] = hi
		}
		m.history.list.SetItems(items)
		if len(items) > 0 {
			m.history.list.Select(len(items) - 1)
		} else {
			m.history.list.Select(-1)
		}
	}
	return nil
}

// handleScrollKeys scrolls history or topic lists.
func (m *model) handleScrollKeys(key string) tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
		// keep current selection and anchor
		return nil
	}
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics {
		delta := -1
		if key == "down" || key == "j" {
			delta = 1
		}
		m.scrollTopics(delta)
	}
	return nil
}

// handlePublishKey publishes the current message to subscribed topics.
func (m *model) handlePublishKey() tea.Cmd {
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
	return nil
}

// handleEnterKey handles Enter for topic input and topic toggling.
func (m *model) handleEnterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopic {
		topic := strings.TrimSpace(m.topics.input.Value())
		if topic != "" && !m.hasTopic(topic) {
			m.topics.items = append(m.topics.items, topicItem{title: topic, subscribed: true})
			m.sortTopics()
			if m.currentMode() == modeTopics {
				m.rebuildActiveTopicList()
			}
			if m.mqttClient != nil {
				m.mqttClient.Subscribe(topic, 0, nil)
			}
			m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
			m.topics.input.SetValue("")
		}
	} else if m.ui.focusOrder[m.ui.focusIndex] == idTopics && m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		m.toggleTopic(m.topics.selected)
		m.ensureTopicVisible()
		if m.currentMode() == modeTopics {
			m.rebuildActiveTopicList()
		}
	}
	return nil
}

// handleArchiveKey archives selected history messages.
func (m *model) handleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if len(m.history.items) == 0 {
			return nil
		}
		archived := false
		for i := len(m.history.items) - 1; i >= 0; i-- {
			it := m.history.items[i]
			if it.isSelected != nil && *it.isSelected {
				key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
				if m.history.store != nil {
					_ = m.history.store.Archive(key)
				}
				m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
				archived = true
			}
		}
		if !archived {
			idx := m.history.list.Index()
			if idx >= 0 && idx < len(m.history.items) {
				it := m.history.items[idx]
				key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
				if m.history.store != nil {
					_ = m.history.store.Archive(key)
				}
				m.history.items = append(m.history.items[:idx], m.history.items[idx+1:]...)
			}
		}
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			it.isSelected = nil
			m.history.items[i] = it
			items[i] = it
		}
		m.history.list.SetItems(items)
		if len(m.history.items) == 0 {
			m.history.list.Select(-1)
		} else if m.history.list.Index() >= len(m.history.items) {
			m.history.list.Select(len(m.history.items) - 1)
		}
		m.history.selectionAnchor = -1
	}
	return nil
}

// handleDeleteKey deletes selected history messages or topics.
func (m *model) handleDeleteKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if len(m.history.items) == 0 {
			return nil
		}
		hasSelection := false
		for i := range m.history.items {
			if m.history.items[i].isSelected != nil && *m.history.items[i].isSelected {
				v := true
				m.history.items[i].isMarkedForDeletion = &v
				hasSelection = true
			}
		}
		if !hasSelection {
			idx := m.history.list.Index()
			if idx >= 0 && idx < len(m.history.items) {
				v := true
				m.history.items[idx].isMarkedForDeletion = &v
			}
		}
		m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
		m.startConfirm("Delete selected messages? [y/n]", "", func() {
			for i := len(m.history.items) - 1; i >= 0; i-- {
				it := m.history.items[i]
				if it.isMarkedForDeletion != nil && *it.isMarkedForDeletion {
					key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
					if m.history.store != nil {
						_ = m.history.store.Delete(key)
					}
					m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
				}
			}
			items := make([]list.Item, len(m.history.items))
			for i, it := range m.history.items {
				it.isSelected = nil
				it.isMarkedForDeletion = nil
				m.history.items[i] = it
				items[i] = it
			}
			m.history.list.SetItems(items)
			if len(m.history.items) == 0 {
				m.history.list.Select(-1)
			} else if m.history.list.Index() >= len(m.history.items) {
				m.history.list.Select(len(m.history.items) - 1)
			}
			m.history.selectionAnchor = -1
		})
		m.confirmCancel = func() {
			for i := range m.history.items {
				m.history.items[i].isMarkedForDeletion = nil
			}
		}
	} else if m.ui.focusOrder[m.ui.focusIndex] == idTopics && m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		idx := m.topics.selected
		name := m.topics.items[idx].title
		m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
		m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
			m.removeTopic(idx)
			if m.currentMode() == modeTopics {
				m.rebuildActiveTopicList()
			}
		})
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
		m.savePlannedTraces()
		return m.setMode(modeConnections)
	case "ctrl+t":
		m.topics.panes.subscribed = paneState{sel: 0, page: 0, index: 0, m: m}
		m.topics.panes.unsubscribed = paneState{sel: 0, page: 0, index: 1, m: m}
		m.topics.panes.active = 0
		m.topics.list.SetSize(m.ui.width/2-2, m.ui.height-4)
		m.rebuildActiveTopicList()
		return m.setMode(modeTopics)
	case "ctrl+p":
		items := []list.Item{}
		for _, pld := range m.message.payloads {
			items = append(items, payloadItem{topic: pld.topic, payload: pld.payload})
		}
		m.message.list = list.New(items, list.NewDefaultDelegate(), m.ui.width-4, m.ui.height-4)
		m.message.list.DisableQuitKeybindings()
		m.message.list.SetShowTitle(false)
		return m.setMode(modePayloads)
	case "ctrl+r":
		m.traces.list.SetSize(m.ui.width-4, m.ui.height-4)
		return m.setMode(modeTracer)
	default:
		return nil
	}
}
