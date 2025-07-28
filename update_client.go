package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"goemqutiti/history"
)

func (m *model) handleStatusMessage(msg statusMessage) tea.Cmd {
	m.appendHistory("", string(msg), "log", string(msg))
	if strings.HasPrefix(string(msg), "Connected") && m.activeConn != "" {
		m.connections.Statuses[m.activeConn] = "connected"
		m.refreshConnectionItems()
		m.subscribeActiveTopics()
	} else if strings.HasPrefix(string(msg), "Connection lost") && m.activeConn != "" {
		m.connections.Statuses[m.activeConn] = "disconnected"
		m.refreshConnectionItems()
	}
	return listenStatus(m.statusChan)
}

func (m *model) handleMQTTMessage(msg MQTTMessage) tea.Cmd {
	m.appendHistory(msg.Topic, msg.Payload, "sub", fmt.Sprintf("Received on %s: %s", msg.Topic, msg.Payload))
	return listenMessages(m.mqttClient.MessageChan)
}

func (m *model) handleClientKey(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg.String() {
	case "ctrl+d":
		m.saveCurrent()
		return tea.Quit
	case "ctrl+c":
		if len(m.selectedHistory) > 0 {
			var idxs []int
			for i := range m.selectedHistory {
				idxs = append(idxs, i)
			}
			sort.Ints(idxs)
			var parts []string
			items := m.history.Items()
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
		} else if len(m.history.Items()) > 0 {
			idx := m.history.Index()
			if idx >= 0 {
				hi := m.history.Items()[idx].(historyItem)
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
			m.connections.Statuses[m.activeConn] = "disconnected"
			m.connections.Errors[m.activeConn] = ""
			m.refreshConnectionItems()
			m.connection = ""
			m.activeConn = ""
			m.mqttClient = nil
		}
	case "space":
		if m.focusOrder[m.focusIndex] == "history" {
			idx := m.history.Index()
			if idx >= 0 {
				if _, ok := m.selectedHistory[idx]; ok {
					delete(m.selectedHistory, idx)
				} else {
					m.selectedHistory[idx] = struct{}{}
				}
				m.selectionAnchor = idx
			}
		}
	case "shift+up":
		if m.focusOrder[m.focusIndex] == "history" {
			if m.selectionAnchor == -1 {
				m.selectionAnchor = m.history.Index()
				m.selectedHistory[m.selectionAnchor] = struct{}{}
			}
			if m.history.Index() > 0 {
				m.history.CursorUp()
				idx := m.history.Index()
				m.updateSelectionRange(idx)
			}
		}
	case "shift+down":
		if m.focusOrder[m.focusIndex] == "history" {
			if m.selectionAnchor == -1 {
				m.selectionAnchor = m.history.Index()
				m.selectedHistory[m.selectionAnchor] = struct{}{}
			}
			if m.history.Index() < len(m.history.Items())-1 {
				m.history.CursorDown()
				idx := m.history.Index()
				m.updateSelectionRange(idx)
			}
		}
	case "tab", "shift+tab":
		if len(m.focusOrder) > 0 {
			step := 1
			if msg.String() == "shift+tab" {
				step = -1
			}
			next := (m.focusIndex + step + len(m.focusOrder)) % len(m.focusOrder)
			id := m.focusOrder[next]
			cmds = append(cmds, m.setFocus(id))
			if id == "topics" {
				if len(m.topics) > 0 {
					m.selectedTopic = 0
				} else {
					m.selectedTopic = -1
				}
			}
		}
	case "left":
		if m.focusOrder[m.focusIndex] == "topics" && len(m.topics) > 0 {
			m.selectedTopic = (m.selectedTopic - 1 + len(m.topics)) % len(m.topics)
		}
	case "right":
		if m.focusOrder[m.focusIndex] == "topics" && len(m.topics) > 0 {
			m.selectedTopic = (m.selectedTopic + 1) % len(m.topics)
		}
	case "ctrl+shift+up":
		id := m.focusOrder[m.focusIndex]
		if id == "message" {
			if m.messageHeight > 1 {
				m.messageHeight--
				m.messageInput.SetHeight(m.messageHeight)
			}
		} else if id == "history" {
			if m.historyHeight > 1 {
				m.historyHeight--
				m.history.SetSize(m.width-4, m.historyHeight)
			}
		}
	case "ctrl+shift+down":
		id := m.focusOrder[m.focusIndex]
		if id == "message" {
			m.messageHeight++
			m.messageInput.SetHeight(m.messageHeight)
		} else if id == "history" {
			m.historyHeight++
			m.history.SetSize(m.width-4, m.historyHeight)
		}
	case "up", "down":
		if m.focusOrder[m.focusIndex] == "history" {
			// keep current selection and anchor
		}
	case "ctrl+s", "ctrl+enter":
		if m.focusOrder[m.focusIndex] == "message" {
			payload := m.messageInput.Value()
			for _, t := range m.topics {
				if t.active {
					m.payloads = append(m.payloads, payloadItem{topic: t.title, payload: payload})
					m.appendHistory(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
					if m.mqttClient != nil {
						m.mqttClient.Publish(t.title, 0, false, payload)
					}
				}
			}
		}
	case "enter":
		if m.focusOrder[m.focusIndex] == "topic" {
			topic := strings.TrimSpace(m.topicInput.Value())
			if topic != "" && !m.hasTopic(topic) {
				m.topics = append(m.topics, topicItem{title: topic, active: true})
				m.sortTopics()
				if m.mqttClient != nil {
					m.mqttClient.Subscribe(topic, 0, nil)
				}
				m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
				m.topicInput.SetValue("")
			}
		} else if m.focusOrder[m.focusIndex] == "topics" && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
			m.toggleTopic(m.selectedTopic)
		}
	case "d":
		if m.focusOrder[m.focusIndex] == "topics" && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
			idx := m.selectedTopic
			name := m.topics[idx].title
			m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), func() {
				m.removeTopic(idx)
			})
		}
	default:
		switch msg.String() {
		case "ctrl+b":
			if err := m.connections.LoadProfiles(""); err != nil {
				m.appendHistory("", err.Error(), "log", err.Error())
			}
			m.refreshConnectionItems()
			m.saveCurrent()
			m.mode = modeConnections
		case "ctrl+t":
			items := []list.Item{}
			for _, tpc := range m.topics {
				items = append(items, topicItem{title: tpc.title, active: tpc.active})
			}
			m.topicsList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
			m.topicsList.DisableQuitKeybindings()
			m.topicsList.SetShowTitle(false)
			m.mode = modeTopics
		case "ctrl+p":
			items := []list.Item{}
			for _, pld := range m.payloads {
				items = append(items, payloadItem{topic: pld.topic, payload: pld.payload})
			}
			m.payloadList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
			m.payloadList.DisableQuitKeybindings()
			m.payloadList.SetShowTitle(false)
			m.mode = modePayloads
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
		if m.focusOrder[m.focusIndex] == "history" {
			var hCmd tea.Cmd
			m.history, hCmd = m.history.Update(msg)
			cmds = append(cmds, hCmd)
		}
	}
	if msg.Type == tea.MouseLeft {
		cmds = append(cmds, m.focusFromMouse(msg.Y))
		if m.focusOrder[m.focusIndex] == "history" {
			idx := m.historyIndexAt(msg.Y)
			if idx >= 0 {
				m.history.Select(idx)
				if msg.Shift {
					if m.selectionAnchor == -1 {
						m.selectionAnchor = m.history.Index()
						m.selectedHistory[m.selectionAnchor] = struct{}{}
					}
					m.updateSelectionRange(idx)
				} else {
					m.selectedHistory = map[int]struct{}{}
					m.selectionAnchor = -1
				}
			}
		}
	}
	m.handleTopicsClick(msg)
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// handleTopicsClick processes mouse events within the topics area.
// The mouse coordinates are adjusted for the viewport offset and compared
// against precomputed chip bounds.
func (m *model) handleTopicsClick(msg tea.MouseMsg) {
	y := msg.Y + m.viewport.YOffset
	idx := m.topicAtPosition(msg.X, y)
	if idx < 0 {
		return
	}
	m.selectedTopic = idx
	if msg.Type == tea.MouseLeft {
		m.toggleTopic(idx)
	} else if msg.Type == tea.MouseRight {
		name := m.topics[idx].title
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
	m.topicInput, cmd = m.topicInput.Update(msg)
	cmds = append(cmds, cmd)
	var cmdMsg tea.Cmd
	m.messageInput, cmdMsg = m.messageInput.Update(msg)
	cmds = append(cmds, cmdMsg)
	var vpCmd tea.Cmd
	skipVP := false
	if m.focusOrder[m.focusIndex] == "history" {
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
		m.viewport, vpCmd = m.viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	var histCmd tea.Cmd
	if m.focusOrder[m.focusIndex] == "history" {
		m.history, histCmd = m.history.Update(msg)
		cmds = append(cmds, histCmd)
	}

	if m.history.FilterState() == list.Filtering {
		q := m.history.FilterInput.Value()
		topics, start, end, text := history.ParseQuery(q)
		msgs := m.store.Search(topics, start, end, text)
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
		}
		m.history.SetItems(items)
	} else {
		items := make([]list.Item, len(m.historyItems))
		for i, it := range m.historyItems {
			items[i] = it
		}
		m.history.SetItems(items)
	}

	cmds = append(cmds, listenStatus(m.statusChan))
	if m.mqttClient != nil {
		cmds = append(cmds, listenMessages(m.mqttClient.MessageChan))
	}
	return tea.Batch(cmds...)
}
