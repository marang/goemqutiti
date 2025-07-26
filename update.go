package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type statusMessage string

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
}

func (m *model) restoreState(name string) {
	if data, ok := m.saved[name]; ok {
		m.topics = data.Topics
		m.payloads = data.Payloads
	} else {
		m.topics = []topicItem{}
		m.payloads = make(map[string]string)
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
		return m.setFocus(chosen)
	}
	if len(m.focusOrder) > 0 {
		return m.setFocus(m.focusOrder[0])
	}
	return nil
}

func (m *model) updateClient(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case statusMessage:
		m.appendHistory("", string(msg), "log", string(msg))
		return listenStatus(m.statusChan)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			m.saveCurrent()
			return tea.Quit
		case "ctrl+c":
			if len(m.history.Items()) > 0 {
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
			if m.focusIndex == 2 && len(m.topics) > 0 {
				m.selectedTopic = (m.selectedTopic - 1 + len(m.topics)) % len(m.topics)
			}
		case "right":
			if m.focusIndex == 2 && len(m.topics) > 0 {
				m.selectedTopic = (m.selectedTopic + 1) % len(m.topics)
			}
		case "ctrl+s", "ctrl+enter":
			if m.focusIndex == 1 {
				payload := m.messageInput.Value()
				for _, t := range m.topics {
					if t.active {
						m.payloads[t.title] = payload
						m.appendHistory(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
						if m.mqttClient != nil {
							m.mqttClient.Publish(t.title, 0, false, payload)
						}
					}
				}
			}
		case "enter":
			if m.focusIndex == 0 {
				topic := strings.TrimSpace(m.topicInput.Value())
				if topic != "" && !m.hasTopic(topic) {
					m.topics = append(m.topics, topicItem{title: topic, active: true})
					if m.mqttClient != nil {
						m.mqttClient.Subscribe(topic, 0, nil)
					}
					m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
					m.topicInput.SetValue("")
					cmds = append(cmds, m.setFocus("message"))
				}
			} else if m.focusIndex == 2 && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
				m.toggleTopic(m.selectedTopic)
			}
		case "d":
			if m.focusIndex == 2 && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
				m.removeTopic(m.selectedTopic)
			}
		default:
			switch msg.String() {
			case "ctrl+m":
				m.connections.LoadProfiles("")
				items := []list.Item{}
				for _, p := range m.connections.Profiles {
					items = append(items, connectionItem{title: p.Name})
				}
				m.connections.ConnectionsList.SetItems(items)
				m.saveCurrent()
				m.mode = modeConnections
			case "ctrl+t":
				items := []list.Item{}
				for _, tpc := range m.topics {
					items = append(items, topicItem{title: tpc.title, active: tpc.active})
				}
				m.topicsList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
				m.topicsList.DisableQuitKeybindings()
				m.topicsList.Title = "Topics"
				m.mode = modeTopics
			case "ctrl+p":
				items := []list.Item{}
				for topic, payload := range m.payloads {
					items = append(items, payloadItem{topic: topic, payload: payload})
				}
				m.payloadList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
				m.payloadList.DisableQuitKeybindings()
				m.payloadList.Title = "Payloads"
				m.mode = modePayloads
			}
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			var hCmd tea.Cmd
			m.history, hCmd = m.history.Update(msg)
			cmds = append(cmds, hCmd)
		}
		if msg.Type == tea.MouseLeft {
			cmds = append(cmds, m.focusFromMouse(msg.Y))
		}
		if m.focusIndex == 2 {
			start := m.elemPos["topics"] + 1
			idx := m.topicAtPosition(msg.X-2, msg.Y-start, m.width-6)
			if idx >= 0 {
				m.selectedTopic = idx
				if msg.Type == tea.MouseLeft {
					m.toggleTopic(idx)
				} else if msg.Type == tea.MouseRight {
					m.removeTopic(idx)
				}
			}
		}
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
	if m.focusIndex > 1 {
		m.history, histCmd = m.history.Update(msg)
		cmds = append(cmds, histCmd)
	}

	cmds = append(cmds, listenStatus(m.statusChan))
	return tea.Batch(cmds...)
}

func (m model) updateConnections(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
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
					client, err := NewMQTTClient(p, m.statusChan)
					if err != nil {
						m.appendHistory("", "", "log", fmt.Sprintf("Failed to connect: %v", err))
					} else {
						m.mqttClient = client
						m.activeConn = p.Name
						m.restoreState(p.Name)
						brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
						m.connection = "Connected to " + brokerURL
						m.mode = modeClient
					}
				}
			case "esc":
				m.mode = modeClient
			}
			break
		}
		switch msg.String() {
		case "esc":
			m.mode = modeClient
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
				client, err := NewMQTTClient(p, m.statusChan)
				if err != nil {
					m.appendHistory("", "", "log", fmt.Sprintf("Failed to connect: %v", err))
				} else {
					m.mqttClient = client
					m.activeConn = p.Name
					m.restoreState(p.Name)
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connection = "Connected to " + brokerURL
					m.mode = modeClient
				}
			}
		case "d":
			i := m.connections.ConnectionsList.Index()
			if i >= 0 {
				m.deleteIndex = i
				m.mode = modeConfirmDelete
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
	m.connections.ConnectionsList, _ = m.connections.ConnectionsList.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
			items := []list.Item{}
			for _, pr := range m.connections.Profiles {
				items = append(items, connectionItem{title: pr.Name})
			}
			m.connections.ConnectionsList.SetItems(items)
			m.mode = modeConnections
			m.connForm = nil
			return m, nil
		}
	}
	f, cmd := m.connForm.Update(msg)
	m.connForm = &f
	return m, tea.Batch(cmd, listenStatus(m.statusChan))
}

func (m model) updateConfirmDelete(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.connections.DeleteConnection(m.deleteIndex)
			items := []list.Item{}
			for _, p := range m.connections.Profiles {
				items = append(items, connectionItem{title: p.Name})
			}
			m.connections.ConnectionsList.SetItems(items)
			m.mode = modeConnections
		case "n", "esc":
			m.mode = modeConnections
		}
	}
	return m, listenStatus(m.statusChan)
}

func (m model) updateTopics(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeClient
		case "d":
			i := m.topicsList.Index()
			if i >= 0 && i < len(m.topics) {
				m.removeTopic(i)
				items := []list.Item{}
				for _, t := range m.topics {
					items = append(items, t)
				}
				m.topicsList.SetItems(items)
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
		case "esc":
			m.mode = modeClient
		case "d":
			i := m.payloadList.Index()
			if i >= 0 {
				items := m.payloadList.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					delete(m.payloads, pi.topic)
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

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.connections.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		m.topicInput.Width = msg.Width - 6
		m.messageInput.SetWidth(msg.Width - 6)
		m.history.SetSize(msg.Width-4, (msg.Height-1)/3)
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
