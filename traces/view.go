package traces

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewTraces lists configured traces and their state.
func (t *Component) viewTraces() string {
	t.api.ResetElemPos()
	t.api.SetElemPos(IDList, 1)
	listView := t.list.View()
	help := ui.InfoStyle.Render("[a] add  [enter] start/stop  [v] view  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := t.api.FocusedID() == IDList
	view := ui.LegendBox(content, "Traces", t.api.Width()-2, 0, ui.ColBlue, focused, -1)
	return t.api.OverlayHelp(view)
}
