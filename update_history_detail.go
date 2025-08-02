package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateHistoryDetail handles input when viewing a long history payload.
func (m *model) updateHistoryDetail(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd := m.setMode(modeClient)
			return cmd
		case "ctrl+d":
			return tea.Quit
		}
	}
	m.history.detail, cmd = m.history.detail.Update(msg)
	return cmd
}
