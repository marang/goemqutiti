package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewHistoryDetail renders the full payload of a history message.
func (m model) viewHistoryDetail() string {
	m.ui.elemPos = map[string]int{}
	lines := strings.Split(m.history.detail.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	lines = append(lines, help)
	content := strings.Join(lines, "\n")
	sp := -1.0
	if m.history.detail.Height < lipgloss.Height(content) {
		sp = m.history.detail.ScrollPercent()
	}
	view := ui.LegendBox(content, "Message", m.ui.width-2, m.ui.height-2, ui.ColGreen, true, sp)
	return m.overlayHelp(view)
}
