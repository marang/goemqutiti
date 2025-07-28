package main

import (
	"fmt"
	"sort"
	"time"

	"goemqutiti/history"
	"goemqutiti/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type connectionItem struct {
	title  string
	status string
	detail string
}

func (c connectionItem) FilterValue() string { return c.title }
func (c connectionItem) Title() string       { return c.title }
func (c connectionItem) Description() string {
	if c.detail != "" {
		return c.status + "\n" + c.detail
	}
	return c.status
}

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
	color := ui.ColBlue
	switch h.kind {
	case "sub":
		label = "SUB"
		color = ui.ColPink
	case "pub":
		label = "PUB"
		color = ui.ColBlue
	default:
		label = "LOG"
		color = ui.ColGray
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

type boxConfig struct {
	height int
}

type layoutConfig struct {
	message boxConfig
	history boxConfig
	topics  boxConfig
}

type model struct {
	mqttClient *MQTTClient

	connection      string
	activeConn      string
	history         list.Model
	historyItems    []historyItem
	store           *history.Index
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

	layout layoutConfig
}

func initialModel(conns *Connections) *model {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.Focus()
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.Cursor.Style = ui.CursorStyle
	ti.TextStyle = ui.FocusedStyle
	// Defer width assignment until we know the terminal size
	ti.Width = 0

	ta := textarea.New()
	ta.Placeholder = "Enter Message"
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false
	ta.SetPromptFunc(0, func(i int) string {
		return fmt.Sprintf("%d> ", i+1)
	})
	promptColor := ui.ColGray
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(promptColor)
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(promptColor)
	ta.Blur()
	ta.Cursor.Style = ui.NoCursor
	// Set width once the WindowSizeMsg arrives
	ta.SetWidth(0)
	ta.SetHeight(6)
	ta.FocusedStyle.CursorLine = ui.FocusedStyle
	ta.BlurredStyle.CursorLine = ui.BlurredStyle

	var connModel Connections
	if conns != nil {
		connModel = *conns
	} else {
		connModel = NewConnectionsModel()
		if err := connModel.LoadProfiles(""); err != nil {
			fmt.Println("Warning:", err)
		}
	}
	connModel.ConnectionsList.SetShowStatusBar(false)
	for _, p := range connModel.Profiles {
		if _, ok := connModel.Statuses[p.Name]; !ok {
			connModel.Statuses[p.Name] = "disconnected"
		}
	}
	items := []list.Item{}
	for _, p := range connModel.Profiles {
		detail := connModel.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: connModel.Statuses[p.Name], detail: detail})
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
	topicsList.SetShowTitle(false)
	payloadList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	payloadList.DisableQuitKeybindings()
	payloadList.SetShowTitle(false)
	vp := viewport.New(0, 0)

	order := []string{"topics", "topic", "message", "history"}
	saved := loadState()

	m := &model{
		history:         hist,
		historyItems:    []historyItem{},
		store:           nil,
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
		layout: layoutConfig{
			message: boxConfig{height: 6},
			history: boxConfig{height: 10},
			topics:  boxConfig{height: 3},
		},
	}
	m.focusMap = map[string]focusable{
		"topic":   &m.topicInput,
		"message": &m.messageInput,
	}
	hDel.m = m
	m.history.SetDelegate(hDel)
	if idx, err := history.Open(""); err == nil {
		m.store = idx
		msgs := idx.Search(nil, time.Time{}, time.Time{}, "")
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
			m.historyItems = append(m.historyItems, items[i].(historyItem))
		}
		m.history.SetItems(items)
	}
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

func (m *model) sortTopics() {
	if len(m.topics) == 0 {
		return
	}
	sel := ""
	if m.selectedTopic >= 0 && m.selectedTopic < len(m.topics) {
		sel = m.topics[m.selectedTopic].title
	}
	sort.SliceStable(m.topics, func(i, j int) bool {
		if m.topics[i].active != m.topics[j].active {
			return m.topics[i].active && !m.topics[j].active
		}
		return m.topics[i].title < m.topics[j].title
	})
	if sel != "" {
		for i, t := range m.topics {
			if t.title == sel {
				m.selectedTopic = i
				break
			}
		}
	}
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
	m.sortTopics()
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
	m.sortTopics()
}

func (m *model) topicAtPosition(x, y int) int {
	for i, b := range m.chipBounds {
		if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
			return i
		}
	}
	return -1
}

func (m *model) historyIndexAt(y int) int {
	rel := y - (m.elemPos["history"] + 1) + m.viewport.YOffset
	if rel < 0 {
		return -1
	}
	h := 2 // historyDelegate height
	idx := rel / h
	start := m.history.Paginator.Page * m.history.Paginator.PerPage
	i := start + idx
	if i >= len(m.history.Items()) || i < 0 {
		return -1
	}
	return i
}

func (m *model) startConfirm(prompt string, action func()) {
	m.confirmPrompt = prompt
	m.confirmAction = action
	m.prevMode = m.mode
	m.mode = modeConfirmDelete
}

func (m *model) subscribeActiveTopics() {
	if m.mqttClient == nil {
		return
	}
	for _, t := range m.topics {
		if t.active {
			m.mqttClient.Subscribe(t.title, 0, nil)
		}
	}
}

func (m *model) refreshConnectionItems() {
	items := []list.Item{}
	for _, p := range m.connections.Profiles {
		status := m.connections.Statuses[p.Name]
		detail := m.connections.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	m.connections.ConnectionsList.SetItems(items)
}
