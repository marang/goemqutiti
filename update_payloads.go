package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updatePayloads manages the stored payloads list.
func (m *model) updatePayloads(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return cmd
		case "delete":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					m.message.payloads = append(m.message.payloads[:i], m.message.payloads[i+1:]...)
					items = append(items[:i], items[i+1:]...)
					m.message.list.SetItems(items)
				}
			}
			return m.connections.ListenStatus()
		case "enter":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					m.topics.input.SetValue(pi.topic)
					m.message.input.SetValue(pi.payload)
					cmd := m.setMode(modeClient)
					return cmd
				}
			}
		}
	}
	m.message.list, cmd = m.message.list.Update(msg)
	return tea.Batch(cmd, m.connections.ListenStatus())
}
