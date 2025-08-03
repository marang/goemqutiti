package connections

import tea "github.com/charmbracelet/bubbletea"

// StatusMessage wraps connection status text for Tea messages.
type StatusMessage string

// ListenStatus retrieves status updates from the status channel.
func ListenStatus(ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return StatusMessage(msg)
	}
}

// FlushStatus drains all pending messages from the status channel.
func FlushStatus(ch chan string) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
