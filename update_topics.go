package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// updateTopics manages the topics list UI.
func (m *model) updateTopics(msg tea.Msg) (model, tea.Cmd) {
	var cmd, fcmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return *m, tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return *m, cmd
		case "left":
			if m.topics.panes.active == 1 {
				fcmd = m.setFocus(idTopicsEnabled)
			}
		case "right":
			if m.topics.panes.active == 0 {
				fcmd = m.setFocus(idTopicsDisabled)
			}
		case "delete":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				name := m.topics.items[i].title
				m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
				m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
					m.removeTopic(i)
					m.rebuildActiveTopicList()
				})
				return *m, listenStatus(m.connections.statusChan)
			}
		case "enter", " ":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				m.toggleTopic(i)
				m.rebuildActiveTopicList()
			}
		}
	}
	m.topics.list, cmd = m.topics.list.Update(msg)
	if m.topics.panes.active == 0 {
		m.topics.panes.subscribed.sel = m.topics.list.Index()
		m.topics.panes.subscribed.page = m.topics.list.Paginator.Page
	} else {
		m.topics.panes.unsubscribed.sel = m.topics.list.Index()
		m.topics.panes.unsubscribed.page = m.topics.list.Paginator.Page
	}
	m.topics.selected = m.indexForPane(m.topics.panes.active, m.topics.list.Index())
	return *m, tea.Batch(fcmd, cmd, listenStatus(m.connections.statusChan))
}
