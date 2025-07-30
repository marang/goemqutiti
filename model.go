package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/marang/goemqutiti/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	idTopics      = "topics"       // topics chip list
	idTopic       = "topic"        // topic input box
	idMessage     = "message"      // message input box
	idHistory     = "history"      // history list
	idConnList    = "conn-list"    // broker list
	idTopicList   = "topic-list"   // topics manager list
	idPayloadList = "payload-list" // payload manager list
	idTraceList   = "trace-list"   // traces manager list
	idHelp        = "help"         // help icon
)

// focusByMode lists focusable elements for each view. The first element is
// focused when switching to the corresponding mode.
var focusByMode = map[appMode][]string{
	modeClient:         {idTopics, idTopic, idMessage, idHistory, idHelp},
	modeConnections:    {idConnList, idHelp},
	modeEditConnection: {idConnList, idHelp},
	modeConfirmDelete:  {idHelp},
	modeTopics:         {idTopicList, idHelp},
	modePayloads:       {idPayloadList, idHelp},
	modeTracer:         {idTraceList, idHelp},
	modeEditTrace:      {idHelp},
	modeViewTrace:      {idHelp},
	modeImporter:       {idHelp},
	modeHelp:           {idHelp},
}

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
	xPos, yPos    int
	width, height int
}

type historyItem struct {
	timestamp time.Time
	topic     string
	payload   string
	kind      string // pub, sub, log
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
	modeEditTrace
	modeViewTrace
	modeImporter
	modeHelp
)

type connectionData struct {
	Topics   []topicItem
	Payloads []payloadItem
}

type boxConfig struct {
	height int
}

type layoutConfig struct {
	message boxConfig
	history boxConfig
	topics  boxConfig
	trace   boxConfig
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
	store           *HistoryStore
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
	cfg    TracerConfig
	tracer *Tracer
	counts map[string]int
	loaded bool
}

func (t *traceItem) FilterValue() string { return t.key }
func (t *traceItem) Title() string       { return t.key }
func (t *traceItem) Description() string {
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
	counts := t.counts
	if t.tracer != nil {
		counts = t.tracer.Counts()
	} else if !t.loaded {
		if c, err := tracerLoadCounts(t.cfg.Profile, t.cfg.Key, t.cfg.Topics); err == nil {
			t.counts = c
			t.loaded = true
			counts = c
		}
	}
	for _, tp := range t.cfg.Topics {
		parts = append(parts, fmt.Sprintf("%s:%d", tp, counts[tp]))
	}
	if len(parts) > 0 {
		status += " " + strings.Join(parts, " ")
	}
	var times []string
	if !t.cfg.Start.IsZero() {
		times = append(times, t.cfg.Start.Format(time.RFC3339))
	}
	if !t.cfg.End.IsZero() {
		times = append(times, t.cfg.End.Format(time.RFC3339))
	}
	if len(times) > 0 {
		status += " " + strings.Join(times, " -> ")
	}
	return status
}

type tracesState struct {
	list    list.Model
	items   []*traceItem
	form    *traceForm
	view    list.Model
	viewKey string
}

type helpState struct {
	vp      viewport.Model
	focused bool
}

func (h *helpState) Focus() tea.Cmd {
	h.focused = true
	return nil
}

func (h *helpState) Blur() { h.focused = false }

func (h helpState) Focused() bool { return h.focused }

func (h helpState) View() string { return "" }

// uiState groups general UI information such as current focus and layout.
type uiState struct {
	focusIndex int            // index of the currently focused element
	modeStack  []appMode      // mode stack, index 0 is current
	width      int            // terminal width
	height     int            // terminal height
	viewport   viewport.Model // scrolling container for the main view
	elemPos    map[string]int // cached Y positions of each box
	focusOrder []string       // order of focusable elements
}

type model struct {
	mqttClient *MQTTClient

	connections  connectionsState
	history      historyState
	topics       topicsState
	message      messageState
	traces       tracesState
	help         helpState
	importWizard *ImportWizard

	ui uiState

	confirmPrompt string
	confirmInfo   string
	confirmAction func()

	layout layoutConfig

	focusables map[string]Focusable
	focus      *FocusMap
}
