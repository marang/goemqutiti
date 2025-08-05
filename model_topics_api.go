package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// ShowClient switches the UI back to the main client view.
func (m *model) ShowClient() tea.Cmd {
	return m.SetMode(constants.ModeClient)
}
