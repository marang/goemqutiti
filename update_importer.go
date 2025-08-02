package main

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/internal/importer"
)

// updateImporter forwards messages to the import wizard when active.
func (m *model) updateImporter(msg tea.Msg) tea.Cmd {
	if m.importWizard == nil {
		return nil
	}
	nm, cmd := m.importWizard.Update(msg)
	if w, ok := nm.(*importer.ImportWizard); ok {
		m.importWizard = w
	}
	return cmd
}
