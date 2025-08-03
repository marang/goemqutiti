package emqutiti

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

	"github.com/marang/emqutiti/importer"
	"github.com/marang/emqutiti/ui"

	"github.com/marang/emqutiti/history"
)

func initConnections(conns *Connections) (connectionsState, error) {
	var connModel Connections
	var loadErr error
	if conns != nil {
		connModel = *conns
	} else {
		connModel = NewConnectionsModel()
		if err := connModel.LoadProfiles(""); err != nil {
			loadErr = err
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
	statusChan := make(chan string, 10)
	saved := loadState()
	cs := connectionsState{
		connection:  "",
		active:      "",
		manager:     connModel,
		form:        nil,
		deleteIndex: 0,
		statusChan:  statusChan,
		saved:       saved,
	}
	return cs, loadErr
}

func initHistory() (historyState, historyDelegate) {
	hDel := historyDelegate{}
	hist := list.New([]list.Item{}, hDel, 0, 0)
	hist.SetShowTitle(false)
	hist.SetShowStatusBar(false)
	hist.SetShowPagination(false)
	hist.DisableQuitKeybindings()
	hs := historyState{
		list:            hist,
		items:           []history.Item{},
		store:           nil,
		selectionAnchor: -1,
		detail:          viewport.New(0, 0),
	}
	if idx, err := history.OpenStore(""); err == nil {
		hs.store = idx
		msgs := idx.Search(false, nil, time.Time{}, time.Time{}, "")
		var items []list.Item
		hs.items, items = messagesToHistoryItems(msgs)
		hs.list.SetItems(items)
	}
	return hs, hDel
}

func initTopics() topicsState {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.Focus()
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.Cursor.Style = ui.CursorStyle
	ti.TextStyle = ui.FocusedStyle
	ti.Width = 0
	topicsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	topicsList.DisableQuitKeybindings()
	topicsList.SetShowTitle(false)
	ts := topicsState{
		input: ti,
		items: []topicItem{},
		list:  topicsList,
		panes: topicsPanes{
			subscribed:   paneState{sel: 0, page: 0, index: 0},
			unsubscribed: paneState{sel: 0, page: 0, index: 1},
			active:       0,
		},
		selected:   -1,
		chipBounds: []chipBound{},
		vp:         viewport.New(0, 0),
	}
	return ts
}

func initMessage() messageState {
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
	ta.SetWidth(0)
	ta.SetHeight(6)
	ta.FocusedStyle.CursorLine = ui.FocusedStyle
	ta.BlurredStyle.CursorLine = ui.BlurredStyle
	ms := messageState{
		input: ta,
	}
	return ms
}

func initTraces() (tracesState, traceMsgDelegate) {
	traceList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	traceList.DisableQuitKeybindings()
	traceList.SetShowTitle(false)
	traceDel := traceMsgDelegate{}
	traceView := list.New([]list.Item{}, traceDel, 0, 0)
	traceView.DisableQuitKeybindings()
	traceView.SetShowTitle(false)
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
	ts := tracesState{
		list:  traceList,
		items: traceData,
		form:  nil,
		view:  traceView,
	}
	return ts, traceDel
}

func initUI(order []string) uiState {
	vp := viewport.New(0, 0)
	return uiState{
		focusIndex: 0,
		modeStack:  []appMode{modeClient},
		width:      0,
		height:     0,
		viewport:   vp,
		elemPos:    map[string]int{},
		focusOrder: order,
	}
}

func initLayout() layoutConfig {
	return layoutConfig{
		message: boxConfig{height: 6},
		history: boxConfig{height: 10},
		topics:  boxConfig{height: 1},
		trace:   boxConfig{height: 10},
	}
}

// initialModel creates the main program model with optional connection data.
func initialModel(conns *Connections) (*model, error) {
	order := append([]string(nil), focusByMode[modeClient]...)
	cs, loadErr := initConnections(conns)
	hs, hDel := initHistory()
	ms := initMessage()
	tr, traceDel := initTraces()
	m := &model{
		connections: cs,
		ui:          initUI(order),
		layout:      initLayout(),
	}
	historyComp := newHistoryComponent(m, hs)
	m.history = historyComp
	msgComp := newMessageComponent(m, ms)
	m.message = msgComp
	m.help = newHelpComponent(m, &m.ui.width, &m.ui.height, &m.ui.elemPos)
	m.confirm = newConfirmComponent(m, m, nil, nil, nil)
	connComp := newConnectionsComponent(m, m.connectionsAPI())
	topicsComp := newTopicsComponent(m)
	m.topics = topicsComp
	m.payloads = newPayloadsComponent(m, &m.connections)
	tracesComp := newTracesComponent(m, tr, m.tracesStore())
	m.traces = tracesComp

	// Collect focusable elements from model and components.
	providers := []FocusableSet{m, connComp, historyComp, topicsComp, msgComp, m.payloads, tracesComp, m.help, m.confirm}
	m.focusables = map[string]Focusable{}
	for _, p := range providers {
		for id, f := range p.Focusables() {
			m.focusables[id] = f
		}
	}
	fitems := make([]Focusable, len(order))
	for i, id := range order {
		fitems[i] = m.focusables[id]
	}
	m.focus = NewFocusMap(fitems)
	m.history.list.SetDelegate(hDel)
	traceDel.t = m.traces
	m.traces.view.SetDelegate(traceDel)
	// Register mode components so that view and update logic can be
	// delegated based on the current application mode.
	m.components = map[appMode]Component{
		modeClient:         component{update: m.updateClient, view: m.viewClient},
		modeConnections:    connComp,
		modeEditConnection: component{update: m.updateForm, view: m.viewForm},
		modeConfirmDelete:  m.confirm,
		modeTopics:         topicsComp,
		modePayloads:       m.payloads,
		modeTracer:         tracesComp,
		modeEditTrace:      component{update: m.traces.UpdateForm, view: m.traces.ViewForm},
		modeViewTrace:      component{update: m.traces.UpdateView, view: m.traces.ViewMessages},
		modeHistoryFilter:  component{update: m.history.UpdateFilter, view: m.history.ViewFilter},
		modeHistoryDetail:  component{update: m.history.UpdateDetail, view: m.history.ViewDetail},
		modeHelp:           m.help,
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
				ApplyEnvVars(&cfg)
			} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
				cfg.Password = env
			}
			if client, err := NewMQTTClient(cfg, nil); err == nil {
				m.mqttClient = client
				m.connections.active = cfg.Name
				m.importer = importer.New(client, importFile)
				m.components[modeImporter] = m.importer
				m.setMode(modeImporter)
			} else {
				return nil, fmt.Errorf("connect error: %w", err)
			}
		}
	}
	m.topics.RebuildActiveTopicList()
	return m, loadErr
}

// Init enables initial Tea behavior such as mouse support.
func (m model) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}
