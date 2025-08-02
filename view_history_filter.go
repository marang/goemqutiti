package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

// viewHistoryFilter displays the history filter form.
func (m model) viewHistoryFilter() string {
	m.ui.elemPos = map[string]int{}
	if m.history.filterForm == nil {
		return ""
	}
	content := lipgloss.NewStyle().Padding(1, 2).Render(m.history.filterForm.View())
	box := ui.LegendBox(content, "Filter", m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(m.ui.width, m.ui.height, lipgloss.Center, lipgloss.Center, box)
}
