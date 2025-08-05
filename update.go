package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/payloads"
	"github.com/marang/emqutiti/topics"
)

// Update routes messages based on the current mode.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.handleWindowSize(msg)
	case topics.ToggleMsg:
		return m, m.handleTopicToggle(msg)
	case payloads.LoadMsg:
		m.topics.SetTopic(msg.Topic)
		m.message.SetPayload(msg.Payload)
		return m, nil
	case tea.KeyMsg:
		if cmd, handled := m.handleKeyNav(msg); handled {
			return m, cmd
		}
	}

	if c, ok := m.components[m.CurrentMode()]; ok {
		cmd := c.Update(msg)
		return m, cmd
	}
	return m, nil
}
