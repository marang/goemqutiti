package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

// viewConnections shows the list of saved broker profiles.
func (m model) viewConnections() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idConnList] = 1
	listView := m.connections.manager.ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [del] delete  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	view := ui.LegendBox(content, "Brokers", m.ui.width-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}
