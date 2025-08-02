package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updatePublish processes publishing progress.
func (w *ImportWizard) updatePublish(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case publishMsg:
		w.index++
		p := float64(w.index) / float64(len(w.rows))
		if p > 1 {
			p = 1
		}
		cmd := w.progress.SetPercent(p)
		if w.index >= len(w.rows) {
			w.finished = true
			return w, cmd
		}
		return w, tea.Batch(cmd, w.nextPublishCmd())
	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyCtrlN:
			if w.finished {
				w.step = stepDone
			}
		case tea.KeyCtrlP:
			if w.finished {
				w.step = stepReview
				w.finished = false
			}
		}
		cmd := w.history.Update(m)
		return w, cmd
	}
	return w, nil
}
