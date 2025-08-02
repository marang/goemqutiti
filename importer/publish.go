package importer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// updatePublish processes publishing progress.
func (m *Model) updatePublish(msg tea.Msg) tea.Cmd {
	switch ev := msg.(type) {
	case publishMsg:
		m.index++
		p := float64(m.index) / float64(len(m.rows))
		if p > 1 {
			p = 1
		}
		cmd := m.progress.SetPercent(p)
		if m.index >= len(m.rows) {
			m.finished = true
			return cmd
		}
		return tea.Batch(cmd, m.nextPublishCmd())
	case tea.KeyMsg:
		switch ev.Type {
		case tea.KeyCtrlN:
			if m.finished {
				m.step = stepDone
			}
		case tea.KeyCtrlP:
			if m.finished {
				m.step = stepReview
				m.finished = false
			}
		}
		cmd := m.history.Update(ev)
		return cmd
	}
	return nil
}

// viewPublish renders the publishing progress step.
func (m *Model) viewPublish(bw, wrap int) string {
	bar := m.progress.View()
	lines := m.published
	limit := m.sampleLimit
	if limit == 0 {
		limit = sampleSize(len(m.rows))
		m.sampleLimit = limit
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	m.history.SetSize(bw, m.historyHeight())
	m.history.SetLines(spacedLines(lines))
	recent := m.history.View()
	if recent != "" {
		recent += "\n"
	}
	headerLine := ""
	if m.finished {
		headerLine = fmt.Sprintf("Published %d messages", len(m.rows))
	} else {
		headerLine = fmt.Sprintf("Publishing %d/%d", m.index, len(m.rows))
	}
	msg := fmt.Sprintf("%s\n%s\n%s", headerLine, bar, recent)
	msg = ansi.Wrap(msg, wrap, " ")
	return ui.LegendBox(msg, "Progress", bw, 0, ui.ColGreen, true, m.history.ScrollPercent())
}
