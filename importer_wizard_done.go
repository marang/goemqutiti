package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
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

// viewDone renders the final step.
func (w *ImportWizard) viewDone(bw, wrap int) string {
	if w.dryRun {
		w.history.SetSize(bw, w.historyHeight())
		w.history.SetLines(spacedLines(w.published))
		out := w.history.View()
		out = ansi.Wrap(out, wrap, " ") + "\n[ctrl+p] back  [q] quit"
		return ui.LegendBox(out, "Dry Run", bw, 0, ui.ColGreen, true, w.history.ScrollPercent())
	} else if w.finished {
		msg := fmt.Sprintf("Published %d messages\n[ctrl+p] back  [q] quit", len(w.rows))
		msg = ansi.Wrap(msg, wrap, " ")
		return ui.LegendBox(msg, "Import", bw, 0, ui.ColBlue, true, -1)
	}
	return ui.LegendBox("Done", "Import", bw, 0, ui.ColBlue, true, -1)
}
