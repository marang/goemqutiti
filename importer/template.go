package importer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// updateTemplate handles the template entry step.
func (m *Model) updateTemplate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.tmpl, cmd = m.tmpl.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			m.step = stepMap
		case tea.KeyCtrlN, tea.KeyEnter:
			if strings.TrimSpace(m.tmpl.Value()) != "" {
				m.step = stepReview
			}
		}
	}
	return cmd
}

// viewTemplate renders the topic template step.
func (m *Model) viewTemplate(bw, wrap int) string {
	names := make([]string, len(m.headers))
	for i, h := range m.headers {
		names[i] = "{" + h + "}"
	}
	help := "Available fields: " + strings.Join(names, " ")
	help = ansi.Wrap(help, wrap, " ")
	content := m.tmpl.View() + "\n" + help + "\n[enter] continue  [ctrl+n] next  [ctrl+p] back"
	return ui.LegendBox(content, "Topic Template", bw, 0, ui.ColBlue, true, -1)
}
