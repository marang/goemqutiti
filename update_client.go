package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goemqutiti/history"
	"goemqutiti/ui"
)

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

func (m *model) handleMQTTMessage(msg MQTTMessage) tea.Cmd {
	m.appendHistory(msg.Topic, msg.Payload, "sub", fmt.Sprintf("Received on %s: %s", msg.Topic, msg.Payload))
	return listenMessages(m.mqttClient.MessageChan)
}

func (m *model) scrollTopics(delta int) {
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	if delta > 0 {
		m.topics.vp.ScrollDown(delta * rowH)
	} else if delta < 0 {
		m.topics.vp.ScrollUp(-delta * rowH)
	}
}

func (m *model) ensureTopicVisible() {
	if m.topics.selected < 0 || m.topics.selected >= len(m.topics.items) {
		return
	}
	var chips []string
	for _, t := range m.topics.items {
		st := ui.ChipStyle
		if !t.active {
			st = ui.ChipInactive
		}
		chips = append(chips, st.Render(t.title))
	}
	_, bounds := layoutChips(chips, m.ui.width-4)
	if m.topics.selected >= len(bounds) {
		return
	}
	b := bounds[m.topics.selected]
	if b.y < m.topics.vp.YOffset {
		m.topics.vp.SetYOffset(b.y)
	} else if b.y+b.h > m.topics.vp.YOffset+m.topics.vp.Height {
		m.topics.vp.SetYOffset(b.y + b.h - m.topics.vp.Height)
	}
}

func (m *model) handleClientKey(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg.String() {
	case "ctrl+d":
		m.saveCurrent()
		return tea.Quit
	case "ctrl+c":
		if len(m.history.selected) > 0 {
			var idxs []int
			for i := range m.history.selected {
				idxs = append(idxs, i)
			}
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
			clipboard.WriteAll(strings.Join(parts, "\n"))
		} else if len(m.history.list.Items()) > 0 {
			idx := m.history.list.Index()
			if idx >= 0 {
				hi := m.history.list.Items()[idx].(historyItem)
				text := hi.payload
				if hi.kind != "log" {
					text = fmt.Sprintf("%s: %s", hi.topic, hi.payload)
				}
				clipboard.WriteAll(text)
			}
		}
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
	case "space":
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			idx := m.history.list.Index()
			if idx >= 0 {
				if _, ok := m.history.selected[idx]; ok {
					delete(m.history.selected, idx)
				} else {
					m.history.selected[idx] = struct{}{}
				}
				m.history.selectionAnchor = idx
			}
		}
	case "shift+up":
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			if m.history.selectionAnchor == -1 {
				m.history.selectionAnchor = m.history.list.Index()
				m.history.selected[m.history.selectionAnchor] = struct{}{}
			}
			if m.history.list.Index() > 0 {
				m.history.list.CursorUp()
				idx := m.history.list.Index()
				m.updateSelectionRange(idx)
			}
		}
	case "shift+down":
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			if m.history.selectionAnchor == -1 {
				m.history.selectionAnchor = m.history.list.Index()
				m.history.selected[m.history.selectionAnchor] = struct{}{}
			}
			if m.history.list.Index() < len(m.history.list.Items())-1 {
				m.history.list.CursorDown()
				idx := m.history.list.Index()
				m.updateSelectionRange(idx)
			}
		}
	case "tab", "shift+tab":
		if len(m.ui.focusOrder) > 0 {
			step := 1
			if msg.String() == "shift+tab" {
				step = -1
			}
			next := (m.ui.focusIndex + step + len(m.ui.focusOrder)) % len(m.ui.focusOrder)
			id := m.ui.focusOrder[next]
			cmds = append(cmds, m.setFocus(id))
			if id == "topics" {
				if len(m.topics.items) > 0 {
					m.topics.selected = 0
					m.ensureTopicVisible()
				} else {
					m.topics.selected = -1
				}
			}
		}
	case "left":
		if m.ui.focusOrder[m.ui.focusIndex] == "topics" && len(m.topics.items) > 0 {
			m.topics.selected = (m.topics.selected - 1 + len(m.topics.items)) % len(m.topics.items)
			m.ensureTopicVisible()
		}
	case "right":
		if m.ui.focusOrder[m.ui.focusIndex] == "topics" && len(m.topics.items) > 0 {
			m.topics.selected = (m.topics.selected + 1) % len(m.topics.items)
			m.ensureTopicVisible()
		}
	case "ctrl+shift+up":
		id := m.ui.focusOrder[m.ui.focusIndex]
		if id == "message" {
			if m.layout.message.height > 1 {
				m.layout.message.height--
				m.message.input.SetHeight(m.layout.message.height)
			}
		} else if id == "history" {
			if m.layout.history.height > 1 {
				m.layout.history.height--
				m.history.list.SetSize(m.ui.width-4, m.layout.history.height)
			}
		} else if id == "topics" {
			if m.layout.topics.height > 1 {
				m.layout.topics.height--
			}
		}
	case "ctrl+shift+down":
		id := m.ui.focusOrder[m.ui.focusIndex]
		if id == "message" {
			m.layout.message.height++
			m.message.input.SetHeight(m.layout.message.height)
		} else if id == "history" {
			m.layout.history.height++
			m.history.list.SetSize(m.ui.width-4, m.layout.history.height)
		} else if id == "topics" {
			m.layout.topics.height++
		}
	case "up", "down":
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			// keep current selection and anchor
		} else if m.ui.focusOrder[m.ui.focusIndex] == "topics" {
			delta := -1
			if msg.String() == "down" {
				delta = 1
			}
			m.scrollTopics(delta)
		}
	case "ctrl+s", "ctrl+enter":
		if m.ui.focusOrder[m.ui.focusIndex] == "message" {
			payload := m.message.input.Value()
			for _, t := range m.topics.items {
				if t.active {
					m.message.payloads = append(m.message.payloads, payloadItem{topic: t.title, payload: payload})
					m.appendHistory(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
					if m.mqttClient != nil {
						m.mqttClient.Publish(t.title, 0, false, payload)
					}
				}
			}
		}
	case "enter":
		if m.ui.focusOrder[m.ui.focusIndex] == "topic" {
			topic := strings.TrimSpace(m.topics.input.Value())
			if topic != "" && !m.hasTopic(topic) {
				m.topics.items = append(m.topics.items, topicItem{title: topic, active: true})
				m.sortTopics()
				if m.mqttClient != nil {
					m.mqttClient.Subscribe(topic, 0, nil)
				}
				m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
				m.topics.input.SetValue("")
			}
		} else if m.ui.focusOrder[m.ui.focusIndex] == "topics" && m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
			m.toggleTopic(m.topics.selected)
			m.ensureTopicVisible()
		}
	case "d":
		if m.ui.focusOrder[m.ui.focusIndex] == "topics" && m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
			idx := m.topics.selected
			name := m.topics.items[idx].title
			m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), func() {
				m.removeTopic(idx)
			})
		}
	default:
		switch msg.String() {
		case "ctrl+b":
			if err := m.connections.manager.LoadProfiles(""); err != nil {
				m.appendHistory("", err.Error(), "log", err.Error())
			}
			m.refreshConnectionItems()
			m.saveCurrent()
			m.ui.mode = modeConnections
		case "ctrl+t":
			items := []list.Item{}
			for _, tpc := range m.topics.items {
				items = append(items, topicItem{title: tpc.title, active: tpc.active})
			}
			m.topics.list = list.New(items, list.NewDefaultDelegate(), m.ui.width-4, m.ui.height-4)
			m.topics.list.DisableQuitKeybindings()
			m.topics.list.SetShowTitle(false)
			m.ui.mode = modeTopics
		case "ctrl+p":
			items := []list.Item{}
			for _, pld := range m.message.payloads {
				items = append(items, payloadItem{topic: pld.topic, payload: pld.payload})
			}
			m.message.list = list.New(items, list.NewDefaultDelegate(), m.ui.width-4, m.ui.height-4)
			m.message.list.DisableQuitKeybindings()
			m.message.list.SetShowTitle(false)
			m.ui.mode = modePayloads
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m *model) handleClientMouse(msg tea.MouseMsg) tea.Cmd {
	var cmds []tea.Cmd
	if msg.Action == tea.MouseActionPress && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			var hCmd tea.Cmd
			m.history.list, hCmd = m.history.list.Update(msg)
			cmds = append(cmds, hCmd)
		} else if m.ui.focusOrder[m.ui.focusIndex] == "topics" {
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
		if m.ui.focusOrder[m.ui.focusIndex] == "history" {
			idx := m.historyIndexAt(msg.Y)
			if idx >= 0 {
				m.history.list.Select(idx)
				if msg.Shift {
					if m.history.selectionAnchor == -1 {
						m.history.selectionAnchor = m.history.list.Index()
						m.history.selected[m.history.selectionAnchor] = struct{}{}
					}
					m.updateSelectionRange(idx)
				} else {
					m.history.selected = map[int]struct{}{}
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
	} else if msg.Type == tea.MouseRight {
		name := m.topics.items[idx].title
		m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), func() {
			m.removeTopic(idx)
		})
	}
}

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

	var cmd tea.Cmd
	m.topics.input, cmd = m.topics.input.Update(msg)
	cmds = append(cmds, cmd)
	var cmdMsg tea.Cmd
	m.message.input, cmdMsg = m.message.input.Update(msg)
	cmds = append(cmds, cmdMsg)
	var vpCmd tea.Cmd
	skipVP := false
	if m.ui.focusOrder[m.ui.focusIndex] == "history" {
		switch mt := msg.(type) {
		case tea.KeyMsg:
			s := mt.String()
			if s == "up" || s == "down" || s == "pgup" || s == "pgdown" {
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
	if m.ui.focusOrder[m.ui.focusIndex] == "history" {
		m.history.list, histCmd = m.history.list.Update(msg)
		cmds = append(cmds, histCmd)
	}

	if m.history.list.FilterState() == list.Filtering {
		q := m.history.list.FilterInput.Value()
		topics, start, end, text := history.ParseQuery(q)
		msgs := m.history.store.Search(topics, start, end, text)
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
		}
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
