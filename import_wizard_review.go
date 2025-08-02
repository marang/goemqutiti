package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateReview handles the review step interaction.
func (w *ImportWizard) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "p":
			w.dryRun = false
			w.index = 0
			w.published = nil
			w.finished = false
			w.history.GotoTop()
			w.step = stepPublish
			return w, tea.Batch(w.progress.SetPercent(0), w.nextPublishCmd())
		case "d":
			w.dryRun = true
			w.index = 0
			w.published = nil
			w.finished = false
			w.history.GotoTop()
			w.step = stepPublish
			return w, tea.Batch(w.progress.SetPercent(0), w.nextPublishCmd())
		case "e":
			w.step = stepMap
		case "q":
			w.step = stepDone
		case "ctrl+p":
			w.step = stepTemplate
			w.tmpl.Focus()
		}
	}
	return w, nil
}
