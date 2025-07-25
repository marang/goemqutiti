package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noCursor     = lipgloss.NewStyle()
	borderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
)

// connectionItem implements list.DefaultItem interface for connection names.
type connectionItem struct{ title string }

func (c connectionItem) FilterValue() string { return c.title }
func (c connectionItem) Title() string       { return c.title }
func (c connectionItem) Description() string { return "" }

type appMode int

const (
	modeClient appMode = iota
	modeConnections
	modeEditConnection
	modeConfirmDelete
)

type model struct {
	mqttClient   *MQTTClient
	connection   string
	messages     []string
	topicInput   textinput.Model
	messageInput textinput.Model
	payloads     map[string]string
	subscribed   bool
	focusIndex   int

	width  int
	height int

	mode        appMode
	connections Connections
	connForm    *connectionForm
	deleteIndex int
}

func initialModel(conns *Connections) model {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.Focus()
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.Cursor.Style = cursorStyle
	ti.TextStyle = focusedStyle
	ti.Width = 40

	mi := textinput.New()
	mi.Placeholder = "Enter Message"
	mi.CharLimit = 128
	mi.Prompt = "> "
	mi.Blur()
	mi.Cursor.Style = noCursor
	mi.TextStyle = blurredStyle
	mi.Width = 80

	var connModel Connections
	if conns != nil {
		connModel = *conns
	} else {
		connModel = NewConnectionsModel()
		connModel.LoadProfiles("")
	}
	connModel.ConnectionsList.SetShowStatusBar(false)
	items := []list.Item{}
	for _, p := range connModel.Profiles {
		items = append(items, connectionItem{title: p.Name})
	}
	connModel.ConnectionsList.SetItems(items)

	return model{
		messages:     make([]string, 0),
		payloads:     make(map[string]string),
		topicInput:   ti,
		messageInput: mi,
		focusIndex:   0,
		mode:         modeClient,
		connections:  connModel,
		width:        0,
		height:       0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (m model) updateClient(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focusIndex == 0 {
				m.topicInput.Blur()
				m.messageInput.Focus()
				m.focusIndex = 1
			} else {
				m.messageInput.Blur()
				m.topicInput.Focus()
				m.focusIndex = 0
			}
		case "enter":
			if !m.subscribed {
				m.messages = append(m.messages, fmt.Sprintf("Subscribed to topic: %s", m.topicInput.Value()))
				m.subscribed = true
			} else {
				topic := m.topicInput.Value()
				payload := m.messageInput.Value()
				m.payloads[topic] = payload
				m.messages = append(m.messages, fmt.Sprintf("Published to %s: %s", topic, payload))
			}
		case "m":
			m.connections.LoadProfiles("")
			items := []list.Item{}
			for _, p := range m.connections.Profiles {
				items = append(items, connectionItem{title: p.Name})
			}
			m.connections.ConnectionsList.SetItems(items)
			m.mode = modeConnections
		}
	}

	m.topicInput, cmd = m.topicInput.Update(msg)
	m.messageInput, _ = m.messageInput.Update(msg)

	return m, cmd
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
					client, err := NewMQTTClient(p)
					if err != nil {
						m.messages = append(m.messages, fmt.Sprintf("Failed to connect: %v", err))
					} else {
						m.mqttClient = client
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
				client, err := NewMQTTClient(p)
				if err != nil {
					m.messages = append(m.messages, fmt.Sprintf("Failed to connect: %v", err))
				} else {
					m.mqttClient = client
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
	return m, cmd
}

func (m model) updateForm(msg tea.Msg) (model, tea.Cmd) {
	if m.connForm == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeConnections
			m.connForm = nil
			return m, nil
		case "enter":
			// save
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
	return m, cmd
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
	return m, nil
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
	default:
		return m, nil
	}
}

func (m model) viewClient() string {
	header := borderStyle.Copy().Width(m.width - 2).Render("GoEmqutiti - MQTT Client")
	info := borderStyle.Copy().Width(m.width - 2).Render("Press 'm' to manage connections")
	conn := borderStyle.Copy().Width(m.width - 2).Render("Connection: " + m.connection)

	msgLines := strings.Join(m.messages, "\n")
	messagesBox := borderStyle.Copy().Width(m.width - 2).Height(m.height / 3).Render(msgLines)

	inputs := lipgloss.JoinVertical(lipgloss.Left,
		"Topic:\n"+m.topicInput.View(),
		"Message:\n"+m.messageInput.View(),
	)
	inputsBox := borderStyle.Copy().Width(m.width - 2).Render(inputs)

	var payloadLines []string
	for topic, payload := range m.payloads {
		payloadLines = append(payloadLines, fmt.Sprintf("- %s: %s", topic, payload))
	}
	payloadBox := borderStyle.Copy().Width(m.width - 2).Render("Stored Payloads:\n" + strings.Join(payloadLines, "\n"))

	content := lipgloss.JoinVertical(lipgloss.Left, header, info, conn, messagesBox, inputsBox, payloadBox)
	return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(1, 1).Render(content)
}

func (m model) viewConnections() string {
	listView := m.connections.ConnectionsList.View()
	help := "[enter] connect  [a]dd [e]dit [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
}

func (m model) viewForm() string {
	if m.connForm == nil {
		return ""
	}
	return m.connForm.View()
}

func (m model) viewConfirmDelete() string {
	var name string
	if m.deleteIndex >= 0 && m.deleteIndex < len(m.connections.Profiles) {
		name = m.connections.Profiles[m.deleteIndex].Name
	}
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	return border.Render(fmt.Sprintf("Delete connection '%s'? [y/n]", name))
}

func (m model) View() string {
	switch m.mode {
	case modeClient:
		return m.viewClient()
	case modeConnections:
		return m.viewConnections()
	case modeEditConnection:
		return m.viewForm()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	default:
		return ""
	}
}
