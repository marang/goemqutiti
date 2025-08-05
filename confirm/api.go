package confirm

import tea "github.com/charmbracelet/bubbletea"

// API exposes confirmation dialogs to components.
type API interface {
	StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func())
}

// Navigator defines navigation helpers required by the confirm dialog.
type Navigator interface {
	SetConfirmMode() tea.Cmd
	SetPreviousMode() tea.Cmd
	Width() int
	Height() int
	ScrollToFocused()
}

// StatusListener provides status updates for components.
type StatusListener interface {
	ListenStatus() tea.Cmd
}
