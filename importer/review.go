package importer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// updateReview handles the review step interaction.
func (m *Model) updateReview(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case constants.KeyP:
			m.dryRun = false
			m.index = 0
			m.published = nil
			m.finished = false
			m.history.GotoTop()
			m.step = stepPublish
			return tea.Batch(m.progress.SetPercent(0), m.nextPublishCmd())
		case constants.KeyD:
			m.dryRun = true
			m.index = 0
			m.published = nil
			m.finished = false
			m.history.GotoTop()
			m.step = stepPublish
			return tea.Batch(m.progress.SetPercent(0), m.nextPublishCmd())
		case constants.KeyE:
			m.step = stepMap
		case constants.KeyQ:
			m.step = stepDone
		case constants.KeyCtrlP:
			m.step = stepTemplate
			m.tmpl.Focus()
		}
	}
	return nil
}

// viewReview renders the review step.
func (m *Model) viewReview(bw, wrap int) string {
	topic := m.tmpl.Value()
	mapping := m.mapping()
	previews := ""
	max := 3
	if len(m.rows) < max {
		max = len(m.rows)
	}
	for i := 0; i < max; i++ {
		t := BuildTopic(topic, renameFields(m.rows[i], mapping))
		p, _ := RowToJSON(m.rows[i], mapping)
		line := fmt.Sprintf("%s -> %s", t, string(p))
		previews += ansi.Wrap(line, wrap, " ") + "\n"
	}
	s := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [ctrl+p] back  [q] quit", len(m.rows), previews)
	return ui.LegendBox(s, "Review", bw, 0, ui.ColBlue, true, -1)
}
