package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateDone handles the final step once publishing is complete.
func (w *ImportWizard) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			w.step = stepReview
			w.finished = false
		}
		if cmd := w.history.Update(km); cmd != nil {
			return w, cmd
		}
		if km.String() == "q" {
			return w, tea.Quit
		}
	}
	return w, nil
}
