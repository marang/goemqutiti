package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// historyModel defines the dependencies historyComponent requires from the model.
type historyModel interface {
	SetMode(appMode) tea.Cmd
	PreviousMode() appMode
	CurrentMode() appMode
	SetFocus(id string) tea.Cmd
	Width() int
	Height() int
	OverlayHelp(string) string
}

var _ historyModel = (*model)(nil)
