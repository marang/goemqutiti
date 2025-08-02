package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateConfirmDelete processes confirmation dialog key presses.
func (m *model) updateConfirmDelete(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "y":
			if m.confirmAction != nil {
				m.confirmAction()
				m.confirmAction = nil
			}
			if m.confirmCancel != nil {
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return tea.Batch(cmds...)
		case "n", "esc":
			if m.confirmCancel != nil {
				m.confirmCancel()
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return tea.Batch(cmds...)
		}
	}
	return listenStatus(m.connections.statusChan)
}
