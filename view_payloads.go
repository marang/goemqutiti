package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// viewPayloads shows stored payloads for reuse.
func (m model) viewPayloads() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idPayloadList] = 1
	listView := m.message.list.View()
	help := ui.InfoStyle.Render("[enter] load  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.ui.focusOrder[m.ui.focusIndex] == idPayloadList
	view := ui.LegendBox(content, "Payloads", m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return m.overlayHelp(view)
}
