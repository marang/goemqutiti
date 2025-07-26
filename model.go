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

type chipBound struct {
	x, y int
	w, h int
}

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
	Payloads []payloadItem
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
	payloads        []payloadItem
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
	chipBounds []chipBound
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
	ta.ShowLineNumbers = false
	ta.SetPromptFunc(0, func(i int) string {
		return fmt.Sprintf("%d> ", i+1)
	})
	promptColor := lipgloss.Color("240")
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(promptColor)
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(promptColor)
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

	order := []string{"topics", "topic", "message", "history"}
	saved := loadState()

	m := &model{
		history:         hist,
		payloads:        []payloadItem{},
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
		chipBounds:      []chipBound{},
		focusOrder:      order,
		saved:           saved,
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

func (m *model) topicAtPosition(x, y int) int {
	for i, b := range m.chipBounds {
		if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
			return i
		}
	}
	return -1
}

func (m *model) startConfirm(prompt string, action func()) {
	m.confirmPrompt = prompt
	m.confirmAction = action
	m.prevMode = m.mode
	m.mode = modeConfirmDelete
}
