package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateImporter(msg tea.Msg) (model, tea.Cmd) {
	if m.importWizard == nil {
		return m, nil
	}
	nm, cmd := m.importWizard.Update(msg)
	if w, ok := nm.(*ImportWizard); ok {
		m.importWizard = w
	}
	return m, cmd
}
