package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// updateConfirmDelete processes confirmation dialog key presses.
func (m *model) updateConfirmDelete(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return *m, tea.Quit
		case "y":
			if m.confirmAction != nil {
				m.confirmAction()
				m.confirmAction = nil
			}
			if m.confirmCancel != nil {
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return *m, tea.Batch(cmds...)
		case "n", "esc":
			if m.confirmCancel != nil {
				m.confirmCancel()
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return *m, tea.Batch(cmds...)
		}
	}
	return *m, listenStatus(m.connections.statusChan)
}

// updateTopics manages the topics list UI.
func (m model) updateTopics(msg tea.Msg) (model, tea.Cmd) {
	var cmd, fcmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return m, cmd
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
				return m, listenStatus(m.connections.statusChan)
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
	return m, tea.Batch(fcmd, cmd, listenStatus(m.connections.statusChan))
}

// updatePayloads manages the stored payloads list.
func (m model) updatePayloads(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return m, cmd
		case "delete":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					m.message.payloads = append(m.message.payloads[:i], m.message.payloads[i+1:]...)
					items = append(items[:i], items[i+1:]...)
					m.message.list.SetItems(items)
				}
			}
			return m, listenStatus(m.connections.statusChan)
		case "enter":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					m.topics.input.SetValue(pi.topic)
					m.message.input.SetValue(pi.payload)
					cmd := m.setMode(modeClient)
					return m, cmd
				}
			}
		}
	}
	m.message.list, cmd = m.message.list.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.connections.statusChan))
}

// updateSelectionRange selects history entries from the anchor to idx.
func (m *model) updateSelectionRange(idx int) {
	start := m.history.selectionAnchor
	end := idx
	if start > end {
		start, end = end, start
	}
	for i := range m.history.items {
		m.history.items[i].isSelected = nil
	}
	for i := start; i <= end && i < len(m.history.items); i++ {
		v := true
		m.history.items[i].isSelected = &v
	}
}

// Update routes messages based on the current mode.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		m.connections.manager.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		// textinput.View() renders the prompt and cursor in addition
		// to the configured width. Reduce the width slightly so the
		// surrounding box stays within the terminal boundaries.
		m.topics.input.Width = msg.Width - 7
		m.message.input.SetWidth(msg.Width - 4)
		m.message.input.SetHeight(m.layout.message.height)
		if m.layout.history.height == 0 {
			m.layout.history.height = (msg.Height-1)/3 + 10
		}
		m.history.list.SetSize(msg.Width-4, m.layout.history.height)
		if m.layout.trace.height == 0 {
			m.layout.trace.height = msg.Height - 6
		}
		m.traces.view.SetSize(msg.Width-4, m.layout.trace.height)
		m.traces.list.SetSize(msg.Width-4, msg.Height-4)
		m.help.vp.Width = msg.Width - 4
		m.help.vp.Height = msg.Height - 4
		m.ui.viewport.Width = msg.Width
		// Reserve two lines for the info header at the top of the view.
		m.ui.viewport.Height = msg.Height - 2
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+up", "ctrl+k":
			m.ui.viewport.ScrollUp(1)
			return m, nil
		case "ctrl+down", "ctrl+j":
			m.ui.viewport.ScrollDown(1)
			return m, nil
		case "tab":
			if m.currentMode() == modeHistoryFilter {
				nm, cmd := m.updateHistoryFilter(msg)
				*m = nm
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Next()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.selected = 0
						m.ensureTopicVisible()
					} else {
						m.topics.selected = -1
					}
				}
				return m, nil
			}
		case "shift+tab":
			if m.currentMode() == modeHistoryFilter {
				nm, cmd := m.updateHistoryFilter(msg)
				*m = nm
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Prev()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.selected = 0
						m.ensureTopicVisible()
					} else {
						m.topics.selected = -1
					}
				}
				return m, nil
			}
		}
		if m.currentMode() != modeHistoryFilter &&
			(msg.String() == "enter" || msg.String() == " " || msg.String() == "space") &&
			m.help.Focused() {
			cmd := m.setMode(modeHelp)
			return m, cmd
		}
	}

	switch m.currentMode() {
	case modeClient:
		cmd := m.updateClient(msg)
		return m, cmd
	case modeConnections:
		nm, cmd := m.updateConnections(msg)
		*m = nm
		return m, cmd
	case modeEditConnection:
		nm, cmd := m.updateForm(msg)
		*m = nm
		return m, cmd
	case modeConfirmDelete:
		nm, cmd := m.updateConfirmDelete(msg)
		*m = nm
		return m, cmd
	case modeTopics:
		nm, cmd := m.updateTopics(msg)
		*m = nm
		return m, cmd
	case modePayloads:
		nm, cmd := m.updatePayloads(msg)
		*m = nm
		return m, cmd
	case modeTracer:
		nm, cmd := m.updateTraces(msg)
		*m = nm
		return m, cmd
	case modeEditTrace:
		nm, cmd := m.updateTraceForm(msg)
		*m = nm
		return m, cmd
	case modeViewTrace:
		nm, cmd := m.updateTraceView(msg)
		*m = nm
		return m, cmd
	case modeImporter:
		nm, cmd := m.updateImporter(msg)
		*m = nm
		return m, cmd
	case modeHistoryFilter:
		nm, cmd := m.updateHistoryFilter(msg)
		*m = nm
		return m, cmd
	case modeHelp:
		nm, cmd := m.updateHelp(msg)
		*m = nm
		return m, cmd
	default:
		return m, nil
	}
}
