package emqutiti

import (
	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/focus"
)

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
