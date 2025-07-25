package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (m model) updateClient(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case statusMessage:
		m.appendHistory("", string(msg), "log", string(msg))
		return m, listenStatus(m.statusChan)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			m.saveCurrent()
			return m, tea.Quit
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
		case "tab":
			switch m.focusIndex {
			case 0:
				m.topicInput.Blur()
				m.messageInput.Focus()
				m.focusIndex = 1
			case 1:
				m.messageInput.Blur()
				m.focusIndex = 2
				if len(m.topics) > 0 {
					m.selectedTopic = 0
				} else {
					m.selectedTopic = -1
				}
			default:
				m.focusIndex = 0
				m.topicInput.Focus()
			}
		case "left":
			if m.focusIndex == 2 && len(m.topics) > 0 {
				m.selectedTopic = (m.selectedTopic - 1 + len(m.topics)) % len(m.topics)
			}
		case "right":
			if m.focusIndex == 2 && len(m.topics) > 0 {
				m.selectedTopic = (m.selectedTopic + 1) % len(m.topics)
			}
		case "enter", " ":
			if m.focusIndex == 0 {
				topic := strings.TrimSpace(m.topicInput.Value())
				if topic != "" {
					m.topics = append(m.topics, topicItem{title: topic, active: true})
					m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
					m.topicInput.SetValue("")
				}
			} else if m.focusIndex == 1 {
				payload := m.messageInput.Value()
				for _, t := range m.topics {
					if t.active {
						m.payloads[t.title] = payload
						m.appendHistory(t.title, payload, "pub", fmt.Sprintf("Published to %s: %s", t.title, payload))
						pl := payloadItem{topic: t.title, payload: payload}
						items := append(m.payloadList.Items(), pl)
						m.payloadList.SetItems(items)
					}
				}
				m.messageInput.SetValue("")
			} else if m.focusIndex == 2 && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
				m.topics[m.selectedTopic].active = !m.topics[m.selectedTopic].active
			}
		case "d":
			if m.focusIndex == 2 && m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
				m.topics = append(m.topics[:m.selectedTopic], m.topics[m.selectedTopic+1:]...)
				if len(m.topics) == 0 {
					m.selectedTopic = -1
				} else if m.selectedTopic >= len(m.topics) {
					m.selectedTopic = len(m.topics) - 1
				}
			}
		default:
			if m.focusIndex > 1 {
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
					m.topicsList.Title = "Topics"
					m.mode = modeTopics
				case "ctrl+p":
					items := []list.Item{}
					for topic, payload := range m.payloads {
						items = append(items, payloadItem{topic: topic, payload: payload})
					}
					m.payloadList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
					m.payloadList.Title = "Payloads"
					m.mode = modePayloads
				}
			}
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			m.history, cmd = m.history.Update(msg)
			return m, tea.Batch(cmd, listenStatus(m.statusChan))
		}
		if m.focusIndex == 2 {
			row := 4
			if msg.Y == row {
				x := msg.X - 2
				cum := 0
				for i, t := range m.topics {
					chip := chipStyle.Render(t.title)
					if !t.active {
						chip = chipInactive.Render(t.title)
					}
					w := lipgloss.Width(chip)
					if x >= cum && x < cum+w {
						m.selectedTopic = i
						if msg.Type == tea.MouseLeft {
							m.topics[i].active = !m.topics[i].active
						} else if msg.Type == tea.MouseRight {
							m.topics = append(m.topics[:i], m.topics[i+1:]...)
							if i >= len(m.topics) {
								m.selectedTopic = len(m.topics) - 1
							}
						}
						break
					}
					cum += w
				}
			}
		}
	}

	m.topicInput, cmd = m.topicInput.Update(msg)
	var cmdMsg tea.Cmd
	m.messageInput, cmdMsg = m.messageInput.Update(msg)

	return m, tea.Batch(cmd, cmdMsg, listenStatus(m.statusChan))
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
				m.topics = append(m.topics[:i], m.topics[i+1:]...)
				items := []list.Item{}
				for _, t := range m.topics {
					items = append(items, t)
				}
				m.topicsList.SetItems(items)
			}
		case "enter", " ":
			i := m.topicsList.Index()
			if i >= 0 && i < len(m.topics) {
				m.topics[i].active = !m.topics[i].active
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.connections.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		return m, nil
	}

	switch m.mode {
	case modeClient:
		return m.updateClient(msg)
	case modeConnections:
		return m.updateConnections(msg)
	case modeEditConnection:
		return m.updateForm(msg)
	case modeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case modeTopics:
		return m.updateTopics(msg)
	case modePayloads:
		return m.updatePayloads(msg)
	default:
		return m, nil
	}
}
