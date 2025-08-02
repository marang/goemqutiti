package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
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

// viewPublish renders the publishing progress step.
func (w *ImportWizard) viewPublish(bw, wrap int) string {
	bar := w.progress.View()
	lines := w.published
	limit := w.sampleLimit
	if limit == 0 {
		limit = sampleSize(len(w.rows))
		w.sampleLimit = limit
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	w.history.SetSize(bw, w.historyHeight())
	w.history.SetLines(spacedLines(lines))
	recent := w.history.View()
	if recent != "" {
		recent += "\n"
	}
	headerLine := ""
	if w.finished {
		headerLine = fmt.Sprintf("Published %d messages", len(w.rows))
	} else {
		headerLine = fmt.Sprintf("Publishing %d/%d", w.index, len(w.rows))
	}
	msg := fmt.Sprintf("%s\n%s\n%s", headerLine, bar, recent)
	msg = ansi.Wrap(msg, wrap, " ")
	return ui.LegendBox(msg, "Progress", bw, 0, ui.ColGreen, true, w.history.ScrollPercent())
}
