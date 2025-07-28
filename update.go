package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"goemqutiti/history"
)

type statusMessage string

type connectResult struct {
	client  *MQTTClient
	profile Profile
	err     error
}

func connectBroker(p Profile, ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch != nil {
			brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
			ch <- fmt.Sprintf("Connecting to %s", brokerURL)
		}
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

func flushStatus(ch chan string) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		default:
			return
		}
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
		m.sortTopics()
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
	hi := historyItem{topic: topic, payload: text, kind: kind}
	m.historyItems = append(m.historyItems, hi)
	items := make([]list.Item, len(m.historyItems))
	for i, it := range m.historyItems {
		items[i] = it
	}
	m.history.SetItems(items)
	m.history.Select(len(items) - 1)
	if m.store != nil {
		m.store.Add(history.Message{Timestamp: time.Now(), Topic: topic, Payload: payload, Kind: kind})
	}
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

func (m model) updateConnections(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case connectResult:
		brokerURL := fmt.Sprintf("%s://%s:%d", msg.profile.Schema, msg.profile.Host, msg.profile.Port)
		if msg.err != nil {
			m.connections.Statuses[msg.profile.Name] = "disconnected"
			m.connections.Errors[msg.profile.Name] = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err)
			m.connection = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err)
			m.refreshConnectionItems()
		} else {
			m.mqttClient = msg.client
			m.activeConn = msg.profile.Name
			if m.store != nil {
				m.store.Close()
			}
			if idx, err := history.Open(msg.profile.Name); err == nil {
				m.store = idx
				msgs := idx.Search(nil, time.Time{}, time.Time{}, "")
				m.historyItems = nil
				items := make([]list.Item, len(msgs))
				for i, mmsg := range msgs {
					items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
					m.historyItems = append(m.historyItems, items[i].(historyItem))
				}
				m.history.SetItems(items)
			}
			m.restoreState(msg.profile.Name)
			m.subscribeActiveTopics()
			m.connection = "Connected to " + brokerURL
			m.connections.Statuses[msg.profile.Name] = "connected"
			m.connections.Errors[msg.profile.Name] = ""
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
					if p.Name == m.activeConn && m.connections.Statuses[p.Name] == "connected" {
						m.mode = modeClient
						return m, nil
					}
					flushStatus(m.statusChan)
					if p.FromEnv {
						applyEnvVars(&p)
					} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
						p.Password = env
					}
					m.connections.Errors[p.Name] = ""
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
				if p.Name == m.activeConn && m.connections.Statuses[p.Name] == "connected" {
					m.mode = modeClient
					return m, nil
				}
				flushStatus(m.statusChan)
				if p.FromEnv {
					applyEnvVars(&p)
				} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
					p.Password = env
				}
				m.connections.Errors[p.Name] = ""
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
					m.connections.refreshList()
					m.refreshConnectionItems()
				})
			}
		case "x":
			if m.mqttClient != nil {
				m.mqttClient.Disconnect()
				m.connections.Statuses[m.activeConn] = "disconnected"
				m.connections.Errors[m.activeConn] = ""
				m.refreshConnectionItems()
				m.connection = ""
				m.activeConn = ""
				m.mqttClient = nil
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
		m.messageInput.SetHeight(m.layout.message.height)
		if m.layout.history.height == 0 {
			m.layout.history.height = (msg.Height-1)/3 + 10
		}
		m.history.SetSize(msg.Width-4, m.layout.history.height)
		m.viewport.Width = msg.Width
		// Reserve two lines for the info header at the top of the view.
		m.viewport.Height = msg.Height - 2
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
