package importer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// updateDone handles the final step once publishing is complete.
func (m *Model) updateDone(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			m.step = stepReview
			m.finished = false
		}
		if cmd := m.history.Update(km); cmd != nil {
			return cmd
		}
		if km.String() == "q" {
			return tea.Quit
		}
	}
	return nil
}

// viewDone renders the final step.
func (m *Model) viewDone(bw, wrap int) string {
	if m.dryRun {
		m.history.SetSize(bw, m.historyHeight())
		m.history.SetLines(spacedLines(m.published))
		out := m.history.View()
		out = ansi.Wrap(out, wrap, " ") + "\n[ctrl+p] back  [q] quit"
		return ui.LegendBox(out, "Dry Run", bw, 0, ui.ColGreen, true, m.history.ScrollPercent())
	} else if m.finished {
		msg := fmt.Sprintf("Published %d messages\n[ctrl+p] back  [q] quit", len(m.rows))
		msg = ansi.Wrap(msg, wrap, " ")
		return ui.LegendBox(msg, "Import", bw, 0, ui.ColBlue, true, -1)
	}
	return ui.LegendBox("Done", "Import", bw, 0, ui.ColBlue, true, -1)
}
