package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewForm renders the add/edit broker form alongside the list.
func (m model) viewForm() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idConnList] = 1
	if m.connections.form == nil {
		return ""
	}
	listView := ui.LegendBox(m.connections.manager.ConnectionsList.View(), "Brokers", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	formLabel := "Add Broker"
	if m.connections.form.index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := ui.LegendBox(m.connections.form.View(), formLabel, m.ui.width/2-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(lipgloss.JoinHorizontal(lipgloss.Top, listView, formView))
}
