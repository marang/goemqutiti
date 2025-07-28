package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/marang/goemqutiti/history"
	"github.com/marang/goemqutiti/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/goemqutiti/config"
	"github.com/marang/goemqutiti/tracer"
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
	modeTracer
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

type connectionsState struct {
	connection  string
	active      string
	manager     Connections
	form        *connectionForm
	deleteIndex int
	statusChan  chan string
	saved       map[string]connectionData
}

type historyState struct {
	list            list.Model
	items           []historyItem
	store           *history.Index
	selected        map[int]struct{}
	selectionAnchor int
}

type topicsState struct {
	input      textinput.Model
	items      []topicItem
	list       list.Model
	selected   int
	chipBounds []chipBound
	vp         viewport.Model
}

type messageState struct {
	input    textarea.Model
	payloads []payloadItem
	list     list.Model // payloadList reused when viewing payloads
}

type traceItem struct {
	key    string
	cfg    tracer.Config
	tracer *tracer.Tracer
}

func (t traceItem) FilterValue() string { return t.key }
func (t traceItem) Title() string       { return t.key }
func (t traceItem) Description() string {
	status := "stopped"
	if t.tracer != nil {
		if t.tracer.Running() {
			status = "running"
		} else if t.tracer.Planned() {
			status = "planned"
		}
	} else if time.Now().Before(t.cfg.Start) {
		status = "planned"
	}
	var parts []string
	counts := map[string]int{}
	if t.tracer != nil {
		counts = t.tracer.Counts()
	}
	for _, tp := range t.cfg.Topics {
		parts = append(parts, fmt.Sprintf("%s:%d", tp, counts[tp]))
	}
	if len(parts) > 0 {
		status += " " + strings.Join(parts, " ")
	}
	return status
}

type tracesState struct {
	list  list.Model
	items []traceItem
}

// uiState groups general UI information such as current focus and layout.
type uiState struct {
	focusIndex int            // index of the currently focused element
	mode       appMode        // current application mode
	prevMode   appMode        // mode prior to confirmations
	width      int            // terminal width
	height     int            // terminal height
	viewport   viewport.Model // scrolling container for the main view
	elemPos    map[string]int // cached Y positions of each box
	focusOrder []string       // order of focusable elements
}

type model struct {
	mqttClient *MQTTClient

	connections connectionsState
	history     historyState
	topics      topicsState
	message     messageState
	traces      tracesState

	ui uiState

	confirmPrompt string
	confirmAction func()

	layout layoutConfig

	focusMap map[string]focusable
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
	traceList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	traceList.DisableQuitKeybindings()
	traceList.SetShowTitle(false)
	vp := viewport.New(0, 0)

	order := []string{"topics", "topic", "message", "history"}
	saved := loadState()
	tracesCfg := loadTraces()
	var traceItems []list.Item
	var traceData []traceItem
	keys := make([]string, 0, len(tracesCfg))
	for k := range tracesCfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		it := traceItem{key: k, cfg: tracesCfg[k]}
		traceItems = append(traceItems, it)
		traceData = append(traceData, it)
	}
	traceList.SetItems(traceItems)

	m := &model{
		connections: connectionsState{
			connection:  "",
			active:      "",
			manager:     connModel,
			form:        nil,
			deleteIndex: 0,
			statusChan:  statusChan,
			saved:       saved,
		},
		history: historyState{
			list:            hist,
			items:           []historyItem{},
			store:           nil,
			selected:        make(map[int]struct{}),
			selectionAnchor: -1,
		},
		topics: topicsState{
			input:      ti,
			items:      []topicItem{},
			list:       topicsList,
			selected:   -1,
			chipBounds: []chipBound{},
			vp:         viewport.New(0, 0),
		},
		message: messageState{
			input:    ta,
			payloads: []payloadItem{},
			list:     payloadList,
		},
		traces: tracesState{
			list:  traceList,
			items: traceData,
		},
		ui: uiState{
			focusIndex: 0,
			mode:       modeClient,
			prevMode:   modeClient,
			width:      0,
			height:     0,
			viewport:   vp,
			elemPos:    map[string]int{},
			focusOrder: order,
		},
		layout: layoutConfig{
			message: boxConfig{height: 6},
			history: boxConfig{height: 10},
			topics:  boxConfig{height: 3},
		},
	}
	m.focusMap = map[string]focusable{
		"topic":   &m.topics.input,
		"message": &m.message.input,
	}
	hDel.m = m
	m.history.list.SetDelegate(hDel)
	if idx, err := history.Open(""); err == nil {
		m.history.store = idx
		msgs := idx.Search(nil, time.Time{}, time.Time{}, "")
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
			m.history.items = append(m.history.items, items[i].(historyItem))
		}
		m.history.list.SetItems(items)
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (m *model) hasTopic(topic string) bool {
	for _, t := range m.topics.items {
		if t.title == topic {
			return true
		}
	}
	return false
}

func (m *model) sortTopics() {
	if len(m.topics.items) == 0 {
		return
	}
	sel := ""
	if m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		sel = m.topics.items[m.topics.selected].title
	}
	sort.SliceStable(m.topics.items, func(i, j int) bool {
		if m.topics.items[i].active != m.topics.items[j].active {
			return m.topics.items[i].active && !m.topics.items[j].active
		}
		return m.topics.items[i].title < m.topics.items[j].title
	})
	if sel != "" {
		for i, t := range m.topics.items {
			if t.title == sel {
				m.topics.selected = i
				break
			}
		}
	}
}

func (m *model) toggleTopic(index int) {
	if index < 0 || index >= len(m.topics.items) {
		return
	}
	t := &m.topics.items[index]
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
	if index < 0 || index >= len(m.topics.items) {
		return
	}
	topic := m.topics.items[index]
	if m.mqttClient != nil {
		m.mqttClient.Unsubscribe(topic.title)
		m.appendHistory(topic.title, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", topic.title))
	}
	m.topics.items = append(m.topics.items[:index], m.topics.items[index+1:]...)
	if len(m.topics.items) == 0 {
		m.topics.selected = -1
	} else if m.topics.selected >= len(m.topics.items) {
		m.topics.selected = len(m.topics.items) - 1
	}
	m.sortTopics()
}

func (m *model) startTrace(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
	item := &m.traces.items[index]
	p, err := config.LoadProfile(item.cfg.Profile, "")
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	if p.FromEnv {
		config.ApplyEnvVars(p)
	} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
		p.Password = env
	}
	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	tr := tracer.New(item.cfg, client)
	if err := tr.Start(); err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		client.Disconnect()
		return
	}
	item.tracer = tr
}

func (m *model) stopTrace(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
	if tr := m.traces.items[index].tracer; tr != nil {
		tr.Stop()
	}
}

func (m *model) anyTraceRunning() bool {
	for i := range m.traces.items {
		if tr := m.traces.items[i].tracer; tr != nil && (tr.Running() || tr.Planned()) {
			return true
		}
	}
	return false
}

func (m *model) savePlannedTraces() {
	data := map[string]tracer.Config{}
	now := time.Now()
	for _, it := range m.traces.items {
		cfg := it.cfg
		if it.tracer != nil {
			cfg = it.tracer.Config()
		}
		if cfg.Start.After(now) {
			data[it.key] = cfg
		}
	}
	saveTraces(data)
}

func (m *model) topicAtPosition(x, y int) int {
	for i, b := range m.topics.chipBounds {
		if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
			return i
		}
	}
	return -1
}

func (m *model) historyIndexAt(y int) int {
	rel := y - (m.ui.elemPos["history"] + 1) + m.ui.viewport.YOffset
	if rel < 0 {
		return -1
	}
	h := 2 // historyDelegate height
	idx := rel / h
	start := m.history.list.Paginator.Page * m.history.list.Paginator.PerPage
	i := start + idx
	if i >= len(m.history.list.Items()) || i < 0 {
		return -1
	}
	return i
}

func (m *model) startConfirm(prompt string, action func()) {
	m.confirmPrompt = prompt
	m.confirmAction = action
	m.ui.prevMode = m.ui.mode
	m.ui.mode = modeConfirmDelete
}

func (m *model) subscribeActiveTopics() {
	if m.mqttClient == nil {
		return
	}
	for _, t := range m.topics.items {
		if t.active {
			m.mqttClient.Subscribe(t.title, 0, nil)
		}
	}
}

func (m *model) refreshConnectionItems() {
	items := []list.Item{}
	for _, p := range m.connections.manager.Profiles {
		status := m.connections.manager.Statuses[p.Name]
		detail := m.connections.manager.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	m.connections.manager.ConnectionsList.SetItems(items)
}
