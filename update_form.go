package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// updateForm handles the add/edit connection form.
func (m *model) updateForm(msg tea.Msg) tea.Cmd {
	if m.connections.form == nil {
		return nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.WindowSizeMsg, tea.MouseMsg:
		m.connections.manager.ConnectionsList, _ = m.connections.manager.ConnectionsList.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			cmd := m.setMode(modeConnections)
			m.connections.form = nil
			return cmd
		case "enter":
			p, err := m.connections.form.Profile()
			if err != nil {
				m.connections.SendStatus(err.Error())
				return m.connections.ListenStatus()
			}
			if m.connections.form.index >= 0 {
				m.connections.manager.EditConnection(m.connections.form.index, p)
			} else {
				m.connections.manager.AddConnection(p)
			}
			m.connections.RefreshConnectionItems()
			cmd := m.setMode(modeConnections)
			m.connections.form = nil
			return cmd
		}
	}
	f, cmd := m.connections.form.Update(msg)
	m.connections.form = &f
	return tea.Batch(cmd, m.connections.ListenStatus())
}
