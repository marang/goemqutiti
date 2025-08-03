package emqutiti

import tea "github.com/charmbracelet/bubbletea"

type navigator interface {
	SetMode(appMode) tea.Cmd
	PreviousMode() appMode
	Width() int
	Height() int
}

type ConfirmNavigator interface {
	SetMode(appMode) tea.Cmd
	PreviousMode() appMode
	Width() int
	Height() int
	ScrollToFocused()
}

type StatusListener interface {
	ListenStatus() tea.Cmd
}
