package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

// viewTopics displays the topic manager list.
func (m model) viewTopics() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idTopicsEnabled] = 1
	m.ui.elemPos[idTopicsDisabled] = 1
	help := ui.InfoStyle.Render("[space] toggle  [del] delete  [esc] back")
	activeView := m.topics.list.View()
	var left, right string
	if m.topics.panes.active == 0 {
		other := list.New(m.unsubscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.unsubscribed.page
		other.Select(m.topics.panes.unsubscribed.sel)
		left = ui.LegendBox(activeView, "Enabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsEnabled, -1)
		right = ui.LegendBox(other.View(), "Disabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	} else {
		other := list.New(m.subscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.subscribed.page
		other.Select(m.topics.panes.subscribed.sel)
		left = ui.LegendBox(other.View(), "Enabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
		right = ui.LegendBox(activeView, "Disabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsDisabled, -1)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, panes, help)
	return m.overlayHelp(content)
}
