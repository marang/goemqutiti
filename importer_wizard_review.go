package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
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

// viewReview renders the review step.
func (w *ImportWizard) viewReview(bw, wrap int) string {
	topic := w.tmpl.Value()
	mapping := w.mapping()
	previews := ""
	max := 3
	if len(w.rows) < max {
		max = len(w.rows)
	}
	for i := 0; i < max; i++ {
		t := BuildTopic(topic, renameFields(w.rows[i], mapping))
		p, _ := RowToJSON(w.rows[i], mapping)
		line := fmt.Sprintf("%s -> %s", t, string(p))
		previews += ansi.Wrap(line, wrap, " ") + "\n"
	}
	s := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [ctrl+p] back  [q] quit", len(w.rows), previews)
	return ui.LegendBox(s, "Review", bw, 0, ui.ColBlue, true, -1)
}
