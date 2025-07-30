package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/config"
	"github.com/marang/goemqutiti/ui"
)

// initialModel creates the main program model with optional connection data.
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
	traceDel := traceMsgDelegate{}
	traceView := list.New([]list.Item{}, traceDel, 0, 0)
	traceView.DisableQuitKeybindings()
	traceView.SetShowTitle(false)
	vp := viewport.New(0, 0)

	order := []string{"topics", "topic", "message", "history", "help"}
	saved := loadState()
	tracesCfg := loadTraces()
	var traceItems []list.Item
	var traceData []*traceItem
	keys := make([]string, 0, len(tracesCfg))
	for k := range tracesCfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		it := &traceItem{key: k, cfg: tracesCfg[k]}
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
			form:  nil,
			view:  traceView,
		},
		help: helpState{
			vp: viewport.New(0, 0),
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
			trace:   boxConfig{height: 10},
		},
	}
	m.focusMap = map[string]focusable{
		"topic":   &m.topics.input,
		"message": &m.message.input,
		"help":    &m.help,
	}
	hDel.m = m
	m.history.list.SetDelegate(hDel)
	traceDel.m = m
	m.traces.view.SetDelegate(traceDel)
	if idx, err := openHistoryStore(""); err == nil {
		m.history.store = idx
		msgs := idx.Search(nil, time.Time{}, time.Time{}, "")
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			items[i] = historyItem{timestamp: mmsg.Timestamp, topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
			m.history.items = append(m.history.items, items[i].(historyItem))
		}
		m.history.list.SetItems(items)
	}

	if importFile != "" {
		var p *Profile
		if profileName != "" {
			for i := range m.connections.manager.Profiles {
				if m.connections.manager.Profiles[i].Name == profileName {
					p = &m.connections.manager.Profiles[i]
					break
				}
			}
		} else if m.connections.manager.DefaultProfileName != "" {
			for i := range m.connections.manager.Profiles {
				if m.connections.manager.Profiles[i].Name == m.connections.manager.DefaultProfileName {
					p = &m.connections.manager.Profiles[i]
					break
				}
			}
		}
		if p == nil && len(m.connections.manager.Profiles) > 0 {
			p = &m.connections.manager.Profiles[0]
		}
		if p != nil {
			cfg := *p
			if cfg.FromEnv {
				config.ApplyEnvVars(&cfg)
			} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
				cfg.Password = env
			}
			if client, err := NewMQTTClient(cfg, nil); err == nil {
				m.mqttClient = client
				m.connections.active = cfg.Name
				m.importWizard = NewImportWizard(client, importFile)
				m.setMode(modeImporter)
			} else {
				fmt.Println("connect error:", err)
			}
		}
	}
	return m
}

// Init enables initial Tea behavior such as mouse support.
func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}
