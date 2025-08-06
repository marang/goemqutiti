package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// updateConnectionForm handles the add/edit connection form.
func (m *model) updateConnectionForm(msg tea.Msg) tea.Cmd {
	if m.connections.Form == nil {
		return nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.WindowSizeMsg, tea.MouseMsg:
		m.connections.Manager.ConnectionsList, _ = m.connections.Manager.ConnectionsList.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case constants.KeyCtrlD:
			return tea.Quit
		case constants.KeyEsc:
			cmd := m.SetMode(constants.ModeConnections)
			m.connections.Form = nil
			return cmd
		case constants.KeyEnter:
			p, err := m.connections.Form.Profile()
			if err != nil {
				m.connections.SendStatus(err.Error())
				return m.connections.ListenStatus()
			}
			if m.connections.Form.Index >= 0 {
				m.connections.Manager.EditConnection(m.connections.Form.Index, p)
			} else {
				m.connections.Manager.AddConnection(p)
			}
			m.connections.RefreshConnectionItems()
			cmd := m.SetMode(constants.ModeConnections)
			m.connections.Form = nil
			return cmd
		}
	}
	f, cmd := m.connections.Form.Update(msg)
	m.connections.Form = &f
	return tea.Batch(cmd, m.connections.ListenStatus())
}
