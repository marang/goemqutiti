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

	connection    string
	activeConn    string
	history       list.Model
	topicInput    textinput.Model
	messageInput  textarea.Model
	payloads      map[string]string
	topics        []topicItem
	topicsList    list.Model
	payloadList   list.Model
	focusIndex    int
	selectedTopic int

	saved map[string]connectionData

	statusChan chan string

	width  int
	height int

	mode        appMode
	connections Connections
	connForm    *connectionForm
	deleteIndex int

	viewport   viewport.Model
	elemPos    map[string]int
	focusMap   map[string]focusable
	focusOrder []string
}

func initialModel(conns *Connections) model {
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

	hist := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	hist.SetShowStatusBar(false)
	hist.SetShowPagination(false)
	hist.DisableQuitKeybindings()
	statusChan := make(chan string, 10)

	topicsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	topicsList.DisableQuitKeybindings()
	payloadList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	payloadList.DisableQuitKeybindings()
	vp := viewport.New(0, 0)

	order := []string{"topic", "message", "topics"}

	m := model{
		history:       hist,
		payloads:      make(map[string]string),
		topicInput:    ti,
		messageInput:  ta,
		topics:        []topicItem{},
		topicsList:    topicsList,
		payloadList:   payloadList,
		focusIndex:    0,
		selectedTopic: -1,
		statusChan:    statusChan,
		mode:          modeClient,
		connections:   connModel,
		width:         0,
		height:        0,
		viewport:      vp,
		elemPos:       map[string]int{},
		focusOrder:    order,
		saved:         make(map[string]connectionData),
	}
	m.focusMap = map[string]focusable{
		"topic":   &m.topicInput,
		"message": &m.messageInput,
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}
