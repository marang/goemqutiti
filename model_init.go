package emqutiti

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/confirm"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/focus"
	"github.com/marang/emqutiti/help"
	"github.com/marang/emqutiti/importer"
	"github.com/marang/emqutiti/layout"
	"github.com/marang/emqutiti/logs"
	"github.com/marang/emqutiti/message"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/traces"
	"github.com/marang/emqutiti/ui"

	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/payloads"
)

type navAdapter struct{ navigator }

func (n navAdapter) SetMode(mode constants.AppMode) tea.Cmd { return n.navigator.SetMode(mode) }
func (n navAdapter) Width() int                             { return n.navigator.Width() }
func (n navAdapter) Height() int                            { return n.navigator.Height() }
func (n navAdapter) PreviousMode() constants.AppMode        { return n.navigator.PreviousMode() }

func initConnections(conns *connections.Connections) (connections.State, error) {
	var connModel connections.Connections
	var loadErr error
	if conns != nil {
		connModel = *conns
	} else {
		connModel = connections.NewConnectionsModel()
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
	statusChan := make(chan string, 10)
	saved := connections.LoadState()
	cs := connections.State{
		Connection:  "",
		Active:      "",
		Manager:     connModel,
		Form:        nil,
		DeleteIndex: 0,
		StatusChan:  statusChan,
		Saved:       saved,
	}
	cs.RefreshConnectionItems()
	return cs, loadErr
}

func initMessage() message.State {
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
	ms := message.State{
		TA: ta,
	}
	return ms
}

func initUI(order []string) uiState {
	vp := viewport.New(0, 0)
	fm := make(map[string]int, len(order))
	for i, id := range order {
		fm[id] = i
	}
	return uiState{
		focusIndex: 0,
		modeStack:  []constants.AppMode{constants.ModeClient},
		width:      0,
		height:     0,
		viewport:   vp,
		elemPos:    map[string]int{},
		focusOrder: order,
		focusMap:   fm,
	}
}

func initLayout() layout.Manager {
	return layout.Manager{
		Message: layout.Box{Height: 6},
		History: layout.Box{Height: 10},
		Topics:  layout.Box{Height: 1},
		Trace:   layout.Box{Height: 10},
	}
}

// initComponents registers focusable elements and mode components.
func initComponents(m *model, order []string, connComp Component) {
	providers := []focus.FocusableSet{m, m.topics, m.message, m.payloads, m.traces, m.help}
	m.focusables = map[string]focus.Focusable{}
	for _, p := range providers {
		for id, f := range p.Focusables() {
			m.focusables[id] = f
		}
	}
	fitems := make([]focus.Focusable, len(order))
	for i, id := range order {
		fitems[i] = m.focusables[id]
	}
	m.focus = focus.NewFocusMap(fitems)
	m.components = map[constants.AppMode]Component{
		constants.ModeClient:         component{update: m.updateClient, view: m.viewClient},
		constants.ModeConnections:    connComp,
		constants.ModeEditConnection: component{update: m.updateConnectionForm, view: m.viewForm},
		constants.ModeConfirmDelete:  m.confirm,
		constants.ModeTopics:         m.topics,
		constants.ModePayloads:       m.payloads,
		constants.ModeTracer:         m.traces,
		constants.ModeEditTrace:      component{update: m.traces.UpdateForm, view: m.traces.ViewForm},
		constants.ModeViewTrace:      component{update: m.traces.UpdateView, view: m.traces.ViewMessages},
		constants.ModeTraceFilter:    component{update: m.traces.UpdateFilter, view: m.traces.ViewFilter},
		constants.ModeHistoryFilter:  component{update: m.history.UpdateFilter, view: m.history.ViewFilter},
		constants.ModeHistoryDetail:  component{update: m.history.UpdateDetail, view: m.history.ViewDetail},
		constants.ModeHelp:           m.help,
		constants.ModeLogs:           m.logs,
	}
}

// initImporter bootstraps the importer for a selected profile.
func initImporter(m *model) error {
	if importFile == "" {
		return nil
	}
	var p *connections.Profile
	if profileName != "" {
		for i := range m.connections.Manager.Profiles {
			if m.connections.Manager.Profiles[i].Name == profileName {
				p = &m.connections.Manager.Profiles[i]
				break
			}
		}
	} else if m.connections.Manager.DefaultProfileName != "" {
		for i := range m.connections.Manager.Profiles {
			if m.connections.Manager.Profiles[i].Name == m.connections.Manager.DefaultProfileName {
				p = &m.connections.Manager.Profiles[i]
				break
			}
		}
	}
	if p == nil && len(m.connections.Manager.Profiles) > 0 {
		p = &m.connections.Manager.Profiles[0]
	}
	if p == nil {
		return nil
	}
	cfg := *p
	if cfg.FromEnv {
		connections.ApplyEnvVars(&cfg)
	}
	connections.ApplyDefaultPassword(&cfg)
	client, err := NewMQTTClient(cfg, nil)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	m.mqttClient = client
	m.connections.Active = cfg.Name
	m.importer = importer.New(client, importFile)
	m.components[constants.ModeImporter] = m.importer
	m.SetMode(constants.ModeImporter)
	return nil
}

// initialModel creates the main program model with optional connection data.
func initialModel(conns *connections.Connections) (*model, error) {
	order := append([]string(nil), focusByMode[constants.ModeClient]...)
	cs, loadErr := initConnections(conns)
	st, herr := history.OpenStore("")
	if herr != nil && loadErr == nil {
		loadErr = herr
	}
	if herr != nil {
		fmt.Fprintf(os.Stderr, "history store error: %v\n", herr)
		st = nil
	}
	ms := initMessage()
	tr := traces.Init()
	m := &model{
		connections: cs,
		ui:          initUI(order),
		layout:      initLayout(),
	}
	m.history = history.NewComponent(historyModelAdapter{m}, st)
	if herr != nil {
		m.history.Append("", "", "log", false, fmt.Sprintf("history store error: %v", herr))
	}
	m.message = message.NewComponent(m, ms)
	m.logs = logs.New(navAdapter{m}, &m.ui.width, &m.ui.height, &m.ui.elemPos)
	m.help = help.New(navAdapter{m}, &m.ui.width, &m.ui.height, &m.ui.elemPos)
	m.confirm = confirm.NewDialog(m, m, nil, nil, nil)
	connComp := connections.NewComponent(navAdapter{m}, m)
	m.topics = topics.New(m)
	m.payloads = payloads.New(m, &m.connections)
	m.traces = traces.NewComponent(m, tr, m.tracesStore())
	initComponents(m, order, connComp)
	if err := initImporter(m); err != nil {
		return nil, err
	}
	m.topics.RebuildActiveTopicList()
	return m, loadErr
}

// Init enables initial Tea behavior such as mouse support.
func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnableMouseCellMotion}
	if profileName == "" {
		if name := m.connections.Manager.DefaultProfileName; name != "" {
			for _, p := range m.connections.Manager.Profiles {
				if p.Name == name {
					cmds = append(cmds, m.Connect(p))
					break
				}
			}
		}
	}
	return tea.Batch(cmds...)
}
