package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewConfirmDelete displays a confirmation dialog.
func (m model) viewConfirmDelete() string {
	m.ui.elemPos = map[string]int{}
	content := m.confirmPrompt
	if m.confirmInfo != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, m.confirmPrompt, m.confirmInfo)
	}
	content = lipgloss.NewStyle().Padding(1, 2).Render(content)
	box := ui.LegendBox(content, "Confirm", m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(m.ui.width, m.ui.height, lipgloss.Center, lipgloss.Center, box)
}
