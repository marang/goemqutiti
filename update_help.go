package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) updateHelp(msg tea.Msg) (model, tea.Cmd) {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "esc":
			cmd := m.setMode(m.previousMode())
			return *m, cmd
		case "ctrl+d":
			return *m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.help.vp, cmd = m.help.vp.Update(msg)
	return *m, cmd
}
