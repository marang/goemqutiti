package emqutiti

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewTraces lists configured traces and their state.
func (t *tracesComponent) viewTraces() string {
	t.m.ui.elemPos = map[string]int{}
	t.m.ui.elemPos[idTraceList] = 1
	listView := t.list.View()
	help := ui.InfoStyle.Render("[a] add  [enter] start/stop  [v] view  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := t.m.ui.focusOrder[t.m.ui.focusIndex] == idTraceList
	view := ui.LegendBox(content, "Traces", t.m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return t.m.overlayHelp(view)
}
