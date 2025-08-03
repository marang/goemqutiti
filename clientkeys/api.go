package clientkeys

import tea "github.com/charmbracelet/bubbletea"

// Handler processes keyboard input in client mode.
type Handler interface {
	// HandleClientKey reacts to a key message and optionally returns a Tea command.
	HandleClientKey(msg tea.KeyMsg) tea.Cmd
}

// HandleClientKey dispatches a key message to the provided handler.
func HandleClientKey(h Handler, msg tea.KeyMsg) tea.Cmd {
	if h == nil {
		return nil
	}
	return h.HandleClientKey(msg)
}
