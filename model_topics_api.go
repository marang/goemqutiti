package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// ShowClient switches the UI back to the main client view.
func (m *model) ShowClient() tea.Cmd {
	return m.SetMode(modeClient)
}
