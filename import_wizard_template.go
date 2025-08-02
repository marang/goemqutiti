package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// updateTemplate handles the template entry step.
func (w *ImportWizard) updateTemplate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.tmpl, cmd = w.tmpl.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			w.step = stepMap
		case tea.KeyCtrlN, tea.KeyEnter:
			if strings.TrimSpace(w.tmpl.Value()) != "" {
				w.step = stepReview
			}
		}
	}
	return w, cmd
}
