package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewTraces lists configured traces and their state.
func (m *model) viewTraces() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idTraceList] = 1
	listView := m.traces.list.View()
	help := ui.InfoStyle.Render("[a] add  [enter] start/stop  [v] view  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.ui.focusOrder[m.ui.focusIndex] == idTraceList
	view := ui.LegendBox(content, "Traces", m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return m.overlayHelp(view)
}
