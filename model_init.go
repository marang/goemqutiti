package emqutiti

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	_ "github.com/marang/emqutiti/clientkeys"
	"github.com/marang/emqutiti/importer"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/traces"
	"github.com/marang/emqutiti/ui"

	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/payloads"
)

type navAdapter struct{ navigator }

func (n navAdapter) SetMode(mode int) tea.Cmd { return n.navigator.SetMode(appMode(mode)) }
func (n navAdapter) Width() int               { return n.navigator.Width() }
func (n navAdapter) Height() int              { return n.navigator.Height() }

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
func initialModel(conns *connections.Connections) (*model, error) {
	order := append([]string(nil), focusByMode[modeClient]...)
	cs, loadErr := initConnections(conns)
	st, _ := history.OpenStore("")
	ms := initMessage()
	tr, traceDel := traces.Init()
	m := &model{
		connections: cs,
		ui:          initUI(order),
		layout:      initLayout(),
	}
	m.history = history.NewComponent(historyModelAdapter{m}, st)
	msgComp := newMessageComponent(m, ms)
	m.message = msgComp
	m.help = newHelpComponent(m, &m.ui.width, &m.ui.height, &m.ui.elemPos)
	m.confirm = newConfirmComponent(m, m, nil, nil, nil)
	connComp := connections.NewComponent(navAdapter{m}, m.connectionsAPI())
	topicsComp := topics.New(m)
	m.topics = topicsComp
	m.payloads = payloads.New(m, &m.connections)
	tracesComp := traces.NewComponent(m, tr, m.tracesStore())
	m.traces = tracesComp

	// Collect focusable elements from model and components.
	providers := []FocusableSet{m, topicsComp, msgComp, m.payloads, tracesComp, m.help, m.confirm}
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
	traceDel.T = m.traces
	m.traces.ViewList().SetDelegate(traceDel)
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
		if p != nil {
			cfg := *p
			if cfg.FromEnv {
				connections.ApplyEnvVars(&cfg)
			} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
				cfg.Password = env
			}
			if client, err := NewMQTTClient(cfg, nil); err == nil {
				m.mqttClient = client
				m.connections.Active = cfg.Name
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
