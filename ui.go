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
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))
	chipInactive = chipStyle.Copy().Foreground(lipgloss.Color("240"))
)

// connectionItem implements list.DefaultItem interface for connection names.
type connectionItem struct{ title string }

func (c connectionItem) FilterValue() string { return c.title }
func (c connectionItem) Title() string       { return c.title }
func (c connectionItem) Description() string { return "" }

type topicItem struct {
	title  string
	active bool
}

func (t topicItem) FilterValue() string { return t.title }
func (t topicItem) Title() string       { return t.title }
func (t topicItem) Description() string {
	if t.active {
		return "enabled"
	}
	return "disabled"
}

type payloadItem struct{ topic, payload string }

func (p payloadItem) FilterValue() string { return p.topic }
func (p payloadItem) Title() string       { return p.topic }
func (p payloadItem) Description() string { return p.payload }

type appMode int

const (
	modeClient appMode = iota
	modeConnections
	modeEditConnection
	modeConfirmDelete
	modeTopics
	modePayloads
)

type model struct {
	mqttClient    *MQTTClient
	connection    string
	messages      []string
	topicInput    textinput.Model
	messageInput  textinput.Model
	payloads      map[string]string
	topics        []topicItem
	topicsList    list.Model
	payloadList   list.Model
	focusIndex    int
	selectedTopic int

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
		messages:      make([]string, 0),
		payloads:      make(map[string]string),
		topicInput:    ti,
		messageInput:  mi,
		topics:        []topicItem{},
		topicsList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		payloadList:   list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		focusIndex:    0,
		selectedTopic: -1,
		mode:          modeClient,
		connections:   connModel,
		width:         0,
		height:        0,
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
					m.messages = append(m.messages, fmt.Sprintf("Subscribed to topic: %s", topic))
					m.topicInput.SetValue("")
				}
			} else if m.focusIndex == 1 {
				payload := m.messageInput.Value()
				for _, t := range m.topics {
					if t.active {
						m.payloads[t.title] = payload
						m.messages = append(m.messages, fmt.Sprintf("Published to %s: %s", t.title, payload))
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
				case "m":
					m.connections.LoadProfiles("")
					items := []list.Item{}
					for _, p := range m.connections.Profiles {
						items = append(items, connectionItem{title: p.Name})
					}
					m.connections.ConnectionsList.SetItems(items)
					m.mode = modeConnections
				case "t":
					items := []list.Item{}
					for _, tpc := range m.topics {
						items = append(items, topicItem{title: tpc.title, active: tpc.active})
					}
					m.topicsList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-4)
					m.topicsList.Title = "Topics"
					m.mode = modeTopics
				case "p":
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
	m.connections.ConnectionsList, _ = m.connections.ConnectionsList.Update(msg)
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
	return m, cmd
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
	return m, cmd
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

func (m model) viewClient() string {
	header := borderStyle.Copy().Width(m.width - 4).Render("GoEmqutiti - MQTT Client")
	info := borderStyle.Copy().Width(m.width - 4).Render("Press 'm' to manage connections")
	conn := borderStyle.Copy().Width(m.width - 4).Render("Connection: " + m.connection)

	var chips []string
	for i, t := range m.topics {
		st := chipStyle
		if !t.active {
			st = chipInactive
		}
		if m.focusIndex == 2 && i == m.selectedTopic {
			st = st.Copy().BorderForeground(lipgloss.Color("212"))
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsBox := borderStyle.Copy().Width(m.width - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, chips...))

	msgLines := strings.Join(m.messages, "\n")
	messagesBox := borderStyle.Copy().Width(m.width - 4).Height(m.height / 3).Render(msgLines)

	inputs := lipgloss.JoinVertical(lipgloss.Left,
		"Topic:\n"+m.topicInput.View(),
		"Message:\n"+m.messageInput.View(),
	)
	inputsBox := borderStyle.Copy().Width(m.width - 4).Render(inputs)

	var payloadLines []string
	for topic, payload := range m.payloads {
		payloadLines = append(payloadLines, fmt.Sprintf("- %s: %s", topic, payload))
	}
	payloadHelp := "Stored Payloads (press 'p' to manage):"
	payloadBox := borderStyle.Copy().Width(m.width - 4).Render(payloadHelp + "\n" + strings.Join(payloadLines, "\n"))

	content := lipgloss.JoinVertical(lipgloss.Left, header, info, conn, topicsBox, messagesBox, inputsBox, payloadBox)
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
	listView := m.connections.ConnectionsList.View()
	formView := m.connForm.View()
	left := borderStyle.Copy().Width(m.width/2 - 2).Render(listView)
	right := borderStyle.Copy().Width(m.width/2 - 2).Render(formView)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) viewConfirmDelete() string {
	var name string
	if m.deleteIndex >= 0 && m.deleteIndex < len(m.connections.Profiles) {
		name = m.connections.Profiles[m.deleteIndex].Name
	}
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	return border.Render(fmt.Sprintf("Delete connection '%s'? [y/n]", name))
}

func (m model) viewTopics() string {
	listView := m.topicsList.View()
	help := "[enter] toggle  [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
}

func (m model) viewPayloads() string {
	listView := m.payloadList.View()
	help := "[enter] load  [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
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
	case modeTopics:
		return m.viewTopics()
	case modePayloads:
		return m.viewPayloads()
	default:
		return ""
	}
}
