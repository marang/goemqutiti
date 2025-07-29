package main

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/importer"
)

func (m model) updateImporter(msg tea.Msg) (model, tea.Cmd) {
	if m.wizard == nil {
		return m, nil
	}
	nm, cmd := m.wizard.Update(msg)
	if w, ok := nm.(*importer.Wizard); ok {
		m.wizard = w
	}
	return m, cmd
}
