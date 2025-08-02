package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// Navigator defines minimal navigation features for components.
type Navigator interface {
	setMode(appMode) tea.Cmd
	previousMode() appMode
}
