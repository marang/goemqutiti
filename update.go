package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update routes messages based on the current mode.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		m.connections.Manager.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		// textinput.View() renders the prompt and cursor in addition
		// to the configured width. Reduce the width slightly so the
		// surrounding box stays within the terminal boundaries.
		m.topics.input.Width = msg.Width - 7
		m.message.input.SetWidth(msg.Width - 4)
		m.message.input.SetHeight(m.layout.message.height)
		if m.layout.history.height == 0 {
			m.layout.history.height = (msg.Height-1)/3 + 10
		}
		m.history.List().SetSize(msg.Width-4, m.layout.history.height)
		if m.layout.trace.height == 0 {
			m.layout.trace.height = msg.Height - 6
		}
		m.traces.view.SetSize(msg.Width-4, m.layout.trace.height)
		m.traces.list.SetSize(msg.Width-4, msg.Height-4)
		m.help.vp.Width = msg.Width - 4
		m.help.vp.Height = msg.Height - 4
		m.history.Detail().Width = msg.Width - 4
		m.history.Detail().Height = msg.Height - 4
		m.ui.viewport.Width = msg.Width
		// Reserve two lines for the info header at the top of the view.
		m.ui.viewport.Height = msg.Height - 2
		return m, nil
	case topicToggleMsg:
		cmd := m.handleTopicToggle(msg)
		return m, cmd
	case loadPayloadMsg:
		m.topics.setTopic(msg.topic)
		m.message.setPayload(msg.payload)
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
				cmd := m.history.UpdateFilter(msg)
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Next()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.SetSelected(0)
						m.topics.EnsureVisible(m.ui.width - 4)
					} else {
						m.topics.SetSelected(-1)
					}
				}
				return m, nil
			}
		case "shift+tab":
			if m.currentMode() == modeHistoryFilter {
				cmd := m.history.UpdateFilter(msg)
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Prev()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.SetSelected(0)
						m.topics.EnsureVisible(m.ui.width - 4)
					} else {
						m.topics.SetSelected(-1)
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

	if c, ok := m.components[m.currentMode()]; ok {
		cmd := c.Update(msg)
		return m, cmd
	}
	return m, nil
}
