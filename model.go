package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type historyItem struct {
	topic   string
	payload string
	kind    string // pub, sub, log
}

func (h historyItem) FilterValue() string { return h.payload }
func (h historyItem) Title() string {
	var label string
	color := lipgloss.Color("63")
	switch h.kind {
	case "sub":
		label = "SUB"
		color = lipgloss.Color("205")
	case "pub":
		label = "PUB"
		color = lipgloss.Color("63")
	default:
		label = "LOG"
		color = lipgloss.Color("240")
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s: %s", label, h.topic, h.payload))
}
func (h historyItem) Description() string { return "" }

type appMode int

const (
	modeClient appMode = iota
	modeConnections
	modeEditConnection
	modeConfirmDelete
	modeTopics
	modePayloads
)

type connectionData struct {
	Topics   []topicItem
	Payloads map[string]string
}

type focusable interface {
	Focus() tea.Cmd
	Blur()
}

type model struct {
	mqttClient *MQTTClient

	connection      string
	activeConn      string
	history         list.Model
	topicInput      textinput.Model
	messageInput    textarea.Model
	payloads        map[string]string
	topics          []topicItem
	topicsList      list.Model
	payloadList     list.Model
	focusIndex      int
	selectedTopic   int
	selectedHistory map[int]struct{}
	selectionAnchor int

	saved map[string]connectionData

	statusChan chan string

	width  int
	height int

	mode        appMode
	connections Connections
	connForm    *connectionForm
	deleteIndex int

	confirmPrompt string
	confirmAction func()
	prevMode      appMode

	viewport   viewport.Model
	elemPos    map[string]int
	focusMap   map[string]focusable
	focusOrder []string
}

func initialModel(conns *Connections) *model {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.Focus()
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.Cursor.Style = cursorStyle
	ti.TextStyle = focusedStyle
	// Defer width assignment until we know the terminal size
	ti.Width = 0

	ta := textarea.New()
	ta.Placeholder = "Enter Message"
	ta.CharLimit = 10000
	ta.Prompt = "> "
	ta.Blur()
	ta.Cursor.Style = noCursor
	// Set width once the WindowSizeMsg arrives
	ta.SetWidth(0)
	ta.SetHeight(6)
	ta.FocusedStyle.CursorLine = focusedStyle
	ta.BlurredStyle.CursorLine = blurredStyle

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

	hDel := historyDelegate{}
	hist := list.New([]list.Item{}, hDel, 0, 0)
	hist.SetShowTitle(false)
	hist.SetShowStatusBar(false)
	hist.SetShowPagination(false)
	hist.DisableQuitKeybindings()
	statusChan := make(chan string, 10)

	topicsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	topicsList.DisableQuitKeybindings()
	payloadList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	payloadList.DisableQuitKeybindings()
	vp := viewport.New(0, 0)

	order := []string{"topic", "message", "history", "topics"}

	m := &model{
		history:         hist,
		payloads:        make(map[string]string),
		topicInput:      ti,
		messageInput:    ta,
		topics:          []topicItem{},
		topicsList:      topicsList,
		payloadList:     payloadList,
		focusIndex:      0,
		selectedTopic:   -1,
		statusChan:      statusChan,
		mode:            modeClient,
		connections:     connModel,
		width:           0,
		height:          0,
		viewport:        vp,
		elemPos:         map[string]int{},
		focusOrder:      order,
		saved:           make(map[string]connectionData),
		selectedHistory: make(map[int]struct{}),
		selectionAnchor: -1,
		prevMode:        modeClient,
	}
	m.focusMap = map[string]focusable{
		"topic":   &m.topicInput,
		"message": &m.messageInput,
	}
	hDel.m = m
	m.history.SetDelegate(hDel)
	return m
}

func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (m *model) hasTopic(topic string) bool {
	for _, t := range m.topics {
		if t.title == topic {
			return true
		}
	}
	return false
}

func (m *model) toggleTopic(index int) {
	if index < 0 || index >= len(m.topics) {
		return
	}
	t := &m.topics[index]
	t.active = !t.active
	if m.mqttClient != nil {
		if t.active {
			m.mqttClient.Subscribe(t.title, 0, nil)
			m.appendHistory(t.title, "", "log", fmt.Sprintf("Subscribed to topic: %s", t.title))
		} else {
			m.mqttClient.Unsubscribe(t.title)
			m.appendHistory(t.title, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", t.title))
		}
	}
}

func (m *model) removeTopic(index int) {
	if index < 0 || index >= len(m.topics) {
		return
	}
	topic := m.topics[index]
	if m.mqttClient != nil {
		m.mqttClient.Unsubscribe(topic.title)
		m.appendHistory(topic.title, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", topic.title))
	}
	m.topics = append(m.topics[:index], m.topics[index+1:]...)
	if len(m.topics) == 0 {
		m.selectedTopic = -1
	} else if m.selectedTopic >= len(m.topics) {
		m.selectedTopic = len(m.topics) - 1
	}
}

func (m *model) topicAtPosition(x, y, width int) int {
	curX, curY := 0, 0
	for i, t := range m.topics {
		chip := chipStyle.Render(t.title)
		if !t.active {
			chip = chipInactive.Render(t.title)
		}
		w := lipgloss.Width(chip)
		if curX+w > width && curX > 0 {
			curY++
			curX = 0
		}
		if y == curY && x >= curX && x < curX+w {
			return i
		}
		curX += w
	}
	return -1
}

func (m *model) startConfirm(prompt string, action func()) {
	m.confirmPrompt = prompt
	m.confirmAction = action
	m.prevMode = m.mode
	m.mode = modeConfirmDelete
}
