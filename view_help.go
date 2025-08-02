package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

func (m *model) viewHelp() string {
	m.ui.elemPos = map[string]int{}
	m.help.vp.SetContent(helpText)
	content := m.help.vp.View()
	sp := -1.0
	if m.help.vp.Height < lipgloss.Height(content) {
		sp = m.help.vp.ScrollPercent()
	}
	box := ui.LegendBox(content, "Help", m.ui.width-2, m.ui.height-2, ui.ColGreen, true, sp)
	return box
}
