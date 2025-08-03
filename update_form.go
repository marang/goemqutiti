package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// updateForm handles the add/edit connection form.
func (m *model) updateForm(msg tea.Msg) tea.Cmd {
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
		case "ctrl+d":
			return tea.Quit
		case "esc":
			cmd := m.setMode(modeConnections)
			m.connections.Form = nil
			return cmd
		case "enter":
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
			cmd := m.setMode(modeConnections)
			m.connections.Form = nil
			return cmd
		}
	}
	f, cmd := m.connections.Form.Update(msg)
	m.connections.Form = &f
	return tea.Batch(cmd, m.connections.ListenStatus())
}
