package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type statusMessage string

type connectResult struct {
	client  *MQTTClient
	profile Profile
	err     error
}

func connectBroker(p Profile, ch chan string) tea.Cmd {
	return func() tea.Msg {
		client, err := NewMQTTClient(p, ch)
		return connectResult{client: client, profile: p, err: err}
	}
}

func listenMessages(ch chan MQTTMessage) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func listenStatus(ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return statusMessage(msg)
	}
}

func (m *model) saveCurrent() {
	if m.activeConn == "" {
		return
	}
	m.saved[m.activeConn] = connectionData{Topics: m.topics, Payloads: m.payloads}
	saveState(m.saved)
}

func (m *model) restoreState(name string) {
	if data, ok := m.saved[name]; ok {
		m.topics = data.Topics
		m.payloads = data.Payloads
	} else {
		m.topics = []topicItem{}
		m.payloads = []payloadItem{}
	}
}

func (m *model) appendHistory(topic, payload, kind, logText string) {
	text := payload
	if kind == "log" {
		text = logText
	}
	items := append(m.history.Items(), historyItem{topic: topic, payload: text, kind: kind})
	m.history.SetItems(items)
	m.history.Select(len(items) - 1)
}

func (m *model) setFocus(id string) tea.Cmd {
	var cmds []tea.Cmd
	for i, name := range m.focusOrder {
		if f, ok := m.focusMap[name]; ok && f != nil {
			if name == id {
				if c := f.Focus(); c != nil {
					cmds = append(cmds, c)
				}
				m.focusIndex = i
			} else {
				f.Blur()
			}
		} else if name == id {
			m.focusIndex = i
		}
	}
	m.scrollToFocused()
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m *model) focusFromMouse(y int) tea.Cmd {
	cy := y + m.viewport.YOffset - 1
	chosen := ""
	maxPos := -1
	for _, id := range m.focusOrder {
		if pos, ok := m.elemPos[id]; ok && cy >= pos && pos > maxPos {
			chosen = id
			maxPos = pos
		}
	}
	if chosen != "" {
		if chosen != m.focusOrder[m.focusIndex] {
			return m.setFocus(chosen)
		}
		return nil
	}
	if len(m.focusOrder) > 0 && m.focusOrder[m.focusIndex] != m.focusOrder[0] {
		return m.setFocus(m.focusOrder[0])
	}
	return nil
}

func (m *model) scrollToFocused() {
	if len(m.focusOrder) == 0 {
		return
	}
	id := m.focusOrder[m.focusIndex]
	pos, ok := m.elemPos[id]
	if !ok {
		return
	}
	offset := pos - 1
	if offset < 0 {
		offset = 0
	}
	if offset < m.viewport.YOffset {
		m.viewport.SetYOffset(offset)
	} else if offset >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(offset - m.viewport.Height + 1)
	}
}

func (m *model) updateClient(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case statusMessage:
		m.appendHistory("", string(msg), "log", string(msg))
		if strings.HasPrefix(string(msg), "Connected") && m.activeConn != "" {
			m.connections.Statuses[m.activeConn] = "connected"
			m.refreshConnectionItems()
		} else if strings.HasPrefix(string(msg), "Connection lost") && m.activeConn != "" {
			m.connections.Statuses[m.activeConn] = "disconnected"
			m.refreshConnectionItems()
		}
		return listenStatus(m.statusChan)
	case MQTTMessage:
		m.appendHistory(msg.Topic, msg.Payload, "sub", fmt.Sprintf("Received on %s: %s", msg.Topic, msg.Payload))
		return listenMessages(m.mqttClient.MessageChan)
	case tea.KeyMsg:
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
		case "up", "down":
			if m.focusOrder[m.focusIndex] == "history" {
				m.selectedHistory = map[int]struct{}{}
				m.selectionAnchor = -1
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
				m.connections.LoadProfiles("")
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
				m.mode = modeTopics
			case "ctrl+p":
				items := []list.Item{}
				for _, pld := range m.payloads {
					items = append(items, payloadItem{topic: pld.topic, payload: pld.payload})
				}
				m.payloadList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
				m.payloadList.DisableQuitKeybindings()
				m.mode = modePayloads
			}
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
			if m.focusOrder[m.focusIndex] == "history" {
				var hCmd tea.Cmd
				m.history, hCmd = m.history.Update(msg)
				cmds = append(cmds, hCmd)
			}
		}
		if msg.Type == tea.MouseLeft {
			cmds = append(cmds, m.focusFromMouse(msg.Y))
		}
		m.handleTopicsClick(msg)
	}

	var cmd tea.Cmd
	m.topicInput, cmd = m.topicInput.Update(msg)
	cmds = append(cmds, cmd)
	var cmdMsg tea.Cmd
	m.messageInput, cmdMsg = m.messageInput.Update(msg)
	cmds = append(cmds, cmdMsg)
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	var histCmd tea.Cmd
	if m.focusOrder[m.focusIndex] == "history" {
		m.history, histCmd = m.history.Update(msg)
		cmds = append(cmds, histCmd)
	}

	cmds = append(cmds, listenStatus(m.statusChan))
	if m.mqttClient != nil {
		cmds = append(cmds, listenMessages(m.mqttClient.MessageChan))
	}
	return tea.Batch(cmds...)
}

// handleTopicsClick processes mouse events within the topics area. The
// mouse coordinates are adjusted for the viewport offset and compared
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

func (m model) updateConnections(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case connectResult:
		if msg.err != nil {
			m.appendHistory("", "", "log", fmt.Sprintf("Failed to connect: %v", msg.err))
			m.connections.Statuses[msg.profile.Name] = "disconnected"
			m.connection = ""
			m.refreshConnectionItems()
		} else {
			m.mqttClient = msg.client
			m.activeConn = msg.profile.Name
			m.restoreState(msg.profile.Name)
			m.subscribeActiveTopics()
			brokerURL := fmt.Sprintf("%s://%s:%d", msg.profile.Schema, msg.profile.Host, msg.profile.Port)
			m.connection = "Connected to " + brokerURL
			m.connections.Statuses[msg.profile.Name] = "connected"
			m.refreshConnectionItems()
			m.mode = modeClient
		}
		return m, listenStatus(m.statusChan)
	case tea.KeyMsg:
		if m.connections.ConnectionsList.FilterState() == list.Filtering {
			switch msg.String() {
			case "enter":
				i := m.connections.ConnectionsList.Index()
				if i >= 0 && i < len(m.connections.Profiles) {
					p := m.connections.Profiles[i]
					envPassword := os.Getenv("MQTT_PASSWORD")
					if envPassword != "" {
						p.Password = envPassword
					}
					m.connections.Statuses[p.Name] = "connecting"
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connection = "Connecting to " + brokerURL
					m.refreshConnectionItems()
					return m, connectBroker(p, m.statusChan)
				}
			}
			break
		}
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "a":
			f := newConnectionForm(Profile{}, -1)
			m.connForm = &f
			m.mode = modeEditConnection
		case "e":
			i := m.connections.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.Profiles) {
				f := newConnectionForm(m.connections.Profiles[i], i)
				m.connForm = &f
				m.mode = modeEditConnection
			}
		case "enter":
			i := m.connections.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.Profiles) {
				p := m.connections.Profiles[i]
				envPassword := os.Getenv("MQTT_PASSWORD")
				if envPassword != "" {
					p.Password = envPassword
				}
				m.connections.Statuses[p.Name] = "connecting"
				brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
				m.connection = "Connecting to " + brokerURL
				m.refreshConnectionItems()
				return m, connectBroker(p, m.statusChan)
			}
		case "d":
			i := m.connections.ConnectionsList.Index()
			if i >= 0 {
				name := m.connections.Profiles[i].Name
				m.startConfirm(fmt.Sprintf("Delete broker '%s'? [y/n]", name), func() {
					m.connections.DeleteConnection(i)
					m.refreshConnectionItems()
				})
			}
		}
	}
	m.connections.ConnectionsList, cmd = m.connections.ConnectionsList.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.statusChan))
}

func (m model) updateForm(msg tea.Msg) (model, tea.Cmd) {
	if m.connForm == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.WindowSizeMsg, tea.MouseMsg:
		m.connections.ConnectionsList, _ = m.connections.ConnectionsList.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			m.mode = modeConnections
			m.connForm = nil
			return m, nil
		case "enter":
			p := m.connForm.Profile()
			if m.connForm.index >= 0 {
				m.connections.EditConnection(m.connForm.index, p)
			} else {
				m.connections.AddConnection(p)
			}
			m.refreshConnectionItems()
			m.mode = modeConnections
			m.connForm = nil
			return m, nil
		}
	}
	f, cmd := m.connForm.Update(msg)
	m.connForm = &f
	return m, tea.Batch(cmd, listenStatus(m.statusChan))
}

func (m *model) updateConfirmDelete(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return *m, tea.Quit
		case "y":
			if m.confirmAction != nil {
				m.confirmAction()
				m.confirmAction = nil
			}
			m.mode = m.prevMode
			m.scrollToFocused()
		case "n", "esc":
			m.mode = m.prevMode
			m.scrollToFocused()
		}
	}
	return *m, listenStatus(m.statusChan)
}

func (m model) updateTopics(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			m.mode = modeClient
		case "d":
			i := m.topicsList.Index()
			if i >= 0 && i < len(m.topics) {
				name := m.topics[i].title
				m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), func() {
					m.removeTopic(i)
					items := []list.Item{}
					for _, t := range m.topics {
						items = append(items, t)
					}
					m.topicsList.SetItems(items)
				})
			}
		case "enter", " ":
			i := m.topicsList.Index()
			if i >= 0 && i < len(m.topics) {
				m.toggleTopic(i)
				items := m.topicsList.Items()
				items[i] = m.topics[i]
				m.topicsList.SetItems(items)
			}
		}
	}
	m.topicsList, cmd = m.topicsList.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.statusChan))
}

func (m model) updatePayloads(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			m.mode = modeClient
		case "d":
			i := m.payloadList.Index()
			if i >= 0 {
				items := m.payloadList.Items()
				if i < len(items) {
					m.payloads = append(m.payloads[:i], m.payloads[i+1:]...)
					items = append(items[:i], items[i+1:]...)
					m.payloadList.SetItems(items)
				}
			}
		case "enter":
			i := m.payloadList.Index()
			if i >= 0 {
				items := m.payloadList.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					m.topicInput.SetValue(pi.topic)
					m.messageInput.SetValue(pi.payload)
					m.mode = modeClient
				}
			}
		}
	}
	m.payloadList, cmd = m.payloadList.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.statusChan))
}

func (m *model) updateSelectionRange(idx int) {
	start := m.selectionAnchor
	end := idx
	if start > end {
		start, end = end, start
	}
	m.selectedHistory = map[int]struct{}{}
	for i := start; i <= end; i++ {
		m.selectedHistory[i] = struct{}{}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.connections.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		// textinput.View() renders the prompt and cursor in addition
		// to the configured width. Reduce the width slightly so the
		// surrounding box stays within the terminal boundaries.
		m.topicInput.Width = msg.Width - 7
		m.messageInput.SetWidth(msg.Width - 4)
		m.history.SetSize(msg.Width-4, (msg.Height-1)/3+10)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 1
		return m, nil
	}

	switch m.mode {
	case modeClient:
		cmd := m.updateClient(msg)
		return m, cmd
	case modeConnections:
		nm, cmd := m.updateConnections(msg)
		*m = nm
		return m, cmd
	case modeEditConnection:
		nm, cmd := m.updateForm(msg)
		*m = nm
		return m, cmd
	case modeConfirmDelete:
		nm, cmd := m.updateConfirmDelete(msg)
		*m = nm
		return m, cmd
	case modeTopics:
		nm, cmd := m.updateTopics(msg)
		*m = nm
		return m, cmd
	case modePayloads:
		nm, cmd := m.updatePayloads(msg)
		*m = nm
		return m, cmd
	default:
		return m, nil
	}
}
