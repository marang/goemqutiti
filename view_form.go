package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// viewForm renders the add/edit broker form alongside the list.
func (m *model) viewForm() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[constants.IDConnList] = 1
	if m.connections.Form == nil {
		return ""
	}
	cw, ch := calcConnectionsSize(m.ui.width/2, m.ui.height)
	m.connections.Manager.ConnectionsList.SetSize(cw, ch)
	listView := ui.LegendBox(m.connections.Manager.ConnectionsList.View(), "Brokers", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	formLabel := "Add Broker"
	if m.connections.Form.Index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := ui.LegendBox(m.connections.Form.View(), formLabel, m.ui.width/2-2, 0, ui.ColBlue, true, -1)
	view := lipgloss.JoinHorizontal(lipgloss.Top, listView, formView)
	return m.overlayHelp(view)
}
