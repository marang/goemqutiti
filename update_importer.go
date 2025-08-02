package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateImporter forwards messages to the import wizard when active.
func (m *model) updateImporter(msg tea.Msg) tea.Cmd {
	if m.importWizard == nil {
		return nil
	}
	nm, cmd := m.importWizard.Update(msg)
	if w, ok := nm.(*ImportWizard); ok {
		m.importWizard = w
	}
	return cmd
}
